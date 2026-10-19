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
make export ID=goldenbrown         # a first draft of any song as a synth.Arrangement
go run ./cmd/synth -id lavalamp -stems   # each part's peak, level, and LUFS, for mixing
go run ./cmd/synth -bake <dir>     # bake everything to Ogg Vorbis (needs oggenc)
make demo                          # Ebiten window: Golden Brown with a beat-synced display
```

`cmd/synth` is `studio.Main` over the example content; a game wraps `studio` around its
own songs and sounds to get the same loop.

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
`Song.Master` → clip. Mix levels (`Part.Level`, `Send.Level`, `Bus.Level`) are dB, 0 is
unity. A bus is a send return (one reverb for many parts) or, with
`Part.Out`, a group (a drum kit compressed together). Master chains should end in a
`Limiter`; the final clip is a safety, not a mastering tool. Effects implement `Effect`
and start
from silence on every call, so an effect value is data, not a stateful object. Each
reports its `Tail`, which sizes the render.

Automation is effects too: `AutoFilter` (Hz), `Fader` (dB), and `Panner` follow `Point`s
through song time, each interpolating in its own domain through `follow`. Being
song-time effects inside a render, they loop seamlessly whatever their values at the seam.
`Loudness` is BS.1770 (integrated and short-term max, LUFS); studio prints it per song
and per stem, and `Song.Target` makes the songs test hold a song within 1 LU of it.

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
is never heard twice. The player plays it as an intro once, then the body forever.
Time-invariant effects (EQ, delay, reverb) loop seamlessly; a compressor on the master is
approximate at the seam. A note starting at or after `Length` fails `Validate`.

### Notation

`Phrase.Notes` reads a step grid: `C4` a note, `Bb3,Db4,F4` a chord, `-` hold, `.` rest,
`D5/E5` a bend or slide (over `Glide`), `F#2*3` a ratchet (three hits in one step, for hat
rolls), `|` a bar line (needs no spaces), `!`/`?` accent and ghost. `Strum` rolls chords,
`Swing` pushes every second step late, and `Humanize` adds reproducible drift. Both count
steps from the song's start (`At / Step`), so a line written in pieces swings and drifts
as one.
`Pattern("C2", "X . x o")` writes drum and strum lines.
`MustKey` panics on a typo so a score fails when it's built.

### Arrangements

Songs are written as `synth.Arrangement` literals, the declarative form the studio
(`specs/studio.md`) edits, and `Song()` compiles one to the `Song` the player and the
bake take. `Tempo` and `Meter` make bars; `Sections` name the form in bars and give the
song its length; `Loop` names the section the loop starts at. A `Track` is a `Part`
whose notes are `Clip`s placed by `Play` at bars (from 1), `Times` over. A clip is a
`Phrase`'s settings with `Step` in beats and `Bars`, one string per bar, each filling its
bar exactly; `Hit` makes the bars a `Pattern` line (a drum, or a chord struck in a
rhythm); `Notes` (in beats) hold what isn't on a grid; `Pedal` holds notes to the bar
line. Song positions and durations inside effects (`Point.At`, `TapeStop`, `Delay.Time`,
`Sidechain.Hits`) are beats, and `Sidechain.Key` names the tracks whose onsets duck it.
`Arrangement.Validate` checks what a Song can't: unknown clips and buses, bars that
don't fill their meter, overlapping plays.

Keep arrangements literal: a field written as a helper call or a shared var is "custom"
to the studio, shown but not editable. Patches modified from a `patches` constructor are
the exception, written as a small func. Plays on one track may not overlap, so lines
with different settings that sound together (a harpsichord's two hands) are tracks of
their own.

`studio.Arrange` drafts an arrangement from any Song (`make export`): bars folded into
clips, grids inferred, patches matched to `patches` calls. It can't recover a humanized
line's seed, so a draft is a start, not a port.

### Playing in a game

`player.New()` creates the process's one audio output (a second call panics on the
desktop, where it's Ebiten's audio context).
A game holds a `player.Mixer` over it: levels, mute, music that doesn't restart when asked
for again, and `Beat()`, which maps the playhead through the loop onto `Song.Meter` so
animation can keep time. Renders happen on goroutines (`Prepare` warms them), so
`PlayMusic` never stalls a frame. `PlayGain` trims one effect, for a sound heard from
across the map. Tests drive the Mixer with a recording `Sink`.

The Player owns one output reading a `bus`: the music and effects `stream`s
summed, then a peak limiter (no look-ahead, so no latency; ceiling -1 dB, or
`player.Ceiling`) so sounds that land together can't clip. Each stream sums its own
voices: per-voice gain ramped over 5ms so nothing clicks, playback rate with Hermite
interpolation, and the loop wrap (`looped`, which blends the bake's overlap into the
loop start as Ebiten's `InfiniteLoop` did). A voice starts on the first frame of the
stream's next mix, so voices started together are sample-aligned. A new effect waits
behind whatever the output has queued. On the desktop the output is an Ebiten player
with a 40ms buffer (`player.Buffer`). In a browser it is an AudioWorklet of the player's
own (`output_js.go`, `worklet.js`): Ebiten's asks for 170ms at a time and pads what isn't
mixed with silence, so it would need a buffer longer than that. The worklet asks for
1024 frames whenever it holds less than 64ms (`player.Latency`), Go mixes them on the
spot, and the first gesture starts it, as a browser requires. A stall longer than the
latency is heard as a click and logged to the console. Browsers without AudioWorklet
(or pages not served securely) fall back to Ebiten with a 250ms buffer. Baked music decodes on its own
goroutine half a second ahead (`ahead`) so decoding stays out of the bus's reads, and
at gain 0 the music stream stands still rather than decoding a loop nobody hears.
`MusicPosition` is what the music stream has mixed less what the bus has mixed but
the output hasn't played. A sound's `Vary` (pitch and gain at random per play, and `Takes`: renders
whose noise is reseeded by `synth.Take`, never the same take twice running) and its
`Voices` cap and `Priority` are applied there, so `Render` stays pure. Stream tests read
the stream directly; they never need an audio device.

### Baking

Rendering a full band takes seconds, and on WebAssembly's one thread it starves the game's
frames, so a game ships baked songs: `bake.Bake` renders each to Ogg Vorbis (via oggenc)
beside a `bake.json` manifest, and `player.New(player.Baked(fsys))` decodes them instead,
streaming music and holding effects in memory. The songs stay the source: the manifest
fingerprints each one (by walking the value, plus `Revision` and the bake's format), a
game's test asserts `bake.Stale` is empty, and the player renders live any song whose
bake doesn't match. Bump `Revision` when a change to rendering changes what songs sound
like. A looping file is the intro, then the first pass only until it settles into the
body (`settle`), then the body rotated to start there, then a little more of it, which
the player blends into the loop start to hide the codec's seam. A sound with `Takes`
bakes a file per take.

### Content

`patches` is a starting instrument library; `songs` holds example content:

- `goldenbrown.go`, the Golden Brown intro: harpsichord sound design and irregular metre
  (`Meter.Bars = {3, 3, 3, 4}`).
- `amberhours.go`, "Amber Hours": an original in Golden Brown's spirit with
  Interstellar's pipe organ, a D minor harpsichord waltz in `{3, 3, 3, 4}`: recorder,
  bass, brushed kit, a ticking clock, and organ pedal, ostinato, and chords opened by an
  `AutoFilter` swell box, in sections A A B B A A after a two-pass intro.
- `playasunrise.go`, "Playa Sunrise": the same `{3, 3, 3, 4}` lilt taken to a desert
  festival at dawn, E mixolydian and lightly swung: handpan keeping Golden Brown's
  interlock, tanpura and didgeridoo drones, djembes and shaker, sitar and bansuri
  sharing a hook with meend bends, a phased pad, wind, and chimes. Its form is Amber
  Hours'.
- `lavalamp.go`, "Lava Lamp Sun": a full band (kit on a group bus, Höfner-style bass,
  double-tracked phased rhythm guitars, wah-fuzz lead with bends, Leslie organ, Juno
  arp). The reference for mixing; balance with `go run ./cmd/synth -id <id> -stems`,
  which reports each part's level while it's sounding.
- `brighterdays.go`, "Brighter Days": Avicii-style progressive house. The reference for
  arrangement dynamics: `Sidechain` keyed by the kick tracks so pumping follows the
  kick, `AutoFilter` automating a build, a riser, and a build quieter than its drop.
- `smokesignal.go`, "Smoke Signal": 2015 festival trap. Half-time drums with ratcheted
  hat rolls, a gliding distorted 808, brass, and a `TapeStop` on the parts that should
  spin down before the drop (an insert per part, so the air horns survive the silence).
- `woodlands.go`, "The Woodlands": the Firefly festival in the woods outside Dover,
  stomp-clap indie folk in G (acoustic guitar, banjo rolls, upright bass, whistled
  verse, a gang singalong with mandolin tremolo and glockenspiel, a shouted "Hey!")
  until nightfall, when crickets, fireflies, an arp, and a pad take a bridge that
  builds back into the last chorus.
- `interstellar.go`, the Interstellar main theme's organ intro: A minor in 6/8, an
  ostinato ticking against an E pedal while the chords move beneath it, the swell box
  breathing open and shut again by the loop's seam, a pedalled grand rolling the chords
  and singing a tune an octave above.
- `platformer.go`, a side-scrolling MMO's effects (jump, hit, meso, level up, UI, ...),
  in `Sounds()`. The reference for sound effects: layers of transient, body, and tail
  through a `Limiter`, balanced by tier (UI quietest, combat and fanfares loudest).
  Anything fired many times a second stays dry, since a reverb's faint tail holds a
  voice for a second and the player keeps only 16. Priority follows the tiers, and
  whatever a game repeats (footsteps, hits, coins) varies and is capped.

Drum-bus compressors undo velocity crescendos; a build's roll or soft kick belongs on
its own part off the bus.

Renders are parallel per part and per note; notes land in their own buffers and are
summed in order, so scheduling never changes a sample.
