package session

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"slices"

	"github.com/cbodonnell/synth"
)

// Automation is a []Point field of an effect in an arrangement, shown as a lane: synth's
// AutoFilter, Fader and Panner, or a game's own effect, found by reflection. Its points
// are in beats.
type Automation struct {
	// The chain it's in: a track's inserts, a bus's effects, or with neither the master's.
	Track, Bus string
	Effect     int
	Field      string
	Name       string // the effect's type, and the field when it isn't Points
	Param      synth.Param
	Points     []synth.Point
}

// Key names the lane across edits.
func (l Automation) Key() string {
	return fmt.Sprintf("%q %q %d %s", l.Track, l.Bus, l.Effect, l.Field)
}

// Path is where the points are, for Doc.Custom.
func (l Automation) Path(a synth.Arrangement) []any {
	switch {
	case l.Track != "":
		return []any{"Tracks", slices.IndexFunc(a.Tracks, func(t synth.Track) bool { return t.Name == l.Track }), "Inserts", l.Effect, l.Field}
	case l.Bus != "":
		return []any{"Buses", slices.IndexFunc(a.Buses, func(b synth.Bus) bool { return b.Name == l.Bus }), "Effects", l.Effect, l.Field}
	}
	return []any{"Master", l.Effect, l.Field}
}

// Automations are the arrangement's lanes: each track's, then each bus's, then the
// master's.
func Automations(a synth.Arrangement) []Automation {
	var out []Automation
	add := func(track, bus string, fx []synth.Effect) {
		for i, e := range fx {
			for _, f := range pointFields(e) {
				pts := effectValue(e).FieldByName(f).Interface().([]synth.Point)
				name := effectValue(e).Type().Name()
				if f != "Points" {
					name += " " + f
				}
				out = append(out, Automation{
					Track: track, Bus: bus, Effect: i, Field: f, Name: name,
					Param: pointParam(e, f, pts), Points: pts,
				})
			}
		}
	}
	for _, t := range a.Tracks {
		add(t.Name, "", t.Inserts)
	}
	for _, b := range a.Buses {
		add("", b.Name, b.Effects)
	}
	add("", "", a.Master)
	return out
}

var pointsType = reflect.TypeFor[[]synth.Point]()

func effectValue(e synth.Effect) reflect.Value {
	v := reflect.ValueOf(e)
	for v.Kind() == reflect.Pointer && !v.IsNil() {
		v = v.Elem()
	}
	return v
}

// pointFields are the effect's []Point fields.
func pointFields(e synth.Effect) []string {
	v := effectValue(e)
	if v.Kind() != reflect.Struct {
		return nil
	}
	var out []string
	for i := range v.NumField() {
		if f := v.Type().Field(i); f.IsExported() && f.Type == pointsType {
			out = append(out, f.Name)
		}
	}
	return out
}

// pointParam is what the points' values are: synth's table for its own effects, or for
// a game's, a range around the values the points hold.
func pointParam(e synth.Effect, field string, pts []synth.Point) synth.Param {
	if ps, ok := synth.Params(effectValue(e).Type()); ok {
		for _, p := range ps {
			if p.Field == field {
				return p
			}
		}
	}
	p := synth.Param{Field: field}
	if len(pts) == 0 {
		p.Max = 1
		return p
	}
	lo, hi := math.Inf(1), math.Inf(-1)
	for _, pt := range pts {
		lo, hi = math.Min(lo, pt.Value), math.Max(hi, pt.Value)
	}
	pad := math.Max((hi-lo)/2, math.Max(math.Abs(hi), 1)/2)
	p.Min, p.Max = lo-pad, hi+pad
	if lo >= 0 {
		p.Min = math.Max(p.Min, 0)
	}
	return p
}

// chain is the effects the lane's effect is in.
func (l Automation) chain(a *synth.Arrangement) ([]synth.Effect, error) {
	var fx []synth.Effect
	switch {
	case l.Track != "":
		i := slices.IndexFunc(a.Tracks, func(t synth.Track) bool { return t.Name == l.Track })
		if i < 0 {
			return nil, fmt.Errorf("no track %q", l.Track)
		}
		fx = a.Tracks[i].Inserts
	case l.Bus != "":
		i := slices.IndexFunc(a.Buses, func(b synth.Bus) bool { return b.Name == l.Bus })
		if i < 0 {
			return nil, fmt.Errorf("no bus %q", l.Bus)
		}
		fx = a.Buses[i].Effects
	default:
		fx = a.Master
	}
	if l.Effect >= len(fx) || !slices.Contains(pointFields(fx[l.Effect]), l.Field) {
		return nil, errors.New("the automation is gone")
	}
	return fx, nil
}

func (l Automation) points(a *synth.Arrangement) ([]synth.Point, error) {
	fx, err := l.chain(a)
	if err != nil {
		return nil, err
	}
	return effectValue(fx[l.Effect]).FieldByName(l.Field).Interface().([]synth.Point), nil
}

func (l Automation) set(a *synth.Arrangement, pts []synth.Point) {
	fx, _ := l.chain(a)
	l.setIn(fx, pts)
}

func (l Automation) setIn(fx []synth.Effect, pts []synth.Point) {
	e := reflect.ValueOf(fx[l.Effect])
	if e.Kind() == reflect.Pointer {
		e.Elem().FieldByName(l.Field).Set(reflect.ValueOf(pts))
		return
	}
	v := reflect.New(e.Type()).Elem()
	v.Set(e)
	v.FieldByName(l.Field).Set(reflect.ValueOf(pts))
	fx[l.Effect] = v.Interface().(synth.Effect)
}

// insert puts p among pts in At order, after any at the same At, and returns where.
func insert(pts []synth.Point, p synth.Point) ([]synth.Point, int) {
	i, _ := slices.BinarySearchFunc(pts, p.At, func(q synth.Point, at float64) int {
		if q.At <= at {
			return -1
		}
		return 1
	})
	return slices.Insert(pts, i, p), i
}

// AddPoint adds a point to the lane, in order, and returns its index.
func AddPoint(a *synth.Arrangement, l Automation, p synth.Point) (int, error) {
	pts, err := l.points(a)
	if err != nil {
		return 0, err
	}
	pts, i := insert(slices.Clone(pts), p)
	l.set(a, pts)
	return i, nil
}

// MovePoint moves the lane's point i to p, past any neighbors it crosses, and returns
// its index there.
func MovePoint(a *synth.Arrangement, l Automation, i int, p synth.Point) (int, error) {
	pts, err := l.points(a)
	if err != nil {
		return 0, err
	}
	if i < 0 || i >= len(pts) {
		return 0, fmt.Errorf("no point %d", i)
	}
	pts, i = insert(slices.Delete(slices.Clone(pts), i, i+1), p)
	l.set(a, pts)
	return i, nil
}

// DeletePoint takes the lane's point i out.
func DeletePoint(a *synth.Arrangement, l Automation, i int) error {
	pts, err := l.points(a)
	if err != nil {
		return err
	}
	if i < 0 || i >= len(pts) {
		return fmt.Errorf("no point %d", i)
	}
	l.set(a, slices.Delete(slices.Clone(pts), i, i+1))
	return nil
}
