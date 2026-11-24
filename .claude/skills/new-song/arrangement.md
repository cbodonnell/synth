# Writing a song as a synth.Arrangement

Everything a song needs, so the other songs don't have to be read. The types live in
`arrangement.go`, `patch.go`, `effects.go`, `automation.go` and `mix.go` if a field
isn't covered here.

## Shape

```go
var id = synth.Arrangement{
	ID: "id", Name: "Name",
	Tempo: 100,             // BPM
	Meter: []int{3},        // beats per bar, repeating; nil is 4/4
	Key:   "D minor",       // for display and the studio's scale; not used in rendering
	Sections: []synth.Section{{Name: "Intro", Bars: 4}, {Name: "A", Bars: 8}},
	Loop:  "A",             // loops from this section to the end; "" plays once
	Tracks: []synth.Track{...},
	Buses:  []synth.Bus{...},
	Master: []synth.Effect{..., synth.Limiter{}},
	Target: -14,            // integrated LUFS the songs test holds within 1 LU
}

func Name() synth.Song { return id.Song() }
```

The sections' bars are the song's length. Register the var in `Arrangements()` in
`songs/songs.go`, after the last song and before `Hits()`.

Meters: 3/4 is `{3}`. 6/8 is `{2}` with `Step: 1.0 / 3`, so each beat is a dotted
quarter. An irregular phrase is a list that repeats: `{3, 3, 3, 4}`.

## Tracks

```go
{
	Name: "bass", Patch: patches.Bass(), Level: -6, Pan: 0,   // Level dB, Pan -1..1
	Inserts: []synth.Effect{...},                           // in order, on this track only
	Sends:   []synth.Send{{Bus: "hall", Level: -10}},       // dB into a return bus
	Out:     "drums",                                       // a group bus instead of the master
	Clips:   []synth.Clip{...},
	Play:    []synth.Play{{Clip: "A", Bar: 5, Times: 4}},   // bars count from 1
}
```

- Plays on one track may not overlap. Lines that sound together with different
  settings, such as a piano's two hands, are separate tracks.
- A track's name seeds its noise. Renaming a track changes its texture.

## Clips

```go
{Name: "A", Step: 0.5, Gate: 0.9, Vel: 0.8, Swing: 0.05, Humanize: 0.005, Seed: 3,
	Bars: []string{"C4 - E4 G4 . . C5 -", ...}}
```

- `Step` is beats per token. Every bar string must fill its bar exactly: 4/4 at
  `Step: 0.5` is 8 tokens, at `0.25` it's 16.
- `Gate`: the fraction of each note that sounds (0 means 1). `Vel`: 0 means 0.8.
- `Strum`: seconds between a chord's notes, lowest first.
- `Swing`: the fraction of a step that every second step is pushed late. 0.33 is a
  triplet shuffle.
- `Glide`: seconds a slide takes.
- `Humanize`: up to this many seconds of timing and velocity drift. It's reproducible
  from `Seed`.
- `Pedal: true` holds every note to the next bar line.
- `Hit: "C2"` makes the bars drum lines: `x` hits, `X` is an accent, `o` is a ghost,
  `x*3` is a ratchet, `.` is a rest. Hit can be a chord (`"C3,E3,G3"`) to strike it in
  a rhythm.
- `Notes: []synth.Note{{At: 0, Dur: 8, Key: 50, Vel: 0.8, Slide: 12}}` holds what isn't
  on a grid. At and Dur are beats from the clip's start, Key is MIDI, and `Length` is
  the clip's length in bars. A riser or a long swell goes here.

Tokens:

| Token | Meaning |
|---|---|
| `C4` | a note |
| `Bb3,Db4,F4` | a chord |
| `-` | holds the previous note another step |
| `.` | a rest |
| `D5/E5` | bends or slides to E5 over `Glide` |
| `F#2*3` | three even hits in the step |
| `C4!` / `C4?` | accent / ghost |
| `\|` | bar line, for the eye only |

Note names are checked when the song is built (`MustKey`), so a typo panics in the test.

## Patches

Run `make patches` for the library with what each patch is and how to play it. Use a
library patch as written when it fits. For a small change, write a func in the song's
file that edits a copy:

```go
func idKick() synth.Patch {
	p := patches.Kick()
	p.Amp.Decay = 0.3
	return p
}
```

When nothing in the library is close, write a new patch. If it could serve other songs,
put it in `patches/`, in the file for its family (`band`, `downtempo`, `edm`, `festival`,
`folk`, `orchestral`, `trap`, or `patches.go` for general ones), or in a new file for a
new family. Give it a doc comment that says what it is and how to play it, since
`make patches` prints that. A patch only this song will use goes in the song's file.

A patch is plain data:

- `Oscs`: `Sine` (with `FM{Ratio, Index, Decay, Sustain}` for bells, electric pianos and
  glass), `Triangle`, `Saw`, `Pulse` (`Width`), `Noise`, `Pluck` (a Karplus-Strong
  string: `String{Decay, Brightness, Pick}`), `Strike` (a piano string).
  - Each oscillator has `Level`, `Semi`, `Cents`, `Pan`, and `Fixed` Hz for drum bodies.
  - `Env` gives one oscillator its own envelope, for a transient such as a pick, a quill
    or a hammer.
- `Unison{Voices, Detune (cents), Spread}`.
- `Amp`: an ADSR in seconds. The zero envelope is an organ's gate.
- `Filter{Mode, Cutoff, Resonance, KeyTrack, Env, EnvAmount (octaves), Velocity}`.
  `Cutoff` 0 bypasses it.
- `LFOs`: `Vibrato` (cents), `Wah` (octaves), `Tremolo`, `PWM`, `AutoPan`. Each has a
  `Shape`, a `Rate` in Hz, and a `Delay` before it fades in.
- `Sweep{Semis, Time}`: a pitch drop or scoop at the start of each note. A kick drum
  uses it.
- `Drive` saturates the sound. `Velocity` sets how much velocity changes the level.
- `Gain` sets the patch's raw level. Library patches sit around 0.05 to 0.3. Set it so
  a note at `Level: 0` reads close to the other patches in `-stems`.

## Buses and effects

```go
Buses: []synth.Bus{
	{Name: "drums", Effects: []synth.Effect{synth.Compressor{Threshold: -18, Ratio: 4, Makeup: 4}}},
	{Name: "hall", Effects: []synth.Effect{synth.Reverb{Size: 0.85, Damp: 0.45, PreDelay: 0.03}}},
	{Name: "echo", Effects: []synth.Effect{synth.Delay{Time: 0.75, Feedback: 0.4, Damp: 2500, PingPong: true}}},
},
```

Effects:

| Effect | Fields |
|---|---|
| `EQ` | `Bands{Kind: Peak/LowCut/HighCut/LowShelf/HighShelf, Freq, Gain, Q}` |
| `Drive` | `Amount, Tone, Mix` |
| `Chorus` | `Rate, Depth, Delay, Feedback, Mix` |
| `Phaser` | `Rate, Depth, Center, Stages, Feedback, Mix` |
| `Delay` | `Time` (beats), `Feedback, Damp, PingPong, Mix` |
| `Reverb` | `Size, Damp, Width, PreDelay, Mix` |
| `Compressor` | `Threshold, Ratio, Attack, Release, Makeup` |
| `Crush` | `Bits, Hold` |
| `Limiter` | `Ceiling, Release` |
| `Sidechain` | `Key` (track names, comma-separated), `Depth, Attack, Release`. Ducks on the key track's notes. |
| `TapeStop` | `At, Length, Until`, all beats from the song's start |

- Automation is in beats from the song's start:
  - `AutoFilter{Mode, Resonance, Points}` (Hz)
  - `Fader{Points}` (dB)
  - `Panner{Points}`

  Each takes `Points: []synth.Point{{At: 32, Value: 400}, {At: 64, Value: 8000}}`.
- The master chain ends in `synth.Limiter{}`. To reach `Target`, raise the track levels
  together, or add `Makeup` on a gentle master compressor. Don't lean on the limiter.
- A drum-bus compressor flattens velocity crescendos. A build's snare roll goes on its
  own track, off the bus.

## Techniques, and where one is written out

Read only the lines you need, found with grep. Don't read whole files.

| Technique | File |
|---|---|
| Kick-keyed pumping, a filter build, a riser | `songs/brighterdays.go` |
| `TapeStop` before a drop | `songs/smokesignal.go` |
| A full band's mix | `songs/lavalamp.go` |
| Irregular meter | `songs/goldenbrown.go` |
| 6/8 | `songs/interstellar.go` |
| Solo piano, pedalled, no loop | `songs/elegy.go` |
