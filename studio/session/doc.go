package session

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/cbodonnell/synth"
)

// Doc is an arrangement as the studio edits it: where its literal is in the Go source,
// the arrangement as last saved, and the edits since, which undo and redo.
type Doc struct {
	Path string
	f    *file

	saved, now synth.Arrangement
	undo, redo []synth.Arrangement
	// Edits are being held together as one step, and the step has been taken.
	held, stepped bool
	dirty         bool
}

// UndoDepth is how many edits can be undone.
const UndoDepth = 200

// ErrChanged refuses a save over a file that changed on disk after it was loaded.
var ErrChanged = errors.New("changed on disk since it was loaded")

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
			f, err := parse(path, src, a.ID)
			if err != nil {
				return nil, err
			}
			if f != nil {
				found = append(found, &Doc{Path: path, f: f})
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
	if err := d.f.check(d.f.lit, reflect.ValueOf(a), a.ID); err != nil {
		return nil, fmt.Errorf("%s isn't the source of this build: %v", d.Path, err)
	}
	d.saved, d.now = a, a
	return d, nil
}

// Now is the arrangement as edited. It's shared: edit it with Edit.
func (d *Doc) Now() synth.Arrangement { return d.now }

// Dirty reports whether there are edits since the last save.
func (d *Doc) Dirty() bool { return d.dirty }

// Edit applies f to a copy of the arrangement, as one step to undo.
func (d *Doc) Edit(f func(a *synth.Arrangement)) {
	next := clone(reflect.ValueOf(d.now)).Interface().(synth.Arrangement)
	f(&next)
	if reflect.DeepEqual(next, d.now) {
		return
	}
	if !d.held || !d.stepped {
		d.undo = append(d.undo, d.now)
		if len(d.undo) > UndoDepth {
			d.undo = d.undo[len(d.undo)-UndoDepth:]
		}
		d.stepped = true
	}
	d.redo = nil
	d.set(next)
}

// History is how many edits there are to undo and to redo.
func (d *Doc) History() (undo, redo int) { return len(d.undo), len(d.redo) }

// Hold makes the edits until Release one step, as a drag or typing into a field is.
func (d *Doc) Hold() { d.held, d.stepped = true, false }

func (d *Doc) Release() { d.held = false }

func (d *Doc) Undo() bool {
	if len(d.undo) == 0 {
		return false
	}
	d.redo = append(d.redo, d.now)
	d.set(d.undo[len(d.undo)-1])
	d.undo = d.undo[:len(d.undo)-1]
	d.stepped = false
	return true
}

func (d *Doc) Redo() bool {
	if len(d.redo) == 0 {
		return false
	}
	d.undo = append(d.undo, d.now)
	d.set(d.redo[len(d.redo)-1])
	d.redo = d.redo[:len(d.redo)-1]
	d.stepped = false
	return true
}

func (d *Doc) set(a synth.Arrangement) {
	d.now = a
	d.dirty = !reflect.DeepEqual(d.now, d.saved)
}

// Source is the file as a save would write it: the source as loaded, with only what
// has been edited since spliced in, read back to the arrangement as edited.
func (d *Doc) Source() ([]byte, error) {
	if !d.dirty {
		return d.f.src, nil
	}
	w := &writer{file: d.f}
	w.diff(d.f.lit, reflect.ValueOf(d.saved), reflect.ValueOf(d.now), true)
	out, err := w.apply()
	if err != nil {
		return nil, fmt.Errorf("%s: %v", d.Path, err)
	}
	back, err := parse(d.Path, out, d.now.ID)
	if err == nil && back == nil {
		err = errors.New("the arrangement is gone")
	}
	if err == nil {
		err = back.check(back.lit, reflect.ValueOf(d.now), d.now.ID)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: the edit doesn't read back: %v", d.Path, err)
	}
	return out, nil
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
	if !bytes.Equal(disk, d.f.src) {
		return fmt.Errorf("%s: %w", d.Path, ErrChanged)
	}
	out, err := d.Source()
	if err != nil {
		return err
	}
	if err := replace(d.Path, out); err != nil {
		return err
	}
	f, err := parse(d.Path, out, d.now.ID)
	if err != nil {
		return err
	}
	d.f, d.saved, d.dirty = f, d.now, false
	return nil
}

// replace writes the file by rename, so it's never seen half written.
func replace(path string, src []byte) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(src); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), info.Mode().Perm()); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// reread finds the literal again after another doc's save changed its file.
func (d *Doc) reread() error {
	src, err := os.ReadFile(d.Path)
	if err != nil {
		return err
	}
	f, err := parse(d.Path, src, d.saved.ID)
	if err == nil && f == nil {
		err = errors.New("the arrangement is gone")
	}
	if err == nil {
		err = f.check(f.lit, reflect.ValueOf(d.saved), d.saved.ID)
	}
	if err != nil {
		return fmt.Errorf("%s: %v", d.Path, err)
	}
	d.f = f
	return nil
}

// Custom reports whether the value at path is written as code rather than a literal,
// so the studio shows it but doesn't edit it. A path is field names and indexes into
// the arrangement as edited: Custom("Tracks", 2, "Sends"). Anything the studio added is
// literal.
func (d *Doc) Custom(path ...any) bool {
	var expr ast.Expr = d.f.lit
	saved, now := reflect.ValueOf(d.saved), reflect.ValueOf(d.now)
	for _, step := range path {
		if expr == nil {
			return false
		}
		lit := compositeOf(expr)
		if lit == nil {
			return true
		}
		for saved.Kind() == reflect.Interface || saved.Kind() == reflect.Pointer {
			saved, now = saved.Elem(), now.Elem()
		}
		switch s := step.(type) {
		case string:
			if !keyed(lit) {
				return true
			}
			saved, now = saved.FieldByName(s), now.FieldByName(s)
			expr = nil
			if kv := field(lit, s); kv != nil {
				expr = kv.Value
			}
		case int:
			if s >= now.Len() {
				return false
			}
			now = now.Index(s)
			j := -1
			if key, ok := itemKey(now); ok {
				for i := range saved.Len() {
					if k, _ := itemKey(saved.Index(i)); k == key {
						j = i
						break
					}
				}
			} else if s < saved.Len() {
				j = s
			}
			if j < 0 || j >= len(lit.Elts) {
				return false
			}
			saved, expr = saved.Index(j), lit.Elts[j]
		}
	}
	return expr != nil && compositeOf(expr) == nil && !d.f.plain(expr)
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
