package session

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/constant"
	"go/format"
	"go/parser"
	"go/token"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/cbodonnell/synth"
)

// file is a Go file holding an arrangement literal.
type file struct {
	src  []byte
	fset *token.FileSet
	tok  *token.File
	lit  *ast.CompositeLit
	// The name the file knows synth by; "" inside synth itself.
	synth string
}

// parse finds the synth.Arrangement literal with the ID in src: nil when there is none.
func parse(path string, src []byte, id string) (*file, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, src, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}
	out := &file{src: src, fset: fset, tok: fset.File(f.Pos())}
	if f.Name.Name != "synth" {
		out.synth = "-"
		for _, imp := range f.Imports {
			if imp.Path.Value != strconv.Quote(Synth) {
				continue
			}
			out.synth = "synth"
			if imp.Name != nil {
				out.synth = imp.Name.Name
			}
		}
		if out.synth == "-" {
			return nil, nil
		}
	}
	var found []*ast.CompositeLit
	ast.Inspect(f, func(n ast.Node) bool {
		lit, ok := n.(*ast.CompositeLit)
		if !ok || !out.names(lit.Type, "Arrangement") {
			return true
		}
		if kv := field(lit, "ID"); kv != nil {
			if v, ok := constantOf(kv.Value); ok && v.Kind() == constant.String && constant.StringVal(v) == id {
				found = append(found, lit)
			}
		}
		return true
	})
	switch len(found) {
	case 0:
		return nil, nil
	case 1:
		out.lit = found[0]
		return out, nil
	}
	return nil, fmt.Errorf("%s: %d arrangements with ID %q", path, len(found), id)
}

// names reports whether expr names synth's type.
func (f *file) names(expr ast.Expr, name string) bool {
	switch e := expr.(type) {
	case *ast.Ident:
		return f.synth == "" && e.Name == name
	case *ast.SelectorExpr:
		x, ok := e.X.(*ast.Ident)
		return ok && f.synth != "" && x.Name == f.synth && e.Sel.Name == name
	}
	return false
}

// synth's effects, which an interface's literal names.
var effects = []reflect.Type{
	reflect.TypeFor[synth.EQ](), reflect.TypeFor[synth.Drive](), reflect.TypeFor[synth.Chorus](),
	reflect.TypeFor[synth.Phaser](), reflect.TypeFor[synth.Delay](), reflect.TypeFor[synth.Reverb](),
	reflect.TypeFor[synth.Compressor](), reflect.TypeFor[synth.Crush](), reflect.TypeFor[synth.Limiter](),
	reflect.TypeFor[synth.Sidechain](), reflect.TypeFor[synth.TapeStop](), reflect.TypeFor[synth.AutoFilter](),
	reflect.TypeFor[synth.Fader](), reflect.TypeFor[synth.Panner](),
}

func (f *file) effect(expr ast.Expr) reflect.Type {
	for _, t := range effects {
		if f.names(expr, t.Name()) {
			return t
		}
	}
	return nil
}

func field(lit *ast.CompositeLit, name string) *ast.KeyValueExpr {
	for _, e := range lit.Elts {
		if kv, ok := e.(*ast.KeyValueExpr); ok {
			if k, ok := kv.Key.(*ast.Ident); ok && k.Name == name {
				return kv
			}
		}
	}
	return nil
}

func keyed(lit *ast.CompositeLit) bool {
	for _, e := range lit.Elts {
		if _, ok := e.(*ast.KeyValueExpr); !ok {
			return false
		}
	}
	return true
}

// compositeOf is the literal expr writes, through a &.
func compositeOf(expr ast.Expr) *ast.CompositeLit {
	if u, ok := expr.(*ast.UnaryExpr); ok && u.Op == token.AND {
		expr = u.X
	}
	lit, _ := expr.(*ast.CompositeLit)
	return lit
}

// constantOf evaluates a constant expression of literals, as the compiler does.
func constantOf(expr ast.Expr) (constant.Value, bool) {
	switch e := expr.(type) {
	case *ast.BasicLit:
		v := constant.MakeFromLiteral(e.Value, e.Kind, 0)
		return v, v.Kind() != constant.Unknown
	case *ast.ParenExpr:
		return constantOf(e.X)
	case *ast.Ident:
		switch e.Name {
		case "true", "false":
			return constant.MakeBool(e.Name == "true"), true
		}
	case *ast.UnaryExpr:
		if x, ok := constantOf(e.X); ok && (e.Op == token.SUB || e.Op == token.ADD) {
			return constant.UnaryOp(e.Op, x, 0), true
		}
	case *ast.BinaryExpr:
		x, okx := constantOf(e.X)
		y, oky := constantOf(e.Y)
		if !okx || !oky {
			return nil, false
		}
		switch e.Op {
		case token.ADD, token.SUB, token.MUL:
			return constant.BinaryOp(x, e.Op, y), true
		case token.QUO:
			if y.Kind() == constant.Int && constant.Sign(y) == 0 || y.Kind() == constant.Float && constant.Sign(y) == 0 {
				return nil, false
			}
			if x.Kind() == constant.Int && y.Kind() == constant.Int {
				return constant.BinaryOp(x, token.QUO_ASSIGN, y), true
			}
			return constant.BinaryOp(x, e.Op, y), true
		}
	}
	return nil, false
}

// scalar is the value of expr as t, when it's constant.
func (f *file) scalar(expr ast.Expr, t reflect.Type) (reflect.Value, bool) {
	if names, ok := enums[t]; ok {
		for i, name := range names {
			if f.names(expr, name) {
				return reflect.ValueOf(i).Convert(t), true
			}
		}
	}
	c, ok := constantOf(expr)
	if !ok {
		return reflect.Value{}, false
	}
	v := reflect.New(t).Elem()
	switch t.Kind() {
	case reflect.Float32, reflect.Float64:
		x, _ := constant.Float64Val(constant.ToFloat(c))
		v.SetFloat(x)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		x, exact := constant.Int64Val(constant.ToInt(c))
		if !exact {
			return v, false
		}
		v.SetInt(x)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		x, exact := constant.Uint64Val(constant.ToInt(c))
		if !exact {
			return v, false
		}
		v.SetUint(x)
	case reflect.String:
		if c.Kind() != constant.String {
			return v, false
		}
		v.SetString(constant.StringVal(c))
	case reflect.Bool:
		if c.Kind() != constant.Bool {
			return v, false
		}
		v.SetBool(constant.BoolVal(c))
	default:
		return v, false
	}
	return v, true
}

// plain reports whether expr is a literal a control can rewrite: a number, a string, a
// bool or one of synth's named constants. Arithmetic and names of the game's own are
// code.
func (f *file) plain(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.BasicLit:
		return true
	case *ast.UnaryExpr:
		_, ok := e.X.(*ast.BasicLit)
		return ok && (e.Op == token.SUB || e.Op == token.ADD)
	case *ast.Ident:
		return e.Name == "true" || e.Name == "false"
	case *ast.SelectorExpr:
		for _, names := range enums {
			if slices.ContainsFunc(names, func(n string) bool { return f.names(e, n) }) {
				return true
			}
		}
	}
	return false
}

// check reports where what expr writes differs from v. What it can't read (calls,
// variables, a game's own types) is taken to be v.
func (f *file) check(expr ast.Expr, v reflect.Value, where string) error {
	if expr == nil {
		if !v.IsZero() {
			return fmt.Errorf("%s: missing", where)
		}
		return nil
	}
	lit := compositeOf(expr)
	if lit == nil {
		got, ok := f.scalar(expr, v.Type())
		if ok && !reflect.DeepEqual(got.Interface(), v.Interface()) {
			return fmt.Errorf("%s: %v in the source, %v here", where, got, v)
		}
		return nil
	}
	switch v.Kind() {
	case reflect.Interface:
		t := f.effect(lit.Type)
		if t == nil {
			return nil
		}
		if v.IsNil() || v.Elem().Type() != t {
			return fmt.Errorf("%s: a %s in the source", where, t.Name())
		}
		return f.check(lit, v.Elem(), where)
	case reflect.Pointer:
		if v.IsNil() {
			return fmt.Errorf("%s: nil here", where)
		}
		return f.check(lit, v.Elem(), where)
	case reflect.Struct:
		if !keyed(lit) {
			return nil
		}
		for _, e := range lit.Elts {
			k := e.(*ast.KeyValueExpr).Key.(*ast.Ident).Name
			if _, ok := v.Type().FieldByName(k); !ok {
				return fmt.Errorf("%s: no field %s", where, k)
			}
		}
		for i := range v.NumField() {
			sf := v.Type().Field(i)
			if !sf.IsExported() {
				continue
			}
			var value ast.Expr
			if kv := field(lit, sf.Name); kv != nil {
				value = kv.Value
			}
			if err := f.check(value, v.Field(i), where+"."+sf.Name); err != nil {
				return err
			}
		}
		return nil
	case reflect.Slice:
		if len(lit.Elts) != v.Len() {
			return fmt.Errorf("%s: %d in the source, %d here", where, len(lit.Elts), v.Len())
		}
		for i, e := range lit.Elts {
			if _, ok := e.(*ast.KeyValueExpr); ok {
				return nil
			}
			if err := f.check(e, v.Index(i), fmt.Sprintf("%s[%d]", where, i)); err != nil {
				return err
			}
		}
		return nil
	}
	return fmt.Errorf("%s: a composite literal for a %s", where, v.Type())
}

// splice replaces the source from..to with text.
type splice struct {
	from, to int
	text     string
}

// writer rewrites a file's arrangement, splicing new source over only what changed.
type writer struct {
	*file
	edits []splice
	code  bool // it writes over code
}

func (w *writer) off(p token.Pos) int { return w.tok.Offset(p) }

func (w *writer) line(p int) int { return w.tok.Line(w.tok.Pos(p)) }

func (w *writer) lineStart(p int) int { return bytes.LastIndexByte(w.src[:p], '\n') + 1 }

// lineEnd is the start of the line after p's when the rest of p's line is no more than
// a comma and a comment, or -1.
func (w *writer) lineEnd(p int) int {
	rest := w.src[p:]
	if i := bytes.IndexByte(rest, '\n'); i >= 0 {
		rest = rest[:i]
	}
	s := strings.TrimLeft(string(rest), " \t")
	s = strings.TrimPrefix(s, ",")
	s = strings.TrimSpace(s)
	if s != "" && !strings.HasPrefix(s, "//") {
		return -1
	}
	return min(p+len(rest)+1, len(w.src))
}

func (w *writer) multiline(lit *ast.CompositeLit) bool {
	return w.line(w.off(lit.Lbrace)) != w.line(w.off(lit.Rbrace))
}

func (w *writer) replace(expr ast.Expr, v reflect.Value, typed bool) {
	w.edits = append(w.edits, splice{w.off(expr.Pos()), w.off(expr.End()), literal(v, typed)})
}

// remove deletes the literal's elements at the indexes: whole lines when they have lines
// of their own, end-of-line comments and all.
func (w *writer) remove(lit *ast.CompositeLit, gone []int) {
	slices.Sort(gone)
	for len(gone) > 0 {
		a, b := gone[0], gone[0]+1
		for len(gone) > b-a && gone[b-a] == b {
			b++
		}
		gone = gone[b-a:]
		from, to := w.off(lit.Elts[a].Pos()), w.off(lit.Elts[b-1].End())
		start := w.lineStart(from)
		if end := w.lineEnd(to); strings.TrimSpace(string(w.src[start:from])) == "" && end >= 0 && w.multiline(lit) {
			w.edits = append(w.edits, splice{start, end, ""})
			continue
		}
		switch {
		case a > 0 && (b == len(lit.Elts) || w.lineEnd(to) >= 0):
			from = w.off(lit.Elts[a-1].End())
		case b < len(lit.Elts):
			to = w.off(lit.Elts[b].Pos())
		}
		w.edits = append(w.edits, splice{from, to, ""})
	}
}

// insert writes items into the literal before element k, or at its end, one to a line
// when the literal has lines, or into the row they fall in when rows says it's written
// several to a line; survivors says whether any element is kept.
func (w *writer) insert(lit *ast.CompositeLit, k int, items []string, survivors, rows bool) {
	if len(items) == 0 {
		return
	}
	lines := strings.Join(items, ",\n") + ",\n"
	inline := strings.Join(items, ", ")
	if k < len(lit.Elts) {
		at := w.off(lit.Elts[k].Pos())
		if start := w.lineStart(at); w.multiline(lit) && !rows && strings.TrimSpace(string(w.src[start:at])) == "" {
			w.edits = append(w.edits, splice{start, start, lines})
		} else {
			w.edits = append(w.edits, splice{at, at, inline + ", "})
		}
		return
	}
	rbrace := w.off(lit.Rbrace)
	after := w.off(lit.Lbrace) + 1
	if len(lit.Elts) > 0 {
		after = w.off(lit.Elts[len(lit.Elts)-1].End())
	}
	between := string(w.src[after:rbrace])
	switch {
	case rows:
		w.edits = append(w.edits, splice{after, after, ", " + inline})
	case strings.Contains(between, "\n"):
		start := w.lineStart(rbrace)
		w.edits = append(w.edits, splice{start, start, lines})
	case !survivors:
		w.edits = append(w.edits, splice{rbrace, rbrace, inline})
	case strings.Contains(between, ","):
		w.edits = append(w.edits, splice{rbrace, rbrace, " " + inline})
	default:
		w.edits = append(w.edits, splice{rbrace, rbrace, ", " + inline})
	}
}

var noteType = reflect.TypeFor[synth.Note]()

// diff splices what changed from old to new into expr, which writes old; typed says
// whether a composite written in its place needs its type.
func (w *writer) diff(expr ast.Expr, old, new reflect.Value, typed bool) {
	if reflect.DeepEqual(old.Interface(), new.Interface()) {
		return
	}
	lit := compositeOf(expr)
	if lit == nil {
		w.code = w.code || !w.plain(expr)
		w.replace(expr, new, typed)
		return
	}
	switch old.Kind() {
	case reflect.Interface, reflect.Pointer:
		if old.IsNil() || new.IsNil() || old.Elem().Type() != new.Elem().Type() {
			w.replace(expr, new, typed)
			return
		}
		w.diff(lit, old.Elem(), new.Elem(), typed)
	case reflect.Struct:
		if !keyed(lit) {
			w.replace(expr, new, typed)
			return
		}
		var gone []int
		var added []string
		for i := range old.NumField() {
			sf := old.Type().Field(i)
			of, nf := old.Field(i), new.Field(i)
			if !sf.IsExported() || reflect.DeepEqual(of.Interface(), nf.Interface()) {
				continue
			}
			kv := field(lit, sf.Name)
			switch {
			case kv == nil:
				added = append(added, sf.Name+": "+literal(nf, true))
			case nf.IsZero():
				gone = append(gone, slices.Index(lit.Elts, ast.Expr(kv)))
			default:
				w.diff(kv.Value, of, nf, true)
			}
		}
		w.remove(lit, gone)
		w.insert(lit, len(lit.Elts), added, len(gone) < len(lit.Elts), false)
	case reflect.Slice:
		keyed := slices.ContainsFunc(lit.Elts, func(e ast.Expr) bool { _, ok := e.(*ast.KeyValueExpr); return ok })
		switch elem := old.Type().Elem(); {
		case keyed, elem == noteType:
			w.replace(expr, new, typed)
		case elem.Kind() != reflect.Struct && elem.Kind() != reflect.Interface && elem.Kind() != reflect.Pointer:
			// A clip's bars, as many as there were, are rewritten where they are.
			if old.Len() != new.Len() || len(lit.Elts) != old.Len() {
				w.replace(expr, new, typed)
				return
			}
			for i := range old.Len() {
				w.diff(lit.Elts[i], old.Index(i), new.Index(i), false)
			}
		default:
			w.list(lit, old, new)
		}
	default:
		w.replace(expr, new, typed)
	}
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

// list matches old's items to new's, by name or by being equal, in order: those matched
// are diffed, those between them paired off and diffed, and the rest removed or
// inserted where they fall.
func (w *writer) list(lit *ast.CompositeLit, old, new reflect.Value) {
	n, m := old.Len(), new.Len()
	typed := old.Type().Elem().Kind() == reflect.Interface
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if same(old.Index(i), new.Index(j)) {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else {
				lcs[i][j] = max(lcs[i+1][j], lcs[i][j+1])
			}
		}
	}
	var gone []int
	type insertion struct {
		at    int
		items []string
	}
	var inserts []insertion
	gap := func(i0, i1, j0, j1 int) {
		k := min(i1-i0, j1-j0)
		for t := range k {
			w.diff(lit.Elts[i0+t], old.Index(i0+t), new.Index(j0+t), typed)
		}
		for i := i0 + k; i < i1; i++ {
			gone = append(gone, i)
		}
		var items []string
		for j := j0 + k; j < j1; j++ {
			items = append(items, literal(new.Index(j), typed))
		}
		if len(items) > 0 {
			inserts = append(inserts, insertion{i1, items})
		}
	}
	i0, j0 := 0, 0
	for i, j := 0, 0; i < n && j < m; {
		switch {
		case same(old.Index(i), new.Index(j)) && lcs[i][j] == lcs[i+1][j+1]+1:
			gap(i0, i, j0, j)
			w.diff(lit.Elts[i], old.Index(i), new.Index(j), typed)
			i, j = i+1, j+1
			i0, j0 = i, j
		case lcs[i+1][j] >= lcs[i][j+1]:
			i++
		default:
			j++
		}
	}
	gap(i0, n, j0, m)
	w.remove(lit, gone)
	for _, ins := range inserts {
		w.insert(lit, ins.at, ins.items, len(gone) < n, w.rows(lit))
	}
}

// rows reports whether a list with lines has items several to a line.
func (w *writer) rows(lit *ast.CompositeLit) bool {
	if !w.multiline(lit) {
		return false
	}
	for i := 1; i < len(lit.Elts); i++ {
		if w.line(w.off(lit.Elts[i].Pos())) == w.line(w.off(lit.Elts[i-1].End())) {
			return true
		}
	}
	return false
}

// apply makes the edited source, gofmted.
func (w *writer) apply() ([]byte, error) {
	slices.SortStableFunc(w.edits, func(a, b splice) int {
		if a.from != b.from {
			return a.from - b.from
		}
		return a.to - b.to
	})
	var out bytes.Buffer
	at := 0
	for _, e := range w.edits {
		if e.from < at {
			return nil, errors.New("edits overlap")
		}
		out.Write(w.src[at:e.from])
		out.WriteString(e.text)
		at = e.to
	}
	out.Write(w.src[at:])
	return format.Source(out.Bytes())
}
