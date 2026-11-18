package synth

import "math"

// Loudness measures s per ITU-R BS.1770: K-weighted, in 400ms blocks overlapping by
// 75%, gated at -70 LUFS and again 10 LU below the level of what passed. It returns
// the integrated loudness and the loudest 3s (short-term) window, in LUFS; silence is
// -Inf. A sound shorter than a window is measured as one window of its own length.
func Loudness(s Stereo) (integrated, shortTermMax float64) {
	power := kPower(s)
	lufs := func(p float64) float64 { return -0.691 + 10*math.Log10(p) }

	var gated []float64
	for _, p := range blockPowers(power, 0.4) {
		if lufs(p) > -70 {
			gated = append(gated, p)
		}
	}
	integrated = math.Inf(-1)
	if len(gated) > 0 {
		relative := lufs(mean(gated)) - 10
		var kept []float64
		for _, p := range gated {
			if lufs(p) > relative {
				kept = append(kept, p)
			}
		}
		integrated = lufs(mean(kept))
	}

	shortTermMax = math.Inf(-1)
	for _, p := range blockPowers(power, 3) {
		shortTermMax = math.Max(shortTermMax, lufs(p))
	}
	return integrated, shortTermMax
}

// kPower is the running sum of each sample's K-weighted power across both channels.
// The coefficients are BS.1770's own, which are for 48kHz.
func kPower(s Stereo) []float64 {
	sum := make([]float64, s.Len()+1)
	for _, ch := range [][]float32{s.L, s.R} {
		k := newKWeighting()
		for i, x := range ch {
			y := k.process(float64(x))
			sum[i+1] += y * y
		}
	}
	for i := 1; i < len(sum); i++ {
		sum[i] += sum[i-1]
	}
	return sum
}

type kWeighting struct{ shelf, highpass biquad }

func newKWeighting() kWeighting {
	return kWeighting{
		shelf:    biquad{b0: 1.53512485958697, b1: -2.69169618940638, b2: 1.19839281085285, a1: -1.69065929318241, a2: 0.73248077421585},
		highpass: biquad{b0: 1, b1: -2, b2: 1, a1: -1.99004745483398, a2: 0.99007225036621},
	}
}

func (k *kWeighting) process(x float64) float64 { return k.highpass.process(k.shelf.process(x)) }

// Gauge reads a buffer as a level meter does, anywhere in it: its peak, and its
// short-term loudness, to the nearest 10ms.
type Gauge struct {
	peak []float32 // each block's
	sum  []float64 // K-weighted power before each block
}

const gaugeBlock = SampleRate / 100

// NewGauge meters s.
func NewGauge(s Stereo) Gauge { return newGauge(span{buf: s}, s.Len(), 1) }

// newGauge meters n samples of which sp holds the sound, scaled by gain.
func newGauge(sp span, n int, gain float32) Gauge {
	blocks := (n + gaugeBlock - 1) / gaugeBlock
	g := Gauge{peak: make([]float32, blocks), sum: make([]float64, blocks+1)}
	power := make([]float64, blocks)
	for _, ch := range [][]float32{sp.buf.L, sp.buf.R} {
		k := newKWeighting()
		for i := sp.at; i < n; i++ {
			var x float32
			if j := i - sp.at; j < len(ch) {
				x = ch[j] * gain
			} else if j > len(ch)+gaugeBlock {
				break
			}
			y := k.process(float64(x))
			power[i/gaugeBlock] += y * y
			g.peak[i/gaugeBlock] = max(g.peak[i/gaugeBlock], abs32(x))
		}
	}
	for b, p := range power {
		g.sum[b+1] = g.sum[b] + p
	}
	return g
}

// Peak is the loudest sample in the window of seconds before frame.
func (g Gauge) Peak(frame int, seconds float64) float32 {
	hi := min(frame/gaugeBlock+1, len(g.peak))
	lo := max(hi-max(samples(seconds)/gaugeBlock, 1), 0)
	var p float32
	for _, x := range g.peak[lo:max(hi, lo)] {
		p = max(p, x)
	}
	return p
}

// ShortTerm is the loudness of the 3s before frame, in LUFS; -Inf in silence.
func (g Gauge) ShortTerm(frame int) float64 {
	const window = 300
	hi := min(frame/gaugeBlock+1, len(g.sum)-1)
	lo := max(hi-window, 0)
	if hi <= lo {
		return math.Inf(-1)
	}
	return -0.691 + 10*math.Log10((g.sum[hi]-g.sum[lo])/(window*gaugeBlock))
}

// blockPowers is the mean power of each window of the given seconds, every 100ms.
func blockPowers(sum []float64, seconds float64) []float64 {
	n := len(sum) - 1
	w := min(samples(seconds), n)
	if w == 0 {
		return nil
	}
	var out []float64
	for start := 0; start+w <= n; start += samples(0.1) {
		out = append(out, (sum[start+w]-sum[start])/float64(w))
	}
	return out
}

func mean(x []float64) float64 {
	var s float64
	for _, v := range x {
		s += v
	}
	return s / float64(len(x))
}
