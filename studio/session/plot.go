package session

import (
	"image"
	"image/color"
	"math"
)

// Plot draws ys, each 0..1 from the bottom, across an image w by h: a line of the
// given width over the area beneath it, and a mark at a fraction of the way across,
// unless it's negative.
func Plot(ys []float64, mark float64, w, h, width int, ink, under color.Color) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	if len(ys) == 0 || w < 2 || h < 2 {
		return img
	}
	at := func(x int) float64 {
		t := float64(x) / float64(w-1) * float64(len(ys)-1)
		i := min(int(t), len(ys)-1)
		y := ys[i]
		if i+1 < len(ys) {
			y += (ys[i+1] - y) * (t - float64(i))
		}
		return (1 - math.Max(0, math.Min(1, y))) * float64(h-1-width)
	}
	half := float64(width) / 2
	prev := at(0)
	for x := range w {
		y := at(x)
		for py := int(y + half); py < h; py++ {
			img.Set(x, py, under)
		}
		lo, hi := math.Min(prev, y), math.Max(prev, y)
		for py := int(lo); py <= int(hi)+width && py < h; py++ {
			img.Set(x, py, ink)
		}
		prev = y
	}
	if mark >= 0 && mark <= 1 {
		x := int(mark * float64(w-1))
		for py := 0; py < h; py += 3 {
			img.Set(x, py, ink)
		}
	}
	return img
}
