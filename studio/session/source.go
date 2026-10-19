package session

import (
	"fmt"
	"go/format"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/patches"
)

// Source writes an arrangement as a Go file in package pkg: a var named after its ID.
// A patch or chain equal to one of synth's patches is written as that call.
func Source(a synth.Arrangement, pkg string) ([]byte, error) {
	body := fmt.Sprintf("var %s = %s\n\nfunc %s() synth.Song { return %s.Song() }\n",
		Ident(a.ID, false), literal(reflect.ValueOf(a), true), Ident(a.ID, true), Ident(a.ID, false))
	imports := `"github.com/cbodonnell/synth"`
	if strings.Contains(body, "patches.") {
		imports += "\n" + `"github.com/cbodonnell/synth/patches"`
	}
	src := fmt.Sprintf("package %s\n\nimport (\n%s\n)\n\n%s", pkg, imports, body)
	out, err := format.Source([]byte(src))
	if err != nil {
		return []byte(src), err
	}
	return out, nil
}

// Ident is a Go name for an ID, exported or not.
func Ident(id string, exported bool) string {
	var b strings.Builder
	upper := exported
	for _, r := range id {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			upper = b.Len() > 0 || exported
			continue
		}
		if upper {
			r = unicode.ToUpper(r)
		}
		b.WriteRune(r)
		upper = false
	}
	return b.String()
}

var patchCalls = map[string]func() synth.Patch{
	"RhythmGuitar": patches.RhythmGuitar, "LeadGuitar": patches.LeadGuitar, "Organ": patches.Organ,
	"HofnerBass": patches.HofnerBass, "SynthArp": patches.SynthArp, "Tom": patches.Tom,
	"Sitar": patches.Sitar, "Tanpura": patches.Tanpura, "Handpan": patches.Handpan,
	"Didgeridoo": patches.Didgeridoo, "HandDrum": patches.HandDrum, "Slap": patches.Slap,
	"Shaker": patches.Shaker, "Wind": patches.Wind, "Banjo": patches.Banjo,
	"Mandolin": patches.Mandolin, "AcousticGuitar": patches.AcousticGuitar,
	"UprightBass": patches.UprightBass, "Whistle": patches.Whistle,
	"Glockenspiel": patches.Glockenspiel, "Tambourine": patches.Tambourine,
	"Firefly": patches.Firefly, "Cricket": patches.Cricket, "GangVocals": patches.GangVocals,
	"Shout": patches.Shout, "SuperSaw": patches.SuperSaw, "SawPad": patches.SawPad,
	"HousePluck": patches.HousePluck, "HousePiano": patches.HousePiano,
	"OffbeatBass": patches.OffbeatBass, "HouseKick": patches.HouseKick, "Clap": patches.Clap,
	"OpenHat": patches.OpenHat, "Harpsichord": patches.Harpsichord,
	"OctaveSynth": patches.OctaveSynth, "Recorder": patches.Recorder,
	"PipeOrgan": patches.PipeOrgan, "Tick": patches.Tick, "Bass": patches.Bass,
	"GrandPiano": patches.GrandPiano, "EPiano": patches.EPiano, "Bell": patches.Bell,
	"Lead": patches.Lead, "Kick": patches.Kick, "Snare": patches.Snare, "HiHat": patches.HiHat,
	"Laser": patches.Laser, "Coin": patches.Coin, "EightOhEight": patches.EightOhEight,
	"TrapKick": patches.TrapKick, "TrapSnare": patches.TrapSnare, "Brass": patches.Brass,
	"AirHorn": patches.AirHorn,
}

// patchCall is the call that makes p, or "".
func patchCall(p synth.Patch) string {
	for name, f := range patchCalls {
		if reflect.DeepEqual(p, f()) {
			return "patches." + name + "()"
		}
	}
	if reflect.DeepEqual(p, patches.Cymbal(p.Amp.Decay)) {
		return "patches.Cymbal(" + number(p.Amp.Decay) + ")"
	}
	if reflect.DeepEqual(p, patches.Riser(p.Amp.Attack)) {
		return "patches.Riser(" + number(p.Amp.Attack) + ")"
	}
	return ""
}

// chainCall writes a chain that begins with patches.GuitarAmp as an append to it, or "".
func chainCall(fx []synth.Effect) string {
	if len(fx) < 2 {
		return ""
	}
	d, ok := fx[0].(synth.Drive)
	amp := patches.GuitarAmp(d.Amount)
	if !ok || len(fx) < len(amp) || !reflect.DeepEqual(fx[:len(amp)], amp) {
		return ""
	}
	call := "patches.GuitarAmp(" + number(d.Amount) + ")"
	if len(fx) == len(amp) {
		return call
	}
	var rest []string
	for _, e := range fx[len(amp):] {
		rest = append(rest, literal(reflect.ValueOf(e), true))
	}
	return "append(" + call + ", " + strings.Join(rest, ", ") + ")"
}

var enums = map[reflect.Type][]string{
	reflect.TypeFor[synth.Wave]():       {"Sine", "Triangle", "Saw", "Pulse", "Noise", "Pluck"},
	reflect.TypeFor[synth.FilterMode](): {"LowPass", "HighPass", "BandPass", "Notch"},
	reflect.TypeFor[synth.BandKind]():   {"Peak", "LowCut", "HighCut", "LowShelf", "HighShelf"},
	reflect.TypeFor[synth.LFOTarget]():  {"Vibrato", "Wah", "Tremolo", "PWM", "AutoPan"},
	reflect.TypeFor[synth.Shape]():      {"SineShape", "TriangleShape", "SquareShape", "SawShape", "RandomShape"},
}

// Past this a literal breaks onto lines of its own.
const lineWidth = 96

// literal writes v as Go; typed says whether a composite needs its type written, which
// it doesn't as an element of a slice of its own type.
func literal(v reflect.Value, typed bool) string {
	switch x := v.Interface().(type) {
	case synth.Patch:
		if call := patchCall(x); call != "" {
			return call
		}
	case []synth.Effect:
		if call := chainCall(x); call != "" {
			return call
		}
	}
	if names, ok := enums[v.Type()]; ok && int(v.Int()) < len(names) {
		return "synth." + names[v.Int()]
	}

	switch v.Kind() {
	case reflect.Interface:
		return literal(v.Elem(), true)
	case reflect.Pointer:
		return "&" + literal(v.Elem(), true)
	case reflect.Float32, reflect.Float64:
		return number(v.Float())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.String:
		return strconv.Quote(v.String())
	case reflect.Slice:
		elem := v.Type().Elem()
		var items []string
		for i := range v.Len() {
			items = append(items, literal(v.Index(i), elem.Kind() == reflect.Interface))
		}
		lines := elem.Kind() == reflect.String && v.Len() > 1
		return composite(typeName(v.Type()), items, lines)
	case reflect.Struct:
		var fields []string
		for i := range v.NumField() {
			f := v.Type().Field(i)
			if !f.IsExported() || v.Field(i).IsZero() {
				continue
			}
			fields = append(fields, f.Name+": "+literal(v.Field(i), true))
		}
		name := ""
		if typed {
			name = typeName(v.Type())
		}
		return composite(name, fields, false)
	}
	panic("studio: can't write a " + v.Type().String())
}

func composite(name string, items []string, lines bool) string {
	one := name + "{" + strings.Join(items, ", ") + "}"
	if !lines && len(one) <= lineWidth && !slices.ContainsFunc(items, func(s string) bool { return strings.Contains(s, "\n") }) {
		return one
	}
	return name + "{\n" + strings.Join(items, ",\n") + ",\n}"
}

func typeName(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Slice:
		return "[]" + typeName(t.Elem())
	case reflect.Pointer:
		return "*" + typeName(t.Elem())
	}
	if t.PkgPath() == "" {
		return t.Name()
	}
	return t.String()
}

// number writes a float at ten significant digits, so arithmetic's last-bit noise
// doesn't reach the source.
func number(x float64) string {
	return strconv.FormatFloat(x, 'g', 10, 64)
}
