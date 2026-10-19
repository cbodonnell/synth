package session

import (
	"math"
	"slices"

	"github.com/cbodonnell/synth"
)

// Layout is a song laid out on the timeline, in seconds.
type Layout struct {
	ID, Name string
	Song     synth.Song // compiled as it plays: muted tracks absent
	// What the mixer shows: the arrangement, or for a song that isn't one, its parts as
	// tracks without clips.
	Arrangement synth.Arrangement
	Tempo       float64
	Meter       []int
	Key         string
	Sections    []Section
	Lanes       []Lane
	// A song that isn't an arrangement: its parts are lanes, each one clip, and nothing
	// in it can be edited here.
	Custom bool
}

type Section struct {
	Name     string
	From, To float64
}

// Lane is a track: a part and the clips it plays.
type Lane struct {
	Name   string
	Out    string // the group bus it plays into, if any
	Level  float64
	Mute   bool
	Blocks []Block
}

// Block is one play of a clip, Times over, with the notes it plays there.
type Block struct {
	Clip     string
	From, To float64
	Hit      bool
	Notes    []synth.Note
}

// Arranged lays out an arrangement.
func Arranged(a synth.Arrangement) Layout {
	l := Layout{
		ID: a.ID, Name: a.Name, Song: a.Song(), Arrangement: a,
		Tempo: a.Tempo, Meter: a.Meter, Key: a.Key,
	}
	m := synth.Meter{Beat: 60 / math.Max(a.Tempo, 1e-9), Bars: a.Meter}

	bar := 0
	for _, s := range a.Sections {
		l.Sections = append(l.Sections, Section{Name: s.Name, From: BarTime(m, bar), To: BarTime(m, bar+s.Bars)})
		bar += s.Bars
	}

	// Every track sounding, so a muted one still shows its notes.
	all := a
	all.Tracks = slices.Clone(a.Tracks)
	for i := range all.Tracks {
		all.Tracks[i].Mute = false
	}
	parts := all.Song().Parts

	for i, t := range a.Tracks {
		lane := Lane{Name: t.Name, Out: t.Out, Level: t.Level, Mute: t.Mute}
		notes := parts[i].Notes
		for _, p := range t.Play {
			c, ok := clip(t, p.Clip)
			if !ok {
				continue
			}
			from := BarTime(m, p.Bar-1)
			to := BarTime(m, p.Bar-1+max(p.Times, 1)*clipBars(c))
			b := Block{Clip: p.Clip, From: from, To: to, Hit: c.Hit != ""}
			for _, n := range notes {
				if n.At >= from-early && n.At < to-early {
					b.Notes = append(b.Notes, n)
				}
			}
			lane.Blocks = append(lane.Blocks, b)
		}
		l.Lanes = append(l.Lanes, lane)
	}
	return l
}

// Humanizing can pull a note this far ahead of the bar its play begins on.
const early = 0.05

func clip(t synth.Track, name string) (synth.Clip, bool) {
	for _, c := range t.Clips {
		if c.Name == name {
			return c, true
		}
	}
	return synth.Clip{}, false
}

func clipBars(c synth.Clip) int {
	if len(c.Bars) > 0 {
		return len(c.Bars)
	}
	return max(c.Length, 1)
}

// Unarranged lays out a song that isn't an arrangement: each part a lane holding one
// clip of all its notes.
func Unarranged(s synth.Song) Layout {
	l := Layout{ID: s.ID, Name: s.Name, Song: s, Custom: true}
	l.Arrangement = synth.Arrangement{ID: s.ID, Name: s.Name, Buses: s.Buses, Master: s.Master, Target: s.Target}
	for _, p := range s.Parts {
		l.Arrangement.Tracks = append(l.Arrangement.Tracks, synth.Track{
			Name: p.Name, Patch: p.Patch, Level: p.Level, Pan: p.Pan,
			Inserts: p.Inserts, Sends: p.Sends, Out: p.Out,
		})
	}
	if Timed(s) {
		l.Tempo, l.Meter = 60/s.Meter.Beat, s.Meter.Bars
	}
	for _, p := range s.Parts {
		end := s.Length
		for _, n := range p.Notes {
			end = math.Max(end, n.At+n.Dur)
		}
		l.Lanes = append(l.Lanes, Lane{
			Name: p.Name, Out: p.Out, Level: p.Level,
			Blocks: []Block{{Clip: p.Name, To: end, Notes: p.Notes}},
		})
	}
	return l
}

// Length is how far the timeline runs: the song, or its last note if that's later.
func (l Layout) Length() float64 {
	end := l.Song.Length
	for _, lane := range l.Lanes {
		for _, b := range lane.Blocks {
			end = math.Max(end, b.To)
		}
	}
	return end
}

// Groups are the buses tracks play into, in the order their first track comes.
func (l Layout) Groups() []string {
	var out []string
	for _, lane := range l.Lanes {
		if lane.Out != "" && !slices.Contains(out, lane.Out) {
			out = append(out, lane.Out)
		}
	}
	return out
}
