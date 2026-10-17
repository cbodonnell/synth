package synth

import (
	"hash/fnv"
	"math"
	"runtime"
	"sync"
)

// Part is one instrument's line: a patch, the notes it plays, and where it goes in the mix.
type Part struct {
	Name    string // also seeds its noise, so renaming a part changes its texture
	Patch   Patch
	Notes   []Note
	Gain    float64 // 0 means 1
	Pan     float64 // -1..1
	Inserts []Effect
	Sends   []Send

	// The bus this part plays into instead of the master: a drum kit's parts into one
	// "drums" bus, compressed together. Sends still go where they say.
	Out string
}

type Send struct {
	Bus   string
	Level float64
}

// Bus is a shared effect return, such as one reverb that several parts send to.
type Bus struct {
	Name    string
	Effects []Effect
	Gain    float64 // 0 means 1
}

// Song is a finished piece, or a one-shot sound effect: anything that renders.
//
// Length is authoritative, not derived from the notes, since the last note usually
// rings past the bar line. With Loop set, the song repeats from LoopStart forever and
// anything ringing past Length wraps onto LoopStart instead of being cut off.
type Song struct {
	ID     string // caches the render; two different songs must not share one
	Name   string
	Parts  []Part
	Buses  []Bus
	Master []Effect

	Length    float64
	Loop      bool
	LoopStart float64

	Meter Meter

	// Integrated loudness in LUFS the song should measure within 1 LU of, rendered as
	// studio plays it: the intro, then the body once more. 0 sets none. Games aim near
	// -16 for music.
	Target float64
}

func gainOr1(g float64) float32 {
	if g == 0 {
		return 1
	}
	return float32(g)
}

// end is when the last note falls silent, ignoring effect tails.
func (s Song) end() float64 {
	end := 0.0
	for _, p := range s.Parts {
		for _, n := range p.Notes {
			end = math.Max(end, n.At+noteLength(p.Patch, n))
		}
	}
	return end
}

func chainTail(fx []Effect) float64 {
	t := 0.0
	for _, e := range fx {
		t += e.Tail()
	}
	return t
}

func (s Song) tail() float64 {
	var parts, buses float64
	for _, p := range s.Parts {
		parts = math.Max(parts, chainTail(p.Inserts))
	}
	for _, b := range s.Buses {
		buses = math.Max(buses, chainTail(b.Effects))
	}
	return parts + buses + chainTail(s.Master)
}

// Render plays the song once from silence until everything has died away.
func Render(s Song) Stereo {
	out := mixdown(s, s.Parts, samples(math.Max(s.Length, s.end())+s.tail()))
	trimSilence(&out, samples(s.Length))
	clip(out)
	return out
}

// Looped is a looping song split the way Ebiten's InfiniteLoopWithIntro plays it.
type Looped struct {
	// Plays once: everything before LoopStart and then the first pass through the
	// loop, which starts from silence rather than from the loop's own tail.
	Intro Stereo
	// Repeats forever. Its own tail is wrapped onto its start, so the seam is continuous.
	Body Stereo
}

func RenderLoop(s Song) Looped {
	total := samples(s.Length)
	start := min(samples(s.LoopStart), total)
	body := total - start
	if body <= 0 {
		return Looped{Intro: Render(s)}
	}
	ring := samples(s.end()+s.tail()) + 1

	intro := mixdown(s, s.Parts, max(total, ring))
	intro = Stereo{L: intro.L[:total], R: intro.R[:total]}

	var loopParts []Part
	for _, p := range s.Parts {
		lp := p
		lp.Notes = nil
		for _, n := range p.Notes {
			if n.At >= s.LoopStart {
				lp.Notes = append(lp.Notes, n)
			}
		}
		loopParts = append(loopParts, lp)
	}
	full := mixdown(s, loopParts, max(total, ring))

	looped := NewStereo(body)
	for i := start; i < full.Len(); i++ {
		j := (i - start) % body
		looped.L[j] += full.L[i]
		looped.R[j] += full.R[i]
	}

	clip(intro)
	clip(looped)
	return Looped{Intro: intro, Body: looped}
}

func mixdown(s Song, parts []Part, n int) Stereo {
	master := NewStereo(n)
	if n == 0 {
		return master
	}

	rendered := make([]Stereo, len(parts))
	var wg sync.WaitGroup
	for i, p := range parts {
		wg.Go(func() { rendered[i] = renderPart(p, n) })
	}
	wg.Wait()

	buses := make(map[string]Stereo, len(s.Buses))
	for _, b := range s.Buses {
		buses[b.Name] = NewStereo(n)
	}
	for i, p := range parts {
		g := gainOr1(p.Gain)
		dst := master
		if bus, ok := buses[p.Out]; ok {
			dst = bus
		}
		dst.add(rendered[i], g)
		for _, send := range p.Sends {
			if bus, ok := buses[send.Bus]; ok {
				bus.add(rendered[i], g*float32(send.Level))
			}
		}
	}
	for _, b := range s.Buses {
		buf := buses[b.Name]
		for _, e := range b.Effects {
			e.Process(buf)
		}
		master.add(buf, gainOr1(b.Gain))
	}
	blockDC(master)
	for _, e := range s.Master {
		e.Process(master)
	}
	return master
}

func renderPart(p Part, n int) Stereo {
	h := fnv.New64a()
	h.Write([]byte(p.Name))
	seed := h.Sum64()

	// Notes render concurrently into buffers of their own and are summed in order,
	// so the result is identical however the work was scheduled.
	notes := make([]Stereo, len(p.Notes))
	work := make(chan int)
	var wg sync.WaitGroup
	for range runtime.GOMAXPROCS(0) {
		wg.Go(func() {
			for ni := range work {
				notes[ni] = renderNote(p, p.Notes[ni], seed, n)
			}
		})
	}
	for ni := range p.Notes {
		work <- ni
	}
	close(work)
	wg.Wait()

	out := NewStereo(n)
	for ni, buf := range notes {
		start := samples(p.Notes[ni].At)
		for i := range buf.L {
			out.L[start+i] += buf.L[i]
			out.R[start+i] += buf.R[i]
		}
	}
	for _, e := range p.Inserts {
		e.Process(out)
	}
	return out
}

func renderNote(p Part, note Note, partSeed uint64, songLength int) Stereo {
	start := samples(note.At)
	length := min(samples(noteLength(p.Patch, note)), songLength-start)
	if length <= 0 {
		return Stereo{}
	}
	buf := NewStereo(length)
	// Seeded by the note itself, not its index, so RenderLoop's loop-only pass
	// renders each note exactly as the full pass does.
	seed := mixSeed(partSeed, math.Float64bits(note.At), math.Float64bits(note.Key))
	for _, v := range unisonVoices(p.Patch.Unison, p.Pan, seed) {
		renderVoice(buf, 0, p.Patch, note, v)
	}
	return buf
}

// ~10Hz: AC coupling, so a pulse wave's off-center average never reaches the speaker.
const dcPole = 0.9987

// Runs before any loop is folded, so the fold itself stays continuous at the seam.
func blockDC(s Stereo) {
	for _, ch := range [][]float32{s.L, s.R} {
		var px, py float32
		for i, x := range ch {
			y := x - px + dcPole*py
			px, py = x, y
			ch[i] = y
		}
	}
}

func clip(s Stereo) {
	for _, ch := range [][]float32{s.L, s.R} {
		for i, x := range ch {
			ch[i] = float32(clamp(float64(x), -1, 1))
		}
	}
}

// Below 16-bit resolution; nothing past this point would reach a listener.
const silence = 1.0 / 65536

func trimSilence(s *Stereo, keep int) {
	n := s.Len()
	for n > keep && abs32(s.L[n-1]) < silence && abs32(s.R[n-1]) < silence {
		n--
	}
	s.L, s.R = s.L[:n], s.R[:n]
}
