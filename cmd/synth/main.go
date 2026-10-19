// synth is the authoring loop over the example songs and sounds; see package studio.
package main

import (
	"github.com/cbodonnell/synth/songs"
	"github.com/cbodonnell/synth/studio"
)

func main() {
	studio.Main(studio.Library{Arrangements: songs.Arrangements(), Sounds: songs.Sounds()})
}
