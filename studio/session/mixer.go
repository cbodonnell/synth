package session

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/cbodonnell/synth"
)

// Resolutions the mixer writes numbers at, so diffs stay readable.
const (
	DBStep  = 0.1
	PanStep = 0.01
)

// Round rounds x to a multiple of step, landing on the float the decimal reads as.
func Round(x, step float64) float64 {
	if inv := math.Round(1 / step); math.Abs(inv*step-1) < 1e-12 {
		return math.Round(x*inv) / inv
	}
	return math.Round(x/step) * step
}

// The fader's scale: dB at points along its travel, 0 at the bottom and 1 at the top.
var faderMarks = []struct{ pos, db float64 }{
	{0, -60}, {0.14, -40}, {0.24, -30}, {0.36, -20}, {0.5, -12}, {0.64, -6}, {0.78, 0}, {0.9, 6}, {1, 12},
}

// FaderMarks are the dB the fader's scale is marked at.
func FaderMarks() []float64 {
	var out []float64
	for _, m := range faderMarks {
		out = append(out, m.db)
	}
	return out
}

// FaderPosition is where on the fader's travel a level in dB sits, 0..1.
func FaderPosition(db float64) float64 {
	if db <= faderMarks[0].db {
		return 0
	}
	for i := 1; i < len(faderMarks); i++ {
		a, b := faderMarks[i-1], faderMarks[i]
		if db <= b.db {
			return a.pos + (db-a.db)/(b.db-a.db)*(b.pos-a.pos)
		}
	}
	return 1
}

// FaderDB is the level in dB at a position on the fader's travel.
func FaderDB(pos float64) float64 {
	pos = math.Max(0, math.Min(1, pos))
	for i := 1; i < len(faderMarks); i++ {
		a, b := faderMarks[i-1], faderMarks[i]
		if pos <= b.pos {
			return a.db + (pos-a.pos)/(b.pos-a.pos)*(b.db-a.db)
		}
	}
	return faderMarks[len(faderMarks)-1].db
}

// ParseDB reads a level typed into the mixer: "-3.5", "-3.5 dB", "+2db".
func ParseDB(s string) (float64, bool) {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.TrimSpace(strings.TrimSuffix(s, "db"))
	x, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(x) || math.IsInf(x, 0) {
		return 0, false
	}
	return Round(x, DBStep), true
}

// FormatPan writes a pan as the mixer shows it: C, L20, R35.
func FormatPan(p float64) string {
	n := int(math.Round(math.Abs(p) * 100))
	switch {
	case n == 0:
		return "C"
	case p < 0:
		return fmt.Sprintf("L%d", n)
	}
	return fmt.Sprintf("R%d", n)
}

// ParsePan reads a pan typed as FormatPan writes it, or as a number from -1 to 1.
func ParsePan(s string) (float64, bool) {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "C" {
		return 0, true
	}
	sign := 0.0
	switch {
	case strings.HasPrefix(s, "L"):
		sign = -1
	case strings.HasPrefix(s, "R"):
		sign = 1
	}
	if sign != 0 {
		n, err := strconv.ParseFloat(strings.TrimSpace(s[1:]), 64)
		if err != nil || n < 0 || n > 100 {
			return 0, false
		}
		return Round(sign*n/100, PanStep), true
	}
	x, err := strconv.ParseFloat(s, 64)
	if err != nil || x < -1 || x > 1 {
		return 0, false
	}
	return Round(x, PanStep), true
}

// AddReturn sends the track to a new bus holding a reverb, and returns the bus's name.
func AddReturn(a *synth.Arrangement, track int) string {
	name := "reverb"
	for n := 2; slices.ContainsFunc(a.Buses, func(b synth.Bus) bool { return b.Name == name }); n++ {
		name = fmt.Sprintf("reverb %d", n)
	}
	a.Buses = append(a.Buses, synth.Bus{Name: name, Effects: []synth.Effect{synth.Reverb{Size: 0.5, Damp: 0.5}}})
	a.Tracks[track].Sends = append(a.Tracks[track].Sends, synth.Send{Bus: name, Level: -12})
	return name
}
