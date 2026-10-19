# Studio

Songs and sounds are Go source, which agents edit easily and people don't: a bass line is
a loop over a form string writing each pass, a level is a linear `Gain: 0.42`, and the only way to hear a
change is `make play` and a wav from the start. The studio is a desktop Ebiten window over
`studio.Main`: a small music workstation (tracks on a timeline, clips in a piano roll, a
mixer, devices) that plays a song as it will ship, edits it by pointing and
listening, and saves back to the same Go source. Any game that wraps `studio.Main` gets it.

## Principles

1. **The source is the format.** No project file, MIDI or JSON. The studio edits a
   `synth.Arrangement` literal in its `.go` file, so a person's edit and an agent's
   are the same kind of diff.
2. **Touch only what was edited.** A save splices new source over the expressions that
   changed, found by `go/ast` position; comments, order and everything else survive
   byte for byte.
3. **What you hear is the game.** Playback is `synth`'s own render, looping as the player
   loops it (`RenderLoop`'s intro, then the body). No preview approximation.
4. **Literal is editable, code is respected.** A field written as a literal is a control.
   Any other expression (a helper call, a shared var, arithmetic) is shown with its value,
   labelled "custom: edit it in code", and kept as it is.
5. **Simple over complete.** One window, four views, MIDI-free. No audio recording,
   plugins, takes, tempo changes or time stretching.

## The arrangement

Procedural songs can't be edited by pointing: a note in the piano roll has no one place in
the source. So the studio edits a declarative song, `synth.Arrangement`, which compiles to
the `Song` the player and the bake already take:

```go
var riff = synth.Arrangement{
	ID: "riff", Name: "Riff",
	Tempo: 116, Key: "D mixolydian",
	Sections: []synth.Section{{"Intro", 2}, {"A", 8}, {"A'", 8}, {"B", 8}, {"A''", 8}},
	Loop:     "A",
	Tracks: []synth.Track{
		{
			Name: "kick", Patch: kickPatch, Level: -7.5, Out: "drums",
			Clips: []synth.Clip{
				{Name: "four", Step: 1.0 / 4, Hit: "C2", Swing: 0.12, Humanize: 0.004, Seed: 26, Bars: []string{
					"X . . . . . x . . . X . . . . .",
				}},
			},
			Play: []synth.Play{{Clip: "four", Bar: 3, Times: 16}, ...},
		},
		...
	},
	Buses:  []synth.Bus{...},
	Master: []synth.Effect{...},
	Target: -16,
}

func Riff() synth.Song { return riff.Song() }
```

- **Time is musical.** `Tempo` (BPM) and `Meter` (beats per bar, repeating, `nil` is 4/4)
  give bars and beats; `Song()` turns them into seconds. Times inside effects that are
  song positions or durations (`Point.At`, `TapeStop.At`/`Length`/`Until`, `Delay.Time`)
  are in beats in an arrangement, so they follow the tempo. A `Tempo` of 0 is a sound
  effect: time is seconds and there are no bars.
- **Sections** name the form in bars, in order: the ruler's markers and the length of the
  song. `Loop` names the section the loop starts at (`LoopStart`); the song loops to its
  end. No `Loop` is a song that plays once.
- **A track** is a `Part`: a `Patch`, `Level` and send levels in dB (0 is unity), `Pan`,
  `Inserts`, `Sends`, `Out`, and `Mute`, which leaves it out of `Song()`.
- **A clip** is a `Phrase`'s notation and settings: `Step` in beats, `Gate`, `Vel`,
  `Strum`, `Swing`, `Glide`, `Humanize`, `Seed`, and `Bars`, one string per bar, each
  filling its bar exactly. With `Hit` set the bars are drum lines in `Pattern` notation
  (`x X o`). A clip of a sound effect, or any line that isn't on a grid, holds `Notes`
  instead, with `At` from the clip's start.
- **`Play`** places a clip on its track at a bar (counted from 1, as the ruler shows it),
  `Times` over back to back. Plays on one track may not overlap.
- **Sidechain** takes `Key: "kick"` to duck on that track's onsets; `Song()` fills `Hits`.
- **Sound effects** add `Vary`, `Voices` and `Priority` as on `Song`. Their level is set by
  a master `Normalize{Peak}` (dB), a synth effect that scales the whole render so its
  loudest sample lands there, so loudness tiers are a device like any other.

Mix levels become dB across synth (`Part.Level`, `Send.Level`, `Bus.Level`; 0 is unity,
which retires "0 means 1"), so the source reads as the fader does. `Validate` checks
arrangements too: unknown clips and buses, bars that don't fill their meter, overlapping
plays, a `Loop` that names no section.

`studio export -id <id>` writes a first draft of an `Arrangement` from any `Song`: each
part a track, its notes cut into bars, repeated bars folded into one clip played many
times, notes on a step written as a grid (within the part's humanize) and others as
`Notes`, patches equal to a `patches` constructor written as that call. A ported song
keeps a test, while it's being ported, that its arrangement's `Song()` has the old song's
notes and mix; then the procedural version is deleted.

## Tool

`studio.Main` opens the studio with `-edit` (`make edit ID=goldenbrown`; a game
wrapping `studio.Main` gets the same flag). Desktop only. The display-free parts are
`studio/session` (the arrangement doc, edits, undo, the source writer, the render cache,
transport arithmetic, the supervisor), whose tests run in CI; `studio/desk` is the window,
built from ebui with a plain dark painter of synth's own.

```
┌─ Transport ───────────────────────────────────────────────────────────────┐
│ ▶ ■ ⟲  12.3.2   116 bpm  4/4  D mixolydian   ⌁ click   -15.8 / -16 LUFS  │
├─ Browser ─┬─ Arrangement (or Mixer) ──────────────────────────────────────┤
│ Songs     │ Intro │ A         │ A'        │ B         │ A''       │ ⟲ A→  │
│ Sounds    │ kick  M S ▮ │▒▒▒▒▒▒▒▒▒│▒▒▒▒▒▒▒▒▒│           │▒▒▒▒▒▒▒▒▒▒│      │
│ Patches   │ guitar M S ▮│▒▒▒▒▒▒▒▒▒│▒▒▒▒▒▒▒▒▒│░░░░░░░░░░│▒▒▒▒▒▒▒▒▒▒│      │
│ Effects   │ ...                      ▲ playhead                          │
├───────────┴─ Detail: clip editor (or device chain) ───────────────────────┤
│ piano roll / drum grid, velocity lane      │ clip settings                │
└───────────────────────────────────────────────────────────────────────────┘
```

| Area | Holds |
|---|---|
| Transport | play/stop, return to start, loop on/off, position as bar.beat.step, tempo, meter, key, metronome, the song's integrated LUFS against `Target`, render progress, Undo, Redo, Save (lit while unsaved), Bake (lit while the bake is stale) |
| Browser | the game's songs and sounds (opening one), `patches` (drag onto a track to set its patch, or onto the arrangement for a new track), synth's effects (drag into a chain) |
| Arrangement | the section ruler with the loop brace, a lane per track (name, mute, solo, level meter, colored by name) with its clips as blocks drawing their notes, group buses folding their tracks under them, automation lanes beneath their track |
| Mixer | a strip per track, bus and the master: fader (dB), pan, send knobs, inserts, out, mute, solo, peak and short-term LUFS meters; the master's integrated LUFS against `Target` |
| Clip editor | the selected clip: a piano roll (keys of the song's `Key` lit), or for a `Hit` clip a drum grid of `.` `o` `x` `X` cells; a velocity lane; the clip's settings and name |
| Device chain | the selected track's patch device, then its inserts; a bus's or the master's effects |

Views: `X` toggles the mixer, `Shift+Tab` the detail between clip and devices, `B` the browser, `A` the automation lanes. Space plays and stops; Return goes to
the start; clicking the ruler moves the playhead; `C` toggles loop; `K` the metronome
(heard in preview only). `M` and `S` mute and solo the selected track. Ctrl/Cmd+Z undoes,
Ctrl/Cmd+Shift+Z redoes, Ctrl/Cmd+S saves, Ctrl/Cmd+D duplicates, Delete deletes. Zoom:
Ctrl/Cmd and wheel or pinch about the cursor; `Z` fits the song. Caps Lock is musical
typing: `A` to `K` play the white keys through the selected
track's patch, `W` `E` `T` `Y` `U` the black, `Z` `X` the octave, `C` `V` the velocity, and
other shortcuts are off while it's on.

### Playback

- **As the game plays it.** With loop on, playback is the intro, then the body forever,
  from wherever the playhead was dropped; Seam (`Shift+C`) drops it two bars before the
  loop's end, to hear the wrap. Off, the song plays to its end and its tail.
- **Edits are heard while playing.** Every edit re-renders what it touched in the
  background and the playing buffer is swapped at the playhead, crossfaded over 5ms.
- **Incremental render.** `synth.Renderer` caches notes (by the patch, the part's seed and
  the note), parts (by their notes, patch and inserts) and buses, so moving a note renders
  one note and its part's inserts, a fader re-mixes, and a tempo change renders
  everything. Its output is byte-identical to `Render` and `RenderLoop`.
- **Solo is the studio's, mute is the song's.** Solo never reaches the source; mute is
  `Track.Mute`, saved.
- `player` gains a preview voice: a buffer played from an offset, looped over a region,
  swappable at its position, reporting the position for the playhead and meters.

### Editing

- **Arrangement.** Drag a clip to move it by bars (Alt drag copies the play, Ctrl/Cmd+D
  repeats it after itself), drag its right edge to change `Times`. Double-click an empty
  lane to add a one-bar clip. A clip played in many places is one clip: editing it edits
  them all; "Make unique" copies it under a new name. Sections
  are renamed, resized by their right edge (moving everything after them), inserted,
  deleted and duplicated with every track's plays in them. Drag the loop brace's start
  onto a section to set `Loop`. Tempo and meter are typed in the transport.
- **Clips.** In the piano roll, draw a note by clicking, stretch it by its end, move with
  drag or the arrows (up/down a semitone, Shift an octave), and set velocity in its lane.
  Notes snap to the clip's `Step`; a grid clip can't hold notes off its grid, so its step
  is changed first (halving it rewrites the bars, losing nothing). Accent and ghost show
  as velocity and are written back as `!`/`?`. Ratchets show as their hits and are
  edited as one note. A slide is drawn from a note to its target. With `Shift`, a click
  stamps the scale's triad on that degree, with `Alt` its seventh. Step input
  (`Ctrl/Cmd+E`): musical typing writes a note at the cursor and steps on, `Tab`
  writes a rest, `-` holds. In a drum grid, a click cycles a cell `.` `x` `X` `o`.
- **Mixer.** Faders, knobs and sends drag, double-click resets, typed values accept dB.
  Adding a send to a new bus creates the bus with a reverb. Out picks a group bus.
- **Devices.** Effects are drawn from a parameter table in synth (`synth.Params`: each
  field's unit, range, scale, and whether it's in beats), falling back to reflection for
  a game's own effects (number fields as knobs, enums as cyclers, `[]Point` as an
  automation lane). The patch device shows oscillators (wave, level, tune, width, FM,
  string), unison, the amp and filter envelopes drawn as curves, the filter with its
  response, LFOs, sweep and drive. Musical typing auditions it.
- **Patches from code.** A patch written as a call (`patches.Kick()`) or a var is custom:
  its device is shown read-only with "Make literal", which writes its current value into
  the track as a literal, cutting it loose from the call. The same for custom effects.
- **Automation.** An automated effect's `Points` are a lane under its track (or bus or
  master): click adds a point, drag moves it (snapped to the step), Delete removes it.
- **Undo.** Every edit is one step on a stack of whole copies of the doc, 200 deep; a drag
  or typing into a field is one step.
- **Numbers are written at the control's resolution** (0.1 dB, 0.01 of pan, 3 significant
  figures elsewhere) so diffs stay readable.

### Sounds

A sound effect is an arrangement with `Tempo` 0: the ruler is in milliseconds, each layer
is a track, clips hold `Notes`, snapping is off, 10ms or 50ms. Its inspector edits `Vary`,
`Voices` and `Priority`; Play auditions it as the player plays it, so a press with `Vary`
plays a fresh take and pitch, and Repeat (`R`) plays it every 300ms to hear the
variation.

## Saving

`session.Load` finds the `.go` file holding the `synth.Arrangement` literal with the song's
`ID` (searching the packages the songs were built from) and keeps its bytes. A save
compares the doc with the doc as last saved, item by item by name (tracks, clips, plays,
sections, buses, effects in order), and splices new source over only what changed:

- A clip's `Bars` are rewritten whole, one string per line; its `Notes` one per line.
- A changed field in a kept item is rewritten in place; a field gone to zero is removed;
  a new one is added at the end of its literal.
- Added items are appended to their list, one per line; deleted ones take their line,
  end-of-line comment and all; a comment on a line of its own stays.
- The result is `gofmt`ed and read back to the same doc, or nothing is written. Files are
  replaced by rename. If the file changed on disk since it was loaded, the save is
  refused and the studio says so.
- Bake renders the songs whose bake is stale into the game's bake directory (passed to
  `studio.Main`), leaving the rest.

## Live reload

Songs are compiled in, so the studio reloads by rebuilding itself.
`-edit` starts a supervisor that runs the studio as a child; the child watches the Go
files of every package in the main package's module and synth's (from `go list -deps`),
and when they change and settle runs `go build` on the main package
(`debug.ReadBuildInfo`). On success it hands off the new binary with its view (song,
playhead, zoom, views open) and exits; on failure it keeps playing the last build and
shows the compiler's output. While there are unsaved edits a finished build waits behind
"Reload, dropping unsaved edits". The studio's own saves need no rebuild. Closing with
unsaved edits asks first.

A song that isn't an arrangement opens read-only: its parts as tracks with their notes as
one clip each, playable, soloable and metered, every control custom.

## Phases

1. **Arrangement.** `synth.Arrangement` and `Song()`, levels in dB, beats in effects,
   `Sidechain.Key`, `Validate`, `studio export`, and porting the example songs (each with its
   equivalence test). No window.
2. **Player.** The window, browser, arrangement view (read-only), transport, the
   incremental renderer and preview voice, solo, meters and loudness, live reload.
3. **Mixer.** The mixer view and its edits, mute, the session doc, undo and saving.
4. **Clips.** The clip editor: piano roll, drum grid, velocity, clip settings, step input.
5. **Arranging.** Moving, copying and repeating plays, sections, the loop, tempo and meter.
6. **Devices.** The parameter table, effect and patch devices, Make literal, automation
   lanes, musical typing.
7. **Sounds.** Arrangements in seconds, `Notes` clips, `Normalize`, `Vary` audition, and
   porting `platformer.go`'s effects.

## Tests

- Round trip: every arrangement loads and saves with no edits to the same bytes.
- Per edit kind (a note drawn, moved, stretched, accented; a step halved; a drum cell
  cycled; a play moved, repeated; a section inserted; a fader, pan, send, insert parameter;
  a patch made literal; an automation point): the edit saved and then saved undone is the
  source again, byte for byte.
- Grid round trip: any clip's bars read as notes and written back are the same strings;
  notes written from the piano roll read back as those notes.
- `Song()`: bars, sections and the loop land on the seconds the meter gives; beats in
  effects scale with tempo; a muted track is absent; `Sidechain.Key` ducks on its
  track's onsets.
- `Renderer` matches `Render` and `RenderLoop` byte for byte after any sequence of edits.
- `export` of each example song compiles and its `Song()` has the original's notes.
- The preview voice swaps buffers without a step larger than its crossfade allows and
  loops its region seamlessly; the supervisor follows handoffs; a save refuses a file
  changed on disk.
- Tests assert function, never content, so no song's values are pinned.
