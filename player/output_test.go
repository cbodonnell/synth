package player

import (
	"testing"
	"time"

	"github.com/cbodonnell/synth"
)

func TestClockRunsOnToWhatWasDelivered(t *testing.T) {
	var c clock
	t0 := time.Now()
	if got := c.now(t0); got != 0 {
		t.Fatalf("before any report: %d, want 0", got)
	}
	c.deliver(4096)
	c.report(1024, t0)
	if got := c.now(t0.Add(10 * time.Millisecond)); got != 1024+synth.SampleRate/100 {
		t.Errorf("10ms on: %d, want %d", got, 1024+synth.SampleRate/100)
	}
	if got := c.now(t0.Add(time.Second)); got != 4096 {
		t.Errorf("a second on: %d, want the 4096 delivered", got)
	}
}
