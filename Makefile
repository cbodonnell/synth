.PHONY: test vet list play score export edit demo

test:
	go test ./...

vet:
	go vet ./...

list:
	go run ./cmd/synth -list

# Render a song or sound to a wav, print its envelope, and play it.
# `make play ID=goldenbrown LOOPS=3 DUMP=1` writes three passes without playing them.
play:
	go run ./cmd/synth -id $(ID) $(if $(LOOPS),-loops $(LOOPS)) $(if $(DUMP),-dump)

score:
	go run ./cmd/synth -id $(ID) -score

# Print a first draft of a song as a synth.Arrangement.
export:
	go run ./cmd/synth -id $(ID) -export

# Open the studio's window, reloading whenever the songs' source changes.
edit:
	go run ./cmd/synth -edit $(if $(ID),-id $(ID))

demo:
	go run ./cmd/demo
