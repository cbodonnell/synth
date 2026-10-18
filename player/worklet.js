// The bus's output in a browser: a ring of interleaved stereo frames, topped up from Go a
// chunk at a time whenever it holds less than the target.
class SynthOutput extends AudioWorkletProcessor {
	constructor(options) {
		super();
		const { chunk, target } = options.processorOptions;
		this.target = target;
		this.capacity = target + chunk;
		this.ring = new Float32Array(this.capacity * 2);
		this.read = 0;
		this.queued = 0;
		this.played = 0;
		this.underran = 0;
		this.started = false;
		this.waiting = false;
		this.port.onmessage = (e) => this.push(e.data);
		this.ask();
	}

	push(buf) {
		let w = (this.read + this.queued) % this.capacity;
		const frames = Math.min(buf.length / 2, this.capacity - this.queued);
		for (let i = 0; i < frames; i++) {
			this.ring[w * 2] = buf[i * 2];
			this.ring[w * 2 + 1] = buf[i * 2 + 1];
			w = (w + 1) % this.capacity;
		}
		this.queued += frames;
		this.started = true;
		this.waiting = false;
		this.ask();
	}

	ask() {
		if (this.waiting || this.queued >= this.target) {
			return;
		}
		this.waiting = true;
		this.port.postMessage({ played: this.played, underran: this.underran });
		this.underran = 0;
	}

	process(inputs, outputs) {
		const [l, r] = outputs[0];
		const n = l.length;
		const frames = Math.min(n, this.queued);
		for (let i = 0; i < frames; i++) {
			l[i] = this.ring[this.read * 2];
			r[i] = this.ring[this.read * 2 + 1];
			this.read = (this.read + 1) % this.capacity;
		}
		l.fill(0, frames);
		r.fill(0, frames);
		this.queued -= frames;
		this.played += frames;
		if (this.started) {
			this.underran += n - frames;
		}
		this.ask();
		return true;
	}
}

registerProcessor("synth-output", SynthOutput);
