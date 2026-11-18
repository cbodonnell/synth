package session

import (
	"fmt"
	"maps"
	"math"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/cbodonnell/aac/golit"
	"github.com/cbodonnell/synth"
)

// Chain is a chain of effects: a track's inserts, a bus's effects, or the master.
type Chain struct {
	Kind  ChainKind
	Index int // the track's or the bus's
}

type ChainKind int

const (
	TrackChain ChainKind = iota
	BusChain
	MasterChain
)

// Path is the chain's path into the arrangement, as Doc.Custom takes it.
func (c Chain) Path() []any {
	switch c.Kind {
	case BusChain:
		return []any{"Buses", c.Index, "Effects"}
	case MasterChain:
		return []any{"Master"}
	}
	return []any{"Tracks", c.Index, "Inserts"}
}

// In is the chain in a, nil when a has no such track or bus.
func (c Chain) In(a *synth.Arrangement) *[]synth.Effect {
	switch {
	case c.Kind == MasterChain:
		return &a.Master
	case c.Kind == BusChain && c.Index >= 0 && c.Index < len(a.Buses):
		return &a.Buses[c.Index].Effects
	case c.Kind == TrackChain && c.Index >= 0 && c.Index < len(a.Tracks):
		return &a.Tracks[c.Index].Inserts
	}
	return nil
}

// AddEffect puts a copy of e into the chain before its effect at, or at its end.
func AddEffect(a *synth.Arrangement, c Chain, at int, e synth.Effect) {
	fx := c.In(a)
	if fx == nil {
		return
	}
	e = clone(reflect.ValueOf(&e).Elem()).Interface().(synth.Effect)
	*fx = slices.Insert(*fx, max(0, min(at, len(*fx))), e)
}

func RemoveEffect(a *synth.Arrangement, c Chain, i int) {
	if fx := c.In(a); fx != nil && i >= 0 && i < len(*fx) {
		if *fx = slices.Delete(*fx, i, i+1); len(*fx) == 0 {
			*fx = nil
		}
	}
}

// MoveEffect moves the chain's effect from one place to another, counted as the chain
// is once it has been taken out.
func MoveEffect(a *synth.Arrangement, c Chain, from, to int) {
	fx := c.In(a)
	if fx == nil || from < 0 || from >= len(*fx) {
		return
	}
	e := (*fx)[from]
	*fx = slices.Delete(*fx, from, from+1)
	*fx = slices.Insert(*fx, max(0, min(to, len(*fx))), e)
}

var effectNames = []string{
	"EQ", "Drive", "Chorus", "Phaser", "Delay", "Reverb", "Compressor", "Crush", "Limiter",
	"Sidechain", "AutoFilter", "Fader", "Panner", "Normalize",
}

// EffectNames are synth's effects as the browser lists them.
func EffectNames() []string { return effectNames }

// NewEffect is one of synth's effects as it starts in a chain.
func NewEffect(name string) (synth.Effect, bool) {
	switch name {
	case "EQ":
		return synth.EQ{Bands: []synth.Band{{Kind: synth.Peak, Freq: 1000}}}, true
	case "Drive":
		return synth.Drive{Amount: 2}, true
	case "Chorus":
		return synth.Chorus{Mix: 0.5}, true
	case "Phaser":
		return synth.Phaser{Rate: 0.5, Depth: 2}, true
	case "Delay":
		return synth.Delay{Time: 0.75, Feedback: 0.3, Mix: 0.3}, true
	case "Reverb":
		return synth.Reverb{Size: 0.5, Damp: 0.5, Mix: 0.3}, true
	case "Compressor":
		return synth.Compressor{Threshold: -18}, true
	case "Crush":
		return synth.Crush{Bits: 8}, true
	case "Limiter":
		return synth.Limiter{}, true
	case "Sidechain":
		return synth.Sidechain{Depth: 0.5}, true
	case "AutoFilter":
		return synth.AutoFilter{}, true
	case "Fader":
		return synth.Fader{}, true
	case "Panner":
		return synth.Panner{}, true
	case "Normalize":
		return synth.Normalize{Peak: -6}, true
	}
	return nil, false
}

// PatchNames are synth's patches that take no arguments, as the browser lists them.
func PatchNames() []string { return slices.Sorted(maps.Keys(patchCalls)) }

// SetPatch sets the track's patch to one of synth's patches, written as its call.
func (d *Doc) SetPatch(track int, name string) bool {
	f, ok := patchCalls[name]
	if !ok || track < 0 || track >= len(d.now.Tracks) {
		return false
	}
	return d.respell(func(a *synth.Arrangement) { a.Tracks[track].Patch = f() },
		map[string]*golit.Spelling{fmt.Sprintf("Tracks[%d].Patch", track): {Kind: golit.Opaque, Source: "patches." + name + "()"}})
}

// AddTrack adds a track playing one of synth's patches, written as its call, and
// returns its name.
func (d *Doc) AddTrack(patch string) (string, bool) {
	f, ok := patchCalls[patch]
	if !ok {
		return "", false
	}
	name := words(patch)
	for n := 2; slices.ContainsFunc(d.now.Tracks, func(t synth.Track) bool { return t.Name == name }); n++ {
		name = fmt.Sprintf("%s %d", words(patch), n)
	}
	i := len(d.now.Tracks)
	ok = d.respell(func(a *synth.Arrangement) { a.Tracks = append(a.Tracks, synth.Track{Name: name, Patch: f()}) },
		map[string]*golit.Spelling{fmt.Sprintf("Tracks[%d].Patch", i): {Kind: golit.Opaque, Source: "patches." + patch + "()"}})
	return name, ok
}

// words splits a Go name into lower-case words: HouseKick is "house kick".
func words(name string) string {
	var b strings.Builder
	for i, r := range name {
		if unicode.IsUpper(r) && i > 0 {
			b.WriteByte(' ')
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

// MakeLiteral writes the value at path, now written as code, as the literal it makes,
// cut loose from the code: the value and anything holding it that is code too.
func (d *Doc) MakeLiteral(path ...any) bool {
	ps := litPath(path)
	if len(ps) == 0 {
		return false
	}
	at := ps[len(ps)-1]
	spell := map[string]*golit.Spelling{}
	for p := range d.spell {
		if golit.Under(at, p) || golit.Under(p, at) {
			spell[p] = nil
		}
	}
	return len(spell) > 0 && d.respell(func(*synth.Arrangement) {}, spell)
}

// Set sets the value at a path into the arrangement, as Doc.Custom takes it, to x
// converted to its type, making any nil pointer on the way. It reports whether the path
// holds a value x converts to.
func Set(a *synth.Arrangement, path []any, x any) bool {
	return set(reflect.ValueOf(a).Elem(), path, reflect.ValueOf(x))
}

func set(v reflect.Value, path []any, x reflect.Value) bool {
	if len(path) == 0 {
		if !x.IsValid() || !x.Type().ConvertibleTo(v.Type()) {
			return false
		}
		v.Set(x.Convert(v.Type()))
		return true
	}
	switch v.Kind() {
	case reflect.Interface:
		if v.IsNil() {
			return false
		}
		c := reflect.New(v.Elem().Type()).Elem()
		c.Set(v.Elem())
		if !set(c, path, x) {
			return false
		}
		v.Set(c)
		return true
	case reflect.Pointer:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return set(v.Elem(), path, x)
	}
	switch s := path[0].(type) {
	case string:
		if v.Kind() != reflect.Struct {
			return false
		}
		f := v.FieldByName(s)
		return f.IsValid() && f.CanSet() && set(f, path[1:], x)
	case int:
		return v.Kind() == reflect.Slice && s >= 0 && s < v.Len() && set(v.Index(s), path[1:], x)
	}
	return false
}

// Get is the value at a path into the arrangement.
func Get(a synth.Arrangement, path []any) (reflect.Value, bool) {
	v := reflect.ValueOf(a)
	for _, step := range path {
		for v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
			if v.IsNil() {
				return reflect.Value{}, false
			}
			v = v.Elem()
		}
		switch s := step.(type) {
		case string:
			if v.Kind() != reflect.Struct {
				return reflect.Value{}, false
			}
			if v = v.FieldByName(s); !v.IsValid() {
				return v, false
			}
		case int:
			if v.Kind() != reflect.Slice || s < 0 || s >= v.Len() {
				return reflect.Value{}, false
			}
			v = v.Index(s)
		}
	}
	return v, true
}

// ControlKind is how a device shows a field.
type ControlKind int

const (
	Knob   ControlKind = iota
	Cycle              // an enum, stepped through its names
	Toggle             // a bool
	Text               // a string, typed
	Points             // automation's points, edited in its lane
	Note               // a value shown, not edited
)

// Control is a field of a device.
type Control struct {
	Label string
	Path  []any // into the arrangement
	Kind  ControlKind
	Param synth.Param // a Knob's, or an Automation's points' values
	Value float64     // a Knob's, or a Cycle's index
	On    bool        // a Toggle's
	Names []string    // a Cycle's
	Text  string      // a Text's, or what an Automation or a Note shows
}

var lfoType = reflect.TypeFor[synth.LFO]()

// Controls are the controls of v, the value at path in the arrangement: synth's own
// parameters where it has a table for v's type, and otherwise what reflection finds.
// Fields holding structs are their fields, labelled under theirs. Only the named fields
// are shown, when there are any.
func Controls(v any, path []any, fields ...string) []Control {
	return controls(reflect.ValueOf(v), path, "", fields)
}

func controls(v reflect.Value, path []any, prefix string, only []string) []Control {
	for v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	t := v.Type()
	table, _ := synth.Params(t)
	var out []Control
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() || (len(only) > 0 && !slices.Contains(only, f.Name)) {
			continue
		}
		fv := v.Field(i)
		c := Control{Label: prefix + f.Name, Path: append(slices.Clone(path), f.Name)}
		if names, ok := enums[f.Type]; ok {
			c.Kind, c.Value, c.Names = Cycle, float64(fv.Int()), names
			out = append(out, c)
			continue
		}
		switch {
		case fv.CanFloat() || fv.CanInt():
			c.Value = number64(fv)
			p, ok := param(table, f.Name)
			switch {
			case t == lfoType && f.Name == "Depth":
				p = synth.LFODepth(synth.LFOTarget(v.FieldByName("Target").Int()))
			case !ok:
				p = guess(f.Name, c.Value, fv.CanInt())
			}
			c.Param = p
		case fv.Kind() == reflect.Bool:
			c.Kind, c.On = Toggle, fv.Bool()
		case fv.Kind() == reflect.String:
			c.Kind, c.Text = Text, fv.String()
		case f.Type == pointsType:
			p, ok := param(table, f.Name)
			if !ok {
				p = guess(f.Name, 1, false)
			}
			c.Kind, c.Param, c.Text = Points, p, fmt.Sprintf("automation, %d points", fv.Len())
		case fv.Kind() == reflect.Struct || fv.Kind() == reflect.Pointer:
			out = append(out, controls(fv, c.Path, c.Label+" ", nil)...)
			continue
		case fv.Kind() == reflect.Slice && fv.Type().Elem().Kind() == reflect.Struct:
			for j := range fv.Len() {
				out = append(out, controls(fv.Index(j), append(slices.Clone(c.Path), j), fmt.Sprintf("%s %d ", c.Label, j+1), nil)...)
			}
			continue
		case fv.Kind() == reflect.Slice:
			if fv.Len() == 0 {
				continue
			}
			c.Kind, c.Text = Note, fmt.Sprintf("%d values", fv.Len())
		default:
			continue
		}
		out = append(out, c)
	}
	return out
}

func param(table []synth.Param, field string) (synth.Param, bool) {
	i := slices.IndexFunc(table, func(p synth.Param) bool { return p.Field == field })
	if i < 0 {
		return synth.Param{}, false
	}
	return table[i], true
}

// guess is a parameter for a game's own effect's field: plain, from 0 to a power of ten
// past twice its value, or as far below 0 when it's negative.
func guess(field string, x float64, whole bool) synth.Param {
	hi := 1.0
	for hi < 2*math.Abs(x) {
		hi *= 10
	}
	p := synth.Param{Field: field, Max: hi}
	if x < 0 {
		p.Min = -hi
	}
	if whole {
		p.Step = 1
	}
	return p
}

func number64(v reflect.Value) float64 {
	if v.CanInt() {
		return float64(v.Int())
	}
	return v.Float()
}

// Group is a part of the patch device: the controls of one of its parts, and what
// it's drawn as.
type Group struct {
	Title    string
	Path     []any
	Controls []Control
	Curve    Curve
}

type Curve int

const (
	NoCurve Curve = iota
	EnvelopeCurve
	ResponseCurve
)

// PatchGroups are the patch device's groups for the patch at path: each oscillator,
// unison, the amp envelope, the filter and its envelope, each LFO, the sweep, and its
// drive and levels. An oscillator shows FM only on a sine and the string only on a
// pluck, unless they're set.
func PatchGroups(p synth.Patch, path []any) []Group {
	at := func(steps ...any) []any { return append(slices.Clone(path), steps...) }
	var out []Group
	for i, o := range p.Oscs {
		var fields []string
		for j := range reflect.TypeFor[synth.Osc]().NumField() {
			f := reflect.TypeFor[synth.Osc]().Field(j)
			switch {
			case f.Name == "FM" && o.Wave != synth.Sine && o.FM == (synth.FM{}):
			case f.Name == "String" && o.Wave != synth.Pluck && o.String == (synth.String{}):
			default:
				fields = append(fields, f.Name)
			}
		}
		out = append(out, Group{Title: fmt.Sprintf("Osc %d", i+1), Path: at("Oscs", i), Controls: Controls(o, at("Oscs", i), fields...)})
	}
	out = append(out,
		Group{Title: "Unison", Path: at("Unison"), Controls: Controls(p.Unison, at("Unison"))},
		Group{Title: "Amp", Path: at("Amp"), Controls: Controls(p.Amp, at("Amp")), Curve: EnvelopeCurve},
		Group{Title: "Filter", Path: at("Filter"), Controls: Controls(p.Filter, at("Filter"), "Mode", "Cutoff", "Resonance", "KeyTrack", "EnvAmount", "Velocity"), Curve: ResponseCurve},
		Group{Title: "Filter env", Path: at("Filter", "Env"), Controls: Controls(p.Filter.Env, at("Filter", "Env")), Curve: EnvelopeCurve},
	)
	for i, l := range p.LFOs {
		out = append(out, Group{Title: fmt.Sprintf("LFO %d", i+1), Path: at("LFOs", i), Controls: Controls(l, at("LFOs", i))})
	}
	return append(out,
		Group{Title: "Sweep", Path: at("Sweep"), Controls: Controls(p.Sweep, at("Sweep"))},
		Group{Title: "Out", Path: path, Controls: Controls(p, path, "Drive", "Velocity", "Gain")},
	)
}

// EnvelopeShape is an envelope's level at n times from note-on through a note held a
// while and released, and the fraction of the way through that the key is released.
func EnvelopeShape(e synth.Envelope, n int) ([]float64, float64) {
	hold := max(e.Attack+e.Decay, 0.05) * 1.5
	total := hold + max(e.Release, 0.005)
	out := make([]float64, n)
	for i := range out {
		out[i] = e.Level(total*float64(i)/float64(max(n-1, 1)), hold)
	}
	return out, hold / total
}

// Response spans 20 Hz to 20 kHz and shows -36 to +18 dB.
const (
	responseLo, responseHi     = 20.0, 20000.0
	responseFloor, responseTop = -36.0, 18.0
)

// FilterResponse is the filter's gain at n frequencies from 20 Hz to 20 kHz spaced by
// octaves, from 0 at -36 dB to 1 at +18: the state-variable filter's analog response at
// its cutoff, without key tracking or its envelope. A cutoff of 0 is no filter.
func FilterResponse(f synth.Filter, n int) []float64 {
	out := make([]float64, n)
	k := 2 - 1.96*math.Max(0, math.Min(1, f.Resonance))
	for i := range out {
		hz := responseLo * math.Pow(responseHi/responseLo, float64(i)/float64(max(n-1, 1)))
		gain := 1.0
		if f.Cutoff > 0 {
			w := hz / math.Max(20, f.Cutoff)
			den := complex(1-w*w, k*w)
			var num complex128
			switch f.Mode {
			case synth.HighPass:
				num = complex(-w*w, 0)
			case synth.BandPass:
				num = complex(0, w)
			case synth.Notch:
				num = complex(1-w*w, 0)
			default:
				num = 1
			}
			gain = abs(num / den)
		}
		db := 20 * math.Log10(math.Max(gain, 1e-9))
		out[i] = math.Max(0, math.Min(1, (db-responseFloor)/(responseTop-responseFloor)))
	}
	return out
}

func abs(c complex128) float64 { return math.Hypot(real(c), imag(c)) }

// Resolution is the step a control writes a value at: its parameter's, or three
// significant figures.
func Resolution(p synth.Param, x float64) float64 {
	if p.Step > 0 {
		return p.Step
	}
	if x == 0 || math.IsInf(x, 0) || math.IsNaN(x) {
		return 1e-3
	}
	return math.Pow(10, math.Floor(math.Log10(math.Abs(x)))-2)
}

// RoundParam rounds a value to its control's resolution.
func RoundParam(p synth.Param, x float64) float64 { return Round(x, Resolution(p, x)) }

// Effective is what a value means: the parameter's Zero, when it's 0.
func Effective(p synth.Param, x float64) float64 {
	if x == 0 {
		return p.Zero
	}
	return x
}

// FormatParam writes a value as its control shows it.
func FormatParam(p synth.Param, x float64) string {
	g := func(x float64) string { return strconv.FormatFloat(x, 'g', 3, 64) }
	switch p.Unit {
	case synth.UnitDB:
		return fmt.Sprintf("%.1f dB", x)
	case synth.UnitHz:
		if math.Abs(x) >= 1000 {
			return g(x/1000) + " kHz"
		}
		return g(x) + " Hz"
	case synth.UnitSeconds:
		if math.Abs(x) < 1 {
			return g(x*1000) + " ms"
		}
		return g(x) + " s"
	case synth.UnitPan:
		return FormatPan(x)
	case synth.UnitCount:
		return strconv.Itoa(int(math.Round(x)))
	case synth.UnitAt:
		return "beat " + g(x)
	case synth.UnitPlain, synth.UnitFraction:
		return g(x)
	}
	return g(x) + " " + p.Unit.String()
}

// ParseParam reads a value typed into a control, in its unit or with one named: "3k",
// "250 ms", "1/3", "-6 dB", "L20". It's kept within the parameter's range and rounded
// to its resolution.
func ParseParam(p synth.Param, s string) (float64, bool) {
	if p.Unit == synth.UnitPan {
		return ParsePan(s)
	}
	s = strings.ToLower(strings.TrimSpace(s))
	scale := 1.0
	for _, u := range []struct {
		suffix string
		scale  float64
	}{
		{"khz", 1000}, {"hz", 1}, {"k", 1000}, {"ms", 0.001}, {"db", 1}, {"oct", 1}, {"st", 1}, {"ct", 1},
		{"beats", 1}, {"beat", 1}, {"s", 1}, {"%", 0.01},
	} {
		if rest, ok := strings.CutSuffix(s, u.suffix); ok {
			s, scale = strings.TrimSpace(rest), u.scale
			break
		}
	}
	s = strings.TrimPrefix(strings.TrimPrefix(s, "beat "), "+")
	var x float64
	if num, den, ok := strings.Cut(s, "/"); ok {
		n, err1 := strconv.ParseFloat(strings.TrimSpace(num), 64)
		m, err2 := strconv.ParseFloat(strings.TrimSpace(den), 64)
		if err1 != nil || err2 != nil || m == 0 {
			return 0, false
		}
		x = n / m
	} else {
		var err error
		if x, err = strconv.ParseFloat(s, 64); err != nil {
			return 0, false
		}
	}
	x *= scale
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return 0, false
	}
	if x != 0 || p.Zero == 0 {
		x = math.Max(p.Min, math.Min(p.Max, x))
	}
	return RoundParam(p, x), true
}

// Position is where a value sits along its control's travel, 0..1; hi stands in for a
// parameter's Max when that's infinite.
func Position(p synth.Param, x, hi float64) float64 {
	lo, top := p.Min, p.Max
	if math.IsInf(top, 1) {
		top = hi
	}
	if top <= lo {
		return 0
	}
	x = Effective(p, x)
	if p.Log && lo > 0 {
		if x <= lo {
			return 0
		}
		return math.Min(1, math.Log(x/lo)/math.Log(top/lo))
	}
	return math.Max(0, math.Min(1, (x-lo)/(top-lo)))
}

// FromPosition is the value at a place along a control's travel, rounded to its
// resolution.
func FromPosition(p synth.Param, pos, hi float64) float64 {
	lo, top := p.Min, p.Max
	if math.IsInf(top, 1) {
		top = hi
	}
	pos = math.Max(0, math.Min(1, pos))
	x := lo + pos*(top-lo)
	if p.Log && lo > 0 {
		x = lo * math.Pow(top/lo, pos)
	}
	return RoundParam(p, x)
}
