package studio

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/cbodonnell/synth"
)

// Arrange writes a first draft of an Arrangement from any Song: each part a track, its
// notes cut into bars, repeated bars folded into one clip played many times. A part
// whose notes sit on a step (within the drift humanizing gives them) is written as a
// grid, and any other as Notes. Humanized timing can't be read back to its seed, so a
// humanized part's draft drifts its own way.
func Arrange(s synth.Song) synth.Arrangement {
	a := synth.Arrangement{
		ID: s.ID, Name: s.Name,
		Vary: s.Vary, Voices: s.Voices, Priority: s.Priority, Target: s.Target,
	}
	if s.Meter.Beat <= 0 {
		return arrangeSound(a, s)
	}
	a.Tempo = round(60/s.Meter.Beat, 6)
	a.Meter = s.Meter.Bars
	m := synth.Meter{Beat: s.Meter.Beat, Bars: s.Meter.Bars}

	length := s.Length / m.Beat
	if !s.Loop {
		length = math.Max(length, songEnd(s)/m.Beat)
	}
	bars := 0
	for m.BarStart(bars) < length-1e-6 {
		bars++
	}
	if s.Loop && s.LoopStart > 0 {
		intro := 0
		for m.BarStart(intro) < s.LoopStart/m.Beat-1e-6 {
			intro++
		}
		a.Sections = []synth.Section{{Name: "Intro", Bars: intro}, {Name: "A", Bars: bars - intro}}
		a.Loop = "A"
	} else {
		a.Sections = []synth.Section{{Name: "A", Bars: max(bars, 1)}}
		if s.Loop {
			a.Loop = "A"
		}
	}

	for _, p := range s.Parts {
		t := track(p)
		t.Inserts = inBeats(t.Inserts, a.Tempo, s.Parts)
		t.Clips, t.Play = arrangeNotes(p.Notes, m, bars)
		a.Tracks = append(a.Tracks, t)
	}
	for _, b := range s.Buses {
		b.Effects = inBeats(b.Effects, a.Tempo, s.Parts)
		a.Buses = append(a.Buses, b)
	}
	a.Master = inBeats(s.Master, a.Tempo, s.Parts)
	return a
}

// A sound effect has no bars: each layer is one clip of Notes in seconds.
func arrangeSound(a synth.Arrangement, s synth.Song) synth.Arrangement {
	a.Sections = []synth.Section{{Name: "A", Bars: 1}}
	for _, p := range s.Parts {
		t := track(p)
		t.Clips = []synth.Clip{{Name: "a", Notes: p.Notes}}
		t.Play = []synth.Play{{Clip: "a", Bar: 1}}
		a.Tracks = append(a.Tracks, t)
	}
	a.Buses, a.Master = s.Buses, s.Master
	return a
}

func track(p synth.Part) synth.Track {
	return synth.Track{
		Name: p.Name, Patch: p.Patch, Level: p.Level, Pan: p.Pan,
		Inserts: p.Inserts, Sends: p.Sends, Out: p.Out,
	}
}

func songEnd(s synth.Song) float64 {
	end := 0.0
	for _, p := range s.Parts {
		for _, n := range p.Notes {
			end = math.Max(end, n.At+n.Dur)
		}
	}
	return end
}

// inBeats rewrites a chain's song times in beats, and keys a Sidechain whose hits are
// exactly some parts' onsets by those parts.
func inBeats(fx []synth.Effect, tempo float64, parts []synth.Part) []synth.Effect {
	if fx == nil {
		return nil
	}
	out := make([]synth.Effect, len(fx))
	for i, e := range fx {
		if sc, ok := e.(synth.Sidechain); ok && len(sc.Hits) > 0 {
			if key := sidechainKey(sc.Hits, parts); key != "" {
				sc.Hits, sc.Key = nil, key
				out[i] = sc
				continue
			}
		}
		out[i] = synth.InBeats(e, tempo)
	}
	return out
}

func sidechainKey(hits []float64, parts []synth.Part) string {
	var names []string
	var notes []synth.Note
	for _, p := range parts {
		on := synth.Onsets(p.Notes)
		if len(on) == 0 || !slices.ContainsFunc(hits, func(h float64) bool { return h == on[0] }) {
			continue
		}
		if all(on, func(t float64) bool { return slices.Contains(hits, t) }) {
			names = append(names, p.Name)
			notes = append(notes, p.Notes...)
		}
	}
	if !slices.Equal(synth.Onsets(notes), hits) {
		return ""
	}
	return strings.Join(names, ", ")
}

func all[T any](s []T, f func(T) bool) bool {
	for _, x := range s {
		if !f(x) {
			return false
		}
	}
	return true
}

// Steps tried for a grid, in beats, coarsest first.
var steps = []float64{4, 3, 2, 1, 1.0 / 2, 1.0 / 3, 1.0 / 4, 1.0 / 6, 1.0 / 8, 1.0 / 12, 1.0 / 16, 1.0 / 24, 1.0 / 32}

// A chord or single note on the grid: its notes in the order they sound.
type token struct {
	step  int // from the song's start
	notes []synth.Note
	holds int // steps it spans
}

type grid struct {
	clip   synth.Clip // the settings, without a name or bars
	tokens []token
}

func arrangeNotes(notes []synth.Note, m synth.Meter, bars int) ([]synth.Clip, []synth.Play) {
	if len(notes) == 0 {
		return nil, nil
	}
	for _, step := range steps {
		if g, ok := fit(notes, m, bars, step); ok {
			return g.clips(m, bars)
		}
	}
	return noteClips(notes, m, bars)
}

// fit reads notes as a grid of step beats, or reports that they aren't one.
func fit(notes []synth.Note, m synth.Meter, bars int, step float64) (grid, bool) {
	for b := range bars {
		if n := float64(m.BarBeats(b)) / step; math.Abs(n-math.Round(n)) > 1e-9 {
			return grid{}, false
		}
	}
	sec := step * m.Beat

	byStep := map[int][]synth.Note{}
	for _, n := range notes {
		k := int(math.Floor(n.At/sec + 1e-6))
		if n.At-float64(k)*sec > math.Min(0.45*sec, 0.15) || n.Key != math.Round(n.Key) {
			return grid{}, false
		}
		byStep[k] = append(byStep[k], n)
	}
	var tokens []token
	for k, ns := range byStep {
		slices.SortStableFunc(ns, func(a, b synth.Note) int { return cmp.Compare(a.At, b.At) })
		tokens = append(tokens, token{step: k, notes: ns})
	}
	slices.SortFunc(tokens, func(a, b token) int { return a.step - b.step })

	var c synth.Clip
	c.Step = step

	// Humanizing only ever drifts late, so the earliest note on each kind of step marks it.
	even, odd := math.Inf(1), math.Inf(1)
	for _, t := range tokens {
		r := t.notes[0].At - float64(t.step)*sec
		if t.step%2 == 0 {
			even = math.Min(even, r)
		} else {
			odd = math.Min(odd, r)
		}
	}
	if !math.IsInf(odd, 1) {
		base := 0.0
		if !math.IsInf(even, 1) {
			base = even
		}
		c.Swing = round(math.Max(odd-base, 0)/sec, 2)
	}
	drift := 0.0
	for _, t := range tokens {
		r := t.notes[0].At - float64(t.step)*sec
		if t.step%2 != 0 {
			r -= c.Swing * sec
		}
		drift = math.Max(drift, r)
	}
	if drift > 0.0005 {
		c.Humanize = round(drift, 3)
	}

	strums := map[float64]bool{}
	for _, t := range tokens {
		if len(t.notes) > 1 {
			strums[round(t.notes[1].At-t.notes[0].At, 6)] = true
		}
	}
	if len(strums) > 1 {
		return grid{}, false
	}
	for s := range strums {
		c.Strum = s
	}
	for i, t := range tokens {
		for k, n := range t.notes {
			if math.Abs(n.At-t.notes[0].At-float64(k)*c.Strum) > 1e-6 {
				return grid{}, false
			}
		}
		if i > 0 && t.step == tokens[i-1].step {
			return grid{}, false
		}
	}

	glide := tokens[0].notes[0].SlideTime
	for _, t := range tokens {
		for _, n := range t.notes {
			if n.SlideTime != glide {
				return grid{}, false
			}
		}
	}
	c.Glide = glide

	gap := func(i int) int {
		if i+1 < len(tokens) {
			return tokens[i+1].step - tokens[i].step
		}
		return 1 << 20
	}
	tightest := 0
	for i := range tokens {
		if gap(i) < gap(tightest) {
			tightest = i
		}
	}
	var gates []float64
	for n := 1; n <= min(gap(tightest), 64); n++ {
		if g := tokens[tightest].notes[0].Dur / (float64(n) * sec); g >= 0.5-1e-9 && g <= 1+1e-9 {
			gates = append(gates, g)
		}
	}
	for _, g := range gates {
		held, ok := holds(tokens, g, sec, c.Strum, gap)
		if !ok {
			continue
		}
		for i := range tokens {
			tokens[i].holds = held[i]
		}
		if math.Abs(g-1) > 1e-6 {
			c.Gate = round(g, 6)
		}
		c.Vel = baseVel(tokens)
		if c.Vel == 0.8 {
			c.Vel = 0
		}
		return grid{clip: c, tokens: tokens}, true
	}
	return grid{}, false
}

func holds(tokens []token, gate, sec, strum float64, gap func(int) int) ([]int, bool) {
	held := make([]int, len(tokens))
	for i, t := range tokens {
		n := int(math.Round(t.notes[0].Dur / (gate * sec)))
		if n < 1 || n > gap(i) {
			return nil, false
		}
		span := float64(n) * sec
		for k, note := range t.notes {
			if math.Abs(note.Dur-math.Max(span*gate-float64(k)*strum, span*0.25)) > 1e-6 {
				return nil, false
			}
		}
		held[i] = n
	}
	return held, true
}

// As Phrase writes them.
const (
	accent = 1.25
	ghost  = 0.55
)

// baseVel is the velocity the line is written at: one that accounts for every note as
// played, accented, or ghosted, or failing that (a humanized line), the median.
func baseVel(tokens []token) float64 {
	var vels []float64
	for _, t := range tokens {
		vels = append(vels, t.notes[0].Vel)
	}
	candidates := []float64{0.8}
	for _, v := range vels {
		candidates = append(candidates, v, v/accent)
	}
	for _, b := range candidates {
		if all(vels, func(v float64) bool { return markOf(v, b) != "~" }) {
			return round(b, 6)
		}
	}
	// Humanizing moves velocity up to 15% either way.
	slices.Sort(vels)
	median := round(vels[len(vels)/2], 2)
	for _, b := range []float64{0.8, median, round(median/accent, 2)} {
		if all(vels, func(v float64) bool {
			_, d := nearest(v, b)
			return d <= 0.151
		}) {
			return b
		}
	}
	return median
}

// nearest is the mark whose level is closest to vel against base, and how far off it
// is as a fraction of that level.
func nearest(vel, base float64) (string, float64) {
	mark, off := "", math.Inf(1)
	for m, level := range map[string]float64{"": base, "!": math.Min(base*accent, 1), "?": base * ghost} {
		if d := math.Abs(vel/level - 1); d < off {
			mark, off = m, d
		}
	}
	return mark, off
}

// markOf is the mark that plays vel against base: "", "!", "?", or "~" for none exactly.
func markOf(vel, base float64) string {
	switch {
	case math.Abs(vel-base) < 1e-6:
		return ""
	case math.Abs(vel-math.Min(base*accent, 1)) < 1e-6:
		return "!"
	case math.Abs(vel-base*ghost) < 1e-6:
		return "?"
	}
	return "~"
}

// mark is a token's accent or ghost against the clip's velocity.
func (g grid) mark(t token) string {
	vel := g.clip.Vel
	if vel == 0 {
		vel = 0.8
	}
	if m := markOf(t.notes[0].Vel, vel); m != "~" {
		return m
	}
	m, _ := nearest(t.notes[0].Vel, vel)
	return m
}

// hit is the one key every token plays alone, for a drum line, or "".
func (g grid) hit() string {
	key := g.tokens[0].notes[0].Key
	for _, t := range g.tokens {
		if len(t.notes) > 1 || t.notes[0].Key != key || t.notes[0].Slide != 0 {
			return ""
		}
	}
	return synth.KeyName(int(key))
}

func (g grid) text(t token, hit string) string {
	mark := g.mark(t)
	if hit != "" {
		switch mark {
		case "!":
			return "X"
		case "?":
			return "o"
		}
		return "x"
	}
	keys := make([]string, len(t.notes))
	for i, n := range t.notes {
		keys[i] = synth.KeyName(int(n.Key))
		if n.Slide != 0 {
			keys[i] += "/" + synth.KeyName(int(n.Key+n.Slide))
		}
	}
	return strings.Join(keys, ",") + mark
}

// clips cuts the grid into runs of bars no note is held across, and folds runs that
// read the same into one clip.
func (g grid) clips(m synth.Meter, bars int) ([]synth.Clip, []synth.Play) {
	perBeat := 1 / g.clip.Step
	stepAt := func(bar int) int { return int(math.Round(m.BarStart(bar) * perBeat)) }
	hit := g.hit()

	var runs []run
	next := 0
	for bar := 0; bar < bars; {
		end := bar + 1
		for next < len(g.tokens) && g.tokens[next].step < stepAt(end) {
			t := g.tokens[next]
			for t.step+t.holds > stepAt(end) && end < bars {
				end++
			}
			next++
		}
		r := run{bar: bar, bars: end - bar}
		cells := make([]string, stepAt(end)-stepAt(bar))
		for i := range cells {
			cells[i] = "."
		}
		empty := true
		for _, t := range g.tokens {
			if t.step < stepAt(bar) || t.step >= stepAt(end) {
				continue
			}
			empty = false
			at := t.step - stepAt(bar)
			cells[at] = g.text(t, hit)
			for h := 1; h < t.holds && at+h < len(cells); h++ {
				cells[at+h] = "-"
			}
		}
		if !empty {
			for b := bar; b < end; b++ {
				r.text = append(r.text, strings.Join(cells[stepAt(b)-stepAt(bar):stepAt(b+1)-stepAt(bar)], " "))
			}
			runs = append(runs, r)
		}
		bar = end
	}
	return fold(runs, func(name string, r run) synth.Clip {
		c := g.clip
		c.Name, c.Hit, c.Bars = name, hit, r.text
		return c
	})
}

type run struct {
	bar, bars int
	text      []string
	notes     []synth.Note
}

func (r run) same(o run) bool {
	return slices.Equal(r.text, o.text) && slices.Equal(r.notes, o.notes) && r.bars == o.bars
}

func fold(runs []run, clip func(name string, r run) synth.Clip) ([]synth.Clip, []synth.Play) {
	runs = group(runs)
	var clips []synth.Clip
	var kinds []run
	var plays []synth.Play
	for _, r := range runs {
		k := slices.IndexFunc(kinds, r.same)
		if k < 0 {
			k = len(kinds)
			kinds = append(kinds, r)
			clips = append(clips, clip(clipName(k), r))
		}
		if n := len(plays); n > 0 && plays[n-1].Clip == clips[k].Name && plays[n-1].Bar-1+max(plays[n-1].Times, 1)*r.bars == r.bar {
			plays[n-1].Times = max(plays[n-1].Times, 1) + 1
			continue
		}
		plays = append(plays, synth.Play{Clip: clips[k].Name, Bar: r.bar + 1})
	}
	return clips, plays
}

// group joins runs that repeat back to back as a unit of up to four, so a two-bar riff
// played four times is one clip played four times.
func group(runs []run) []run {
	var out []run
	for i := 0; i < len(runs); {
		unit, reps := 1, 1
		for k := 2; k <= 4; k++ {
			if n := repeats(runs[i:], k); n >= 2 && n*k > unit*reps {
				unit, reps = k, n
			}
		}
		for range reps {
			out = append(out, join(runs[i:i+unit]))
			i += unit
		}
	}
	return out
}

// repeats counts how many times rs's first k runs, all grid and back to back, play over.
func repeats(rs []run, k int) int {
	joined := func(i int) bool {
		return rs[i].text != nil && (i == 0 || rs[i-1].bar+rs[i-1].bars == rs[i].bar)
	}
	n := 0
	for (n+1)*k <= len(rs) {
		for j := range k {
			at := n*k + j
			if !joined(at) || !rs[at].same(rs[j]) {
				return n
			}
		}
		n++
	}
	return n
}

func join(rs []run) run {
	r := rs[0]
	for _, o := range rs[1:] {
		r.bars += o.bars
		r.text = append(slices.Clone(r.text), o.text...)
	}
	return r
}

func clipName(k int) string {
	name := string(rune('a' + k%26))
	if k >= 26 {
		name += fmt.Sprint(k / 26)
	}
	return name
}

// noteClips writes notes as they are, in runs of bars no note rings across.
func noteClips(notes []synth.Note, m synth.Meter, bars int) ([]synth.Clip, []synth.Play) {
	notes = slices.Clone(notes)
	slices.SortStableFunc(notes, func(a, b synth.Note) int { return cmp.Compare(a.At, b.At) })
	start := func(bar int) float64 { return m.BarStart(bar) * m.Beat }

	var runs []run
	next := 0
	for bar := 0; bar < bars; {
		end := bar + 1
		var in []synth.Note
		for next < len(notes) && (notes[next].At < start(end)-1e-9 || end >= bars) {
			n := notes[next]
			for n.At+n.Dur > start(end)+1e-9 && end < bars {
				end++
			}
			n.At = round((n.At-start(bar))/m.Beat, 9)
			n.Dur = round(n.Dur/m.Beat, 9)
			n.SlideTime = round(n.SlideTime, 9)
			in = append(in, n)
			next++
		}
		if len(in) > 0 {
			runs = append(runs, run{bar: bar, bars: end - bar, notes: in})
		}
		bar = end
	}
	return fold(runs, func(name string, r run) synth.Clip {
		c := synth.Clip{Name: name, Notes: r.notes}
		if r.bars > 1 {
			c.Length = r.bars
		}
		return c
	})
}

func round(x float64, digits int) float64 {
	p := math.Pow(10, float64(digits))
	return math.Round(x*p) / p
}
