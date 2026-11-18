package synth

import "math"

// Part is one instrument's line: a patch, the notes it plays, and where it goes in the mix.
type Part struct {
	Name    string // also seeds its noise, so renaming a part changes its texture
	Patch   Patch
	Notes   []Note
	Level   float64 // dB; 0 is unity
	Pan     float64 // -1..1
	Inserts []Effect
	Sends   []Send

	// The bus this part plays into instead of the master: a drum kit's parts into one
	// "drums" bus, compressed together. Sends still go where they say.
	Out string
}

type Send struct {
	Bus   string
	Level float64 // dB; 0 is unity
}

// Bus is a shared effect return, such as one reverb that several parts send to.
type Bus struct {
	Name    string
	Effects []Effect
	Level   float64 // dB; 0 is unity
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

	// For a sound effect: how plays differ, how many may sound at once (0 is no limit;
	// past it the oldest play of this sound gives way), and which sounds the player keeps
	// when too many play at once (higher survives; ties go to the newest).
	Vary     Variation
	Voices   int
	Priority int

	// Integrated loudness in LUFS the song should measure within 1 LU of, rendered as
	// studio plays it: the intro, then the body once more. 0 sets none. Games aim near
	// -16 for music.
	Target float64

	take int
}

// Variation hides a sound's repetition. Pitch and Gain move each play at random within
// ±their range; Takes are renders whose noise is seeded apart, played in turn so the
// texture changes too. All of it happens in the player, so Render stays pure.
type Variation struct {
	Pitch float64 // semitones
	Gain  float64 // dB
	Takes int     // 0 and 1 mean one
}

// Take is take n of a sound with Vary.Takes: the same song, its noise seeded apart.
// Take 0 is the song itself.
func Take(s Song, n int) Song {
	s.take = n
	return s
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
