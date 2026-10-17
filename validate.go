package synth

import (
	"errors"
	"fmt"
)

func Validate(s Song) error {
	var errs []error
	if s.Loop && s.Length <= 0 {
		errs = append(errs, errors.New("a looping song needs a Length"))
	}
	if s.Loop && (s.LoopStart < 0 || s.LoopStart >= s.Length) {
		errs = append(errs, fmt.Errorf("LoopStart %g outside [0, %g)", s.LoopStart, s.Length))
	}

	if s.Loop && s.Vary != (Variation{}) {
		errs = append(errs, errors.New("Vary is for sound effects, not a looping song"))
	}
	if s.Vary.Pitch < 0 || s.Vary.Gain < 0 || s.Vary.Takes < 0 || s.Voices < 0 {
		errs = append(errs, errors.New("Vary and Voices can't be negative"))
	}

	buses := map[string]bool{}
	for _, b := range s.Buses {
		if buses[b.Name] {
			errs = append(errs, fmt.Errorf("bus %q declared twice", b.Name))
		}
		buses[b.Name] = true
	}

	for _, p := range s.Parts {
		for _, send := range p.Sends {
			if !buses[send.Bus] {
				errs = append(errs, fmt.Errorf("part %q: send to unknown bus %q", p.Name, send.Bus))
			}
		}
		if p.Out != "" && !buses[p.Out] {
			errs = append(errs, fmt.Errorf("part %q: out to unknown bus %q", p.Name, p.Out))
		}
		if err := ValidatePatch(p.Patch); err != nil {
			errs = append(errs, fmt.Errorf("part %q: %w", p.Name, err))
		}
		for i, n := range p.Notes {
			if n.Key < 0 || n.Key > 127 {
				errs = append(errs, fmt.Errorf("part %q note %d: key %g out of [0, 127]", p.Name, i, n.Key))
			}
			if n.Dur <= 0 {
				errs = append(errs, fmt.Errorf("part %q note %d: Dur %g must be > 0", p.Name, i, n.Dur))
			}
			if n.At < 0 {
				errs = append(errs, fmt.Errorf("part %q note %d: At %g is before the song starts", p.Name, i, n.At))
			}
			if s.Loop && n.At >= s.Length {
				errs = append(errs, fmt.Errorf("part %q note %d: At %g is past the loop's end", p.Name, i, n.At))
			}
		}
	}
	chains := map[string][]Effect{"master": s.Master}
	for _, p := range s.Parts {
		chains[fmt.Sprintf("part %q", p.Name)] = p.Inserts
	}
	for _, b := range s.Buses {
		chains[fmt.Sprintf("bus %q", b.Name)] = b.Effects
	}
	for name, fx := range chains {
		for i, e := range fx {
			if a, ok := e.(automated); ok && !pointsInOrder(a.points()) {
				errs = append(errs, fmt.Errorf("%s effect %d: Points out of order", name, i))
			}
		}
	}
	return errors.Join(errs...)
}

func ValidatePatch(p Patch) error {
	var errs []error
	if len(p.Oscs) == 0 {
		errs = append(errs, errors.New("no oscillators"))
	}
	if p.Unison.Voices > 16 {
		errs = append(errs, fmt.Errorf("%d unison voices; 16 is the most", p.Unison.Voices))
	}
	errs = append(errs, validateEnvelope("Amp", p.Amp))
	for i, o := range p.Oscs {
		if o.Level < 0 {
			errs = append(errs, fmt.Errorf("osc %d: negative Level", i))
		}
		if o.Width < 0 || o.Width >= 1 {
			errs = append(errs, fmt.Errorf("osc %d: Width %g outside [0, 1)", i, o.Width))
		}
		if o.FM.Index != 0 && o.Wave != Sine {
			errs = append(errs, fmt.Errorf("osc %d: FM needs a Sine, not %s", i, o.Wave))
		}
		if o.FM.Index != 0 && o.FM.Ratio <= 0 {
			errs = append(errs, fmt.Errorf("osc %d: FM Ratio must be > 0", i))
		}
		if o.Env != nil {
			errs = append(errs, validateEnvelope(fmt.Sprintf("osc %d Env", i), *o.Env))
		}
	}
	f := p.Filter
	if f.Cutoff < 0 || f.Cutoff > SampleRate/2 {
		errs = append(errs, fmt.Errorf("Filter Cutoff %g outside [0, %d]", f.Cutoff, SampleRate/2))
	}
	if f.Resonance < 0 || f.Resonance > 1 {
		errs = append(errs, fmt.Errorf("Filter Resonance %g outside [0, 1]", f.Resonance))
	}
	errs = append(errs, validateEnvelope("Filter Env", f.Env))
	for i, l := range p.LFOs {
		if l.Rate <= 0 {
			errs = append(errs, fmt.Errorf("LFO %d: Rate must be > 0", i))
		}
	}
	return errors.Join(errs...)
}

func validateEnvelope(name string, e Envelope) error {
	if e.Attack < 0 || e.Decay < 0 || e.Release < 0 {
		return fmt.Errorf("%s: negative time", name)
	}
	if e.Sustain < 0 || e.Sustain > 1 {
		return fmt.Errorf("%s: Sustain %g outside [0, 1]", name, e.Sustain)
	}
	return nil
}
