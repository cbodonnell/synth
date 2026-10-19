//go:build !unix

package synth

import "errors"

func mapFloats(string, int) ([]float32, func(), error) {
	return nil, nil, errors.New("no mapped files here")
}
