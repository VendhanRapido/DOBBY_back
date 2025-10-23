package implementation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type YamlReaderTestSuite struct {
	suite.Suite
	tempDir string
}

func (s *YamlReaderTestSuite) SetupTest() {
	s.tempDir = s.T().TempDir()
}

func TestYamlReaderTestSuite(t *testing.T) {
	suite.Run(t, new(YamlReaderTestSuite))
}

type TestConfig struct {
	Name    string   `yaml:"name"`
	Version string   `yaml:"version"`
	Items   []string `yaml:"items"`
	Nested  struct {
		Key   string `yaml:"key"`
		Value int    `yaml:"value"`
	} `yaml:"nested"`
}

func (s *YamlReaderTestSuite) TestReadFile_Success() {
	yamlContent := `name: test-app
version: 1.0.0
items:
  - item1
  - item2
  - item3
nested:
  key: test-key
  value: 42`

	filePath := filepath.Join(s.tempDir, "test.yaml")
	err := os.WriteFile(filePath, []byte(yamlContent), 0644)
	s.Require().NoError(err)

	reader := NewYamlReader(filePath)
	var config TestConfig
	err = reader.ReadFile(&config)

	s.Require().NoError(err)
	s.Assert().Equal("test-app", config.Name)
	s.Assert().Equal("1.0.0", config.Version)
	s.Assert().Len(config.Items, 3)
	s.Assert().Equal("item1", config.Items[0])
	s.Assert().Equal("test-key", config.Nested.Key)
	s.Assert().Equal(42, config.Nested.Value)
}

func (s *YamlReaderTestSuite) TestReadFile_FileNotFound() {
	reader := NewYamlReader("/path/to/nonexistent/file.yaml")
	var config TestConfig
	err := reader.ReadFile(&config)

	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "failed to open file")
}

func (s *YamlReaderTestSuite) TestReadFile_InvalidYAML() {
	invalidYAML := `name: test-app
version: 1.0.0
items:
  - item1
  - item2
  invalid yaml here: [`

	filePath := filepath.Join(s.tempDir, "invalid.yaml")
	err := os.WriteFile(filePath, []byte(invalidYAML), 0644)
	s.Require().NoError(err)

	reader := NewYamlReader(filePath)
	var config TestConfig
	err = reader.ReadFile(&config)

	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "error parsing YAML file")
}

func (s *YamlReaderTestSuite) TestReadFile_EmptyFile() {
	filePath := filepath.Join(s.tempDir, "empty.yaml")
	err := os.WriteFile(filePath, []byte(""), 0644)
	s.Require().NoError(err)

	reader := NewYamlReader(filePath)
	var config TestConfig
	err = reader.ReadFile(&config)
	s.Assert().NoError(err)
	s.Assert().Equal(TestConfig{}, config)
}

func (s *YamlReaderTestSuite) TestReadFile_ErrorReadingFile() {
	dirPath := filepath.Join(s.tempDir, "dir_as_file.yaml")
	err := os.Mkdir(dirPath, 0755)
	s.Require().NoError(err)

	reader := NewYamlReader(dirPath)
	var config TestConfig
	err = reader.ReadFile(&config)

	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "error reading file")
}
