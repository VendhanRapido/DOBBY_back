package testutils

import "errors"

type FaultyRequestBodyReader struct{}

func (FaultyRequestBodyReader) Read(bytesArray []byte) (n int, err error) {
	return 0, errors.New("faulty request body reader")
}
