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
		shelf := biquad{b0: 1.53512485958697, b1: -2.69169618940638, b2: 1.19839281085285, a1: -1.69065929318241, a2: 0.73248077421585}
		highpass := biquad{b0: 1, b1: -2, b2: 1, a1: -1.99004745483398, a2: 0.99007225036621}
		for i, x := range ch {
			y := highpass.process(shelf.process(float64(x)))
			sum[i+1] += y * y
		}
	}
	for i := 1; i < len(sum); i++ {
		sum[i] += sum[i-1]
	}
	return sum
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
