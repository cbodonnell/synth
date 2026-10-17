package synth

import "math"

// Meter is a song's pulse, for keeping visuals in time. Nothing in rendering reads it.
// Bars lists beats per bar and repeats, so irregular metres are plain data:
// Golden Brown's thirteen-beat phrase is {3, 3, 3, 4}.
type Meter struct {
	Beat float64 // seconds per beat
	Bars []int   // nil means 4/4
}

type Position struct {
	Bar   int     // counted from 0 since the song began
	Beat  int     // within the bar, from 0
	Phase float64 // through the current beat, 0..1
}

func (m Meter) bars() []int {
	if len(m.Bars) == 0 {
		return []int{4}
	}
	return m.Bars
}

func (m Meter) Cycle() int {
	n := 0
	for _, b := range m.bars() {
		n += b
	}
	return n
}

func (m Meter) At(t float64) Position {
	if m.Beat <= 0 || t < 0 {
		return Position{}
	}
	beats := t / m.Beat
	whole := int(math.Floor(beats))
	phase := beats - float64(whole)

	bars := m.bars()
	cycle := m.Cycle()
	bar := (whole / cycle) * len(bars)
	rem := whole % cycle
	for _, b := range bars {
		if rem < b {
			break
		}
		rem -= b
		bar++
	}
	return Position{Bar: bar, Beat: rem, Phase: phase}
}

// SongTime maps how long a looping song has played onto where it is in the score.
func (s Song) SongTime(played float64) float64 {
	body := s.Length - s.LoopStart
	if !s.Loop || body <= 0 || played < s.LoopStart {
		return played
	}
	return s.LoopStart + math.Mod(played-s.LoopStart, body)
}
