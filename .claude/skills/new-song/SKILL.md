---
name: new-song
description: Write a new original song for synth as a synth.Arrangement in songs/, from the user's description of what they want. Use when the user asks for a new song, track, theme or piece of music.
---

# New song

Write one original song from the user's brief and check it headless. The existing
songs stay out of your context. This file, `arrangement.md` and `skeleton.go` (both
next to this file) cover the API.

## Rules

- Don't read the files in `songs/` or `songs/README.md`. A song file pulls the new song
  toward its style, and you don't need it. The one exception is a technique listed at
  the end of `arrangement.md` that you can't get right otherwise. Then grep that file
  for the few lines you need.
- The song is original: melody, chords and hooks of your own. A genre or an
  artist's style is fine as a direction. Never reproduce an existing tune.
- The doc comment on the arrangement says plainly what the song is: genre, key, tempo,
  instruments, and what each section does. No scene-setting.

## 1. Get the brief

Never start without the user's description of what to create. If the request has none,
ask for it and stop. Ask about these, and let them skip any:

- genre or style, and any reference artists
- mood and energy, and where it plays (a game's menu, a level, a boss, credits)
- length, and whether it loops
- instruments they want or don't want

Then propose a plan in a few lines and wait for the user to approve it or change it:

- ID and name
- key, tempo, meter
- form: each section with its bars and what happens in it, and the loop point
- instruments: a library patch or a new one for each, with what each part plays
- target loudness: -14 for a full mix, around -18 for something sparse and quiet

Only write once the user agrees to the plan.

## 2. Instruments

Run `make patches` and pick from it. Where the library has nothing close, write a new
patch, as `arrangement.md` describes. Library patches go in `patches/` with a doc
comment; one-off patches and small tweaks go in the song's file.

## 3. Write it

1. Copy `skeleton.go` to `songs/<id>.go` and rename `untitled`/`Untitled` throughout.
   Replace everything else. The skeleton only shows the shape.
2. Add the var to `Arrangements()` in `songs/songs.go`, after the last song.
3. Write the sections in order. Keep the arrangement literal: values written in place,
   not helper calls or shared vars, so the studio can edit them. Patch tweaks as small
   funcs are the exception.

Things to get right:

- Each section should be different from the one before: drop or add parts, change the
  register, open a filter. Two sections with the same parts at the same level sound like
  one long section.
- A loop's last bar should lead back into its first.
- Keep the bass and the kick from fighting: put them on different steps, or key a
  `Sidechain` on the bass to the kick.
- Give each part its own register and its own place in the stereo field (pan).

## 4. Check it

Run each of these, read what it prints, and fix what's wrong before moving on.

- `go test ./songs/`: validates every bar's length, overlaps and note names, renders the
  song, and holds it within 1 LU of `Target`.
- `make score ID=<id>`: the piano roll. Check that parts enter and leave where the plan
  says, that lines sit in their registers, and that no part collides with another.
- `go run ./cmd/synth -id <id> -stems`: each part's level while it's sounding. The lead
  and the kick should be loudest. Pads, hats and texture sit 6 to 12 LU under them. A
  part far below the rest is inaudible.
- `make play ID=<id> DUMP=1`: the whole mix's LUFS, peak and an envelope over time. The
  envelope should rise and fall with the form. A flat line means the sections don't
  differ enough.
- `make vet`.

If a new patch went into `patches/`, run `go test ./...` too.

## 5. Report

Tell the user:

- what the song is: form, key, tempo and instruments
- any new patches and where they live
- the numbers: LUFS against target, and the stems
- what to listen for, since you can't hear it: the balance of any part the stems left
  in doubt, how a new patch sounds, and whether the loop seam is smooth

They listen with `make play ID=<id>` or in the studio with `make edit ID=<id>`.
