package filereader

import (
	"github.com/google/wire"
)

func ProvideFileReaderFactory() FileReaderFactory {
	return NewFileReader
}

var WireSet = wire.NewSet(
	ProvideFileReaderFactory,
)
