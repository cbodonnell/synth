package synth

import (
	"cmp"
	"encoding/binary"
	"hash/fnv"
	"hash/maphash"
	"math"
	"reflect"
	"runtime"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

// Renderer renders as Render and RenderLoop do, byte for byte, and keeps what it
// rendered: each note (by its part's patch, pan and seed, and the note), each part (by
// its notes, patch and inserts) and each bus (by what feeds it and its effects). A render
// reuses whatever the last one shared with it, so moving a note renders that note and
// its part's inserts, and a level change only re-mixes. What neither of the last two
// renders used is let go. Safe for concurrent use, though renders take turns.
//
// By default everything is held in memory. Budget caps that, letting go of what's
// cheapest to render again first, and Spill moves the large buffers (parts and buses)
// into memory-mapped files, which the system pages out when memory runs short, as a
// DAW's frozen tracks are.
type Renderer struct {
	busy sync.Mutex // held by the render running

	mu     sync.Mutex
	gen    uint64
	cache  map[uint64]*cached
	memory int64 // bytes of cached buffers held in memory
	mapped int64 // and in mapped files

	budget int64
	spill  string

	done, total atomic.Int64
}

type cached struct {
	span span
	gen  uint64
	// What rendering it again would take: its own work, and what it's made from.
	cost    time.Duration
	bytes   int64
	release func() // unmaps a spilled buffer; nil in memory
}

type RendererOption func(*Renderer)

// Budget holds the cache to this many bytes of memory once a render is done. A render
// in progress may go over it, since what it's using can't be let go.
func Budget(bytes int64) RendererOption { return func(r *Renderer) { r.budget = bytes } }

// Spill keeps buffers of a megabyte or more in memory-mapped files in dir, which are
// deleted at once and so never outlive the process. Where files can't be mapped, they
// stay in memory.
func Spill(dir string) RendererOption { return func(r *Renderer) { r.spill = dir } }

// Below this a buffer isn't worth a file of its own.
const spillBytes = 1 << 20

func NewRenderer(opts ...RendererOption) *Renderer {
	r := &Renderer{cache: map[uint64]*cached{}}
	for _, o := range opts {
		o(r)
	}
	return r
}

// Held is how many bytes the cache holds in memory and in mapped files.
func (r *Renderer) Held() (memory, mapped int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.memory, r.mapped
}

// Render plays the song once from silence until everything has died away.
func Render(s Song) Stereo { return (*Renderer)(nil).Render(s) }

func RenderLoop(s Song) Looped { return (*Renderer)(nil).RenderLoop(s) }

func (r *Renderer) Render(s Song) Stereo {
	defer r.begin(true)()
	out := r.mixdown(s, s.Parts, renderLength(s))
	trimSilence(&out, samples(s.Length))
	clip(out)
	return out
}

func renderLength(s Song) int { return samples(math.Max(s.Length, s.end()) + s.tail()) }

// Looped is a looping song split the way the player plays it: Intro once, then Body forever.
type Looped struct {
	// Plays once: everything before LoopStart and then the first pass through the
	// loop, which starts from silence rather than from the loop's own tail.
	Intro Stereo
	// Repeats forever. Its own tail is wrapped onto its start, so the seam is continuous.
	Body Stereo
}

func (r *Renderer) RenderLoop(s Song) Looped {
	total := samples(s.Length)
	start := min(samples(s.LoopStart), total)
	body := total - start
	if body <= 0 {
		return Looped{Intro: r.Render(s)}
	}
	defer r.begin(true)()
	n := loopLength(s)

	intro := r.mixdown(s, s.Parts, n)
	intro = Stereo{L: intro.L[:total], R: intro.R[:total]}

	var loopParts []Part
	for _, p := range s.Parts {
		lp := p
		lp.Notes = nil
		for _, n := range p.Notes {
			if n.At >= s.LoopStart {
				lp.Notes = append(lp.Notes, n)
			}
		}
		loopParts = append(loopParts, lp)
	}
	full := r.mixdown(s, loopParts, n)

	looped := NewStereo(body)
	for i := start; i < full.Len(); i++ {
		j := (i - start) % body
		looped.L[j] += full.L[i]
		looped.R[j] += full.R[i]
	}

	clip(intro)
	clip(looped)
	return Looped{Intro: intro, Body: looped}
}

func loopLength(s Song) int { return max(samples(s.Length), samples(s.end()+s.tail())+1) }

// Gauges meter each part on its own, its notes through its inserts and at its level,
// and each bus, its effects' output at its level, as the song's first pass mixes them.
// After a Render or RenderLoop of the song all of it is already rendered.
func (r *Renderer) Gauges(s Song) (parts, buses []Gauge) {
	defer r.begin(false)()
	n := renderLength(s)
	if s.Loop && min(samples(s.LoopStart), samples(s.Length)) < samples(s.Length) {
		n = loopLength(s)
	}
	rendered, keys, costs := r.parts(s, s.Parts, n)
	parts = make([]Gauge, len(s.Parts))
	var wg sync.WaitGroup
	for i, p := range s.Parts {
		wg.Go(func() { parts[i] = newGauge(rendered[i], n, float32(dbToGain(p.Level))) })
	}
	bus := r.buses(s, s.Parts, rendered, keys, costs, n)
	buses = make([]Gauge, len(s.Buses))
	for i, b := range s.Buses {
		wg.Go(func() { buses[i] = newGauge(bus[i], n, float32(dbToGain(b.Level))) })
	}
	wg.Wait()
	return parts, buses
}

// Progress is how far the render running has come, 0..1, counted in notes.
func (r *Renderer) Progress() float64 {
	total := r.total.Load()
	if total == 0 {
		return 1
	}
	return float64(r.done.Load()) / float64(total)
}

// begin starts a render, a new one or more of the last, and returns what ends it:
// letting go of everything neither it nor the render before it used, then of whatever
// is over the budget.
func (r *Renderer) begin(fresh bool) func() {
	if r == nil {
		return func() {}
	}
	r.busy.Lock()
	r.mu.Lock()
	if fresh {
		r.gen++
		r.done.Store(0)
		r.total.Store(0)
	}
	r.mu.Unlock()
	return func() {
		defer r.busy.Unlock()
		r.mu.Lock()
		defer r.mu.Unlock()
		for k, c := range r.cache {
			if c.gen+1 < r.gen {
				r.drop(k, c)
			}
		}
		r.shed(r.gen + 1)
	}
}

func (r *Renderer) drop(k uint64, c *cached) {
	delete(r.cache, k)
	if c.release != nil {
		c.release()
		r.mapped -= c.bytes
	} else {
		r.memory -= c.bytes
	}
}

// shed lets go of buffers held in memory from renders before gen until the cache is
// within its budget, those cheapest to render again for their size first. Buffers this
// render has used are held, since it may be reading them.
func (r *Renderer) shed(gen uint64) {
	if r.budget <= 0 || r.memory <= r.budget {
		return
	}
	var keys []uint64
	for k, c := range r.cache {
		if c.release == nil && c.bytes > 0 && c.gen < gen {
			keys = append(keys, k)
		}
	}
	perByte := func(c *cached) float64 { return float64(c.cost) / float64(c.bytes) }
	slices.SortFunc(keys, func(a, b uint64) int {
		return cmp.Compare(perByte(r.cache[a]), perByte(r.cache[b]))
	})
	for _, k := range keys {
		if r.memory <= r.budget {
			return
		}
		r.drop(k, r.cache[k])
	}
}

// span is a buffer kept without its silent ends: everything before at, and after
// at+buf.Len(), is exactly zero.
type span struct {
	at  int
	buf Stereo
}

func (s Stereo) addSpan(src span, gain float32) {
	Stereo{L: s.L[src.at:], R: s.R[src.at:]}.add(src.buf, gain)
}

// keep stores s without its silent ends: in a mapped file if it's large and the
// renderer spills, in memory otherwise.
func (r *Renderer) keep(s Stereo) (sp span, bytes int64, release func()) {
	lo, hi := 0, s.Len()
	for lo < hi && s.L[lo] == 0 && s.R[lo] == 0 {
		lo++
	}
	for hi > lo && s.L[hi-1] == 0 && s.R[hi-1] == 0 {
		hi--
	}
	n := hi - lo
	bytes = int64(n) * 8
	if r.spill != "" && bytes >= spillBytes {
		if f, release, err := mapFloats(r.spill, 2*n); err == nil {
			copy(f, s.L[lo:hi])
			copy(f[n:], s.R[lo:hi])
			return span{at: lo, buf: Stereo{L: f[:n:n], R: f[n:]}}, bytes, release
		}
	}
	if n == s.Len() {
		return span{buf: s}, bytes, nil
	}
	return span{at: lo, buf: Stereo{L: slices.Clone(s.L[lo:hi]), R: slices.Clone(s.R[lo:hi])}}, bytes, nil
}

// reuse is the buffer cached under k, or render's, which it keeps, with what rendering
// it again would cost: render reports what its inputs cost, and its own time is added.
// A key that couldn't be made always renders.
func (r *Renderer) reuse(k key, render func() (Stereo, time.Duration)) (span, time.Duration) {
	if r == nil || !k.ok {
		buf, _ := render()
		return span{buf: buf}, 0
	}
	sum := k.h.Sum64()
	r.mu.Lock()
	if c, ok := r.cache[sum]; ok {
		c.gen = r.gen
		r.mu.Unlock()
		return c.span, c.cost
	}
	r.mu.Unlock()
	began := time.Now()
	buf, inputs := render()
	cost := time.Since(began) + inputs
	sp, bytes, release := r.keep(buf)
	r.mu.Lock()
	defer r.mu.Unlock()
	// Rendered twice at once: the first kept is the one others may be reading.
	if c, ok := r.cache[sum]; ok {
		if release != nil {
			release()
		}
		c.gen = r.gen
		return c.span, c.cost
	}
	r.cache[sum] = &cached{span: sp, gen: r.gen, cost: cost, bytes: bytes, release: release}
	if release != nil {
		r.mapped += bytes
	} else {
		r.memory += bytes
	}
	r.shed(r.gen)
	return sp, cost
}

func (r *Renderer) mixdown(s Song, parts []Part, n int) Stereo {
	master := NewStereo(n)
	if n == 0 {
		return master
	}
	rendered, keys, costs := r.parts(s, parts, n)
	for i, p := range parts {
		if !slices.ContainsFunc(s.Buses, func(b Bus) bool { return b.Name == p.Out }) {
			master.addSpan(rendered[i], float32(dbToGain(p.Level)))
		}
	}
	for i, buf := range r.buses(s, parts, rendered, keys, costs, n) {
		master.addSpan(buf, float32(dbToGain(s.Buses[i].Level)))
	}
	blockDC(master)
	for _, e := range s.Master {
		e.Process(master)
	}
	return master
}

func (r *Renderer) parts(s Song, parts []Part, n int) ([]span, []key, []time.Duration) {
	rendered := make([]span, len(parts))
	keys := make([]key, len(parts))
	costs := make([]time.Duration, len(parts))
	var wg sync.WaitGroup
	for i, p := range parts {
		wg.Go(func() { rendered[i], keys[i], costs[i] = r.part(p, s.take, n) })
	}
	wg.Wait()
	return rendered, keys, costs
}

// buses are each bus's output, before its level: what plays into it and sends to it,
// through its effects.
func (r *Renderer) buses(s Song, parts []Part, rendered []span, keys []key, costs []time.Duration, n int) []span {
	type feed struct {
		part int
		gain float32
	}
	feeds := make(map[string][]feed, len(s.Buses))
	for _, b := range s.Buses {
		feeds[b.Name] = []feed{}
	}
	for i, p := range parts {
		g := float32(dbToGain(p.Level))
		if _, ok := feeds[p.Out]; ok {
			feeds[p.Out] = append(feeds[p.Out], feed{i, g})
		}
		for _, send := range p.Sends {
			if _, ok := feeds[send.Bus]; ok {
				feeds[send.Bus] = append(feeds[send.Bus], feed{i, g * float32(dbToGain(send.Level))})
			}
		}
	}
	out := make([]span, len(s.Buses))
	for bi, b := range s.Buses {
		k := newKey()
		k.int(n)
		k.value(reflect.ValueOf(b.Effects))
		for _, f := range feeds[b.Name] {
			k.key(keys[f.part])
			k.float(float64(f.gain))
		}
		out[bi], _ = r.reuse(k, func() (Stereo, time.Duration) {
			var inputs time.Duration
			buf := NewStereo(n)
			for _, f := range feeds[b.Name] {
				buf.addSpan(rendered[f.part], f.gain)
				inputs += costs[f.part]
			}
			for _, e := range b.Effects {
				e.Process(buf)
			}
			return buf, inputs
		})
	}
	return out
}

func (r *Renderer) part(p Part, take, n int) (span, key, time.Duration) {
	h := fnv.New64a()
	h.Write([]byte(p.Name))
	seed := h.Sum64()
	if take > 0 {
		seed = mixSeed(seed, uint64(take))
	}

	voice := newKey()
	voice.value(reflect.ValueOf(p.Patch))
	voice.float(p.Pan)
	voice.uint(seed)

	k := newKey()
	k.key(voice)
	k.int(n)
	k.value(reflect.ValueOf(p.Inserts))
	k.value(reflect.ValueOf(p.Notes))

	sp, cost := r.reuse(k, func() (Stereo, time.Duration) {
		if r != nil {
			r.total.Add(int64(len(p.Notes)))
		}
		// Notes render concurrently into buffers of their own and are summed in order,
		// so the result is identical however the work was scheduled.
		notes := make([]span, len(p.Notes))
		costs := make([]time.Duration, len(p.Notes))
		work := make(chan int)
		var wg sync.WaitGroup
		for range runtime.GOMAXPROCS(0) {
			wg.Go(func() {
				for ni := range work {
					notes[ni], costs[ni] = r.note(p, p.Notes[ni], voice, seed, n)
					if r != nil {
						r.done.Add(1)
					}
				}
			})
		}
		for ni := range p.Notes {
			work <- ni
		}
		close(work)
		wg.Wait()

		out := NewStereo(n)
		for _, note := range notes {
			for i := range note.buf.L {
				out.L[note.at+i] += note.buf.L[i]
				out.R[note.at+i] += note.buf.R[i]
			}
		}
		for _, e := range p.Inserts {
			e.Process(out)
		}
		var inputs time.Duration
		for _, c := range costs {
			inputs += c
		}
		return out, inputs
	})
	return sp, k, cost
}

func (r *Renderer) note(p Part, note Note, voice key, partSeed uint64, songLength int) (span, time.Duration) {
	start := samples(note.At)
	length := min(samples(noteLength(p.Patch, note)), songLength-start)
	if length <= 0 {
		return span{}, 0
	}
	k := newKey()
	k.key(voice)
	k.int(length)
	k.value(reflect.ValueOf(note))
	sp, cost := r.reuse(k, func() (Stereo, time.Duration) {
		buf := NewStereo(length)
		// Seeded by the note itself, not its index, so RenderLoop's loop-only pass
		// renders each note exactly as the full pass does.
		seed := mixSeed(partSeed, math.Float64bits(note.At), math.Float64bits(note.Key))
		for _, v := range unisonVoices(p.Patch.Unison, p.Pan, seed) {
			renderVoice(buf, 0, p.Patch, note, v)
		}
		return buf, 0
	})
	sp.at += start
	return sp, cost
}

var keySeed = maphash.MakeSeed()

// key hashes everything a buffer renders from: exactly, floats by their bits. A value it
// can't see into (a func or a map in a game's own effect) leaves it unusable.
type key struct {
	h  *maphash.Hash
	ok bool
}

func newKey() key {
	h := &maphash.Hash{}
	h.SetSeed(keySeed)
	return key{h: h, ok: true}
}

func (k *key) uint(x uint64) {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], x)
	k.h.Write(b[:])
}

func (k *key) int(x int)       { k.uint(uint64(x)) }
func (k *key) float(x float64) { k.uint(math.Float64bits(x)) }

func (k *key) key(o key) {
	k.ok = k.ok && o.ok
	k.uint(o.h.Sum64())
}

func (k *key) value(v reflect.Value) {
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			k.uint(0)
			return
		}
		k.uint(1)
		if v.Kind() == reflect.Interface {
			k.h.WriteString(v.Elem().Type().String())
		}
		k.value(v.Elem())
	case reflect.Struct:
		for i := range v.NumField() {
			k.value(v.Field(i))
		}
	case reflect.Slice, reflect.Array:
		k.int(v.Len())
		for i := range v.Len() {
			k.value(v.Index(i))
		}
	case reflect.String:
		k.int(v.Len())
		k.h.WriteString(v.String())
	case reflect.Float32, reflect.Float64:
		k.float(v.Float())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		k.int(int(v.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		k.uint(v.Uint())
	case reflect.Bool:
		if v.Bool() {
			k.uint(1)
		} else {
			k.uint(0)
		}
	case reflect.Invalid:
		k.uint(2)
	default:
		k.ok = false
	}
}
