package player

import (
	"sync"

	"github.com/cbodonnell/synth"
)

// Preview is the studio's voice: a buffer played from any frame, looping over a region,
// and swapped for a new render at its position. Every change crossfades over a few
// milliseconds, so an edit heard while playing never clicks.
type Preview struct {
	mu       sync.Mutex
	cur, old *take
	fade     int // frames left of the crossfade from old to cur
	heard    func() int64
	log      []mark
	mixed    int64
	playing  bool
	stopping bool
}

// take is one buffer playing, from start to end and round again.
type take struct {
	buf        synth.Stereo
	pos        int
	start, end int
}

// mark is where a run of the bus began reading the buffer from, so the frame being
// heard can be traced back to the buffer.
type mark struct {
	bus int64
	pos int
}

const marks = 256

func newPreview() *Preview { return &Preview{heard: func() int64 { return 0 }} }

// Play plays buf from frame from. If loopEnd is past loopStart it plays loopStart to
// loopEnd forever once it reaches loopEnd; otherwise it plays to buf's end and stops.
func (v *Preview) Play(buf synth.Stereo, from, loopStart, loopEnd int) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.crossfade(&take{buf: buf, pos: max(from, 0), start: loopStart, end: loopEnd})
}

// Swap plays buf in place of what's playing, looping as Play does, from the frame at
// maps the playing frame to; nil keeps the frame.
func (v *Preview) Swap(buf synth.Stereo, loopStart, loopEnd int, at func(frame int) int) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if !v.playing || v.stopping {
		return
	}
	pos := v.cur.pos
	if at != nil {
		pos = at(pos)
	}
	t := &take{buf: buf, pos: pos, start: loopStart, end: loopEnd}
	if t.loops() && t.pos >= t.end {
		t.pos = t.start + (t.pos-t.start)%(t.end-t.start)
	}
	v.old, v.cur, v.fade = v.cur, t, rampFrames
}

func (v *Preview) crossfade(t *take) {
	if v.playing && !v.stopping {
		v.old, v.fade = v.cur, rampFrames
	} else {
		v.old, v.fade = nil, rampFrames
	}
	v.cur, v.playing, v.stopping = t, true, false
	// What's heard next is the new take; until then, it's where it starts.
	v.log = append(v.log[:0], mark{bus: v.mixed, pos: t.pos})
}

// Stop fades out.
func (v *Preview) Stop() {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.playing && !v.stopping {
		v.stopping, v.old, v.fade = true, nil, rampFrames
	}
}

// Position is the buffer frame being heard, and whether the preview is playing.
func (v *Preview) Position() (int, bool) {
	heard := v.heard()
	v.mu.Lock()
	defer v.mu.Unlock()
	if !v.playing {
		return 0, false
	}
	m := mark{bus: -1}
	for i := len(v.log) - 1; i >= 0; i-- {
		if v.log[i].bus <= heard {
			m = v.log[i]
			break
		}
	}
	if m.bus < 0 {
		if len(v.log) == 0 {
			return 0, false
		}
		m = v.log[0]
		heard = m.bus
	}
	return m.pos + int(heard-m.bus), true
}

func (v *Preview) mark(bus int64, pos int) {
	if len(v.log) == marks {
		v.log = append(v.log[:0], v.log[marks/2:]...)
	}
	v.log = append(v.log, mark{bus: bus, pos: pos})
}

func (t *take) loops() bool { return t.end > t.start }

// next is the take's next frame; false once it has played out.
func (t *take) next() (l, r float32, ok bool) {
	if t.pos >= t.buf.Len() {
		return 0, 0, false
	}
	l, r = t.buf.L[t.pos], t.buf.R[t.pos]
	t.pos++
	if t.loops() && t.pos >= t.end {
		t.pos = t.start
	}
	return l, r, true
}

// mix adds the preview's next len(out)/2 frames into out.
func (v *Preview) mix(out []float32) {
	v.mu.Lock()
	defer v.mu.Unlock()
	at := v.mixed
	v.mixed += int64(len(out) / 2)
	if !v.playing {
		return
	}
	v.mark(at, v.cur.pos)
	for i := range len(out) / 2 {
		in := float32(1)
		if v.fade > 0 {
			v.fade--
			x := 1 - float32(v.fade)/rampFrames
			in = x
			if v.stopping {
				in = 1 - x
			}
			if v.old != nil {
				l, r, _ := v.old.next()
				out[i*2] += l * (1 - x)
				out[i*2+1] += r * (1 - x)
			}
			if v.fade == 0 {
				v.old = nil
				if v.stopping {
					v.playing = false
					return
				}
			}
		}
		was := v.cur.pos
		l, r, ok := v.cur.next()
		if !ok {
			v.playing = false
			return
		}
		out[i*2] += l * in
		out[i*2+1] += r * in
		if v.cur.pos < was {
			v.mark(at+int64(i+1), v.cur.pos)
		}
	}
}
