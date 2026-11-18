package session

import (
	"math"
	"slices"

	"github.com/cbodonnell/synth"
)

// Snaps are the steps a clip of notes can snap to, in its time (beats, or seconds in a
// sound effect); the first, 0, is off.
func Snaps(sound bool) []float64 {
	if sound {
		return []float64{0, 0.01, 0.05}
	}
	return []float64{0, 1.0 / 16, 1.0 / 4}
}

// fine is the resolution a time is written at with snapping off.
const fine = 1e-4

// Snap is t on the step, or at the resolution times are written at with no step.
func Snap(t, step float64) float64 {
	if step <= 0 {
		step = fine
	}
	return Round(t, step)
}

// NotesSpan is how much time the editor shows for a clip of notes: its bars in a song,
// and in a sound a little past its last note's end.
func NotesSpan(a synth.Arrangement, c synth.Clip, bar int) float64 {
	if a.Tempo > 0 {
		m := synth.Meter{Beat: 1, Bars: a.Meter}
		return m.BarStart(bar+max(c.Length, 1)) - m.BarStart(bar)
	}
	return math.Max(0.25, math.Ceil(soundEnd(c)*1.25/0.25)*0.25)
}

// DrawNote adds n after the clip's notes. Notes keep the order they're written in, so a
// save rewrites only the notes an edit touched.
func DrawNote(notes []synth.Note, n synth.Note) []synth.Note {
	return append(slices.Clone(notes), n)
}

// MoveNotes moves the selected notes by a time and semitones, as far as the span from 0
// and the keyboard allow.
func MoveNotes(notes []synth.Note, sel []int, by, semis, span float64) []synth.Note {
	lo, hi, klo, khi := math.Inf(1), math.Inf(-1), math.Inf(1), math.Inf(-1)
	for _, i := range sel {
		if i >= 0 && i < len(notes) {
			n := notes[i]
			lo, hi = math.Min(lo, n.At), math.Max(hi, n.At)
			klo, khi = math.Min(klo, math.Min(n.Key, n.Key+n.Slide)), math.Max(khi, math.Max(n.Key, n.Key+n.Slide))
		}
	}
	if math.IsInf(lo, 1) {
		return notes
	}
	by = math.Max(-lo, math.Min(by, span-fine-hi))
	semis = math.Max(-klo, math.Min(semis, 127-khi))
	out := slices.Clone(notes)
	for _, i := range sel {
		if i >= 0 && i < len(out) {
			out[i].At = Snap(out[i].At+by, fine)
			out[i].Key += semis
		}
	}
	return out
}

// StretchNote holds note i for dur, at least step.
func StretchNote(notes []synth.Note, i int, dur, step float64) []synth.Note {
	if i < 0 || i >= len(notes) {
		return notes
	}
	out := slices.Clone(notes)
	out[i].Dur = math.Max(Snap(dur, fine), math.Max(step, fine))
	return out
}

// DeleteNotes takes the selected notes out.
func DeleteNotes(notes []synth.Note, sel []int) []synth.Note {
	var out []synth.Note
	for i, n := range notes {
		if !slices.Contains(sel, i) {
			out = append(out, n)
		}
	}
	return out
}
