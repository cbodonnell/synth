package session

import (
	"errors"
	"io/fs"
	"os"
	"slices"
	"sync"

	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/bake"
)

// Baker keeps a game's bake directory: what in it is stale, and a rebake of that, each
// run on a goroutine.
type Baker struct {
	Dir string

	mu    sync.Mutex
	state BakeState
	runs  int // checks and bakes begun, so only the latest's result is kept
}

type BakeState struct {
	Stale   []string // the IDs a bake would change
	Checked bool
	Baking  bool
	Song    string // the one rendering
	Done    int    // songs rendered, of Total
	Total   int
	Err     error // how the last bake failed
}

func NewBaker(dir string) *Baker { return &Baker{Dir: dir} }

func (b *Baker) State() BakeState {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}

// Check finds what's stale for the library, unless a bake is running, which checks when
// it's done.
func (b *Baker) Check(lib Library) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.state.Baking {
		return
	}
	b.runs++
	run := b.runs
	go func() {
		stale, err := b.stale(lib.All())
		b.mu.Lock()
		defer b.mu.Unlock()
		if run == b.runs {
			b.state.Stale, b.state.Checked, b.state.Err = stale, err == nil, err
		}
	}()
}

func (b *Baker) stale(songs []synth.Song) ([]string, error) {
	stale, err := bake.Stale(os.DirFS(b.Dir), songs...)
	if errors.Is(err, fs.ErrNotExist) {
		stale, err = nil, nil
		for _, s := range songs {
			stale = append(stale, s.ID)
		}
	}
	return stale, err
}

// Bake renders the library's stale songs into the directory, leaving the rest.
func (b *Baker) Bake(lib Library) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.state.Baking {
		return
	}
	b.runs++
	b.state = BakeState{Stale: b.state.Stale, Checked: b.state.Checked, Baking: true}
	stale := b.state.Stale
	go func() {
		songs := lib.All()
		total := 0
		for _, s := range songs {
			if slices.Contains(stale, s.ID) {
				total++
			}
		}
		b.mu.Lock()
		b.state.Total = total
		b.mu.Unlock()
		err := bake.BakeEach(b.Dir, func(id string) {
			b.mu.Lock()
			defer b.mu.Unlock()
			if b.state.Song != "" {
				b.state.Done++
			}
			b.state.Song = id
		}, songs...)
		stale, serr := b.stale(songs)
		b.mu.Lock()
		defer b.mu.Unlock()
		b.state = BakeState{Stale: stale, Checked: serr == nil, Err: errors.Join(err, serr)}
	}()
}

// Saved is the library as its source has it now: the build's, with the docs' saved
// arrangements in place of the ones they were loaded from.
func (s *Docs) Saved(lib Library) Library {
	if s == nil {
		return lib
	}
	lib.Arrangements = slices.Clone(lib.Arrangements)
	for i, a := range lib.Arrangements {
		if d, ok := s.open[a.ID]; ok {
			lib.Arrangements[i] = d.lit.Value
		}
	}
	return lib
}
