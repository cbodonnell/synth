# Improvements

The additions that would do the most for `synth`, ranked by value to a game against the
cost of building them. The engine already covers subtractive, plucked, and two-operator
FM sound well; the largest gaps are in what happens at runtime (where FMOD and Wwise
live) and in how a sound changes over time (automation and modulation).

Every proposal keeps the library's principles:

- No audio assets. Anything table-like (wavetables, impulse responses) is synthesized.
- `Render` stays pure. Randomness is seeded from data, never from `math/rand` or time.
- Effects stay values that start from silence on every call.
- Tests assert function, never content.
- Breaking changes are fine; prefer them to shims.

## Summary

| #  | Addition                                     | Area      | Value  | Cost   |
|----|----------------------------------------------|-----------|--------|--------|
| 1  | One mixing stream in the player              | player    | high   | medium |
| 2  | Sound-effect variation                       | player    | high   | low*   |
| 3  | Per-sound voice limits                       | player    | high   | low*   |
| 4  | Music crossfades and bar-quantized switches  | player    | high   | medium |
| 5  | Layered music (vertical remixing)            | player    | high   | medium |
| 6  | Ducking music under sound effects            | player    | medium | low*   |
| 7  | Positional pan and distance                  | player    | medium | low*   |
| 8  | General automation (`Fader`, `Point.Value`)  | synth     | high   | low    |
| 9  | Loudness metering and targets                | studio    | high   | low    |
| 10 | Tempo-relative time                          | synth     | medium | low    |
| 11 | Oscillator sync, ring mod, noise colors      | synth     | medium | low    |
| 12 | Second filter and a ladder model             | synth     | medium | medium |
| 13 | Modulation matrix                            | synth     | medium | high   |
| 14 | Wavetable oscillator                         | synth     | medium | medium |
| 15 | Multi-operator FM                            | synth     | medium | medium |
| 16 | Missing effects                              | synth     | medium | varies |
| 17 | MIDI export, then import                     | studio    | medium | medium |
| 18 | Chord, scale, and arpeggio helpers           | synth     | low    | low    |
| 19 | Watch mode in `studio`                       | studio    | low    | low    |

\* once #1 exists.

---

## Tier 1: runtime (the game-audio gap)

### 1. One mixing stream in the player

**Why.** Right now every effect is its own `ebaudio.Player`, and the music is another one.
Ebiten mixes them, but it can't pan a player, change its pitch, or start two of them on
the same sample. Most of tier 1 needs one or more of those, so this change comes first.

**Shape.** `Player` owns two long-lived `io.Reader`s, one for music and one for effects
(or one combined reader). Each reader sums the voices currently playing from their cached
PCM, applying per-voice gain, pan, and playback rate, with gains smoothed over a few
milliseconds so changes never click. The Ebiten context sees one player per stream.

- Voices start on the next buffer the reader fills, which is sample-accurate relative
  to one another.
- Playback rate uses cubic (Hermite) interpolation, as `Pluck` already does.
- Baked music still streams: a voice's source is a `ReadSeeker` (decoded Vorbis or
  PCM), not necessarily an in-memory slice.
- The loop wrap (`intro`/`body`) moves from `NewInfiniteLoopWithIntroF32` into the voice.
  Keep the bake's 2048-sample overlap and crossfade at the seam the same way Ebiten
  does now.

**Tests.** A recording reader: two voices started together stay sample-aligned; a gain
change has no step larger than a smoothed ramp allows; a looping voice's seam is
continuous; rate 2 halves the length.

### 2. Sound-effect variation

**Why.** A footstep, hit, or pickup fired over and over sounds identical every time,
which is the "machine-gun" effect every game audio tool exists to hide.

**Shape.** Two cheap layers, both declared on the content:

```go
type Variation struct {
    Pitch float64 // semitones, ± around the original, applied as playback rate
    Gain  float64 // dB, ±
    Takes int     // distinct renders, each seeded differently; 0 and 1 mean one
}

type Song struct {
    ...
    Vary Variation
}
```

- `Pitch` and `Gain` are applied per play in the mixing stream (#1). They cost nothing
  to render.
- `Takes` renders the song N times with a seed that feeds the noise (`ID#n`), so a
  noisy hit changes its texture as well as its pitch. Takes round-robin, never
  repeating the last one. `bake` writes every take, and the manifest gains a count.
- The per-play randomness lives in the player, not in `Render`, so rendering stays pure.

**Tests.** Successive plays never repeat a take; pitch stays within its bound; a
song with no `Vary` plays exactly as it does today.

### 3. Per-sound voice limits

**Why.** `maxEffects = 16` is global and drops the oldest sound. Ten pickups in one
frame can push out a level-up fanfare.

**Shape.** `Song.Voices int` (0 means unlimited) caps the concurrent plays of one
sound, stealing that sound's own oldest voice. The global cap stays, but it should steal
by priority (`Song.Priority int`, then age) rather than age alone. Platformer sounds
already sort into tiers, so priority follows the tier.

**Tests.** The (N+1)th play of a capped sound replaces its own oldest voice; a
high-priority sound survives a flood of low-priority ones.

### 4. Music crossfades and bar-quantized switches

**Why.** `PlayMusic` currently stops the old song and starts the new one: a hard cut, at
any moment. Area changes and combat transitions want either a crossfade or a cut on the
next bar.

**Shape.**

```go
type Transition struct {
    Fade time.Duration // 0 is a cut
    On   Boundary      // Now, Beat, Bar, or Loop (the end of the current pass)
}

func (m *Mixer) PlayMusicWith(s synth.Song, t Transition)
```

The boundary is computed from the current song's `Meter` through `SongTime`, as `Beat()`
already does, then converted to a sample offset in the mixing stream. That makes the
switch sample-exact, with no frame jitter. A song still rendering when its boundary
arrives waits for the boundary after the render lands, not for the next frame.

**Tests.** Driven through the recording `Sink`: a `Bar` switch starts the new song at a
bar line of the old one; a fade's summed gain stays bounded.

### 5. Layered music (vertical remixing)

**Why.** Intensity layering (add drums when enemies appear, strip back to the pad in a
menu) is the most common adaptive-music technique, and the songs are already split into
parts.

**Shape.** `Song.Layers map[string][]string`, from a layer name to part names. Parts in
no layer form the base. `RenderLoop` renders each layer as its own `Looped`, through the
same buses and master. The mixer plays them as aligned voices (#1), and
`Mixer.SetLayer(name, gain, fade)` moves one.

The caveat is that bus and master processing are non-linear (compressors, limiters), so
the layers summed back together don't exactly equal the full mix. Accept the difference,
or render each layer with the master chain bypassed and apply a limiter in the stream.
Choose after hearing both.

**Tests.** All layers at full gain stay within a small error of the full render (or
exactly equal it with the master bypassed); every layer loops seamlessly.

### 6. Ducking

`Song.Duck float64` (dB) on important sounds pulls the music down while they play,
attacking quickly and releasing slowly, inside the music stream. It's `Sidechain` at the
mixer level. Test that the music recovers fully after the sound ends.

### 7. Positional pan and distance

`Mixer.PlayAt(s, pan, distance)`: equal-power pan in the stream, and gain (plus an
optional low-pass for far sounds) from distance. `PlayGain` already exists for the "heard
from across the map" case; this generalizes it.

---

## Tier 2: shaping over time

### 8. General automation

**Why.** `AutoFilter` is the only thing that moves over time. Fades, swells, a send
opening into a drop, and a pad's pan drifting all get faked today with velocity or extra
parts.

**Shape.** Generalize `Point` to `{At, Value float64}`, and evaluate it with one shared
`curve(points, t)` that interpolates in the right domain (log for Hz, dB for gain). Then
add automation as effects, so it works wherever an insert does (parts, buses, master):

```go
type Fader struct{ Points []Point } // dB
type Panner struct{ Points []Point } // -1..1
```

`AutoFilter.Points` moves to `Value` (Hz). Send-level automation is a `Fader` as the first
insert on the bus's input, or later a `Send.Points`. All of these are time-invariant
relative to the song, so loops stay seamless as long as the curve's value at `Length`
matches its value at `LoopStart`. `Validate` should check that.

**Tests.** A fader from -∞ to 0 dB rises monotonically; the curve passes through every
point; a loop whose curve doesn't match at the seam fails `Validate`.

### 9. Loudness metering and targets

**Why.** `-stems` reports levels, but nothing says whether Brighter Days is louder than
Amber Hours, or whether a jump is louder than a fanfare. Games set loudness targets
(around -16 LUFS integrated for music, by tier for effects).

**Shape.** `synth.Loudness(Stereo) (integrated, shortTermMax float64)`, implementing
ITU-R BS.1770 (K-weighting, 400 ms blocks, gating). `studio` prints it per song and with
`-stems`. An optional `Song.Target` (LUFS) lets a test check every song lands within
±1 LU of its target. That's a functional assertion, not content.

**Tests.** A 1 kHz sine at -20 dBFS measures about -23 LUFS (the standard's reference
point); silence is -∞.

### 10. Tempo-relative time

`Meter.Beats(n float64) float64` and `Meter.Bars(n)` return seconds, and `Delay.Time` and
`LFO.Rate` docs point to them. That covers tempo-synced delays and LFOs without adding a
tempo concept to effects. A tempo map (ritardando) is a larger change, since notes are in
absolute seconds. Defer it until a song needs one.

---

## Tier 3: a wider palette

### 11. Oscillator sync, ring mod, noise colors

These are cheap and fit the existing `Osc`:

- `Osc.Sync float64`: a ratio above 1 hard-syncs this oscillator to the played pitch,
  giving the classic tearing lead. PolyBLEP already handles the reset discontinuity.
- `Osc.Ring int`: multiply by the oscillator at that index instead of adding.
- `PinkNoise` and `BrownNoise` waves (Voss-McCartney and leaky-integrated white), for
  wind, surf, and hiss without filtering white noise into shape.

Tests: sync's fundamental is the played pitch; pink noise falls about 3 dB per octave.

### 12. Second filter and a ladder model

`Patch.Filter2` in series or parallel, with `Filter.Model` choosing `SVF` (today's) or
`Ladder`, a 4-pole with tanh stages that self-oscillates and saturates the way analog
basses do. The test is that self-oscillation is in tune at high resonance with key
tracking.

### 13. Modulation matrix

Replace the fixed `LFOTarget` set with routes:

```go
type Mod struct {
    From   Source // LFO(i), Env(i), Velocity, Key, NoteRandom
    To     Dest   // Pitch, Cutoff, Resonance, Level, Pan, Width, FMIndex, OscLevel(i), ...
    Amount float64
}
Patch.Mods []Mod
Patch.Envs []Envelope // extra envelopes for Env(i)
```

Today's LFO targets become routes, which is a breaking change to every patch. Do it in
one pass with the patches library. `NoteRandom` is seeded like noise (part name, note
time, and key), so it stays pure. This is the largest synth change; build it after #11,
which is the work most likely to show what routing is actually needed.

### 14. Wavetable oscillator

`Wavetable{Frames [][]float64 (harmonic amplitudes), Position float64}`, a modulatable
dest once #13 exists. Frames are generated from harmonic spectra at init, mip-mapped per
octave against aliasing, so no assets ship. A handful of built-in tables (formant sweeps,
PWM-like, organ drawbars) covers most uses.

### 15. Multi-operator FM

Generalize `FM` to up to four operators with a few fixed algorithms (stack, two pairs,
three-into-one, and feedback on the top). That reaches electric pianos, bells, and FM
basses, which two operators can't. Each operator has its own envelope, extending the
`Decay`/`Sustain` index shape that's already there.

### 16. Missing effects

Roughly in order of usefulness here:

- `Gate`: tightens the drum bus and makes gated-reverb snares.
- `Rotary`: a Leslie (two speeds, horn and drum, Doppler and AM). Lava Lamp's organ
  currently fakes it.
- `Flanger`: a short modulated delay with feedback. It may share code with `Chorus`.
- `Saturator` / `Tape`: soft clipping with level-dependent EQ, and optional wow and
  flutter.
- `Convolution`: a reverb from an impulse response synthesized from decaying, filtered
  noise. It makes plates, rooms, and springs possible, but costs FFT convolution.
- `Transient`: attack and sustain shaping without a threshold.

Each needs only the existing effect tests: silence stays silent, and the tail dies within
`Tail()`.

---

## Tier 4: workflow

### 17. MIDI export, then import

Export first (`studio -midi`): notes are already absolute, and `Meter` gives bars, so a
song opens in any DAW for checking. That's one track per part, with velocity, and slides
written as pitch bend. Import comes later (`synth.ReadMIDI(r) map[string][]Note`, a
track's name to its notes) for sketching in a DAW and finishing in Go.

### 18. Chord, scale, and arpeggio helpers

```go
synth.Chord("Dm7", 4)            // keys
synth.Scale("E", Mixolydian)      // for walking lines
synth.Arp(chord, Up, steps, rate) // []Note
```

Phrase notation could accept chord symbols in brackets (`[Dm7]`). This is low priority,
because the scores are readable already, but it would shorten the organ and pad parts.

### 19. Watch mode

`studio -watch -id X` re-renders and replays when a Go file changes, by rebuilding
through `go run` in a loop. It's the closest cheap substitute for turning a knob in a
DAW.

---

## Non-goals

- **Sample playback.** It contradicts the premise. Covered instead by physical models
  (#16 convolution, a future modal synth) and better noise (#11).
- **Real-time synthesis of whole songs.** Rendering ahead (and baking) is what keeps
  WebAssembly viable. Runtime change comes from the mixing stream (#1) working on
  rendered audio.
- **A GUI.** `studio` and the demo are the tooling.

## Suggested order

1. #8 automation and #9 loudness: small, self-contained, and they improve every song.
2. #1 mixing stream, then #2 variation and #3 voice limits: the biggest win for games.
3. #4 transitions and #5 layers: adaptive music.
4. #11, then #12 and #16: palette quick wins.
5. #13 mod matrix with #14 and #15 riding on it.
6. #17 to #19 as wanted.
