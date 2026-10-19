package session

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"
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
}

// Watch finds the program's main package and the directories of its source. The
// builds go in out.
func Watch(out string) (*Watcher, error) {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Path == "" {
		return nil, errors.New("no build info to rebuild from")
	}
	cmd := exec.Command("go", "list", "-deps", "-f", "{{if .Module}}{{.Module.Path}}\t{{.Dir}}{{end}}", info.Path)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	list, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go list %s: %v: %s", info.Path, err, strings.TrimSpace(stderr.String()))
	}
	w := &Watcher{Builds: make(chan Build, 8), main: info.Path, out: out}
	for line := range strings.SplitSeq(strings.TrimSpace(string(list)), "\n") {
		module, dir, ok := strings.Cut(line, "\t")
		if ok && (module == info.Main.Path || module == Synth) {
			w.dirs = append(w.dirs, dir)
		}
	}
	return w, nil
}

// Fingerprint names every Go file in the directories with its size and modification
// time, so it changes when one is saved, added or removed. Tests are left out.
func Fingerprint(dirs ...string) string {
	var lines []string
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
			lines = append(lines, fmt.Sprintf("%s %d %d", filepath.Join(dir, e.Name()), info.Size(), info.ModTime().UnixNano()))
		}
	}
	slices.Sort(lines)
	sum := sha256.Sum256([]byte(strings.Join(lines, "\n")))
	return hex.EncodeToString(sum[:])
}

// Run watches until the process ends, reporting each build on Builds.
func (w *Watcher) Run() {
	last := Fingerprint(w.dirs...)
	for n := 1; ; {
		time.Sleep(400 * time.Millisecond)
		now := Fingerprint(w.dirs...)
		if now == last {
			continue
		}
		// Wait for a save that writes several files to finish.
		for {
			time.Sleep(300 * time.Millisecond)
			again := Fingerprint(w.dirs...)
			if again == now {
				break
			}
			now = again
		}
		last = now
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
		if Fingerprint(w.dirs...) != now {
			os.Remove(binary)
			continue
		}
		w.Builds <- Build{State: Built, Binary: binary}
	}
}
