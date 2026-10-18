//go:build js

package player

import (
	_ "embed"
	"fmt"
	"syscall/js"
	"time"

	"github.com/cbodonnell/synth"
)

//go:embed worklet.js
var workletJS string

// The worklet asks for this much of the bus at a time.
const workletChunk = 1024

// Ebiten's worklet asks for 8192 frames (170ms) at a time and pads what isn't buffered
// with silence, so its output needs more than that mixed ahead.
const ebitenWebBuffer = 250 * time.Millisecond

func newOutput(b *bus, buffer, latency time.Duration) output {
	if !js.Global().Get("AudioWorkletNode").Truthy() {
		return newEbitenOutput(b, max(buffer, ebitenWebBuffer))
	}
	return newWorkletOutput(b, latency)
}

// workletOutput plays the bus through an AudioWorklet of its own, which asks for a small
// chunk of it whenever it runs low, so a sound is heard about latency after it starts.
type workletOutput struct {
	bus   *bus
	ctx   js.Value
	port  js.Value
	clock clock
	buf   []byte
}

func newWorkletOutput(b *bus, latency time.Duration) *workletOutput {
	o := &workletOutput{bus: b, buf: make([]byte, workletChunk*bytesPerFrame)}
	opts := js.Global().Get("Object").New()
	opts.Set("sampleRate", synth.SampleRate)
	o.ctx = js.Global().Get("AudioContext").New(opts)

	blob := js.Global().Get("Blob").New([]any{workletJS}, map[string]any{"type": "text/javascript"})
	url := js.Global().Get("URL").Call("createObjectURL", blob)
	target := max(int(latency.Seconds()*synth.SampleRate), workletChunk)

	var loaded, failed js.Func
	loaded = js.FuncOf(func(js.Value, []js.Value) any {
		js.Global().Get("URL").Call("revokeObjectURL", url)
		node := js.Global().Get("AudioWorkletNode").New(o.ctx, "synth-output", map[string]any{
			"numberOfInputs":     0,
			"outputChannelCount": []any{2},
			"processorOptions":   map[string]any{"chunk": workletChunk, "target": target},
		})
		o.port = node.Get("port")
		o.port.Set("onmessage", js.FuncOf(func(_ js.Value, args []js.Value) any {
			d := args[0].Get("data")
			o.clock.report(int64(d.Get("played").Float()), time.Now())
			if u := d.Get("underran").Int(); u > 0 {
				warn(fmt.Sprintf("synth: the audio ran dry for %.0fms", float64(u)*1000/synth.SampleRate))
			}
			// Mixing can wait on the music's decoder, which a callback mustn't.
			go o.send()
			return nil
		}))
		node.Call("connect", o.ctx.Get("destination"))
		loaded.Release()
		failed.Release()
		return nil
	})
	failed = js.FuncOf(func(_ js.Value, args []js.Value) any {
		warn("synth: the audio worklet failed to load: " + args[0].Call("toString").String())
		loaded.Release()
		failed.Release()
		return nil
	})
	o.ctx.Get("audioWorklet").Call("addModule", url).Call("then", loaded, failed)
	o.follow()
	return o
}

func (o *workletOutput) send() {
	n, _ := o.bus.Read(o.buf)
	o.clock.deliver(n / bytesPerFrame)
	u8 := js.Global().Get("Uint8Array").New(n)
	js.CopyBytesToJS(u8, o.buf[:n])
	f32 := js.Global().Get("Float32Array").New(u8.Get("buffer"))
	o.port.Call("postMessage", f32, []any{u8.Get("buffer")})
}

func (o *workletOutput) played() int64 { return o.clock.now(time.Now()) }

// follow resumes the audio on a gesture, since a browser starts audio only on one.
func (o *workletOutput) follow() {
	doc := js.Global().Get("document")
	resume := js.FuncOf(func(js.Value, []js.Value) any {
		if o.ctx.Get("state").String() != "running" {
			o.ctx.Call("resume")
		}
		return nil
	})
	for _, e := range []string{"touchend", "keyup", "mouseup"} {
		doc.Call("addEventListener", e, resume)
	}
}

func warn(msg string) { js.Global().Get("console").Call("warn", msg) }
