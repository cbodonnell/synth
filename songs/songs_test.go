package songs

import (
	"testing"

	"github.com/cbodonnell/synth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEveryArrangementIsValid(t *testing.T) {
	for _, a := range Arrangements() {
		assert.NoError(t, a.Validate(), a.ID)
	}
}

func TestEverySongAndSoundRenders(t *testing.T) {
	seen := map[string]bool{}
	for _, s := range append(Songs(), Sounds()...) {
		require.Falsef(t, seen[s.ID], "%s: duplicate ID", s.ID)
		seen[s.ID] = true
		require.NoError(t, synth.Validate(s), s.ID)

		var out synth.Stereo
		if s.Loop {
			l := synth.RenderLoop(s)
			assert.Greater(t, l.Body.Peak(), float32(0), s.ID)
			out = synth.Concat(l.Intro, l.Body)
		} else {
			out = synth.Render(s)
			assert.Greater(t, out.Peak(), float32(0), s.ID)
		}
		if s.Target != 0 {
			lufs, _ := synth.Loudness(out)
			assert.InDeltaf(t, s.Target, lufs, 1, "%s: loudness", s.ID)
		}
	}
}
