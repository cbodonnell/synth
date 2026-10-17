package synth

import (
	"bytes"
	"encoding/binary"
	"math"
)

// WAV is 16-bit stereo, for auditioning outside the game.
func WAV(s Stereo) []byte {
	const channels, bits = 2, 16
	data := make([]byte, s.Len()*channels*bits/8)
	for i := range s.L {
		binary.LittleEndian.PutUint16(data[i*4:], uint16(toInt16(s.L[i])))
		binary.LittleEndian.PutUint16(data[i*4+2:], uint16(toInt16(s.R[i])))
	}

	var buf bytes.Buffer
	w := func(v any) { binary.Write(&buf, binary.LittleEndian, v) }
	buf.WriteString("RIFF")
	w(uint32(36 + len(data)))
	buf.WriteString("WAVEfmt ")
	w(uint32(16))
	w(uint16(1))
	w(uint16(channels))
	w(uint32(SampleRate))
	w(uint32(SampleRate * channels * bits / 8))
	w(uint16(channels * bits / 8))
	w(uint16(bits))
	buf.WriteString("data")
	w(uint32(len(data)))
	buf.Write(data)
	return buf.Bytes()
}

func toInt16(x float32) int16 {
	return int16(math.Round(clamp(float64(x), -1, 1) * 32767))
}

// Concat joins buffers end to end.
func Concat(parts ...Stereo) Stereo {
	n := 0
	for _, p := range parts {
		n += p.Len()
	}
	out := Stereo{L: make([]float32, 0, n), R: make([]float32, 0, n)}
	for _, p := range parts {
		out.L = append(out.L, p.L...)
		out.R = append(out.R, p.R...)
	}
	return out
}
