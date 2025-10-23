package filereader

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type FileReaderTestSuite struct {
	suite.Suite
	tempDir string
}

func (s *FileReaderTestSuite) SetupTest() {
	s.tempDir = s.T().TempDir()
}

func TestFileReaderTestSuite(t *testing.T) {
	suite.Run(t, new(FileReaderTestSuite))
}

func (s *FileReaderTestSuite) TestNewFileReader_YamlType() {
	reader, err := NewFileReader("yaml", "test.yaml")
	s.Require().NoError(err)
	s.Assert().NotNil(reader)
}

func (s *FileReaderTestSuite) TestNewFileReader_UnsupportedType() {
	reader, err := NewFileReader("json", "test.json")
	s.Assert().Error(err)
	s.Assert().Nil(reader)
	s.Assert().Contains(err.Error(), "unsupported store type: json")
}

func (s *FileReaderTestSuite) TestNewFileReader_EmptyType() {
	reader, err := NewFileReader("", "test.yaml")
	s.Assert().Error(err)
	s.Assert().Nil(reader)
	s.Assert().Contains(err.Error(), "unsupported store type:")
}
