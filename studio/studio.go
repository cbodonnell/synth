// Package studio is the authoring loop as a command any game can wrap around its own
// songs and sounds: open the studio's window on them; render one to a wav, print its
// envelope, and play it; or print its piano roll, or each part's level, or a draft of it
// as an Arrangement; or bake them all for shipping.
package studio

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/bake"
	"github.com/cbodonnell/synth/studio/desk"
	"github.com/cbodonnell/synth/studio/session"
)

// Library is a game's music and sound: its arrangements, any songs written as code
// instead, and its sound effects.
type Library = session.Library

// Main parses the command line and runs it over the library.
func Main(lib Library) {
	edit := flag.Bool("edit", false, "open the studio's window, reloading as the source changes")
	child := flag.Bool("child", false, "run the window in this process (the supervisor's child)")
	handoff := flag.String("handoff", "", "where the window leaves a handoff when it has rebuilt itself")
	view := flag.String("view", "", "where the window opens, as a reload hands it on")
	cache := flag.Int64("cache", 256, "megabytes of memory the window's renders may keep")
	id := flag.String("id", "", "song or sound to render (see -list)")
	list := flag.Bool("list", false, "list songs and sounds")
	loops := flag.Int("loops", 2, "times through a looping song's body")
	dump := flag.Bool("dump", false, "write the wav without playing it")
	out := flag.String("out", os.TempDir(), "directory for wav files")
	score := flag.Bool("score", false, "print a piano roll instead of rendering")
	stems := flag.Bool("stems", false, "print each part's level instead of rendering the mix")
	bakeDir := flag.String("bake", "", "bake every song and sound into this directory")
	export := flag.Bool("export", false, "print a first draft of the song as a synth.Arrangement")
	pkg := flag.String("package", "songs", "the package an export is written for")
	flag.Parse()

	lookup := func(id string) (synth.Song, bool) {
		for _, s := range lib.All() {
			if s.ID == id {
				return s, true
			}
		}
		return synth.Song{}, false
	}

	switch {
	case *edit && !*child:
		bin, err := os.Executable()
		if err != nil {
			fail("%v", err)
		}
		os.Exit(session.Supervise(bin, os.Args[1:]))

	case *edit:
		v := session.View{ID: *id, Loop: true, Browser: true}
		if *view != "" {
			var err error
			if v, err = session.ParseView(*view); err != nil {
				fail("-view: %v", err)
			}
		}
		err := desk.Run(desk.Config{Library: lib, View: v, Handoff: *handoff, Cache: *cache << 20})
		if errors.Is(err, desk.ErrReload) {
			os.Exit(session.ReloadExit)
		}
		if err != nil {
			fail("%v", err)
		}

	case *bakeDir != "":
		began := time.Now()
		all := lib.All()
		if err := bake.Bake(*bakeDir, all...); err != nil {
			fail("%v", err)
		}
		fmt.Printf("baked %d into %s in %v\n", len(all), *bakeDir, time.Since(began).Round(time.Millisecond))

	case *list:
		for _, s := range lib.Music() {
			fmt.Printf("%-12s %-24s %5.1fs  loop from %.1fs\n", s.ID, s.Name, s.Length, s.LoopStart)
		}
		for _, s := range lib.Sounds {
			fmt.Printf("%-12s sound\n", s.ID)
		}

	case *id != "":
		s, ok := lookup(*id)
		if !ok {
			fail("unknown id %q; try -list", *id)
		}
		if err := synth.Validate(s); err != nil {
			fail("%s: %v", *id, err)
		}
		if *score {
			printScore(s)
			return
		}
		if *export {
			src, err := Source(Arrange(s), *pkg)
			if err != nil {
				fail("%s: %v", *id, err)
			}
			os.Stdout.Write(src)
			return
		}
		if *stems {
			printStems(s)
			return
		}

		began := time.Now()
		buf := render(s, *loops)
		took := time.Since(began)

		if err := os.MkdirAll(*out, 0o755); err != nil {
			fail("%v", err)
		}
		path := filepath.Join(*out, s.ID+".wav")
		if err := os.WriteFile(path, synth.WAV(buf), 0o644); err != nil {
			fail("%s: %v", *id, err)
		}
		printEnvelope(s.ID, buf, took)
		fmt.Println(path)

		if !*dump {
			if err := play(path); err != nil {
				fail("couldn't play %s: %v", path, err)
			}
		}

	default:
		flag.Usage()
		os.Exit(2)
	}
}

// A looping song renders as Ebiten will play it, so the seam is audible here first.
func render(s synth.Song, loops int) synth.Stereo {
	if !s.Loop {
		return synth.Render(s)
	}
	l := synth.RenderLoop(s)
	parts := []synth.Stereo{l.Intro}
	for range loops - 1 {
		parts = append(parts, l.Body)
	}
	return synth.Concat(parts...)
}

// Each part alone through its own inserts, buses, and the master, so the numbers are
// what that part contributes to the mix. RMS counts only the stretches where the part
// is sounding, so a part that plays for four bars of twenty isn't read as quiet.
func printStems(s synth.Song) {
	const window = synth.SampleRate / 20
	for _, p := range s.Parts {
		solo := s
		solo.Parts = []synth.Part{p}
		solo.Loop = false
		buf := synth.Render(solo)

		var sum float64
		var active int
		for start := 0; start+window <= buf.Len(); start += window {
			var w float64
			for i := start; i < start+window; i++ {
				w += float64(buf.L[i])*float64(buf.L[i]) + float64(buf.R[i])*float64(buf.R[i])
			}
			w /= 2 * window
			if w > 1e-5 {
				sum += w
				active++
			}
		}
		rms := -120.0
		if active > 0 {
			rms = 10 * math.Log10(sum/float64(active))
		}
		lufs, _ := synth.Loudness(buf)
		fmt.Printf("%-18s peak %5.2f  rms %6.1fdB  %6.1f LUFS  sounding %3.0f%%\n",
			p.Name, buf.Peak(), rms, lufs, 100*float64(active*window)/float64(max(buf.Len(), 1)))
	}
}

func play(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("afplay", path)
	case "linux":
		cmd = exec.Command("aplay", path)
	default:
		return fmt.Errorf("no player wired up for GOOS=%s", runtime.GOOS)
	}
	return cmd.Run()
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func printEnvelope(id string, s synth.Stereo, took time.Duration) {
	integrated, short := synth.Loudness(s)
	fmt.Printf("%s   %.2fs  peak %.2f  %.1f LUFS (short-term max %.1f)  rendered in %v\n",
		id, s.Seconds(), s.Peak(), integrated, short, took.Round(time.Millisecond))
	if s.Len() == 0 {
		return
	}

	const width, height = 64, 4
	env := make([]float32, width)
	bucket := max((s.Len()+width-1)/width, 1)
	for i := range s.L {
		b := i / bucket
		env[b] = max(env[b], abs(s.L[i]), abs(s.R[i]))
	}

	shades := []rune{' ', '░', '▒', '▓', '█'}
	for row := height - 1; row >= 0; row-- {
		lo, hi := float32(row)/height, float32(row+1)/height
		var line strings.Builder
		switch row {
		case height - 1:
			line.WriteString(" 1.0 ┤")
		case 0:
			line.WriteString(" 0.0 ┼")
		default:
			line.WriteString("     │")
		}
		for _, a := range env {
			switch {
			case a >= hi:
				line.WriteRune(shades[4])
			case a > lo:
				line.WriteRune(shades[1+int((a-lo)/(hi-lo)*3)])
			default:
				line.WriteRune(' ')
			}
		}
		fmt.Println(line.String())
	}
	fmt.Println("     └" + strings.Repeat("─", width))
}

func abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
