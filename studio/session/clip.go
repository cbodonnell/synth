package session

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/cbodonnell/synth"
)

// Grid is a clip's bars as the clip editor edits them: one line of events, each struck
// on a step and held for the steps after it. A line is one voice, so an event drawn
// under a held one cuts the held one short there.
type Grid struct {
	Hit    bool
	Bars   []int   // steps in each bar
	Events []Event // in step order, none reaching into the next
	// The bars as written, and their tokens, so a bar left alone is written as it was.
	src    []string
	tokens [][]string
}

// Event is a note, a chord, or a Hit clip's hit, struck on Step and sounding Len steps.
type Event struct {
	Step, Len int
	// The note or chord; none for a hit of the clip's Hit.
	Keys   []Pitch
	Hits   int // a ratchet's hits, spread over the event; 0 is one
	Accent int // 1 accented (! or X), -1 ghosted (? or o)
	raw    string
}

// Pitch is one key of an event, and how far it slides.
type Pitch struct {
	Key, Slide int
	name       string // as written
}

// Ref names a note: its event's step, and its key.
type Ref struct{ Step, Key int }

// Velocities the notation's accents play at, as a fraction of the clip's Vel.
const (
	AccentVel = 1.25
	GhostVel  = 0.55
)

// ReadGrid reads a clip's bars. A clip of Notes has none.
func ReadGrid(c synth.Clip) (Grid, error) {
	if len(c.Bars) == 0 {
		return Grid{}, errors.New("no bars: its notes are off the grid")
	}
	g := Grid{Hit: c.Hit != "", src: c.Bars}
	cur := -1
	for _, b := range c.Bars {
		tokens := strings.Fields(strings.ReplaceAll(b, "|", " "))
		g.Bars = append(g.Bars, len(tokens))
		g.tokens = append(g.tokens, tokens)
	}
	step := 0
	for _, tokens := range g.tokens {
		for _, tok := range tokens {
			switch tok {
			case ".":
				cur = -1
			case "-":
				if cur >= 0 {
					g.Events[cur].Len++
				}
			default:
				e, err := parseToken(tok, g.Hit)
				if err != nil {
					return Grid{}, err
				}
				e.Step, e.Len = step, 1
				g.Events = append(g.Events, e)
				cur = len(g.Events) - 1
			}
			step++
		}
	}
	return g, nil
}

func parseToken(tok string, hit bool) (Event, error) {
	e := Event{raw: tok}
	body := tok
	switch {
	case strings.HasSuffix(body, "!"):
		e.Accent, body = 1, body[:len(body)-1]
	case strings.HasSuffix(body, "?"):
		e.Accent, body = -1, body[:len(body)-1]
	}
	if b, count, ok := strings.Cut(body, "*"); ok {
		n, err := strconv.Atoi(count)
		if err != nil || n < 1 {
			return e, fmt.Errorf("bad ratchet %q", tok)
		}
		body, e.Hits = b, n
	}
	if hit {
		switch body {
		case "x":
			return e, nil
		case "X":
			e.Accent = 1
			return e, nil
		case "o":
			e.Accent = -1
			return e, nil
		}
	}
	for name := range strings.SplitSeq(body, ",") {
		from, to, slides := strings.Cut(name, "/")
		k, err := synth.ParseKey(from)
		if err != nil {
			return e, err
		}
		p := Pitch{Key: int(k), name: name}
		if slides {
			t, err := synth.ParseKey(to)
			if err != nil {
				return e, err
			}
			p.Slide = int(t) - p.Key
		}
		e.Keys = append(e.Keys, p)
	}
	return e, nil
}

func (e Event) same(o Event) bool {
	return e.Accent == o.Accent && max(e.Hits, 1) == max(o.Hits, 1) &&
		slices.EqualFunc(e.Keys, o.Keys, func(a, b Pitch) bool { return a.Key == b.Key && a.Slide == b.Slide })
}

// token writes the event as the notation does: as it was written while it says the same.
func (e Event) token(hit bool, sc Scale) string {
	if e.raw != "" {
		if was, err := parseToken(e.raw, hit); err == nil && was.same(e) {
			return e.raw
		}
	}
	ratchet := ""
	if e.Hits > 1 {
		ratchet = "*" + strconv.Itoa(e.Hits)
	}
	if hit && len(e.Keys) == 0 {
		return map[int]string{1: "X", 0: "x", -1: "o"}[e.Accent] + ratchet
	}
	var names []string
	for _, p := range e.Keys {
		names = append(names, p.write(sc))
	}
	return strings.Join(names, ",") + ratchet + map[int]string{1: "!", 0: "", -1: "?"}[e.Accent]
}

func (p Pitch) write(sc Scale) string {
	if p.name != "" {
		if was, err := parseToken(p.name, false); err == nil && len(was.Keys) == 1 &&
			was.Keys[0].Key == p.Key && was.Keys[0].Slide == p.Slide {
			return p.name
		}
	}
	if p.Slide != 0 {
		return sc.Spell(p.Key) + "/" + sc.Spell(p.Key+p.Slide)
	}
	return sc.Spell(p.Key)
}

// Write is the grid's bars: a bar left as it was is written as it was, and new notes
// are spelled as sc spells them.
func (g Grid) Write(sc Scale) []string {
	n := g.Steps()
	flat := make([]string, n)
	sounding := make([]bool, n)
	for i := range flat {
		flat[i] = "."
	}
	for _, e := range g.Events {
		flat[e.Step] = e.token(g.Hit, sc)
		for s := e.Step; s < e.Step+e.Len && s < n; s++ {
			sounding[s] = true
			if s > e.Step {
				flat[s] = "-"
			}
		}
	}
	var out []string
	at := 0
	for i, steps := range g.Bars {
		bar := flat[at : at+steps]
		if i < len(g.tokens) && len(g.tokens[i]) == steps {
			// A hold after a rest is a rest: keep it as written.
			for k := range bar {
				s := at + k
				if g.tokens[i][k] == "-" && !sounding[s] && (s == 0 || !sounding[s-1]) {
					bar[k] = "-"
				}
			}
			if slices.Equal(bar, g.tokens[i]) {
				out = append(out, g.src[i])
				at += steps
				continue
			}
		}
		out = append(out, strings.Join(bar, " "))
		at += steps
	}
	return out
}

// Steps is how many steps the clip holds.
func (g Grid) Steps() int {
	n := 0
	for _, b := range g.Bars {
		n += b
	}
	return n
}

// BarStep is the step bar b (from 0) of the clip begins on.
func (g Grid) BarStep(b int) int {
	n := 0
	for _, s := range g.Bars[:min(b, len(g.Bars))] {
		n += s
	}
	return n
}

// At is the index of the event sounding on step, or -1.
func (g Grid) At(step int) int {
	for i, e := range g.Events {
		if step >= e.Step && step < e.Step+e.Len {
			return i
		}
	}
	return -1
}

// Start is the index of the event struck on step, or -1.
func (g Grid) Start(step int) int {
	return slices.IndexFunc(g.Events, func(e Event) bool { return e.Step == step })
}

// Has reports whether the note is in the grid.
func (g Grid) Has(r Ref) bool {
	i := g.Start(r.Step)
	return i >= 0 && slices.ContainsFunc(g.Events[i].Keys, func(p Pitch) bool { return p.Key == r.Key })
}

func (g Grid) clone() Grid {
	g.Bars = slices.Clone(g.Bars)
	g.Events = slices.Clone(g.Events)
	for i := range g.Events {
		g.Events[i].Keys = slices.Clone(g.Events[i].Keys)
	}
	return g
}

// next is the step the event after step is struck on, or the grid's end.
func (g Grid) next(step int) int {
	for _, e := range g.Events {
		if e.Step > step {
			return e.Step
		}
	}
	return g.Steps()
}

// place strikes e: into the chord already struck there, or as a new event cutting
// short the one held over its step and held no further than the next.
func (g *Grid) place(e Event) {
	if e.Step < 0 || e.Step >= g.Steps() {
		return
	}
	if i := g.Start(e.Step); i >= 0 {
		ev := &g.Events[i]
		n := len(ev.Keys)
		for _, p := range e.Keys {
			if !slices.ContainsFunc(ev.Keys, func(q Pitch) bool { return q.Key == p.Key }) {
				ev.Keys = append(ev.Keys, p)
			}
		}
		if len(ev.Keys) > n {
			slices.SortStableFunc(ev.Keys, func(a, b Pitch) int { return a.Key - b.Key })
		}
		return
	}
	if i := g.At(e.Step); i >= 0 {
		g.Events[i].Len = e.Step - g.Events[i].Step
	}
	e.Len = max(1, min(e.Len, g.next(e.Step)-e.Step))
	i, _ := slices.BinarySearchFunc(g.Events, e.Step, func(x Event, s int) int { return x.Step - s })
	g.Events = slices.Insert(g.Events, i, e)
}

// Draw strikes keys on step, held for steps: a note, a chord, or with no keys a hit.
func (g Grid) Draw(step, steps int, keys ...int) Grid {
	g = g.clone()
	e := Event{Step: step, Len: steps}
	for _, k := range keys {
		e.Keys = append(e.Keys, Pitch{Key: k})
	}
	g.place(e)
	return g
}

// Rest silences step: the event struck there is gone, and one held over it stops short.
func (g Grid) Rest(step int) Grid {
	g = g.clone()
	if i := g.Start(step); i >= 0 {
		g.Events = slices.Delete(g.Events, i, i+1)
	}
	if i := g.At(step); i >= 0 {
		g.Events[i].Len = step - g.Events[i].Step
	}
	return g
}

// Hold holds whatever sounds on the step before step over it too.
func (g Grid) Hold(step int) Grid {
	if step <= 0 || step >= g.Steps() {
		return g
	}
	g = g.Rest(step)
	if i := g.At(step - 1); i >= 0 {
		g.Events[i].Len = step + 1 - g.Events[i].Step
	}
	return g
}

// Delete takes the notes out, and any event left without one.
func (g Grid) Delete(refs ...Ref) Grid {
	g = g.clone()
	for _, r := range refs {
		i := g.Start(r.Step)
		if i < 0 {
			continue
		}
		e := &g.Events[i]
		e.Keys = slices.DeleteFunc(e.Keys, func(p Pitch) bool { return p.Key == r.Key })
		if len(e.Keys) == 0 {
			g.Events = slices.Delete(g.Events, i, i+1)
		}
	}
	return g
}

// Move moves the notes by steps and semitones, as far as the clip and the keyboard
// allow, and returns where they went. Notes struck together stay together.
func (g Grid) Move(refs []Ref, steps, semis int) (Grid, []Ref) {
	lo, hi, klo, khi := g.Steps(), -1, 128, -1
	for _, r := range refs {
		if g.Has(r) {
			lo, hi = min(lo, r.Step), max(hi, r.Step)
			klo, khi = min(klo, r.Key), max(khi, r.Key)
		}
	}
	if hi < 0 {
		return g, refs
	}
	steps = max(-lo, min(steps, g.Steps()-1-hi))
	semis = max(-klo, min(semis, 127-khi))
	if steps == 0 && semis == 0 {
		return g, refs
	}
	var moved []Event
	for _, e := range g.Events {
		m := Event{Step: e.Step + steps, Len: e.Len, Hits: e.Hits, Accent: e.Accent}
		for _, p := range e.Keys {
			if !slices.Contains(refs, Ref{e.Step, p.Key}) {
				continue
			}
			if semis != 0 {
				p = Pitch{Key: p.Key + semis, Slide: p.Slide}
			}
			m.Keys = append(m.Keys, p)
		}
		if len(m.Keys) == len(e.Keys) {
			m.raw = e.raw
		}
		if len(m.Keys) > 0 {
			moved = append(moved, m)
		}
	}
	g = g.Delete(refs...)
	var out []Ref
	for _, m := range moved {
		g.place(m)
		for _, p := range m.Keys {
			out = append(out, Ref{m.Step, p.Key})
		}
	}
	return g, out
}

// Stretch holds the event struck on step for steps, no further than the next.
func (g Grid) Stretch(step, steps int) Grid {
	i := g.Start(step)
	if i < 0 {
		return g
	}
	g = g.clone()
	g.Events[i].Len = max(1, min(steps, g.next(step)-step))
	return g
}

// Accent accents (1), plays plainly (0) or ghosts (-1) the event struck on step.
func (g Grid) Accent(step, accent int) Grid {
	i := g.Start(step)
	if i < 0 {
		return g
	}
	g = g.clone()
	g.Events[i].Accent = max(-1, min(1, accent))
	return g
}

// Ratchet makes the event struck on step hits even hits.
func (g Grid) Ratchet(step, hits int) Grid {
	i := g.Start(step)
	if i < 0 {
		return g
	}
	g = g.clone()
	g.Events[i].Hits = max(1, min(hits, 16))
	if g.Events[i].Hits == 1 {
		g.Events[i].Hits = 0
	}
	return g
}

// Slide slides the note to the key; to its own key, it doesn't slide.
func (g Grid) Slide(r Ref, to int) Grid {
	i := g.Start(r.Step)
	if i < 0 {
		return g
	}
	g = g.clone()
	for k, p := range g.Events[i].Keys {
		if p.Key == r.Key {
			g.Events[i].Keys[k].Slide = to - p.Key
		}
	}
	return g
}

// Cycle steps a drum grid's cell round . x X o.
func (g Grid) Cycle(step int) Grid {
	i := g.Start(step)
	switch {
	case i < 0:
		return g.Draw(step, 1)
	case len(g.Events[i].Keys) > 0 || g.Events[i].Accent < 0:
		return g.Rest(step)
	case g.Events[i].Accent == 0:
		return g.Accent(step, 1)
	}
	return g.Accent(step, -1)
}

// Halve makes every step two, so a step half as long writes the same notes.
func (g Grid) Halve() Grid {
	g = g.clone()
	g.src, g.tokens = nil, nil
	for i := range g.Bars {
		g.Bars[i] *= 2
	}
	for i := range g.Events {
		g.Events[i].Step *= 2
		g.Events[i].Len *= 2
	}
	return g
}

// Double makes every two steps one, when every event starts and ends on a step that
// twice as long has.
func (g Grid) Double() (Grid, bool) {
	for _, b := range g.Bars {
		if b%2 != 0 {
			return g, false
		}
	}
	for _, e := range g.Events {
		if e.Step%2 != 0 || e.Len%2 != 0 {
			return g, false
		}
	}
	g = g.clone()
	g.src, g.tokens = nil, nil
	for i := range g.Bars {
		g.Bars[i] /= 2
	}
	for i := range g.Events {
		g.Events[i].Step /= 2
		g.Events[i].Len /= 2
	}
	return g, true
}

// Velocity is what the event plays at in a clip of base velocity vel, as Phrase plays it.
func (e Event) Velocity(vel float64) float64 {
	if vel <= 0 {
		vel = 0.8
	}
	switch e.Accent {
	case 1:
		vel *= AccentVel
	case -1:
		vel *= GhostVel
	}
	return min(vel, 1)
}
