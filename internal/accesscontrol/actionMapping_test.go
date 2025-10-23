package accesscontrol

import (
	"errors"
	"fmt"
	"testing"

	"github.com/roppenlabs/dobby-service/internal/utils/filereader"
	"github.com/stretchr/testify/suite"
	gomock "go.uber.org/mock/gomock"
)

type ActionMapTestSuite struct {
	suite.Suite
	ctrl              *gomock.Controller
	mockFileReader    *filereader.MockFileReader
	actionMap         ActionMapping
	fileReaderFactory *filereader.MockFileReader
}

func TestActionMapTestSuite(t *testing.T) {
	suite.Run(t, &ActionMapTestSuite{})
}

func (a *ActionMapTestSuite) SetupTest() {
	a.ctrl = gomock.NewController(a.T())
	a.mockFileReader = filereader.NewMockFileReader(a.ctrl)

	mockFactory := func(storeType string, filePath string) (filereader.FileReader, error) {
		a.Assert().Equal("yaml", storeType)
		return a.mockFileReader, nil
	}

	a.fileReaderFactory = filereader.NewMockFileReader(a.ctrl)
	a.actionMap = NewActionMapping(mockFactory)

	expectedRBACFile := RBACFile{
		Modules: []Module{
			{
				Name: "kafka",
				Operations: []Operation{
					{ID: "op-create-topic", RBACAction: "write"},
					{ID: "op-delete-topic", RBACAction: "write"},
					{ID: "op-list-topics", RBACAction: "read"},
				},
			},
		},
	}

	a.mockFileReader.EXPECT().ReadFile(gomock.Any()).SetArg(0, expectedRBACFile).Return(nil)
	err := a.actionMap.LoadRBACMap("staticfiles/database.yaml")
	a.Require().NoError(err)
}

func (a *ActionMapTestSuite) TestLoadRBACMapFileNotFound() {
	invalidPath := "non_existent_file.yaml"
	mockFactory := func(storeType string, filePath string) (filereader.FileReader, error) {
		return nil, errors.New("file not found")
	}
	actionMap := NewActionMapping(mockFactory)
	err := actionMap.LoadRBACMap(invalidPath)
	a.Error(err)
	a.Contains(err.Error(), "failed to read RBAC file")
}

func (a *ActionMapTestSuite) TestLoadRBACMapInvalidFormat() {
	invalidPath := "invaliddatabase_test.yaml"
	mockFactory := func(storeType string, filePath string) (filereader.FileReader, error) {
		a.Assert().Equal("yaml", storeType)
		a.Assert().Equal(invalidPath, filePath)
		return a.mockFileReader, nil
	}
	a.mockFileReader.EXPECT().
		ReadFile(gomock.Any()).
		Return(errors.New("failed to parse YAML"))
	actionMap := NewActionMapping(mockFactory)
	err := actionMap.LoadRBACMap(invalidPath)
	a.Error(err)
	a.Contains(err.Error(), "failed to parse RBAC file")
}

func (a *ActionMapTestSuite) TestLoadRBACMap() {
	impl := a.actionMap.(*actionMappingImpl)
	fmt.Println("test: ", impl.rbacMap)
	var rbacMap = map[string]map[string]string{
		"kafka": {
			"op-create-topic": "write",
			"op-delete-topic": "write",
			"op-list-topics":  "read",
		},
	}

	var moduleActionOperationMap = map[string]map[string][]string{
		"kafka": {
			"read":  {"op-list-topics"},
			"write": {"op-create-topic", "op-delete-topic"},
		},
	}
	a.Equal(rbacMap, impl.rbacMap)
	a.Equal(moduleActionOperationMap, impl.moduleActionOperationMap)
}

func (a *ActionMapTestSuite) TestGetValidOpIdsRead() {
	module := "kafka"
	action := "read"
	ops, err := a.actionMap.GetValidOpIds(module, []string{action})
	a.NoError(err)
	a.True(ops["op-list-topics"])
	a.False(ops["op-create-topic"])
}

func (a *ActionMapTestSuite) TestGetValidOpIdsWrite() {
	module := "kafka"
	action := "write"
	ops, err := a.actionMap.GetValidOpIds(module, []string{action})
	a.NoError(err)
	a.False(ops["op-list-topics"])
	a.True(ops["op-create-topic"])
}

func (a *ActionMapTestSuite) TestGetValidOpIdsUndefinedAction() {
	module := "kafka"
	action := "undefined"
	message := fmt.Sprintf("action %q not found in module %q", action, module)
	_, err := a.actionMap.GetValidOpIds(module, []string{action})
	a.Error(err)
	a.Equal(message, err.Error())
}
func (a *ActionMapTestSuite) TestGetValidOpIdsUndefinedModule() {
	module := "undefined"
	action := "read"
	message := fmt.Sprintf("module %q not found", module)
	_, err := a.actionMap.GetValidOpIds(module, []string{action})
	a.Error(err)
	a.Equal(message, err.Error())
}

func (a *ActionMapTestSuite) TestGetValidAction() {
	module := "kafka"
	operation := "op-list-topics"
	action, err := a.actionMap.GetValidAction(module, operation)
	a.NoError(err)
	a.Equal("read", action)
}

func (a *ActionMapTestSuite) TestGetValidActionUndefinedModule() {
	module := "undefined"
	operation := "op-list-topics"
	message := fmt.Sprintf("module %q not found", module)
	_, err := a.actionMap.GetValidAction(module, operation)
	a.Error(err)
	a.Equal(message, err.Error())
}

func (a *ActionMapTestSuite) TestGetValidActionUndefinedOperation() {
	module := "kafka"
	operation := "undefined"
	message := fmt.Sprintf("operation ID %q not found in module %q", operation, module)
	_, err := a.actionMap.GetValidAction(module, operation)
	a.Error(err)
	a.Equal(message, err.Error())
}
