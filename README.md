# synth

Music and sound effects for [Ebiten](https://ebitengine.org) games, synthesized at
runtime from plain Go values. No audio files.

```go
p := player.New()
mixer := player.NewMixer(p)
mixer.PlayMusic(songs.GoldenBrown())

// in Update:
if pos, ok := mixer.Beat(); ok && pos.Beat == 0 {
	// on the downbeat
}
mixer.Play(synth.Hit("coin", patches.Coin(), synth.MustKey("E6"), 0.1))
```

A song is an arrangement: tracks of clips played at bars, through patches and buses.

```go
var riff = synth.Arrangement{
	ID: "riff", Tempo: 116, Meter: []int{3},
	Sections: []synth.Section{{Name: "A", Bars: 4}},
	Loop:     "A",
	Tracks: []synth.Track{{
		Name: "keys", Patch: patches.Harpsichord(),
		Sends: []synth.Send{{Bus: "room", Level: -14}},
		Clips: []synth.Clip{{Name: "vamp", Step: 0.5, Strum: 0.007, Bars: []string{
			"Bb2 - F3 - Bb2 -",
			". Bb3,Db4,F4 . Bb3,Db4,F4 . .",
		}}},
		Play: []synth.Play{{Clip: "vamp", Bar: 1, Times: 2}},
	}},
	Buses: []synth.Bus{{Name: "room", Effects: []synth.Effect{synth.Reverb{Size: 0.5}}}},
}

mixer.PlayMusic(riff.Song())
```

Sound design covers band-limited saw and pulse (with PWM), triangle, sine with FM, noise,
and Karplus-Strong plucked strings; unison with detune and stereo spread; ADSR envelopes
per patch and per oscillator; a resonant multimode filter with its own envelope, key
tracking, and velocity; LFOs on pitch, cutoff, level, pulse width, and pan; pitch sweeps
and slides, including bent plucked strings; drive; and EQ, chorus, phaser, delay, reverb,
compression, kick-keyed sidechain pumping, automated filter sweeps, tape stops, a
look-ahead limiter, and bitcrush as inserts, send buses, group buses, or master effects.

A game ships its songs baked, since a full band takes seconds to render (and far more on
WebAssembly). The songs stay the source; a stale bake is caught by `bake.Stale` in a test
and rendered live in the meantime:

```go
bake.Bake("assets/baked", songs...)              // at authoring time; needs oggenc
p := player.New(player.Baked(os.DirFS("assets/baked")))
```

```sh
make play ID=goldenbrown   # render, print the envelope, play
make score ID=goldenbrown  # piano roll
make export ID=goldenbrown # a song as a draft Arrangement
make demo                  # the Ebiten demo
```
