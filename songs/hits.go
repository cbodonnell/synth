package songs

import (
	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/patches"
)

// Hits are single notes of the instrument library, for hearing a patch alone.
func Hits() []synth.Arrangement {
	return []synth.Arrangement{
		kickSound, snareSound, hihatSound, laserSound, bellSound, epianoSound, bassSound,
		harpsichordSound, coinSound,
	}
}

var kickSound = synth.Arrangement{
	ID:   "kick",
	Name: "kick",
	Tracks: []synth.Track{
		{
			Name:  "kick",
			Patch: patches.Kick(),
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 0.3, Key: 36, Vel: 1}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
}

var snareSound = synth.Arrangement{
	ID:   "snare",
	Name: "snare",
	Tracks: []synth.Track{
		{
			Name:  "snare",
			Patch: patches.Snare(),
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 0.15, Key: 48, Vel: 1}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
}

var hihatSound = synth.Arrangement{
	ID:   "hihat",
	Name: "hihat",
	Tracks: []synth.Track{
		{
			Name:  "hihat",
			Patch: patches.HiHat(),
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 0.04, Key: 72, Vel: 1}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
}

var laserSound = synth.Arrangement{
	ID:   "laser",
	Name: "laser",
	Tracks: []synth.Track{
		{
			Name:  "laser",
			Patch: patches.Laser(),
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 0.2, Key: 72, Vel: 1}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
}

var bellSound = synth.Arrangement{
	ID:   "bell",
	Name: "bell",
	Tracks: []synth.Track{
		{
			Name:  "bell",
			Patch: patches.Bell(),
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 1, Key: 76, Vel: 1}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
}

var epianoSound = synth.Arrangement{
	ID:   "epiano",
	Name: "epiano",
	Tracks: []synth.Track{
		{
			Name:  "epiano",
			Patch: patches.EPiano(),
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 1, Key: 57, Vel: 1}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
}

var bassSound = synth.Arrangement{
	ID:   "bass",
	Name: "bass",
	Tracks: []synth.Track{
		{
			Name:  "bass",
			Patch: patches.Bass(),
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 0.5, Key: 28, Vel: 1}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
}

var harpsichordSound = synth.Arrangement{
	ID:   "harpsichord",
	Name: "harpsichord",
	Tracks: []synth.Track{
		{
			Name:  "harpsichord",
			Patch: patches.Harpsichord(),
			Clips: []synth.Clip{{Name: "a", Notes: []synth.Note{{Dur: 1, Key: 58, Vel: 1}}}},
			Play:  []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
}

var coinSound = synth.Arrangement{
	ID:   "coin",
	Name: "coin",
	Tracks: []synth.Track{
		{
			Name:  "coin",
			Patch: patches.Coin(),
			Clips: []synth.Clip{
				{
					Name:  "a",
					Notes: []synth.Note{{Dur: 0.063, Key: 83, Vel: 0.8}, {At: 0.07, Dur: 0.126, Key: 88, Vel: 0.8}},
				},
			},
			Play: []synth.Play{{Clip: "a", Bar: 1}},
		},
	},
}
