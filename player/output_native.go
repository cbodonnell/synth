//go:build !js

package player

import "time"

func newOutput(b *bus, buffer, _ time.Duration) output { return newEbitenOutput(b, buffer) }
