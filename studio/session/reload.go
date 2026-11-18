package session

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

// Handoff is what a studio that has rebuilt itself leaves for the supervisor: the new
// build, and the arguments that open it where this one was.
type Handoff struct {
	Binary string
	Args   []string
}

func (h Handoff) Write(path string) error {
	b, err := json.Marshal(h)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

func ReadHandoff(path string) (Handoff, error) {
	var h Handoff
	b, err := os.ReadFile(path)
	if err != nil {
		return h, err
	}
	return h, json.Unmarshal(b, &h)
}

// ReloadExit is the exit code of a studio that has left a handoff.
const ReloadExit = 75

// Supervise runs the studio built at bin as a child with these arguments, then each
// build it hands off to, until one exits for another reason, and returns that exit code.
// The child is given -child and -handoff ahead of args.
func Supervise(bin string, args []string) int {
	dir, err := os.MkdirTemp("", "studio-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer os.RemoveAll(dir)
	handoff := filepath.Join(dir, "handoff.json")
	// Ctrl-C reaches the child too; this process waits for it to exit.
	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt)
	defer signal.Stop(interrupts)
	for {
		cmd := exec.Command(bin, append([]string{"-child", "-handoff", handoff}, args...)...)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		err := cmd.Run()
		var exit *exec.ExitError
		if err != nil && !errors.As(err, &exit) {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		if code := cmd.ProcessState.ExitCode(); code != ReloadExit {
			return code
		}
		h, err := ReadHandoff(handoff)
		if err != nil {
			fmt.Fprintln(os.Stderr, "reload:", err)
			return 1
		}
		bin, args = h.Binary, h.Args
	}
}

// BuildState is where a rebuild is.
type BuildState int

const (
	Watching BuildState = iota
	Building
	Failed
	Built
)

type Build struct {
	State  BuildState
	Output string // the compiler's, when it failed
	Binary string // the new build, when it's done
}

// Synth is synth's module, whose source the studio also watches.
const Synth = "github.com/cbodonnell/synth"

// Watcher rebuilds the running program whenever the source it was built from changes
// and settles: every package it imports from its own module and from synth's.
type Watcher struct {
	Builds chan Build
	main   string
	dirs   []string
	out    string

	mu    sync.Mutex
	own   map[string]string // the files the studio saved, as it left them
	again bool
}

// Sources finds the running program's main package and the directories of its source:
// every package it imports from its own module and from synth's.
func Sources() (main string, dirs []string, err error) {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Path == "" {
		return "", nil, errors.New("no build info to find the source from")
	}
	cmd := exec.Command("go", "list", "-deps", "-f", "{{if .Module}}{{.Module.Path}}\t{{.Dir}}{{end}}", info.Path)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	list, err := cmd.Output()
	if err != nil {
		return "", nil, fmt.Errorf("go list %s: %v: %s", info.Path, err, strings.TrimSpace(stderr.String()))
	}
	for line := range strings.SplitSeq(strings.TrimSpace(string(list)), "\n") {
		module, dir, ok := strings.Cut(line, "\t")
		if ok && (module == info.Main.Path || module == Synth) {
			dirs = append(dirs, dir)
		}
	}
	return info.Path, dirs, nil
}

// Watch watches the program's source, putting the builds in out.
func Watch(out string) (*Watcher, error) {
	main, dirs, err := Sources()
	if err != nil {
		return nil, err
	}
	return &Watcher{Builds: make(chan Build, 8), main: main, dirs: dirs, out: out, own: map[string]string{}}, nil
}

// Dirs are the directories of the program's source.
func (w *Watcher) Dirs() []string { return w.dirs }

// Saved notes a file the studio has saved. It already plays what it saved, so the file
// as it left it needs no rebuild.
func (w *Watcher) Saved(path string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.own[path] = files(filepath.Dir(path))[path]
}

// Rebuild builds again, as when a build waiting to be reloaded is missing a save.
func (w *Watcher) Rebuild() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.again = true
}

// changed reports whether the source differs from what was built, leaving out the
// studio's own saves.
func (w *Watcher) changed(built, now map[string]string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.again {
		w.again = false
		return true
	}
	for path, stat := range now {
		if built[path] != stat && w.own[path] != stat {
			return true
		}
	}
	for path := range built {
		if _, ok := now[path]; !ok {
			return true
		}
	}
	return false
}

// files are the Go files in the directories, each with its size and modification time,
// which change when it's saved. Tests are left out.
func files(dirs ...string) map[string]string {
	out := map[string]string{}
	for _, dir := range dirs {
		entries, _ := os.ReadDir(dir)
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			out[filepath.Join(dir, e.Name())] = fmt.Sprintf("%d %d", info.Size(), info.ModTime().UnixNano())
		}
	}
	return out
}

// Run watches until the process ends, reporting each build on Builds.
func (w *Watcher) Run() {
	built := files(w.dirs...)
	for n := 1; ; {
		time.Sleep(400 * time.Millisecond)
		now := files(w.dirs...)
		if !w.changed(built, now) {
			continue
		}
		// Wait for a save that writes several files to finish.
		for {
			time.Sleep(300 * time.Millisecond)
			again := files(w.dirs...)
			if maps.Equal(again, now) {
				break
			}
			now = again
		}
		built = now
		w.Builds <- Build{State: Building}
		binary := filepath.Join(w.out, fmt.Sprintf("studio-%d", n))
		n++
		output, err := exec.Command("go", "build", "-o", binary, w.main).CombinedOutput()
		if err != nil {
			os.Remove(binary)
			w.Builds <- Build{State: Failed, Output: strings.TrimSpace(string(output))}
			continue
		}
		// Source saved during the build isn't in it: build again.
		if !maps.Equal(files(w.dirs...), now) {
			os.Remove(binary)
			built = nil
			continue
		}
		w.Builds <- Build{State: Built, Binary: binary}
	}
}
