package session

import (
	"math"
	"os"
	"testing"
	"time"

	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/patches"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeChild makes the test binary stand in for the studio under Supervise: started as
// "first" it hands off to itself as "second", which exits with 7.
const fakeChild = "STUDIO_FAKE_CHILD"

func TestMain(m *testing.M) {
	if os.Getenv(fakeChild) != "" {
		args := os.Args[1:]
		if len(args) != 4 || args[0] != "-child" || args[1] != "-handoff" {
			os.Exit(3)
		}
		switch args[3] {
		case "first":
			if err := (Handoff{Binary: os.Args[0], Args: []string{"second"}}).Write(args[2]); err != nil {
				os.Exit(4)
			}
			os.Exit(ReloadExit)
		case "second":
			os.Exit(7)
		}
		os.Exit(5)
	}
	os.Exit(m.Run())
}

func TestSuperviseFollowsHandoffs(t *testing.T) {
	t.Setenv(fakeChild, "1")
	assert.Equal(t, 7, Supervise(os.Args[0], []string{"first"}), "the second build's exit")
}

func waltz() synth.Arrangement {
	return synth.Arrangement{
		ID: "waltz", Tempo: 120, Meter: []int{3},
		Sections: []synth.Section{{Name: "A", Bars: 2}, {Name: "B", Bars: 4}},
		Loop:     "B",
		Tracks: []synth.Track{
			{
				Name: "bell", Patch: patches.Bell(),
				Clips: []synth.Clip{{Name: "up", Step: 1, Bars: []string{"C5 E5 G5"}}},
				Play:  []synth.Play{{Clip: "up", Bar: 1, Times: 6}},
			},
			{
				Name: "bass", Patch: patches.Bass(), Mute: true,
				Clips: []synth.Clip{{Name: "root", Step: 1, Bars: []string{"C2 - -", "G1 - -"}}},
				Play:  []synth.Play{{Clip: "root", Bar: 3, Times: 2}},
			},
		},
	}
}

func TestTransportCountsBarsBeatsAndSteps(t *testing.T) {
	s := waltz().Song()
	assert.Equal(t, "1.1.1", Format(s, 0))
	assert.Equal(t, "2.1.1", Format(s, 1.5))
	assert.Equal(t, "2.3.3", Format(s, 2.75))
	assert.Equal(t, 6, Bars(s))
	assert.Equal(t, 6.0, Seam(s), "two bars before the end")
}

func TestLayoutPlacesPlaysAtTheirBars(t *testing.T) {
	l := Arranged(waltz())
	require.Len(t, l.Lanes, 2)
	assert.Equal(t, []Section{{"A", 0, 3}, {"B", 3, 9}}, l.Sections)
	bass := l.Lanes[1]
	assert.True(t, bass.Mute)
	require.Len(t, bass.Blocks, 1)
	assert.Equal(t, 3.0, bass.Blocks[0].From)
	assert.Equal(t, 9.0, bass.Blocks[0].To)
	assert.Len(t, bass.Blocks[0].Notes, 4, "a muted track still shows its notes")
	assert.Len(t, l.Song.Parts, 1, "but isn't heard")
}

func TestTakeMapsTheLoopBackOntoTheSong(t *testing.T) {
	s := waltz().Song()
	take := Render(synth.NewRenderer(), s, Mix{Loop: true})
	start, end := take.Region()
	assert.Equal(t, take.Intro, start)
	assert.Equal(t, take.Buffer.Len(), end)
	assert.InDelta(t, s.LoopStart+1, take.Time(take.Intro+synth.SampleRate), 1e-9)
	assert.InDelta(t, 2.0, take.Time(take.Frame(2)), 1e-9)

	again := Render(synth.NewRenderer(), s, Mix{Loop: true, Click: true})
	assert.Equal(t, again.Intro+100, again.Follow(take, take.Intro+100), "a re-render stays in the body")

	once := Render(synth.NewRenderer(), s, Mix{})
	assert.Equal(t, once.Frame(s.LoopStart)+100, once.Follow(take, take.Intro+100), "and leaves it to play through")
	start, end = once.Region()
	assert.Zero(t, end-start, "a song played through doesn't loop")
	assert.GreaterOrEqual(t, once.Buffer.Len(), int(s.Length*synth.SampleRate))
}

func TestSoloAndClickAreTheStudios(t *testing.T) {
	a := waltz()
	a.Tracks[1].Mute = false
	s := a.Song()
	r := synth.NewRenderer()
	solo := Render(r, s, Mix{Solo: []string{"bass"}})
	require.Len(t, solo.Song.Parts, 1)
	assert.Zero(t, solo.Parts["bass"].Peak(int(0.5*synth.SampleRate), 0.1), "the bass waits for bar 3")
	assert.NotZero(t, solo.Parts["bass"].Peak(int(4*synth.SampleRate), 0.1))

	plain := Render(r, s, Mix{})
	click := Render(r, s, Mix{Click: true})
	assert.NotEqual(t, plain.Buffer, click.Buffer)
	assert.Equal(t, plain.Loudness, click.Loudness, "the click isn't measured")
	assert.False(t, math.IsInf(plain.Loudness, 0))
}

func TestRendersKeepsOnlyTheLatest(t *testing.T) {
	b := NewRenders()
	s := waltz().Song()
	b.Request(s, Mix{})
	b.Request(s, Mix{Loop: true})
	require.Eventually(t, func() bool {
		busy, _ := b.Busy()
		return !busy
	}, 10*time.Second, 10*time.Millisecond)
	take, ok := b.Done()
	require.True(t, ok)
	assert.True(t, take.Loops)
	_, ok = b.Done()
	assert.False(t, ok, "once")
}
