package songs

import (
	"testing"

	"github.com/cbodonnell/synth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEverySongAndSoundRenders(t *testing.T) {
	seen := map[string]bool{}
	for _, s := range append(Songs(), Sounds()...) {
		require.Falsef(t, seen[s.ID], "%s: duplicate ID", s.ID)
		seen[s.ID] = true
		require.NoError(t, synth.Validate(s), s.ID)

		if s.Loop {
			l := synth.RenderLoop(s)
			assert.Greater(t, l.Body.Peak(), float32(0), s.ID)
			continue
		}
		assert.Greater(t, synth.Render(s).Peak(), float32(0), s.ID)
	}
}
