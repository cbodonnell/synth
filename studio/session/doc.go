package session

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/cbodonnell/aac/edit"
	"github.com/cbodonnell/aac/golit"
	"github.com/cbodonnell/synth"
)

// Doc is an arrangement as the studio edits it: its literal in the Go source, read as it
// was last saved, and the edits since, which undo and redo.
type Doc struct {
	Path string
	lit  *golit.Lit[synth.Arrangement]
	src  []byte

	now synth.Arrangement
	// The values of now written as code, by golit's path in now: the source's code, less
	// what has been edited or made literal, and the calls the studio has written.
	spell map[string]golit.Spelling
	hist  *edit.History[state]
	// Edits are being held together as one step, and the step has been taken.
	held, stepped bool
	dirty         bool
}

// state is an arrangement as the undo history keeps it.
type state struct {
	a     synth.Arrangement
	spell map[string]golit.Spelling
}

func (s state) Clone() state {
	return state{clone(reflect.ValueOf(s.a)).Interface().(synth.Arrangement), maps.Clone(s.spell)}
}

// UndoDepth is how many edits can be undone.
const UndoDepth = edit.MaxHistory

// ErrChanged refuses a save over a file that changed on disk after it was loaded.
var ErrChanged = edit.ErrChanged

// Load finds the Go file in dirs holding a's literal (a synth.Arrangement with its ID),
// which must write a as it is: the source a was built from.
func Load(a synth.Arrangement, dirs []string) (*Doc, error) {
	var found []*Doc
	for _, dir := range dirs {
		entries, _ := os.ReadDir(dir)
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}
			path := filepath.Join(dir, name)
			src, err := os.ReadFile(path)
			if err != nil || !bytes.Contains(src, []byte(strconv.Quote(a.ID))) {
				continue
			}
			f, err := golit.Parse(src)
			if err != nil {
				return nil, fmt.Errorf("%s: %v", path, err)
			}
			switch n := len(arrangements(f, a.ID)); {
			case n == 1:
				found = append(found, &Doc{Path: path, src: src})
			case n > 1:
				return nil, fmt.Errorf("%s: %d arrangements with ID %q", path, n, a.ID)
			}
		}
	}
	switch len(found) {
	case 0:
		return nil, fmt.Errorf("no synth.Arrangement literal with ID %q in its source", a.ID)
	case 1:
	default:
		return nil, fmt.Errorf("%d arrangements with ID %q in its source", len(found), a.ID)
	}
	d := found[0]
	if err := d.read(d.src, a); err != nil {
		return nil, err
	}
	if !reflect.DeepEqual(d.lit.Value, a) {
		return nil, fmt.Errorf("%s isn't the source of this build", d.Path)
	}
	d.now, d.spell = a, codeOf(d.lit.Spelled)
	d.hist = edit.NewHistory(state{a, d.spell})
	return d, nil
}

// read reads the literal from src, what isn't a literal taken from built.
func (d *Doc) read(src []byte, built synth.Arrangement) error {
	f, err := golit.Parse(src)
	if err != nil {
		return fmt.Errorf("%s: %v", d.Path, err)
	}
	lit, err := golit.ReadAt(f, findArrangement(built.ID), built, spellings)
	if err != nil {
		return fmt.Errorf("%s: %v", d.Path, err)
	}
	d.lit, d.src = lit, src
	return nil
}

// Now is the arrangement as edited. It's shared: edit it with Edit.
func (d *Doc) Now() synth.Arrangement { return d.now }

// Dirty reports whether there are edits since the last save.
func (d *Doc) Dirty() bool { return d.dirty }

// Edit applies f to a copy of the arrangement, as one step to undo, and reports
// whether it changed anything.
func (d *Doc) Edit(f func(a *synth.Arrangement)) bool {
	next := d.copy()
	f(&next)
	spell, _ := carry(d.now, next, d.spell)
	return d.take(next, spell)
}

// respell applies f as Edit does, then spells the values at golit's paths in the result
// as code, or as literals where the spelling is nil.
func (d *Doc) respell(f func(a *synth.Arrangement), spell map[string]*golit.Spelling) bool {
	next := d.copy()
	f(&next)
	out, _ := carry(d.now, next, d.spell)
	for p, s := range spell {
		maps.DeleteFunc(out, func(q string, _ golit.Spelling) bool { return golit.Under(q, p) || golit.Under(p, q) })
		if s != nil {
			out[p] = *s
		}
	}
	return d.take(next, out)
}

// ErrCode refuses an edit that would write over code.
var ErrCode = errors.New("it's written as code: edit it there")

// Arrange applies f to a copy of the arrangement as Edit does, when f says it can and
// the result validates and leaves every value written as code as it is.
func (d *Doc) Arrange(f func(a *synth.Arrangement) error) error {
	next := d.copy()
	if err := f(&next); err != nil {
		return err
	}
	if reflect.DeepEqual(next, d.now) {
		return nil
	}
	if err := next.Validate(); err != nil {
		return err
	}
	spell, overwritten := carry(d.now, next, d.spell)
	if overwritten {
		return ErrCode
	}
	d.take(next, spell)
	return nil
}

func (d *Doc) copy() synth.Arrangement { return state{a: d.now}.Clone().a }

// take makes next the arrangement, spelled so, as a step to undo, reporting whether it
// changed.
func (d *Doc) take(next synth.Arrangement, spell map[string]golit.Spelling) bool {
	if reflect.DeepEqual(next, d.now) && maps.EqualFunc(spell, d.spell, sameSpelling) {
		return false
	}
	if d.held && d.stepped {
		d.hist.Amend(state{next, spell})
	} else {
		d.hist.Push(state{next, spell})
		d.stepped = true
	}
	d.set(next, spell)
	return true
}

// CanUndo and CanRedo report whether there's an edit to undo and to redo.
func (d *Doc) CanUndo() bool { return d.hist.CanUndo() }

func (d *Doc) CanRedo() bool { return d.hist.CanRedo() }

// Hold makes the edits until Release one step, as a drag or typing into a field is.
func (d *Doc) Hold() { d.held, d.stepped = true, false }

func (d *Doc) Release() { d.held = false }

func (d *Doc) Undo() bool {
	s, ok := d.hist.Undo()
	if ok {
		d.set(s.a, s.spell)
		d.stepped = false
	}
	return ok
}

func (d *Doc) Redo() bool {
	s, ok := d.hist.Redo()
	if ok {
		d.set(s.a, s.spell)
		d.stepped = false
	}
	return ok
}

func (d *Doc) set(a synth.Arrangement, spell map[string]golit.Spelling) {
	d.now, d.spell = a, spell
	d.dirty = !reflect.DeepEqual(d.now, d.lit.Value) || !maps.EqualFunc(d.spell, codeOf(d.lit.Spelled), sameSpelling)
}

// Source is the file as a save would write it: the source as loaded, with only what
// has been edited since spliced in, read back to the arrangement as edited.
func (d *Doc) Source() ([]byte, error) {
	out, _, err := d.write()
	return out, err
}

func (d *Doc) write() ([]byte, *golit.Lit[synth.Arrangement], error) {
	if !d.dirty {
		return d.src, d.lit, nil
	}
	c := compare(d.lit.Value, d.now, d.lit.Spelled, false)
	spelled := maps.Clone(d.spell)
	for src := range codeOf(d.lit.Spelled) {
		if p, f := c.find(src); f != gone {
			if _, ok := d.spell[p]; !ok {
				spelled[p] = golit.Spelling{Kind: golit.Plain}
			}
		}
	}
	out, back, err := d.lit.Write(d.now, spelled, c.origins)
	if err == nil {
		var fixed []byte
		if fixed, err = patchesImport(out); err == nil && !bytes.Equal(fixed, out) {
			out = fixed
			var f *golit.File
			if f, err = golit.Parse(out); err == nil {
				back, err = golit.ReadAt(f, findArrangement(d.now.ID), d.now, spellings)
			}
		}
	}
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %v", d.Path, err)
	}
	return out, back, nil
}

// Save writes the edits into the file, unless it has changed on disk since it was
// loaded.
func (d *Doc) Save() error {
	if !d.dirty {
		return nil
	}
	disk, err := os.ReadFile(d.Path)
	if err != nil {
		return err
	}
	if !bytes.Equal(disk, d.src) {
		return fmt.Errorf("%s: %w", d.Path, ErrChanged)
	}
	out, back, err := d.write()
	if err != nil {
		return err
	}
	if err := edit.WriteFile(d.Path, out); err != nil {
		return err
	}
	d.lit, d.src, d.dirty = back, out, false
	return nil
}

// reread finds the literal again after another doc's save changed its file.
func (d *Doc) reread() error {
	src, err := os.ReadFile(d.Path)
	if err != nil {
		return err
	}
	saved := d.lit.Value
	if err := d.read(src, saved); err != nil {
		return err
	}
	if !reflect.DeepEqual(d.lit.Value, saved) {
		return fmt.Errorf("%s: the arrangement changed", d.Path)
	}
	return nil
}

// Custom reports whether the value at path is written as code rather than a literal,
// so the studio shows it but doesn't edit it. A path is field names and indexes into
// the arrangement as edited: Custom("Tracks", 2, "Sends"). What the studio added is
// literal, but for a patch set by its call.
func (d *Doc) Custom(path ...any) bool {
	for _, p := range litPath(path) {
		if _, ok := d.spell[p]; ok {
			return true
		}
	}
	return false
}

// clone copies v deeply, so an edit to the copy never reaches v.
func clone(v reflect.Value) reflect.Value {
	switch v.Kind() {
	case reflect.Slice:
		if v.IsNil() {
			return v
		}
		c := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
		for i := range v.Len() {
			c.Index(i).Set(clone(v.Index(i)))
		}
		return c
	case reflect.Struct:
		c := reflect.New(v.Type()).Elem()
		c.Set(v)
		for i := range v.NumField() {
			if c.Field(i).CanSet() {
				c.Field(i).Set(clone(v.Field(i)))
			}
		}
		return c
	case reflect.Interface:
		if v.IsNil() {
			return v
		}
		c := reflect.New(v.Type()).Elem()
		c.Set(clone(v.Elem()))
		return c
	case reflect.Pointer:
		if v.IsNil() {
			return v
		}
		c := reflect.New(v.Type().Elem())
		c.Elem().Set(clone(v.Elem()))
		return c
	}
	return v
}

// Docs are the arrangements the studio has open, each loaded from its source the first
// time it's asked for.
type Docs struct {
	dirs []string
	open map[string]*Doc
}

// NewDocs finds arrangements' source in dirs.
func NewDocs(dirs []string) *Docs { return &Docs{dirs: dirs, open: map[string]*Doc{}} }

// Open is the doc of the arrangement, as edited if it's been opened before.
func (s *Docs) Open(a synth.Arrangement) (*Doc, error) {
	if d, ok := s.open[a.ID]; ok {
		return d, nil
	}
	d, err := Load(a, s.dirs)
	if err != nil {
		return nil, err
	}
	s.open[a.ID] = d
	return d, nil
}

// Save saves the doc, and finds the others in its file again.
func (s *Docs) Save(d *Doc) error {
	if err := d.Save(); err != nil {
		return err
	}
	var errs []error
	for _, o := range s.open {
		if o != d && o.Path == d.Path {
			errs = append(errs, o.reread())
		}
	}
	return errors.Join(errs...)
}

// SaveAll saves every doc with edits.
func (s *Docs) SaveAll() error {
	var errs []error
	for _, id := range s.Dirty() {
		errs = append(errs, s.Save(s.open[id]))
	}
	return errors.Join(errs...)
}

// Dirty is the IDs of the docs with edits not yet saved.
func (s *Docs) Dirty() []string {
	var ids []string
	for id, d := range s.open {
		if d.dirty {
			ids = append(ids, id)
		}
	}
	return ids
}
