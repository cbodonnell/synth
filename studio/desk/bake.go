package desk

import (
	"fmt"
	"strings"
)

// checkBake finds what's stale in the game's bake, as the songs are saved.
func (d *desk) checkBake() {
	if d.baker != nil {
		d.baker.Check(d.docs.Saved(d.lib))
	}
}

// bake rebakes the songs whose bake is stale, as they're saved.
func (d *desk) bake() {
	if d.baker != nil {
		d.baker.Bake(d.docs.Saved(d.lib))
	}
}

// bakeStatus is a line on the bake for the status bar, while there's news.
func (d *desk) bakeStatus() string {
	if d.baker == nil {
		return ""
	}
	s := d.baker.State()
	switch {
	case s.Baking && s.Song != "":
		return fmt.Sprintf("baking %s (%d of %d)…", s.Song, s.Done+1, s.Total)
	case s.Baking:
		return "baking…"
	case s.Err != nil:
		first, _, _ := strings.Cut(s.Err.Error(), "\n")
		return "bake failed: " + first
	case len(s.Stale) > 0:
		return fmt.Sprintf("%d stale in %s", len(s.Stale), d.baker.Dir)
	}
	return ""
}
