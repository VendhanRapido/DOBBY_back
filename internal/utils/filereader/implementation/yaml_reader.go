package implementation

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

type YamlReader struct {
	filePath string
}

func NewYamlReader(filePath string) *YamlReader {
	return &YamlReader{
		filePath: filePath,
	}
}

func (y *YamlReader) ReadFile(data interface{}) error {
	file, err := os.Open(y.filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", y.filePath, err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", y.filePath, err)
	}

	err = yaml.Unmarshal(bytes, data)
	if err != nil {
		return fmt.Errorf("error parsing YAML file %s: %w", y.filePath, err)
	}

	return nil
}
