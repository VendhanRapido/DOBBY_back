package filereader

import (
	"fmt"
	"github.com/roppenlabs/dobby-service/internal/utils/filereader/implementation"
)

type FileReaderFactory func(storeType string, filePath string) (FileReader, error)

type FileReader interface {
	ReadFile(data interface{}) error
}

func NewFileReader(storeType string, filePath string) (FileReader, error) {
	switch storeType {
	case "yaml":
		return implementation.NewYamlReader(filePath), nil
	default:
		return nil, fmt.Errorf("unsupported store type: %s", storeType)
	}
}
