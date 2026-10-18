package player

import (
	"encoding/binary"
	"io"
	"math"
	"sync"

	"github.com/cbodonnell/synth"
)

// Ebiten's F32 format: little-endian float32, interleaved stereo.
const bytesPerFrame = 8

// Every gain change ramps over this many frames, so none clicks.
const rampFrames = synth.SampleRate * 5 / 1000

// source is a voice's audio, read in order as interleaved stereo frames.
type source interface {
	// read fills dst and returns the frames written; fewer than fit means it has ended.
	read(dst []float32) int
}

type seekable interface {
	source
	seek(frame int) error
}

type memory struct {
	pcm []float32
	pos int
}

func (m *memory) read(dst []float32) int {
	n := copy(dst[:len(dst)/2*2], m.pcm[m.pos:])
	m.pos += n
	return n / 2
}

func (m *memory) seek(frame int) error {
	m.pos = min(frame*2, len(m.pcm))
	return nil
}

// decoded reads F32 bytes, from a Vorbis decoder.
type decoded struct {
	rs  io.ReadSeeker
	buf []byte
}

func (d *decoded) read(dst []float32) int {
	want := len(dst) / 2 * bytesPerFrame
	if len(d.buf) < want {
		d.buf = make([]byte, want)
	}
	n, _ := io.ReadFull(d.rs, d.buf[:want])
	n -= n % bytesPerFrame
	floats(dst, d.buf[:n])
	return n / bytesPerFrame
}

func (d *decoded) seek(frame int) error {
	_, err := d.rs.Seek(int64(frame*bytesPerFrame), io.SeekStart)
	return err
}

func floats(dst []float32, b []byte) {
	for i := range len(b) / 4 {
		dst[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[i*4:]))
	}
}

// Like Ebiten's InfiniteLoop: what the source holds past the loop's end is blended
// into the loop's start over this many frames, which hides a codec's seam.
const blendFrames = 256

// looped plays its source to end, then from start to end forever.
type looped struct {
	src        seekable
	start, end int
	pos        int
	after      []float32
	blending   bool
}

func (l *looped) read(dst []float32) int {
	frames := len(dst) / 2
	total := 0
	stuck := false
	for total < frames {
		want := min(frames-total, l.end-l.pos)
		out := dst[total*2 : (total+want)*2]
		n := l.src.read(out)
		if l.blending {
			for i := range n {
				p := l.pos + i - l.start
				if p >= len(l.after)/2 {
					break
				}
				rate := 1 - float32(p)/float32(len(l.after)/2)
				for c := range 2 {
					out[i*2+c] = l.after[p*2+c]*rate + out[i*2+c]*(1-rate)
				}
			}
		}
		l.pos += n
		total += n
		if n == want && l.pos < l.end {
			continue
		}
		if n == 0 && stuck {
			return total
		}
		stuck = n == 0
		if l.after == nil {
			buf := make([]float32, blendFrames*2)
			l.after = buf[:l.src.read(buf)*2]
		}
		l.blending = len(l.after) > 0
		if l.src.seek(l.start) != nil {
			return total
		}
		l.pos = l.start
	}
	return total
}

// level is a gain that ramps to each new target.
type level struct {
	cur, target, step float32
	left              int
}

func newLevel(g float64) level { return level{cur: float32(g), target: float32(g)} }

func (l *level) set(g float64) {
	l.target = float32(g)
	l.left = rampFrames
	l.step = (l.target - l.cur) / rampFrames
}

func (l *level) next() float32 {
	if l.left > 0 {
		l.left--
		l.cur += l.step
		if l.left == 0 {
			l.cur = l.target
		}
	}
	return l.cur
}

type voice struct {
	id       string
	src      source
	rate     float64
	gain     level
	priority int
	stopping bool  // fading out, then gone
	start    int64 // the stream frame it began on; -1 until then
	done     bool

	in     []float32
	at, n  int
	ended  bool
	window [4][2]float32 // frames k-1 .. k+2 around the read position k+frac
	frac   float64
	primed bool
	pad    int
}

func newVoice(id string, src source, rate, gain float64) *voice {
	return &voice{id: id, src: src, rate: rate, gain: newLevel(gain), start: -1, in: make([]float32, 1024*2)}
}

func (v *voice) pull() ([2]float32, bool) {
	if v.at == v.n {
		if v.ended {
			return [2]float32{}, false
		}
		v.n, v.at = v.src.read(v.in), 0
		v.ended = v.n < len(v.in)/2
		if v.n == 0 {
			return [2]float32{}, false
		}
	}
	f := [2]float32{v.in[v.at*2], v.in[v.at*2+1]}
	v.at++
	return f, true
}

// frame is the next output frame, read at the voice's rate with 4-point Hermite
// interpolation between source frames.
func (v *voice) frame() ([2]float32, bool) {
	if v.rate == 1 {
		return v.pull()
	}
	advance := func() {
		v.window[0], v.window[1], v.window[2] = v.window[1], v.window[2], v.window[3]
		f, ok := v.pull()
		if !ok {
			v.pad++
		}
		v.window[3] = f
	}
	if !v.primed {
		v.primed = true
		for range 3 {
			advance()
		}
	}
	for v.frac >= 1 {
		advance()
		v.frac--
	}
	// The third frame of padding is the read position itself: the source has played out.
	if v.pad >= 3 {
		return [2]float32{}, false
	}
	var out [2]float32
	x := float32(v.frac)
	for c := range 2 {
		y0, y1, y2, y3 := v.window[0][c], v.window[1][c], v.window[2][c], v.window[3][c]
		c1 := 0.5 * (y2 - y0)
		c2 := y0 - 2.5*y1 + 2*y2 - 0.5*y3
		c3 := 0.5*(y3-y0) + 1.5*(y1-y2)
		out[c] = ((c3*x+c2)*x+c1)*x + y1
	}
	v.frac += v.rate
	return out, true
}

func (v *voice) mix(out []float32) {
	for i := range len(out) / 2 {
		f, ok := v.frame()
		if !ok {
			v.done = true
			return
		}
		g := v.gain.next()
		out[i*2] += f[0] * g
		out[i*2+1] += f[1] * g
		if v.stopping && v.gain.left == 0 {
			v.done = true
			return
		}
	}
}

func (v *voice) stop() {
	v.stopping = true
	v.gain.set(0)
	if v.start < 0 {
		v.done = true
	}
}

// closer is a source holding something to let go of once its voice is gone.
type closer interface{ close() }

func release(v *voice) {
	if c, ok := v.src.(closer); ok {
		c.close()
	}
}

// stream sums its voices into one part of the bus. A voice started between mixes
// begins on the first frame of the next, so voices started together stay sample-aligned.
type stream struct {
	mu     sync.Mutex
	voices []*voice // oldest first
	gain   level
	pos    int64
	max    int  // voices at once; 0 is no limit
	hold   bool // at gain 0, stand still instead of playing on unheard
	mixed  []float32
}

func newStream(gain float64, max int) *stream {
	return &stream{gain: newLevel(gain), max: max}
}

// Read is the stream on its own, as Ebiten's F32 format.
func (s *stream) Read(b []byte) (int, error) {
	n := len(b) / bytesPerFrame
	if cap(s.mixed) < n*2 {
		s.mixed = make([]float32, n*2)
	}
	mixed := s.mixed[:n*2]
	s.mix(mixed)
	putFloats(b, mixed)
	return n * bytesPerFrame, nil
}

// mix overwrites out with the stream's next len(out)/2 frames.
func (s *stream) mix(out []float32) {
	clear(out)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.hold && s.silent() {
		return
	}
	n := len(out) / 2
	live := s.voices[:0]
	for _, v := range s.voices {
		if v.done {
			release(v)
			continue
		}
		if v.start < 0 {
			v.start = s.pos
		}
		v.mix(out)
		if v.done {
			release(v)
			continue
		}
		live = append(live, v)
	}
	clear(s.voices[len(live):])
	s.voices = live
	for i := range n {
		g := s.gain.next()
		out[i*2] *= g
		out[i*2+1] *= g
	}
	s.pos += int64(n)
}

func (s *stream) silent() bool { return s.gain.left == 0 && s.gain.cur == 0 }

// held is true while a holding stream stands still at gain 0.
func (s *stream) held() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.hold && s.silent()
}

// frames is how many frames the stream has mixed.
func (s *stream) frames() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pos
}

func putFloats(b []byte, x []float32) {
	for i, f := range x {
		binary.LittleEndian.PutUint32(b[i*4:], math.Float32bits(f))
	}
}

// play adds v, first making room: a sound past its own limit gives up its oldest play,
// and past the stream's limit the lowest priority gives way, oldest first. A new voice
// that would be the one to give way isn't added.
func (s *stream) play(v *voice, limit int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if limit > 0 {
		var same []*voice
		for _, o := range s.live() {
			if o.id == v.id {
				same = append(same, o)
			}
		}
		for len(same) >= limit {
			same[0].stop()
			same = same[1:]
		}
	}
	if live := s.live(); s.max > 0 && len(live) >= s.max {
		victim := live[0]
		for _, o := range live {
			if o.priority < victim.priority {
				victim = o
			}
		}
		if v.priority < victim.priority {
			return false
		}
		victim.stop()
	}
	s.voices = append(s.voices, v)
	return true
}

func (s *stream) live() []*voice {
	var live []*voice
	for _, v := range s.voices {
		if !v.stopping && !v.done {
			live = append(live, v)
		}
	}
	return live
}

func (s *stream) stop(v *voice) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v.stop()
}

func (s *stream) setGain(g float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gain.set(g)
}

// started is the stream frame v began on, once it has.
func (s *stream) started(v *voice) (int64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return v.start, v.start >= 0
}
