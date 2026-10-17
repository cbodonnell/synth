package synth

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tone(w Wave) Patch {
	return Patch{
		Oscs: []Osc{{Wave: w}},
		Amp:  Envelope{Attack: 0.005, Decay: 0.1, Sustain: 0.6, Release: 0.05},
	}
}

func oneNote(p Patch, key, at, dur float64) Song {
	return Song{ID: "t", Parts: []Part{{Name: "p", Patch: p, Notes: []Note{{At: at, Dur: dur, Key: key}}}}}
}

// Everything that draws on noise, plus effects with state, in one song.
func busySong() Song {
	p := tone(Saw)
	p.Oscs = append(p.Oscs, Osc{Wave: Noise, Level: 0.3}, Osc{Wave: Pluck, Semi: 12})
	p.Unison = Unison{Voices: 3, Detune: 20, Spread: 1}
	p.LFOs = []LFO{{Target: Wah, Shape: RandomShape, Rate: 8, Depth: 1}}
	p.Filter = Filter{Cutoff: 1200, Resonance: 0.5}
	return Song{
		ID: "busy",
		Parts: []Part{{
			Name: "p", Patch: p,
			Notes:   Phrase{Step: 0.1, Humanize: 0.01}.Notes("C4 E4,G4 - . C5"),
			Inserts: []Effect{Chorus{}, Drive{Amount: 2}},
			Sends:   []Send{{Bus: "verb", Level: 0.5}},
		}},
		Buses:  []Bus{{Name: "verb", Effects: []Effect{Reverb{Size: 0.4}, Delay{Time: 0.05, Feedback: 0.4}}}},
		Master: []Effect{Compressor{Threshold: -12}},
	}
}

func TestRenderIsDeterministic(t *testing.T) {
	require.Equal(t, Render(busySong()), Render(busySong()))
}

func TestRenderEmptySongIsEmpty(t *testing.T) {
	assert.Zero(t, Render(Song{}).Len())
}

func TestRenderStaysInRange(t *testing.T) {
	s := busySong()
	s.Parts[0].Gain = 50
	out := Render(s)
	require.NotZero(t, out.Len())
	assert.LessOrEqual(t, out.Peak(), float32(1))
}

func TestNoteIsSilentBeforeItStarts(t *testing.T) {
	s := oneNote(tone(Saw), 60, 0.1, 0.1)
	s.Length = 0.3
	out := Render(s)
	before := samples(0.1)
	for i := range before {
		require.Zerof(t, out.L[i], "sample %d", i)
	}
	assert.NotZero(t, Stereo{L: out.L[before:], R: out.R[before:]}.Peak())
}

func TestEveryWaveSounds(t *testing.T) {
	for _, w := range []Wave{Sine, Triangle, Saw, Pulse, Noise, Pluck} {
		out := Render(oneNote(tone(w), 60, 0, 0.2))
		assert.Greaterf(t, out.Peak(), float32(0.05), "%s", w)
	}
}

func TestEnvelopeEndsAtExactlyZero(t *testing.T) {
	for _, e := range []Envelope{
		{},
		{Attack: 0.01, Decay: 0.1, Sustain: 0.5, Release: 0.2},
		{Decay: 0.05},
	} {
		dur := 0.3
		end := e.length(dur)
		assert.Zero(t, e.level(0, dur))
		assert.Zero(t, e.level(end, dur))
		assert.Greater(t, e.level(e.attack(), dur), 0.0)
	}
}

func TestEnvelopeReachesItsTargets(t *testing.T) {
	e := Envelope{Attack: 0.01, Decay: 0.1, Sustain: 0.5, Release: 0.2}
	assert.InDelta(t, 1, e.level(0.01, 1), 1e-9)
	assert.InDelta(t, 0.5, e.level(0.11, 1), 1e-9)
	assert.InDelta(t, 0.5, e.level(0.9, 1), 1e-9)
}

// Autocorrelation peak near the expected period, refined between samples.
func measureHz(t *testing.T, x []float32, want float64) float64 {
	t.Helper()
	period := SampleRate / want
	lo, hi := int(period*0.8), int(period*1.25)
	const n = 8192
	corr := func(lag int) float64 {
		var s float64
		for i := range n {
			s += float64(x[i]) * float64(x[i+lag])
		}
		return s
	}
	best, bl := math.Inf(-1), lo
	for lag := lo; lag <= hi; lag++ {
		if c := corr(lag); c > best {
			best, bl = c, lag
		}
	}
	a, b, c := corr(bl-1), best, corr(bl+1)
	return SampleRate / (float64(bl) + 0.5*(a-c)/(a-2*b+c))
}

func TestPluckIsInTune(t *testing.T) {
	for _, name := range []string{"Bb1", "F3", "Db5", "A6"} {
		key := MustKey(name)
		p := Patch{Oscs: []Osc{{Wave: Pluck, String: String{Brightness: 0.8, Pick: 0.12}}}}
		out := Render(oneNote(p, key, 0, 1))
		got := measureHz(t, out.L[2400:], Hz(key))
		assert.InDeltaf(t, 0, 1200*math.Log2(got/Hz(key)), 3, "%s: %.2fHz", name, got)
	}
}

func loopSong() Song {
	pad := tone(Triangle)
	pad.Amp.Release = 0.4
	return Song{
		ID: "loop",
		Parts: []Part{{Name: "p", Patch: pad, Notes: Phrase{Step: 0.25}.Notes(
			"C4 . E4 . | G4",
		)}},
		Length:    1.25,
		Loop:      true,
		LoopStart: 0.5,
	}
}

func TestRenderLoopLengths(t *testing.T) {
	require.NoError(t, Validate(loopSong()))
	l := RenderLoop(loopSong())
	assert.Equal(t, samples(1.25), l.Intro.Len())
	assert.Equal(t, samples(0.75), l.Body.Len())
}

func TestRenderLoopWrapsTheTailOntoTheLoopStart(t *testing.T) {
	l := RenderLoop(loopSong())
	assert.Greater(t, l.Body.Peak(), float32(0))
	// The last note rings past Length; it must be heard at the top of the body.
	assert.NotZero(t, l.Body.L[0])
}

func TestRenderLoopSeamIsContinuous(t *testing.T) {
	l := RenderLoop(loopSong())
	b := l.Body
	var biggest float32
	for i := 1; i < b.Len(); i++ {
		biggest = max(biggest, abs32(b.L[i]-b.L[i-1]))
	}
	assert.LessOrEqual(t, abs32(b.L[0]-b.L[b.Len()-1]), biggest, "body end to body start")
	assert.LessOrEqual(t, abs32(b.L[0]-l.Intro.L[l.Intro.Len()-1]), biggest, "intro end to body start")
}

// The intro's first pass through the loop and the steady loop must agree wherever no tail
// is wrapping, or the second time round would sound like a different take.
func TestLoopBodyMatchesTheFirstPass(t *testing.T) {
	s := busySong()
	s.Parts[0].Notes = Repeat(s.Parts[0].Notes, 3, 1)
	s.Length, s.Loop, s.LoopStart = 3, true, 1
	l := RenderLoop(s)
	off := l.Intro.Len() - l.Body.Len()
	mid := l.Body.Len() / 2
	assert.InDelta(t, l.Intro.L[off+mid], l.Body.L[mid], 0.02)
}

func TestReverbDiesWithinItsTail(t *testing.T) {
	r := Reverb{Size: 0.8, Damp: 0.3}
	buf := NewStereo(samples(r.Tail() + 0.5))
	buf.L[0], buf.R[0] = 1, 1
	r.Process(buf)
	start := samples(r.Tail())
	assert.Less(t, Stereo{L: buf.L[start:], R: buf.R[start:]}.Peak(), float32(0.001))
}

func TestEffectsKeepSilenceSilent(t *testing.T) {
	for _, e := range []Effect{
		EQ{Bands: []Band{{Kind: LowShelf, Freq: 100, Gain: 6}, {Kind: Peak, Freq: 1000, Gain: -3}}},
		Drive{Amount: 3, Tone: 4000}, Chorus{Feedback: 0.5}, Delay{Time: 0.1, Feedback: 0.5, PingPong: true},
		Reverb{Size: 1}, Compressor{Threshold: -20, Makeup: 6}, Crush{Bits: 4, Hold: 3},
	} {
		buf := NewStereo(2000)
		e.Process(buf)
		assert.Zerof(t, buf.Peak(), "%T", e)
	}
}

func TestParseKey(t *testing.T) {
	for name, want := range map[string]float64{"C4": 60, "A4": 69, "Bb2": 46, "C#-1": 1, "Cb4": 59, "G9": 127} {
		got, err := ParseKey(name)
		require.NoError(t, err, name)
		assert.Equal(t, want, got, name)
	}
	for _, bad := range []string{"", "H4", "C", "Cx4", "4C"} {
		_, err := ParseKey(bad)
		assert.Error(t, err, bad)
	}
}

func TestPhraseHoldsRestsChordsAndBars(t *testing.T) {
	notes := Phrase{Step: 0.5}.Notes("C4 - | . E4,G4 | C5")
	require.Len(t, notes, 4)
	assert.Equal(t, 1.0, notes[0].Dur, "a hold extends the note")
	assert.Equal(t, notes[1].At, notes[2].At, "a chord's notes start together")
	assert.Equal(t, 1.5, notes[1].At, "a rest takes a step; a bar line takes none")
	assert.Equal(t, 2.0, notes[3].At)
}

func TestPhraseAccentsAndGhosts(t *testing.T) {
	notes := Phrase{Step: 0.5}.Notes("C4 C4! C4?")
	assert.Greater(t, notes[1].Vel, notes[0].Vel)
	assert.Less(t, notes[2].Vel, notes[0].Vel)
}

func TestMeterIrregularBars(t *testing.T) {
	m := Meter{Beat: 1, Bars: []int{3, 3, 3, 4}}
	assert.Equal(t, Position{Bar: 0, Beat: 2, Phase: 0.5}, m.At(2.5))
	assert.Equal(t, Position{Bar: 3, Beat: 0}, m.At(9))
	assert.Equal(t, Position{Bar: 3, Beat: 3}, m.At(12))
	assert.Equal(t, Position{Bar: 4, Beat: 0}, m.At(13))
}

func TestSongTimeUnwindsTheLoop(t *testing.T) {
	s := Song{Length: 10, Loop: true, LoopStart: 4}
	assert.Equal(t, 3.0, s.SongTime(3))
	assert.Equal(t, 9.0, s.SongTime(9))
	assert.Equal(t, 5.0, s.SongTime(11))
	assert.Equal(t, 11.0, Song{Length: 10}.SongTime(11), "a song that doesn't loop just runs on")
}

func TestValidate(t *testing.T) {
	require.NoError(t, Validate(busySong()))

	s := busySong()
	s.Parts[0].Sends = []Send{{Bus: "nowhere"}}
	assert.Error(t, Validate(s))

	s = busySong()
	s.Parts[0].Patch.Amp.Sustain = 2
	assert.Error(t, Validate(s))

	s = busySong()
	s.Parts[0].Patch.Oscs[0].FM = FM{Ratio: 2, Index: 1}
	assert.Error(t, Validate(s), "FM on a saw")

	s = busySong()
	s.Master = append(s.Master, Fader{Points: []Point{{At: 2, Value: 0}, {At: 1, Value: -6}}})
	assert.Error(t, Validate(s), "automation out of order")

	assert.Error(t, Validate(Song{Loop: true}))
}

func TestFollowPassesThroughEveryPoint(t *testing.T) {
	p := []Point{{At: 0.5, Value: 200}, {At: 1, Value: 4000}, {At: 1, Value: 100}, {At: 3, Value: 1600}}
	assert.Equal(t, 200.0, follow(p, 0, exponential), "holds before the first point")
	assert.Equal(t, 1600.0, follow(p, 9, exponential), "holds after the last point")
	for _, pt := range []Point{p[0], p[3]} {
		assert.InDelta(t, pt.Value, follow(p, pt.At, exponential), 1e-9)
	}
	assert.InDelta(t, 400, follow(p, 2, exponential), 1e-9, "exponential between points")
}

func TestFaderFromSilenceRisesMonotonically(t *testing.T) {
	buf := NewStereo(samples(1.2))
	for i := range buf.L {
		buf.L[i], buf.R[i] = 1, 1
	}
	Fader{Points: []Point{{At: 0.1, Value: math.Inf(-1)}, {At: 1.1, Value: 0}}}.Process(buf)
	assert.Zero(t, buf.L[0])
	for i := 1; i < buf.Len(); i++ {
		require.GreaterOrEqualf(t, buf.L[i], buf.L[i-1], "sample %d", i)
	}
	assert.Equal(t, float32(1), buf.L[buf.Len()-1])
}

func TestLoudness(t *testing.T) {
	// EBU Tech 3341's first case: a 1kHz sine at -23dBFS in both channels reads -23 LUFS.
	buf := NewStereo(samples(20))
	a := dbToGain(-23)
	for i := range buf.L {
		x := float32(a * math.Sin(2*math.Pi*1000*float64(i)/SampleRate))
		buf.L[i], buf.R[i] = x, x
	}
	integrated, short := Loudness(buf)
	assert.InDelta(t, -23, integrated, 0.1)
	assert.InDelta(t, -23, short, 0.1)

	integrated, short = Loudness(NewStereo(samples(5)))
	assert.True(t, math.IsInf(integrated, -1))
	assert.True(t, math.IsInf(short, -1))
}

func TestEncodings(t *testing.T) {
	s := NewStereo(10)
	assert.Len(t, WAV(s), 44+40)
}

func TestLimiterHoldsTheCeiling(t *testing.T) {
	buf := NewStereo(samples(0.5))
	for i := range buf.L {
		x := float32(0.3 * math.Sin(2*math.Pi*220*float64(i)/SampleRate))
		if i > samples(0.2) && i < samples(0.25) {
			x *= 6
		}
		buf.L[i], buf.R[i] = x, -x
	}
	Limiter{Ceiling: -3}.Process(buf)
	assert.LessOrEqual(t, float64(buf.Peak()), dbToGain(-3)+1e-6)
	assert.Greater(t, abs32(buf.L[samples(0.1)+55]), float32(0.25), "quiet passages pass untouched")
}

func TestPluckBendArrivesInTune(t *testing.T) {
	from, to := MustKey("D4"), MustKey("E4")
	p := Patch{Oscs: []Osc{{Wave: Pluck, String: String{Decay: 8, Brightness: 0.6, Pick: 0.15}}}}
	s := oneNote(p, from, 0, 1)
	s.Parts[0].Notes[0].Slide = to - from
	s.Parts[0].Notes[0].SlideTime = 0.1
	out := Render(s)

	end := measureHz(t, out.L[samples(0.5):], Hz(to))
	assert.InDelta(t, 0, 1200*math.Log2(end/Hz(to)), 5)
}

func TestPhraseRatchetsSplitTheStep(t *testing.T) {
	notes := Phrase{Step: 0.3}.Notes(Pattern("C4", "x*3 X*2"))
	require.Len(t, notes, 5)
	assert.InDelta(t, 0.1, notes[1].At, 1e-9)
	assert.InDelta(t, 0.2, notes[2].At, 1e-9)
	assert.InDelta(t, 0.45, notes[4].At, 1e-9)
	assert.Greater(t, notes[3].Vel, notes[0].Vel, "the accent survives the ratchet")
}

func TestTapeStopEndsInSilence(t *testing.T) {
	buf := NewStereo(samples(1))
	for i := range buf.L {
		buf.L[i] = float32(0.5 * math.Sin(2*math.Pi*300*float64(i)/SampleRate))
		buf.R[i] = buf.L[i]
	}
	TapeStop{At: 0.2, Length: 0.3, Until: 0.8}.Process(buf)
	assert.NotZero(t, Stereo{L: buf.L[:samples(0.2)], R: buf.R[:samples(0.2)]}.Peak(), "before the stop is untouched")
	held := Stereo{L: buf.L[samples(0.5):samples(0.8)], R: buf.R[samples(0.5):samples(0.8)]}
	assert.Zero(t, held.Peak())
	assert.NotZero(t, Stereo{L: buf.L[samples(0.8):], R: buf.R[samples(0.8):]}.Peak(), "and resumes after")
}

func TestTakesReseedTheNoise(t *testing.T) {
	s := oneNote(tone(Noise), 60, 0, 0.1)
	assert.Equal(t, Render(s), Render(Take(s, 0)))
	assert.NotEqual(t, Render(s), Render(Take(s, 1)))
	assert.Equal(t, Render(Take(s, 1)), Render(Take(s, 1)))
}
