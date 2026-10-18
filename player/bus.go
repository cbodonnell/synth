package player

import (
	"math"
	"runtime"
	"sync"

	"github.com/cbodonnell/synth"
)

// The master limiter's defaults: a ceiling a little under full scale, and a release
// slow enough that its recovery isn't heard as pumping.
const (
	defaultCeiling = -1.0 // dB
	limiterRelease = 0.15 // seconds
)

// bus is the one reader Ebiten plays: music and effects summed, then limited, so sounds
// that pile up together can't clip however many land at once.
type bus struct {
	mu             sync.Mutex
	music, effects *stream
	limit          limiter
	pos            int64
	m, e           []float32
}

func newBus(music, effects *stream, ceilingDB float64) *bus {
	return &bus{music: music, effects: effects, limit: newLimiter(ceilingDB, limiterRelease)}
}

func (b *bus) Read(p []byte) (int, error) {
	n := len(p) / bytesPerFrame
	b.mu.Lock()
	defer b.mu.Unlock()
	if cap(b.m) < n*2 {
		b.m, b.e = make([]float32, n*2), make([]float32, n*2)
	}
	m, e := b.m[:n*2], b.e[:n*2]
	b.music.mix(m)
	b.effects.mix(e)
	for i := range m {
		m[i] += e[i]
	}
	b.limit.process(m)
	b.pos += int64(n)
	putFloats(p, m)
	return n * bytesPerFrame, nil
}

// musicAt is how far the music stream and the bus had mixed, read together.
func (b *bus) musicAt() (music, mixed int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.music.frames(), b.pos
}

// limiter is a peak limiter without look-ahead, so it adds no latency: its gain drops at
// once to whatever keeps a frame under the ceiling, then recovers exponentially.
type limiter struct {
	ceiling, gain, recover float32
}

func newLimiter(ceilingDB, release float64) limiter {
	return limiter{
		ceiling: float32(math.Pow(10, ceilingDB/20)),
		gain:    1,
		recover: float32(1 - math.Exp(-1/(release*synth.SampleRate))),
	}
}

func (l *limiter) process(x []float32) {
	for i := 0; i < len(x); i += 2 {
		l.gain += (1 - l.gain) * l.recover
		if peak := max(abs32(x[i]), abs32(x[i+1])); peak*l.gain > l.ceiling {
			l.gain = l.ceiling / peak
		}
		x[i] *= l.gain
		x[i+1] *= l.gain
	}
}

func abs32(x float32) float32 { return math.Float32frombits(math.Float32bits(x) &^ (1 << 31)) }

// A decoder reads this far ahead of playback, in chunks.
const (
	aheadFrames = 2048
	aheadChunks = 12 // about half a second
)

// ahead reads its source on its own goroutine, a little ahead of playback, so decoding
// a long song happens beside the bus's reads instead of inside them.
type ahead struct {
	full, free chan []float32
	quit       chan struct{}
	once       sync.Once
	cur, spent []float32
}

func newAhead(src source) *ahead {
	a := &ahead{
		full: make(chan []float32, aheadChunks),
		free: make(chan []float32, aheadChunks+1),
		quit: make(chan struct{}),
	}
	go a.fill(src)
	return a
}

func (a *ahead) fill(src source) {
	defer close(a.full)
	for {
		var buf []float32
		select {
		case buf = <-a.free:
		default:
			buf = make([]float32, aheadFrames*2)
		}
		n := src.read(buf)
		if n > 0 {
			select {
			case a.full <- buf[:n*2]:
			case <-a.quit:
				return
			}
		}
		if n < aheadFrames {
			return
		}
		// WebAssembly never preempts a goroutine; this lets the game in between chunks.
		runtime.Gosched()
	}
}

// read waits for the decoder when it has fallen behind: a late frame is a glitch, but
// a frame of invented silence would end a song early.
func (a *ahead) read(dst []float32) int {
	frames, total := len(dst)/2, 0
	for total < frames {
		if len(a.cur) == 0 {
			if a.spent != nil {
				select {
				case a.free <- a.spent[:cap(a.spent)]:
				default:
				}
				a.spent = nil
			}
			c, ok := <-a.full
			if !ok {
				break
			}
			a.cur, a.spent = c, c
		}
		n := copy(dst[total*2:frames*2], a.cur)
		a.cur = a.cur[n:]
		total += n / 2
	}
	return total
}

func (a *ahead) close() { a.once.Do(func() { close(a.quit) }) }
