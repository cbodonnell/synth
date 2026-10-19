//go:build unix

package synth

import (
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

// mapFloats is n floats in a file of their own in dir, mapped into memory, and what
// unmaps them. The file is deleted at once: the mapping holds it until it's unmapped.
func mapFloats(dir string, n int) ([]float32, func(), error) {
	f, err := os.CreateTemp(dir, "synth-*.pcm")
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	defer os.Remove(f.Name())
	size := n * 4
	if err := f.Truncate(int64(size)); err != nil {
		return nil, nil, err
	}
	b, err := unix.Mmap(int(f.Fd()), 0, size, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		return nil, nil, err
	}
	return unsafe.Slice((*float32)(unsafe.Pointer(&b[0])), n), func() { unix.Munmap(b) }, nil
}
