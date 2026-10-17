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

A song is parts playing notes through patches:

```go
riff := synth.Phrase{Step: 0.18, Strum: 0.007}.Notes(
	"Bb2 - F3 - Bb2 - | . Bb3,Db4,F4 . Bb3,Db4,F4",
)
song := synth.Song{
	ID:     "riff",
	Parts:  []synth.Part{{Name: "keys", Patch: patches.Harpsichord(), Notes: riff,
		Sends: []synth.Send{{Bus: "room", Level: 0.2}}}},
	Buses:  []synth.Bus{{Name: "room", Effects: []synth.Effect{synth.Reverb{Size: 0.5}}}},
	Length: 1.8, Loop: true,
}
```

Sound design covers band-limited saw and pulse (with PWM), triangle, sine with FM, noise,
and Karplus-Strong plucked strings; unison with detune and stereo spread; ADSR envelopes
per patch and per oscillator; a resonant multimode filter with its own envelope, key
tracking, and velocity; LFOs on pitch, cutoff, level, pulse width, and pan; pitch sweeps
and slides, including bent plucked strings; drive; and EQ, chorus, phaser, delay, reverb,
compression, kick-keyed sidechain pumping, automated filter sweeps, tape stops, a
look-ahead limiter, and bitcrush as inserts, send buses, group buses, or master effects.

```sh
make play ID=goldenbrown   # render, print the envelope, play
make score ID=goldenbrown  # piano roll
make demo                  # the Ebiten demo
```
