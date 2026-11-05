package session

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"reflect"
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
	reflect.TypeFor[synth.Fader](), reflect.TypeFor[synth.Panner](),
}

// spellings are how an arrangement's source spells what golit can't alone: synth's named
// constants and its effects. Anything else that isn't a literal (a patch's call, a game's
// var, arithmetic) is code.
var spellings = func() golit.Spellings {
	sp := golit.Spellings{Consts: map[reflect.Type]golit.Consts{}, Types: map[string]reflect.Type{}, Inline: true, Declared: true}
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
// matched by name or by being equal, in order, and those between two matches paired off.
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
	gap := func(i0, i1, j0, j1 int) {
		for j := j0; j < j1; j++ {
			out[j] = -1
			if i := i0 + j - j0; i < i1 {
				out[j] = i
			}
		}
	}
	i0, j0 := 0, 0
	for i, j := 0, 0; i < n && j < m; {
		switch {
		case same(old.Index(i), now.Index(j)) && lcs[i][j] == lcs[i+1][j+1]+1:
			gap(i0, i, j0, j)
			out[j] = i
			i, j = i+1, j+1
			i0, j0 = i, j
		case lcs[i+1][j] >= lcs[i][j+1]:
			i++
		default:
			j++
		}
	}
	gap(i0, n, j0, m)
	return out
}

// changes walks what changed from old (the source's, at src) to now (at path): it gives
// golit each list's origins, and reports whether a change writes over code.
type changes struct {
	spelled map[string]golit.Spelling
	origins map[string][]int
	code    bool
}

func (c *changes) walk(old, now reflect.Value, src, path string) {
	if reflect.DeepEqual(old.Interface(), now.Interface()) {
		return
	}
	if s, ok := c.spelled[src]; ok && code(s) {
		c.code = true
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

// join is a field's path under path, as golit writes paths.
func join(path, field string) string {
	if path == "" {
		return field
	}
	return path + "." + field
}

// diff is what changed from saved to now.
func diff(lit *golit.Lit[synth.Arrangement], now synth.Arrangement) *changes {
	c := &changes{spelled: lit.Spelled, origins: map[string][]int{}}
	c.walk(reflect.ValueOf(lit.Value), reflect.ValueOf(now), "", "")
	return c
}

// sourcePath is the source's path for a path into the arrangement as edited (field names
// and indexes, items found by name), and false for one the source doesn't have.
func sourcePath(saved, now reflect.Value, path []any) ([]string, bool) {
	var out []string
	src := ""
	for _, step := range path {
		for saved.Kind() == reflect.Interface || saved.Kind() == reflect.Pointer {
			if saved.IsNil() || now.IsNil() {
				return out, false
			}
			saved, now = saved.Elem(), now.Elem()
		}
		switch s := step.(type) {
		case string:
			if saved.Kind() != reflect.Struct {
				return out, false
			}
			saved, now = saved.FieldByName(s), now.FieldByName(s)
			src = join(src, s)
		case int:
			if saved.Kind() != reflect.Slice || s >= now.Len() {
				return out, false
			}
			i := s
			if matched(saved.Type()) {
				i = match(saved, now)[s]
			}
			if i < 0 || i >= saved.Len() {
				return out, false
			}
			saved, now = saved.Index(i), now.Index(s)
			src += fmt.Sprintf("[%d]", i)
		}
		out = append(out, src)
	}
	return out, true
}
