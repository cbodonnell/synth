package session

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/constant"
	"go/format"
	"go/token"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/cbodonnell/aac/golit"
	"github.com/cbodonnell/synth"
)

// synth's effects, which an interface's literal names.
var effects = []reflect.Type{
	reflect.TypeFor[synth.EQ](), reflect.TypeFor[synth.Drive](), reflect.TypeFor[synth.Chorus](),
	reflect.TypeFor[synth.Phaser](), reflect.TypeFor[synth.Delay](), reflect.TypeFor[synth.Reverb](),
	reflect.TypeFor[synth.Compressor](), reflect.TypeFor[synth.Crush](), reflect.TypeFor[synth.Limiter](),
	reflect.TypeFor[synth.Sidechain](), reflect.TypeFor[synth.TapeStop](), reflect.TypeFor[synth.AutoFilter](),
	reflect.TypeFor[synth.Fader](), reflect.TypeFor[synth.Panner](), reflect.TypeFor[synth.Normalize](),
}

// spellings are how an arrangement's source spells what golit can't alone: synth's named
// constants and its effects. Anything else that isn't a literal (a patch's call, a game's
// var, arithmetic) is code.
var spellings = func() golit.Spellings {
	sp := golit.Spellings{Consts: map[reflect.Type]golit.Consts{}, Types: map[string]reflect.Type{}, Inline: true, Declared: true, Hug: true}
	for t, names := range enums {
		var b strings.Builder
		b.WriteString("package synth\n\nconst (\n")
		for i, name := range names {
			fmt.Fprintf(&b, "\t%s %s = %d\n", name, t.Name(), i)
		}
		b.WriteString(")\n")
		c, err := golit.ReadConsts(t, []byte(b.String()))
		if err != nil {
			panic(err)
		}
		sp.Consts[t] = c
	}
	for _, t := range effects {
		sp.Types["synth."+t.Name()] = t
	}
	return sp
}()

// findArrangement finds the synth.Arrangement literal with the ID in a file.
func findArrangement(id string) func(*golit.File) (ast.Expr, bool) {
	return func(f *golit.File) (ast.Expr, bool) {
		lits := arrangements(f, id)
		if len(lits) != 1 {
			return nil, false
		}
		return lits[0], true
	}
}

// arrangements are the synth.Arrangement literals with the ID in a file.
func arrangements(f *golit.File, id string) []*ast.CompositeLit {
	name := "Arrangement"
	if f.AST.Name.Name != "synth" {
		pkg := ""
		for _, imp := range f.AST.Imports {
			if imp.Path.Value == strconv.Quote(Synth) {
				pkg = "synth"
				if imp.Name != nil {
					pkg = imp.Name.Name
				}
			}
		}
		if pkg == "" {
			return nil
		}
		name = pkg + ".Arrangement"
	}
	var found []*ast.CompositeLit
	ast.Inspect(f.AST, func(n ast.Node) bool {
		lit, ok := n.(*ast.CompositeLit)
		if !ok || lit.Type == nil || f.Source(lit.Type) != name {
			return true
		}
		for _, e := range lit.Elts {
			kv, ok := e.(*ast.KeyValueExpr)
			if k, isIdent := kv.Key.(*ast.Ident); !ok || !isIdent || k.Name != "ID" {
				continue
			}
			if b, ok := kv.Value.(*ast.BasicLit); ok && b.Kind == token.STRING {
				if v := constant.MakeFromLiteral(b.Value, b.Kind, 0); constant.StringVal(v) == id {
					found = append(found, lit)
				}
			}
		}
		return true
	})
	return found
}

// code reports whether a spelling is code the studio doesn't rewrite.
func code(s golit.Spelling) bool {
	return s.Kind == golit.Var || s.Kind == golit.Call || s.Kind == golit.Opaque
}

// itemKey is what names an item in a list: a Name, or a Send's Bus. Items without one
// are matched by being equal.
func itemKey(v reflect.Value) (string, bool) {
	for v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return "", false
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return "", false
	}
	if v.Type() == reflect.TypeFor[synth.Send]() {
		return v.FieldByName("Bus").String(), true
	}
	if f := v.FieldByName("Name"); f.IsValid() && f.Kind() == reflect.String {
		return f.String(), true
	}
	return "", false
}

func same(a, b reflect.Value) bool {
	ka, oka := itemKey(a)
	kb, okb := itemKey(b)
	if oka && okb {
		return ka == kb
	}
	return reflect.DeepEqual(a.Interface(), b.Interface())
}

// matched lists only items of composites: a list of strings or numbers (a clip's bars)
// is edited by index.
func matched(t reflect.Type) bool {
	switch t.Elem().Kind() {
	case reflect.Struct, reflect.Interface, reflect.Pointer:
		return true
	}
	return false
}

// match is, for each of now's items, the index of old's it was, or -1 for one added:
// matched by name or by being equal, in order, then those moved, and those left between
// two matches paired off.
func match(old, now reflect.Value) []int {
	n, m := old.Len(), now.Len()
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if same(old.Index(i), now.Index(j)) {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else {
				lcs[i][j] = max(lcs[i+1][j], lcs[i][j+1])
			}
		}
	}
	out := make([]int, m)
	for j := range out {
		out[j] = -1
	}
	claimed := make([]bool, n)
	type pair struct{ i, j int }
	var anchors []pair
	for i, j := 0, 0; i < n && j < m; {
		switch {
		case same(old.Index(i), now.Index(j)) && lcs[i][j] == lcs[i+1][j+1]+1:
			out[j], claimed[i] = i, true
			anchors = append(anchors, pair{i, j})
			i, j = i+1, j+1
		case lcs[i+1][j] >= lcs[i][j+1]:
			i++
		default:
			j++
		}
	}
	for j := range out {
		for i := 0; out[j] < 0 && i < n; i++ {
			if !claimed[i] && same(old.Index(i), now.Index(j)) {
				out[j], claimed[i] = i, true
			}
		}
	}
	anchors = append(anchors, pair{n, m})
	i0, j0 := 0, 0
	for _, a := range anchors {
		i := i0
		for j := j0; j < a.j; j++ {
			for i < a.i && claimed[i] {
				i++
			}
			if out[j] < 0 && i < a.i {
				out[j], claimed[i] = i, true
			}
		}
		i0, j0 = a.i+1, a.j+1
	}
	return out
}

// changes walks what changed from old (at src) to now (at path): each list's origins,
// for golit, and where each of old's values went.
type changes struct {
	spelled map[string]golit.Spelling
	origins map[string][]int
	at      map[string]string // the path in now of each of old's values walked
	same    map[string]bool   // old's values that are as they were, all beneath them too
	// Only the values with a spelling beneath them are walked, when set.
	only bool
}

func (c *changes) walk(old, now reflect.Value, src, path string) {
	if c.only && !c.below(src) {
		return
	}
	c.at[src] = path
	if reflect.DeepEqual(old.Interface(), now.Interface()) {
		c.same[src] = true
		return
	}
	if s, ok := c.spelled[src]; ok && code(s) {
		return
	}
	switch old.Kind() {
	case reflect.Interface, reflect.Pointer:
		if !old.IsNil() && !now.IsNil() && old.Elem().Type() == now.Elem().Type() {
			c.walk(old.Elem(), now.Elem(), src, path)
		}
	case reflect.Struct:
		for i := range old.NumField() {
			if f := old.Type().Field(i); f.IsExported() {
				c.walk(old.Field(i), now.Field(i), join(src, f.Name), join(path, f.Name))
			}
		}
	case reflect.Slice:
		if !matched(old.Type()) {
			for i := range min(old.Len(), now.Len()) {
				c.walk(old.Index(i), now.Index(i), fmt.Sprintf("%s[%d]", src, i), fmt.Sprintf("%s[%d]", path, i))
			}
			return
		}
		at := match(old, now)
		c.origins[path] = at
		for j, i := range at {
			if i >= 0 {
				c.walk(old.Index(i), now.Index(j), fmt.Sprintf("%s[%d]", src, i), fmt.Sprintf("%s[%d]", path, j))
			}
		}
	}
}

func (c *changes) below(src string) bool {
	for p := range c.spelled {
		if golit.Under(p, src) {
			return true
		}
	}
	return false
}

// Where a value of old went in now.
type fate int

const (
	kept fate = iota
	changed
	gone
)

// find is where the value at src went in now.
func (c *changes) find(src string) (string, fate) {
	for p := src; ; p = parent(p) {
		if n, ok := c.at[p]; ok {
			switch {
			case c.same[p]:
				return n + src[len(p):], kept
			case p == src:
				return n, changed
			}
			return "", gone
		}
		if p == "" {
			return "", gone
		}
	}
}

// parent is the path of the value holding the one at p.
func parent(p string) string {
	if i := strings.LastIndexAny(p, ".[("); i >= 0 {
		return p[:i]
	}
	return ""
}

// join is a field's path under path, as golit writes paths.
func join(path, field string) string {
	if path == "" {
		return field
	}
	return path + "." + field
}

// compare walks what changed from old, spelled so, to now.
func compare(old, now synth.Arrangement, spelled map[string]golit.Spelling, only bool) *changes {
	c := &changes{spelled: spelled, origins: map[string][]int{}, at: map[string]string{}, same: map[string]bool{}, only: only}
	c.walk(reflect.ValueOf(old), reflect.ValueOf(now), "", "")
	return c
}

// carry follows spellings from prev to next, an edit of it. A spelling whose value the
// edit changed is dropped, so the value is written plainly, and reported.
func carry(prev, next synth.Arrangement, spelled map[string]golit.Spelling) (map[string]golit.Spelling, bool) {
	c := compare(prev, next, spelled, true)
	out := map[string]golit.Spelling{}
	overwritten := false
	for p, s := range spelled {
		switch n, f := c.find(p); f {
		case kept:
			out[n] = s
		case changed:
			overwritten = true
		}
	}
	return out, overwritten
}

// codeOf is the spellings of a literal that are code.
func codeOf(spelled map[string]golit.Spelling) map[string]golit.Spelling {
	out := map[string]golit.Spelling{}
	for p, s := range spelled {
		if code(s) && !strings.Contains(p, "(") {
			out[p] = s
		}
	}
	return out
}

func sameSpelling(a, b golit.Spelling) bool { return a.Kind == b.Kind && a.Source == b.Source }

// litPath is golit's path of each step of a path into an arrangement: field names and
// indexes.
func litPath(path []any) []string {
	var out []string
	p := ""
	for _, step := range path {
		switch s := step.(type) {
		case string:
			p = join(p, s)
		case int:
			p += fmt.Sprintf("[%d]", s)
		}
		out = append(out, p)
	}
	return out
}

// patchesImport imports synth's patches into a file that calls them, and drops the
// import from one that no longer does, as a patch set from the browser and undone leaves
// it.
func patchesImport(src []byte) ([]byte, error) {
	f, err := golit.Parse(src)
	if err != nil {
		return nil, err
	}
	var spec *ast.ImportSpec
	var decl *ast.GenDecl
	for _, dcl := range f.AST.Decls {
		if g, ok := dcl.(*ast.GenDecl); ok && g.Tok == token.IMPORT {
			for _, s := range g.Specs {
				if is := s.(*ast.ImportSpec); is.Path.Value == strconv.Quote(Synth+"/patches") {
					spec, decl = is, g
				}
			}
		}
	}
	name := "patches"
	if spec != nil && spec.Name != nil {
		name = spec.Name.Name
	}
	used := false
	ast.Inspect(f.AST, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok {
			if x, ok := sel.X.(*ast.Ident); ok && x.Name == name {
				used = true
			}
		}
		return !used
	})
	switch {
	case spec == nil && used:
		return golit.AddImports(src, map[string]string{Synth + "/patches": "patches"})
	case spec != nil && !used:
		var n ast.Node = spec
		if !decl.Lparen.IsValid() {
			n = decl
		}
		from, to := f.Fset.Position(n.Pos()).Offset, f.Fset.Position(n.End()).Offset
		from = bytes.LastIndexByte(src[:from], '\n') + 1
		if i := bytes.IndexByte(src[to:], '\n'); i >= 0 {
			to += i + 1
		} else {
			to = len(src)
		}
		return format.Source(slices.Concat(src[:from], src[to:]))
	}
	return src, nil
}
