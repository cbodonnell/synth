package session

import (
	"reflect"
	"testing"

	"github.com/cbodonnell/synth"
	"github.com/cbodonnell/synth/songs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func copyOf(a synth.Arrangement) synth.Arrangement {
	return clone(reflect.ValueOf(a)).Interface().(synth.Arrangement)
}

func TestSectionsMoveWhatFollowsThem(t *testing.T) {
	for _, a := range songs.Arrangements() {
		want := a.Song()
		beat := 60 / a.Tempo
		for i, s := range a.Sections {
			b := copyOf(a)
			n := NewSectionBars(b)
			InsertSection(&b, i, n)
			require.NoError(t, b.Validate(), "%s: a section inserted before %s", a.ID, s.Name)
			grown := b.Song()
			m := synth.Meter{Beat: 1, Bars: a.Meter}
			assert.InDelta(t, want.Length+m.BarStart(n)*beat, grown.Length, 1e-9, a.ID)
			require.NoError(t, DeleteSection(&b, i))
			assert.Equal(t, want, b.Song(), "%s: %s inserted and deleted", a.ID, s.Name)

			b = copyOf(a)
			require.NoError(t, ResizeSection(&b, i, s.Bars+SectionBars(b)))
			require.NoError(t, b.Validate(), "%s: %s grown", a.ID, s.Name)
			require.NoError(t, ResizeSection(&b, i, s.Bars))
			assert.Equal(t, want, b.Song(), "%s: %s grown and shrunk", a.ID, s.Name)

			b = copyOf(a)
			DuplicateSection(&b, i)
			if assert.NoError(t, b.Validate(), "%s: %s duplicated", a.ID, s.Name) {
				from, to := BarTime(want.Meter, SectionBar(a, i)-1), BarTime(want.Meter, SectionBar(a, i)-1+s.Bars)
				assert.InDelta(t, want.Length+to-from, b.Song().Length, 1e-9, a.ID)
			}
		}
	}
}

func TestDeletingTheLoopsSectionLoopsFromTheNext(t *testing.T) {
	a := copyOf(waltz())
	require.NoError(t, DeleteSection(&a, 1))
	assert.Equal(t, "", a.Loop)
	a = copyOf(waltz())
	InsertSection(&a, 2, 3)
	require.NoError(t, RenameSection(&a, 1, "Verse"))
	assert.Equal(t, "Verse", a.Loop)
	require.NoError(t, DeleteSection(&a, 1))
	assert.Equal(t, a.Sections[1].Name, a.Loop)
	assert.Error(t, RenameSection(&a, 0, a.Sections[1].Name))
}

func TestArrangingKeepsOnlyWhatValidates(t *testing.T) {
	d, _ := fixture(t)
	bell := d.Now().Tracks[0].Play[0]

	assert.Error(t, d.Arrange(func(a *synth.Arrangement) error {
		a.Tracks[1].Play[0].Bar = 4
		return nil
	}), "runs past the end")
	assert.ErrorIs(t, d.Arrange(func(a *synth.Arrangement) error {
		a.Tracks[0].Level = -3
		return nil
	}), ErrCode, "a var")
	assert.False(t, d.Dirty())

	require.NoError(t, d.Arrange(func(a *synth.Arrangement) error {
		SetTimes(a, 0, 0, 3)
		RepeatPlay(a, 0, 0)
		return nil
	}))
	plays := d.Now().Tracks[0].Play
	require.Len(t, plays, 2)
	assert.Equal(t, bell.Bar+3, plays[1].Bar)
	assert.Error(t, d.Arrange(func(a *synth.Arrangement) error {
		a.Tracks[0].Play[1].Bar = 2
		return nil
	}), "overlaps")

	require.NoError(t, d.Arrange(func(a *synth.Arrangement) error {
		MakeUnique(a, 0, 1)
		DeletePlay(a, 0, 0)
		return nil
	}))
	tr := d.Now().Tracks[0]
	assert.Len(t, tr.Clips, 1, "a clip nothing plays goes with its last play")
	assert.Equal(t, tr.Clips[0].Name, tr.Play[0].Clip)

	require.NoError(t, d.Arrange(func(a *synth.Arrangement) error {
		AddClip(a, 1, 1)
		return nil
	}))
	_, err := d.Source()
	require.NoError(t, err)
}

func TestMetersAndTemposReadAsTheyreWritten(t *testing.T) {
	for _, m := range [][]int{nil, {3}, {2}, {3, 3, 3, 4}, {7}} {
		got, ok := ParseMeter(FormatMeter(m))
		require.True(t, ok, FormatMeter(m))
		assert.Equal(t, m, got)
	}
	for _, s := range []string{"", "0/4", "3+", "x", "5/8"} {
		_, ok := ParseMeter(s)
		assert.False(t, ok, s)
	}
	v, ok := ParseTempo(" 92.504 bpm")
	assert.True(t, ok)
	assert.Equal(t, 92.5, v)
	_, ok = ParseTempo("0")
	assert.False(t, ok)
}
