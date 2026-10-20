package session

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/cbodonnell/synth"
)

// unique is base, or base with a number after it, whichever taken doesn't hold.
func unique(base string, taken func(string) bool) string {
	name := base
	for n := 2; taken(name); n++ {
		name = fmt.Sprintf("%s %d", base, n)
	}
	return name
}

// CopyPlay plays the play's clip again at bar, and returns the copy's index.
func CopyPlay(a *synth.Arrangement, track, play, bar int) int {
	t := &a.Tracks[track]
	p := t.Play[play]
	p.Bar = bar
	t.Play = append(t.Play, p)
	return len(t.Play) - 1
}

// RepeatPlay plays the play again right after itself, and returns the copy's index.
func RepeatPlay(a *synth.Arrangement, track, play int) int {
	t := a.Tracks[track]
	p := t.Play[play]
	c, _ := clip(t, p.Clip)
	return CopyPlay(a, track, play, p.Bar+max(p.Times, 1)*clipBars(c))
}

// SetTimes plays the play's clip times over.
func SetTimes(a *synth.Arrangement, track, play, times int) {
	p := &a.Tracks[track].Play[play]
	switch {
	case times <= 1 && p.Times <= 1:
	case times <= 1:
		p.Times = 1
	default:
		p.Times = times
	}
}

// DeletePlay takes the play out, and its clip with it when nothing else plays it.
func DeletePlay(a *synth.Arrangement, track, play int) {
	t := &a.Tracks[track]
	name := t.Play[play].Clip
	t.Play = slices.Delete(t.Play, play, play+1)
	if !slices.ContainsFunc(t.Play, func(p synth.Play) bool { return p.Clip == name }) {
		t.Clips = slices.DeleteFunc(t.Clips, func(c synth.Clip) bool { return c.Name == name })
	}
}

// MakeUnique copies the play's clip under a new name for the play alone, and returns the
// name.
func MakeUnique(a *synth.Arrangement, track, play int) string {
	t := &a.Tracks[track]
	c, ok := clip(*t, t.Play[play].Clip)
	if !ok {
		return ""
	}
	c = cloneClip(c)
	c.Name = unique(c.Name, func(s string) bool {
		return slices.ContainsFunc(t.Clips, func(c synth.Clip) bool { return c.Name == s })
	})
	t.Clips = append(t.Clips, c)
	t.Play[play].Clip = c.Name
	return c.Name
}

func cloneClip(c synth.Clip) synth.Clip {
	c.Bars = slices.Clone(c.Bars)
	c.Notes = slices.Clone(c.Notes)
	return c
}

// AddClip writes an empty one-bar clip into the track, played at bar, and returns its
// name. A track of drums gets a drum line.
func AddClip(a *synth.Arrangement, track, bar int) string {
	t := &a.Tracks[track]
	c := synth.Clip{Step: 1.0 / StepsPerBeat}
	if len(t.Clips) > 0 && !slices.ContainsFunc(t.Clips, func(c synth.Clip) bool { return c.Hit == "" }) {
		c.Hit = t.Clips[0].Hit
	}
	m := synth.Meter{Beat: 1, Bars: a.Meter}
	c.Bars = []string{strings.TrimSpace(strings.Repeat(". ", m.BarBeats(max(bar-1, 0))*StepsPerBeat))}
	c.Name = unique("clip", func(s string) bool {
		return slices.ContainsFunc(t.Clips, func(c synth.Clip) bool { return c.Name == s })
	})
	t.Clips = append(t.Clips, c)
	t.Play = append(t.Play, synth.Play{Clip: c.Name, Bar: bar})
	return c.Name
}

// SectionBar is the bar, from 1, section i begins on.
func SectionBar(a synth.Arrangement, i int) int {
	bar := 1
	for _, s := range a.Sections[:i] {
		bar += s.Bars
	}
	return bar
}

// SectionBars is how many bars a section grows and shrinks by: a whole turn of an
// irregular meter, so what follows keeps its place in it.
func SectionBars(a synth.Arrangement) int { return max(len(a.Meter), 1) }

// NewSectionBars is how long a new section is: four bars, or the turns of the meter
// that make at least four.
func NewSectionBars(a synth.Arrangement) int {
	n := SectionBars(a)
	return (4 + n - 1) / n * n
}

type sections []synth.Section

func (s sections) taken(name string) bool {
	return slices.ContainsFunc(s, func(s synth.Section) bool { return s.Name == name })
}

// InsertSection writes a new section of bars before section i (at the end when i is
// past the last), moving everything after it on.
func InsertSection(a *synth.Arrangement, i, bars int) {
	i = max(0, min(i, len(a.Sections)))
	bar := SectionBar(*a, i)
	name := unique("New", sections(a.Sections).taken)
	reflow(a, func(from, to int) []int {
		if from >= bar-1 {
			return []int{from + bars}
		}
		return []int{from}
	})
	a.Sections = slices.Insert(a.Sections, i, synth.Section{Name: name, Bars: bars})
}

// DeleteSection takes section i out with everything played in it, closing the gap. A
// loop starting there starts at the section after it.
func DeleteSection(a *synth.Arrangement, i int) error {
	if len(a.Sections) <= 1 {
		return errors.New("a song needs a section")
	}
	from := SectionBar(*a, i) - 1
	n := a.Sections[i].Bars
	cut(a, from, n)
	if a.Loop == a.Sections[i].Name {
		a.Loop = ""
		if i+1 < len(a.Sections) {
			a.Loop = a.Sections[i+1].Name
		}
	}
	a.Sections = slices.Delete(a.Sections, i, i+1)
	return nil
}

// DuplicateSection writes a copy of section i after it, playing what plays wholly
// inside it, and its automation.
func DuplicateSection(a *synth.Arrangement, i int) {
	s := a.Sections[i]
	from := SectionBar(*a, i) - 1
	to := from + s.Bars
	reflow(a, func(f, t int) []int {
		switch {
		case f >= to:
			return []int{f + s.Bars}
		case f >= from && t <= to:
			return []int{f, f + s.Bars}
		}
		return []int{f}
	})
	name := s.Name
	for sections(a.Sections).taken(name) {
		name += "'"
	}
	a.Sections = slices.Insert(a.Sections, i+1, synth.Section{Name: name, Bars: s.Bars})
}

// ResizeSection makes section i bars long, moving what follows it; shrinking it drops
// what played in the bars cut from its end.
func ResizeSection(a *synth.Arrangement, i, bars int) error {
	if bars < 1 {
		return errors.New("a section needs a bar")
	}
	end := SectionBar(*a, i) - 1 + a.Sections[i].Bars
	by := bars - a.Sections[i].Bars
	switch {
	case by > 0:
		reflow(a, func(from, to int) []int {
			if from >= end {
				return []int{from + by}
			}
			return []int{from}
		})
	case by < 0:
		cut(a, end+by, -by)
	}
	a.Sections[i].Bars = bars
	return nil
}

// RenameSection renames section i, and the loop with it.
func RenameSection(a *synth.Arrangement, i int, name string) error {
	name = strings.TrimSpace(name)
	old := a.Sections[i].Name
	switch {
	case name == "":
		return errors.New("a section needs a name")
	case name != old && sections(a.Sections).taken(name):
		return fmt.Errorf("there's already a section %q", name)
	}
	a.Sections[i].Name = name
	if a.Loop == old {
		a.Loop = name
	}
	return nil
}

// cut takes bars from..from+n (from 0) out, with what plays in them, closing the gap.
func cut(a *synth.Arrangement, from, n int) {
	reflow(a, func(f, t int) []int {
		switch {
		case t <= from:
			return []int{f}
		case f >= from+n:
			return []int{f - n}
		}
		return nil
	})
}

// reflow moves what the song plays bar by bar: where each pass of a clip goes, given the
// bars it spans (from 0, end excluded), as none, one or more starts. Automation follows
// the bar each point is in. Passes of one play still back to back stay one play.
func reflow(a *synth.Arrangement, move func(from, to int) []int) {
	m := synth.Meter{Beat: 1, Bars: a.Meter}
	beats := func(x float64) []float64 {
		b := m.At(x).Bar
		var out []float64
		for _, nb := range move(b, b+1) {
			out = append(out, x+(m.BarStart(nb)-m.BarStart(b)))
		}
		return out
	}
	for i := range a.Tracks {
		t := &a.Tracks[i]
		t.Play = replay(*t, move)
		retime(t.Inserts, beats)
	}
	for i := range a.Buses {
		retime(a.Buses[i].Effects, beats)
	}
	retime(a.Master, beats)
}

func replay(t synth.Track, move func(from, to int) []int) []synth.Play {
	var out []synth.Play
	for _, p := range t.Play {
		c, _ := clip(t, p.Clip)
		n := clipBars(c)
		var copies [][]int // each copy's passes' starts, from 0
		for k := range max(p.Times, 1) {
			from := p.Bar - 1 + k*n
			for j, to := range move(from, from+n) {
				for len(copies) <= j {
					copies = append(copies, nil)
				}
				copies[j] = append(copies[j], to)
			}
		}
		for _, starts := range copies {
			slices.Sort(starts)
			for len(starts) > 0 {
				run := 1
				for run < len(starts) && starts[run] == starts[0]+run*n {
					run++
				}
				q := synth.Play{Clip: p.Clip, Bar: starts[0] + 1, Times: run}
				if run == 1 && p.Times == 0 {
					q.Times = 0
				}
				out = append(out, q)
				starts = starts[run:]
			}
		}
	}
	return out
}

// retime moves the song positions literal effects hold, in beats.
func retime(fx []synth.Effect, move func(x float64) []float64) {
	first := func(x float64) float64 {
		if xs := move(x); len(xs) > 0 {
			return xs[0]
		}
		return x
	}
	for i, e := range fx {
		switch v := e.(type) {
		case synth.AutoFilter:
			v.Points = movePoints(v.Points, move)
			fx[i] = v
		case synth.Fader:
			v.Points = movePoints(v.Points, move)
			fx[i] = v
		case synth.Panner:
			v.Points = movePoints(v.Points, move)
			fx[i] = v
		case synth.TapeStop:
			v.At, v.Until = first(v.At), first(v.Until)
			fx[i] = v
		case synth.Sidechain:
			if v.Hits != nil {
				var hits []float64
				for _, h := range v.Hits {
					hits = append(hits, move(h)...)
				}
				slices.Sort(hits)
				v.Hits = hits
				fx[i] = v
			}
		}
	}
}

func movePoints(points []synth.Point, move func(x float64) []float64) []synth.Point {
	if points == nil {
		return nil
	}
	var out []synth.Point
	for _, p := range points {
		for _, at := range move(p.At) {
			out = append(out, synth.Point{At: at, Value: p.Value})
		}
	}
	slices.SortStableFunc(out, func(a, b synth.Point) int {
		switch {
		case a.At < b.At:
			return -1
		case a.At > b.At:
			return 1
		}
		return 0
	})
	return out
}

// TempoStep is the resolution tempo is written at.
const TempoStep = 0.01

// ParseTempo reads a tempo typed into the transport: "120", "92.5 bpm".
func ParseTempo(s string) (float64, bool) {
	s = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(strings.ToLower(s)), "bpm"))
	x, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(x) || x < 20 || x > 400 {
		return 0, false
	}
	return Round(x, TempoStep), true
}

// FormatMeter writes beats per bar as the transport shows them: 4/4, or 3+3+3+4.
func FormatMeter(bars []int) string {
	if len(bars) == 0 {
		return "4/4"
	}
	if len(bars) == 1 {
		return fmt.Sprintf("%d/4", bars[0])
	}
	var parts []string
	for _, b := range bars {
		parts = append(parts, strconv.Itoa(b))
	}
	return strings.Join(parts, "+")
}

// ParseMeter reads a meter typed as FormatMeter writes it; 4/4 is nil.
func ParseMeter(s string) ([]int, bool) {
	s = strings.ReplaceAll(s, " ", "")
	if n, ok := strings.CutSuffix(s, "/4"); ok {
		s = n
	}
	var bars []int
	for f := range strings.SplitSeq(s, "+") {
		n, err := strconv.Atoi(f)
		if err != nil || n < 1 || n > 32 {
			return nil, false
		}
		bars = append(bars, n)
	}
	if len(bars) == 1 && bars[0] == 4 {
		return nil, true
	}
	return bars, true
}
