package synth

import (
	"math"
	"reflect"
	"testing"
)

func TestParamsNameNumberFields(t *testing.T) {
	for typ, ps := range params {
		for _, p := range ps {
			f, ok := typ.FieldByName(p.Field)
			if !ok {
				t.Errorf("%s has no field %s", typ, p.Field)
				continue
			}
			switch f.Type.Kind() {
			case reflect.Float64, reflect.Int:
			default:
				if f.Type != reflect.TypeFor[[]Point]() {
					t.Errorf("%s.%s is not a number", typ, p.Field)
				}
			}
			if p.Min > p.Max || (p.Log && p.Min <= 0) {
				t.Errorf("%s.%s has range %g..%g", typ, p.Field, p.Min, p.Max)
			}
		}
	}
}

func TestNormalizeLandsOnPeak(t *testing.T) {
	s := Stereo{L: []float32{0.1, -0.4, 0.2}, R: []float32{0, 0.3, -0.1}}
	Normalize{Peak: -6}.Process(s)
	if got := 20 * math.Log10(float64(-s.L[1])); math.Abs(got+6) > 1e-4 {
		t.Errorf("peak %.3f dB, want -6", got)
	}
	silent := Stereo{L: make([]float32, 4), R: make([]float32, 4)}
	Normalize{}.Process(silent)
	if silent.L[0] != 0 {
		t.Error("silence changed")
	}
}
