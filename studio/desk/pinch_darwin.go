package desk

import (
	"math"
	"sync/atomic"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

// Ebiten doesn't hear a trackpad's pinch on macOS, where it arrives as a magnify event
// rather than a wheel, so the window listens for those itself.
var magnified atomic.Uint64 // the float64 bits of the magnification not yet taken

// listenPinch adds a monitor of magnify events to the application, on the main thread
// once its run loop is going.
func listenPinch() {
	lib, err := purego.Dlopen("/usr/lib/libSystem.B.dylib", purego.RTLD_GLOBAL)
	if err != nil {
		return
	}
	mainQueue, err := purego.Dlsym(lib, "_dispatch_main_q")
	if err != nil {
		return
	}
	var dispatchAsync func(queue, context, work uintptr)
	purego.RegisterLibFunc(&dispatchAsync, lib, "dispatch_async_f")
	dispatchAsync(mainQueue, 0, purego.NewCallback(func(uintptr) {
		const maskMagnify = 1 << 30 // NSEventMaskMagnify
		monitor := objc.NewBlock(func(_ objc.Block, event objc.ID) objc.ID {
			m := objc.Send[float64](event, objc.RegisterName("magnification"))
			for {
				old := magnified.Load()
				if magnified.CompareAndSwap(old, math.Float64bits(math.Float64frombits(old)+m)) {
					return event
				}
			}
		})
		objc.ID(objc.GetClass("NSEvent")).Send(objc.RegisterName("addLocalMonitorForEventsMatchingMask:handler:"), uint64(maskMagnify), monitor)
	}))
}

// takePinch is how much the trackpad has been pinched open (positive) or shut since it
// was last asked.
func takePinch() float64 { return math.Float64frombits(magnified.Swap(0)) }
