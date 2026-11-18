package synth

import (
	"math"
	"reflect"
)

// Unit is what a parameter measures.
type Unit int

const (
	UnitPlain Unit = iota
	UnitHz
	UnitDB
	UnitSeconds
	UnitBeats // a duration that an Arrangement writes in beats
	UnitAt    // a song position that an Arrangement writes in beats
	UnitFraction
	UnitOctaves
	UnitSemitones
	UnitCents
	UnitPan
	UnitCount
)

func (u Unit) String() string {
	return [...]string{"", "Hz", "dB", "s", "beats", "beat", "", "oct", "st", "ct", "", ""}[u]
}

// Param describes a number field of an effect or a patch, for a control drawn from it.
// A Param for a []Point field (an automation's Points) describes the points' values.
type Param struct {
	Field    string
	Unit     Unit
	Min, Max float64
	Log      bool    // the control moves exponentially between Min and Max
	Zero     float64 // what 0 means, where the field documents 0 as something else
	Step     float64 // the control's resolution; 0 is three significant figures
}

// Params describes the number fields of synth's effects and of a Patch and its parts,
// in declared order. It reports false for a type it has no table for (a game's own
// effect), which a control falls back to reflection for.
func Params(t reflect.Type) ([]Param, bool) {
	p, ok := params[t]
	return p, ok
}

// LFODepth describes an LFO's Depth, whose unit follows its Target.
func LFODepth(t LFOTarget) Param {
	switch t {
	case Vibrato:
		return Param{Field: "Depth", Unit: UnitCents, Max: 100}
	case Wah:
		return Param{Field: "Depth", Unit: UnitOctaves, Max: 4}
	case AutoPan:
		return Param{Field: "Depth", Unit: UnitPan, Max: 1, Step: 0.01}
	}
	return Param{Field: "Depth", Unit: UnitFraction, Max: 1}
}

func db(field string, lo, hi float64) Param {
	return Param{Field: field, Unit: UnitDB, Min: lo, Max: hi, Step: 0.1}
}

func hz(field string, lo, hi float64) Param {
	return Param{Field: field, Unit: UnitHz, Min: lo, Max: hi, Log: true}
}

func secs(field string, hi float64) Param { return Param{Field: field, Unit: UnitSeconds, Max: hi} }

func frac(field string) Param { return Param{Field: field, Unit: UnitFraction, Max: 1} }

func mix(zero float64) Param { return Param{Field: "Mix", Unit: UnitFraction, Max: 1, Zero: zero} }

func with(p Param, zero float64) Param {
	p.Zero = zero
	return p
}

var params = map[reflect.Type][]Param{
	reflect.TypeFor[Band](): {
		hz("Freq", 20, 20000), db("Gain", -24, 24),
		{Field: "Q", Unit: UnitPlain, Min: 0.1, Max: 10, Log: true, Zero: 0.707},
	},
	reflect.TypeFor[Drive](): {
		{Field: "Amount", Max: 10}, hz("Tone", 200, 20000), mix(1),
	},
	reflect.TypeFor[Chorus](): {
		with(hz("Rate", 0.05, 10), 0.8),
		{Field: "Depth", Unit: UnitSeconds, Max: 0.01, Zero: 0.003},
		{Field: "Delay", Unit: UnitSeconds, Min: 0.002, Max: 0.04, Zero: 0.012},
		{Field: "Feedback", Unit: UnitFraction, Max: 0.9}, mix(1),
	},
	reflect.TypeFor[Phaser](): {
		hz("Rate", 0.02, 10), {Field: "Depth", Unit: UnitOctaves, Max: 4},
		with(hz("Center", 100, 5000), 700),
		{Field: "Stages", Unit: UnitCount, Min: 2, Max: 12, Zero: 6, Step: 1},
		{Field: "Feedback", Unit: UnitFraction, Max: 0.9}, mix(0.5),
	},
	reflect.TypeFor[Delay](): {
		{Field: "Time", Unit: UnitBeats, Min: 1.0 / 16, Max: 4, Log: true},
		{Field: "Feedback", Unit: UnitFraction, Max: 0.95}, hz("Damp", 500, 20000), mix(1),
	},
	reflect.TypeFor[Reverb](): {
		frac("Size"), frac("Damp"), with(frac("Width"), 1), secs("PreDelay", 0.2), mix(1),
	},
	reflect.TypeFor[Compressor](): {
		db("Threshold", -60, 0),
		{Field: "Ratio", Min: 1, Max: 20, Log: true, Zero: 4},
		{Field: "Attack", Unit: UnitSeconds, Min: 0.0005, Max: 0.1, Log: true, Zero: 0.005},
		{Field: "Release", Unit: UnitSeconds, Min: 0.01, Max: 1, Log: true, Zero: 0.1},
		db("Makeup", 0, 24),
	},
	reflect.TypeFor[Crush](): {
		{Field: "Bits", Unit: UnitCount, Min: 1, Max: 16, Step: 1},
		{Field: "Hold", Unit: UnitCount, Min: 1, Max: 32, Step: 1},
	},
	reflect.TypeFor[Limiter](): {
		with(db("Ceiling", -12, 0), -1),
		{Field: "Release", Unit: UnitSeconds, Min: 0.01, Max: 1, Log: true, Zero: 0.08},
	},
	reflect.TypeFor[Sidechain](): {
		frac("Depth"),
		{Field: "Attack", Unit: UnitSeconds, Min: 0.0005, Max: 0.05, Log: true, Zero: 0.004},
		{Field: "Release", Unit: UnitSeconds, Min: 0.01, Max: 1, Log: true, Zero: 0.2},
	},
	reflect.TypeFor[TapeStop](): {
		{Field: "At", Unit: UnitAt, Max: math.Inf(1)},
		{Field: "Length", Unit: UnitBeats, Max: 8},
		{Field: "Until", Unit: UnitAt, Max: math.Inf(1)},
	},
	reflect.TypeFor[AutoFilter](): {frac("Resonance"), hz("Points", 20, 20000)},
	reflect.TypeFor[Fader]():      {db("Points", -60, 12)},
	reflect.TypeFor[Panner]():     {{Field: "Points", Unit: UnitPan, Min: -1, Max: 1, Step: 0.01}},
	reflect.TypeFor[Normalize]():  {db("Peak", -30, 0)},

	reflect.TypeFor[Patch](): {
		{Field: "Drive", Max: 10}, frac("Velocity"),
		{Field: "Gain", Max: 2, Zero: 1},
	},
	reflect.TypeFor[Osc](): {
		with(frac("Level"), 1),
		{Field: "Semi", Unit: UnitSemitones, Min: -24, Max: 24, Step: 1},
		{Field: "Cents", Unit: UnitCents, Min: -100, Max: 100, Step: 1},
		hz("Fixed", 20, 10000),
		{Field: "Width", Unit: UnitFraction, Min: 0.05, Max: 0.95, Zero: 0.5},
		frac("Phase"),
		{Field: "Pan", Unit: UnitPan, Min: -1, Max: 1, Step: 0.01},
	},
	reflect.TypeFor[FM](): {
		{Field: "Ratio", Min: 0.25, Max: 16, Log: true},
		{Field: "Index", Max: 20}, secs("Decay", 5), frac("Sustain"),
	},
	reflect.TypeFor[String](): {
		{Field: "Decay", Unit: UnitSeconds, Min: 0.1, Max: 20, Log: true, Zero: 4},
		frac("Brightness"), {Field: "Pick", Unit: UnitFraction, Max: 0.5},
	},
	reflect.TypeFor[Envelope](): {
		secs("Attack", 5), secs("Decay", 5), frac("Sustain"), secs("Release", 10),
	},
	reflect.TypeFor[Unison](): {
		{Field: "Voices", Unit: UnitCount, Min: 1, Max: 8, Step: 1},
		{Field: "Detune", Unit: UnitCents, Max: 100}, frac("Spread"),
	},
	reflect.TypeFor[Filter](): {
		hz("Cutoff", 20, 20000),
		frac("Resonance"), frac("KeyTrack"),
		{Field: "EnvAmount", Unit: UnitOctaves, Min: -6, Max: 6},
		{Field: "Velocity", Unit: UnitOctaves, Max: 4},
	},
	reflect.TypeFor[LFO](): {
		hz("Rate", 0.05, 20), {Field: "Depth", Max: 1}, secs("Delay", 5),
	},
	reflect.TypeFor[Sweep](): {
		{Field: "Semis", Unit: UnitSemitones, Min: -48, Max: 48},
		secs("Time", 1),
	},
}
