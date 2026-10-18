package bake

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"hash"
	"reflect"
	"strconv"

	"github.com/cbodonnell/synth"
)

// format changes whenever a bake's files change shape, so every song is baked again.
const format = 3

// floatDigits is how many significant digits of a float the fingerprint sees, far finer
// than anything audible and far coarser than a last-bit difference.
const floatDigits = 9

// Fingerprint hashes everything a song renders from, synth.Revision for how it renders,
// and the bake's own format. It walks the value rather than printing it, since %v would hash pointers'
// addresses and couldn't tell two effects with the same fields apart. Floats are hashed to
// floatDigits significant digits: a song computed on another architecture can differ in
// the last bit (arm64 fuses multiply-adds, wasm doesn't), and that mustn't make the bake
// stale there.
func Fingerprint(s synth.Song) string {
	h := sha256.New()
	writeUint(h, synth.Revision)
	writeUint(h, format)
	walk(h, reflect.ValueOf(s))
	return hex.EncodeToString(h.Sum(nil)[:16])
}

func walk(h hash.Hash, v reflect.Value) {
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			h.Write([]byte{0})
			return
		}
		h.Write([]byte{1})
		if v.Kind() == reflect.Interface {
			h.Write([]byte(v.Elem().Type().String()))
		}
		walk(h, v.Elem())
	case reflect.Struct:
		for i := range v.NumField() {
			walk(h, v.Field(i))
		}
	case reflect.Slice, reflect.Array:
		writeUint(h, uint64(v.Len()))
		for i := range v.Len() {
			walk(h, v.Index(i))
		}
	case reflect.String:
		writeUint(h, uint64(v.Len()))
		h.Write([]byte(v.String()))
	case reflect.Float32, reflect.Float64:
		h.Write(strconv.AppendFloat(nil, v.Float(), 'g', floatDigits, 64))
		h.Write([]byte{0})
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		writeUint(h, uint64(v.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		writeUint(h, v.Uint())
	case reflect.Bool:
		if v.Bool() {
			h.Write([]byte{1})
		} else {
			h.Write([]byte{0})
		}
	default:
		panic("bake: can't fingerprint a " + v.Kind().String())
	}
}

func writeUint(h hash.Hash, x uint64) {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], x)
	h.Write(b[:])
}
