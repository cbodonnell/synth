//go:build !darwin

package desk

// Elsewhere a trackpad's pinch arrives as Ctrl and the wheel, which zooms already.
func listenPinch() {}

func takePinch() float64 { return 0 }
