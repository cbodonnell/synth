package player

import (
	"io"
	"sync"
	"time"

	"github.com/cbodonnell/synth"
	ebaudio "github.com/hajimehoshi/ebiten/v2/audio"
)

// output plays the bus from the moment it's made.
type output interface {
	// How many frames of the bus have been heard.
	played() int64
}

type ebitenOutput struct{ p *ebaudio.Player }

func newEbitenOutput(r io.Reader, buffer time.Duration) *ebitenOutput {
	p, err := ebaudio.NewContext(synth.SampleRate).NewPlayerF32(r)
	if err != nil {
		panic(err)
	}
	p.SetBufferSize(buffer)
	p.Play()
	return &ebitenOutput{p}
}

func (o *ebitenOutput) played() int64 {
	return int64(o.p.Position().Seconds() * synth.SampleRate)
}

// clock is a device's playhead as last reported, run on between reports by the wall
// clock but never past what has been handed to the device.
type clock struct {
	mu        sync.Mutex
	heard     int64
	delivered int64
	at        time.Time
}

func (c *clock) report(heard int64, at time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.heard, c.at = heard, at
}

func (c *clock) deliver(frames int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.delivered += int64(frames)
}

func (c *clock) now(t time.Time) int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.at.IsZero() {
		return 0
	}
	h := c.heard + int64(t.Sub(c.at).Seconds()*synth.SampleRate)
	return min(h, c.delivered)
}
