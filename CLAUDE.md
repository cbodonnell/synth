# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`synth` is a music and sound-effects library for Ebiten games. No audio asset ever ships:
instruments, scores, and mixes are Go values, and everything is synthesized at runtime.
It grew out of the NES-shaped synth in `~/code/pets` (`internal/ui/audio`) and keeps its
principles while widening the palette.

## Conventions

- **Comments**: none, except to document genuinely non-obvious behavior. Never narrate
  what the code already says.
- **Tests**: cover core boundary cases only. Never assert content: songs, patches, and
  mixes change freely, and a test that pins them is a tax on editing them. Audio tests
  assert function (a loop's seam is continuous, a string is in tune), never exact samples.
- Prefer deleting to accumulating, in code, tests, and comments alike.
- Breaking changes are acceptable at this stage. Prefer them over compatibility shims.
- Never commit changes to git. The user handles that.

## Commands

```sh
make test                          # everything
make vet
make list                          # songs and sounds
make play ID=goldenbrown           # render to a wav, print the envelope, play it (afplay)
make play ID=kick DUMP=1           # write the wav only
make score ID=goldenbrown          # piano roll, one column per half beat
go run ./cmd/synth -id lavalamp -stems   # each part's peak and level, for mixing
make demo                          # Ebiten window: Golden Brown with a beat-synced display
```

Never launch the demo to check work; cover it with tests and `make play ... DUMP=1`.

## Architecture

### Rendering is pure

`synth.Render(song)` is a pure function: the same Song renders byte-identical samples.
Noise comes from splitmix64 seeded by the part's name and each note's own time and key,
never `math/rand` or a note's index (an index changes when `RenderLoop` filters the notes,
and the loop would no longer match its first pass). Parts render concurrently into their
own buffers and are summed in order, so concurrency never changes the output.

### The signal path

`Patch` (oscillators, unison, amp envelope, state-variable filter with its own envelope,
LFOs, pitch sweep, drive) → `Part` (notes, inserts, sends, pan) → `Bus` → DC block →
`Song.Master` → clip. A bus is a send return (one reverb for many parts) or, with
`Part.Out`, a group (a drum kit compressed together). Master chains should end in a
`Limiter`; the final clip is a safety, not a mastering tool. Effects implement `Effect`
and start
from silence on every call, so an effect value is data, not a stateful object. Each
reports its `Tail`, which sizes the render.

Oscillators are band-limited (PolyBLEP) saw and pulse, triangle, sine with two-operator
FM, seeded white noise, and `Pluck`, a Karplus-Strong string tuned with an allpass for the
fractional delay. A string whose pitch moves (slide, sweep, vibrato) switches to a
Hermite-interpolated read, which bends smoothly where the allpass would click.
Envelope curves are exponential but normalized to land exactly on target, so every note
ends at a true zero, with a 2ms minimum attack and 5ms minimum release against clicks.

### Loops

`Song.Length` is authoritative. `RenderLoop` returns `Intro` (everything before
`LoopStart` plus the first pass, from silence) and `Body` (the loop with its own tail
wrapped onto its start). Body is rendered from only the loop's notes, so the intro's tail
is never heard twice. This is exactly Ebiten's `NewInfiniteLoopWithIntroF32` shape.
Time-invariant effects (EQ, delay, reverb) loop seamlessly; a compressor on the master is
approximate at the seam. A note starting at or after `Length` fails `Validate`.

### Notation

`Phrase.Notes` reads a step grid: `C4` a note, `Bb3,Db4,F4` a chord, `-` hold, `.` rest,
`D5/E5` a bend or slide (over `Glide`), `F#2*3` a ratchet (three hits in one step, for hat
rolls), `|` a bar line (needs no spaces), `!`/`?` accent and ghost. `Strum` rolls chords,
`Swing` pushes every second step late, and `Humanize` adds reproducible drift.
`Pattern("C2", "X . x o")` writes drum and strum lines.
`MustKey` panics on a typo so a score fails when it's built.

### Playing in a game

`player.New()` creates the process's one Ebiten audio context (a second call panics).
A game holds a `player.Mixer` over it: levels, mute, music that doesn't restart when asked
for again, and `Beat()`, which maps the playhead through the loop onto `Song.Meter` so
animation can keep time. Renders happen on goroutines (`Prepare` warms them), so
`PlayMusic` never stalls a frame. Tests drive the Mixer with a recording `Sink`.

### Content

`patches` is a starting instrument library; `songs` holds example content:

- `goldenbrown.go`, the Golden Brown intro: harpsichord sound design and irregular metre
  (`Meter.Bars = {3, 3, 3, 4}`).
- `amberhours.go`, "Amber Hours": an original in Golden Brown's spirit with
  Interstellar's pipe organ, a D minor harpsichord waltz in `{3, 3, 3, 4}`: recorder,
  bass, brushed kit, a ticking clock, and organ pedal, ostinato, and chords opened by an
  `AutoFilter` swell box. Its form is a string of sections (`"AAAABBAA"`), each part
  writing its line pass by pass.
- `playasunrise.go`, "Playa Sunrise": the same `{3, 3, 3, 4}` lilt taken to a desert
  festival at dawn, E mixolydian and lightly swung: handpan keeping Golden Brown's
  interlock, tanpura and didgeridoo drones, djembes and shaker, sitar and bansuri
  sharing a hook with meend bends, a phased pad, wind, and chimes. It shares `form`
  with Amber Hours.
- `lavalamp.go`, "Lava Lamp Sun": a full band (kit on a group bus, Höfner-style bass,
  double-tracked phased rhythm guitars, wah-fuzz lead with bends, Leslie organ, Juno
  arp). The reference for mixing; balance with `go run ./cmd/synth -id <id> -stems`,
  which reports each part's level while it's sounding.
- `brighterdays.go`, "Brighter Days": Avicii-style progressive house. The reference for
  arrangement dynamics: `Sidechain` keyed by `Onsets(kick notes)` so pumping follows the
  kick, `AutoFilter` automating a build, a riser, and a build quieter than its drop.
- `smokesignal.go`, "Smoke Signal": 2015 festival trap. Half-time drums with ratcheted
  hat rolls, a gliding distorted 808, brass, and a `TapeStop` on the parts that should
  spin down before the drop (an insert per part, so the air horns survive the silence).
- `platformer.go`, a side-scrolling MMO's effects (jump, hit, meso, level up, UI, ...),
  in `Sounds()`. The reference for sound effects: layers of transient, body, and tail
  through a `Limiter`, balanced by tier (UI quietest, combat and fanfares loudest).
  Anything fired many times a second stays dry, since a reverb's faint tail holds a
  voice for a second and the player keeps only 16.

Drum-bus compressors undo velocity crescendos; a build's roll or soft kick belongs on
its own part off the bus.

Renders are parallel per part and per note; notes land in their own buffers and are
summed in order, so scheduling never changes a sample.
