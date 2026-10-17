package studio

import (
	"fmt"
	"math"
	"strings"

	"github.com/cbodonnell/synth"
)

// scoreCols keeps a system, its pitch labels, and bar lines within 100 columns.
const scoreCols = 84

// One column per half beat, so an eighth-note line reads one token per column.
func printScore(s synth.Song) {
	step := s.Meter.Beat / 2
	if step <= 0 {
		step = 0.125
	}
	cols := int(math.Ceil(s.Length/step - 1e-9))
	if cols <= 0 {
		for _, p := range s.Parts {
			for _, n := range p.Notes {
				cols = max(cols, int(math.Ceil((n.At+n.Dur)/step)))
			}
		}
	}

	var barlines []int
	bars := s.Meter.Bars
	if len(bars) == 0 {
		bars = []int{4}
	}
	for c, k := 0, 0; c < cols; k++ {
		c += 2 * bars[k%len(bars)]
		barlines = append(barlines, c)
	}
	isBar := map[int]bool{}
	for _, b := range barlines {
		isBar[b] = true
	}

	fmt.Printf("%s: %s   %d columns of %.0fms, %.1fs", s.ID, s.Name, cols, step*1000, s.Length)
	if s.Loop {
		fmt.Printf(", loops from column %d", int(math.Round(s.LoopStart/step)))
	}
	fmt.Println()

	for _, p := range s.Parts {
		if len(p.Notes) == 0 {
			continue
		}
		lo, hi := math.MaxInt, math.MinInt
		for _, n := range p.Notes {
			k := int(math.Round(n.Key))
			lo, hi = min(lo, k), max(hi, k)
		}

		grid := make([][]byte, hi-lo+1)
		for i := range grid {
			grid[i] = []byte(strings.Repeat(" ", cols))
		}
		for _, n := range p.Notes {
			row := int(math.Round(n.Key)) - lo
			at := int(math.Round(n.At / step))
			held := max(int(math.Round(n.Dur/step)), 1)
			for c := at; c < at+held && c < cols; c++ {
				if c == at {
					grid[row][c] = '#'
				} else if grid[row][c] == ' ' {
					grid[row][c] = '-'
				}
			}
		}

		fmt.Printf("\n  %s  (%s, %d notes)\n", p.Name, p.Patch.Name, len(p.Notes))
		for start := 0; start < cols; {
			end := start
			for _, b := range barlines {
				if b > start && b-start <= scoreCols {
					end = b
				}
			}
			if end == start {
				end = min(start+scoreCols, cols)
			}
			end = min(end, cols)
			for k := hi; k >= lo; k-- {
				var b strings.Builder
				fmt.Fprintf(&b, "  %-4s|", synth.KeyName(k))
				for c := start; c < end; c++ {
					b.WriteByte(grid[k-lo][c])
					if isBar[c+1] {
						b.WriteByte('|')
					}
				}
				fmt.Println(b.String())
			}
			fmt.Println()
			start = end
		}
	}
}
