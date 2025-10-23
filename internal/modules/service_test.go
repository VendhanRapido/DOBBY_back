package modules

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/roppenlabs/dobby-service/internal/accesscontrol"
	"github.com/roppenlabs/dobby-service/internal/types"
	"github.com/roppenlabs/dobby-service/internal/utils/filereader"
	"go.uber.org/mock/gomock"
)

type ServiceTestSuite struct {
	suite.Suite
	ctrl              *gomock.Controller
	mockService       *MockService
	mockEnforcer      *accesscontrol.MockAccessControlService
	mockActionMapping *accesscontrol.MockActionMapping
	mockFileReader    *filereader.MockFileReader
}

type testService struct {
	*serviceImpl
}

func newTestService(modules map[string]*Module, schemas map[string]*Schema, actionMapper *accesscontrol.MockActionMapping) *testService {
	return &testService{&serviceImpl{
		modules:      modules,
		schemas:      schemas,
		actionMapper: actionMapper,
	}}
}

func (s *ServiceTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockService = NewMockService(s.ctrl)
	s.mockEnforcer = accesscontrol.NewMockAccessControlService(s.ctrl)
	s.mockActionMapping = accesscontrol.NewMockActionMapping(s.ctrl)
	s.mockFileReader = filereader.NewMockFileReader(s.ctrl)
}

func (s *ServiceTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestMiddlewareTestSutie(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

func (s *ServiceTestSuite) TestNewService() {
	mockFactory := func(storeType string, filePath string) (filereader.FileReader, error) {
		s.Assert().Equal("yaml", storeType)
		return s.mockFileReader, nil
	}
	s.mockFileReader.EXPECT().ReadFile(gomock.Any()).Return(nil).AnyTimes()
	service := NewService(s.mockActionMapping, mockFactory)
	s.NotNil(service)
}
func (s *ServiceTestSuite) TestServiceImpl_GetOperations() {
	modules := map[string]*Module{
		"m1": {
			ID:          "m1",
			Name:        "mod1",
			Description: "desc",
			Operations: []Operations{
				{ID: "op1", Name: "Operation 1", Description: "desc", SchemaId: "schema1"},
			},
		},
	}
	allowedActions := []string{"read", "write"}

	s.mockActionMapping.
		EXPECT().
		GetValidOpIds("mod1", allowedActions).
		Return(map[string]bool{"op1": true}, nil)

	svc := newTestService(modules, map[string]*Schema{}, s.mockActionMapping)
	resp, err := svc.GetFilteredOperations("m1", allowedActions)
	s.Nil(err)
	s.Equal("op1", resp[0].ID)
}

func (s *ServiceTestSuite) TestServiceImpl_GetOperations_ModuleNotFound() {
	svc := newTestService(map[string]*Module{}, map[string]*Schema{}, s.mockActionMapping)
	allowedActions := []string{"read", "write"}
	_, err := svc.GetFilteredOperations("notfound", allowedActions)
	s.Equal(err, types.NewNotFoundError("module not found", MODULE_NOT_FOUND, "The requested module could not be found"))
}

func (s *ServiceTestSuite) TestGetModules() {
	modules := map[string]*Module{
		"m1": {ID: "m1", Name: "kafka", Description: "desc", Resources: nil, Operations: nil},
	}
	svc := newTestService(modules, map[string]*Schema{}, s.mockActionMapping)
	got, err := svc.GetModules([]string{"kafka"})
	s.Assert().Nil(err)
	s.Assert().Len(got, 1)
	s.Assert().Equal("m1", got[0].ID)
}

func (s *ServiceTestSuite) TestGetModulesEmpty() {
	svc := newTestService(map[string]*Module{}, map[string]*Schema{}, s.mockActionMapping)
	got, err := svc.GetModules([]string{"developer"})
	s.Assert().Nil(err)
	s.Assert().Len(got, 0)
}

func (s *ServiceTestSuite) TestServiceImpl_LoadSchemas_FactoryError() {
	mockFactory := func(storeType string, filePath string) (filereader.FileReader, error) {
		s.Equal("yaml", storeType)
		s.Equal(SchemaYAML_FILE_PATH, filePath)
		return nil, errors.New("failed to create file reader")
	}

	svc := &serviceImpl{
		modules:           map[string]*Module{},
		schemas:           map[string]*Schema{},
		fileReaderFactory: mockFactory,
	}

	err := svc.loadSchemas()

	s.Require().Error(err)
	s.Contains(err.Error(), "failed to create file reader")
	s.Len(svc.schemas, 0)
}

func (s *ServiceTestSuite) TestResolveSchema() {
	modules := map[string]*Module{
		"m1": {ID: "m1", Name: "mod1", Resources: []Resources{{ID: "r1", Name: "res1", Type: "f1"}}, Operations: nil},
	}
	schemas := map[string]*Schema{
		"s1": {ID: "s1", Fields: []Field{{Name: "f1", Type: "dropdown", Resolver: []Resolver{{Type: "resources"}}}}},
	}
	svc := newTestService(modules, schemas, s.mockActionMapping)
	schema := schemas["s1"]
	module := modules["m1"]
	field := schema.Fields[0]
	resolver := field.Resolver[0]
	opts := svc.ResolveSchema(&resolver, module, &field)
	s.Assert().Len(opts, 1)
	s.Assert().Equal("r1", opts[0].ID)
}

func (s *ServiceTestSuite) TestGetOperationDetail_Success() {
	modules := map[string]*Module{
		"m1": {
			ID: "m1", Name: "mod1", Operations: []Operations{{ID: "op1", Name: "op1", Description: "desc", SchemaId: "s1"}},
			Resources: []Resources{{ID: "r1", Name: "res1"}},
		},
	}
	schemas := map[string]*Schema{
		"s1": {
			ID: "s1",
			Fields: []Field{
				{
					Name: "f1",
					Type: "dropdown",
					Resolver: []Resolver{
						{
							Type: "resources",
						},
					},
					Validations: Validations{
						Required: true,
					},
				},
			},
		},
	}

	svc := newTestService(modules, schemas, s.mockActionMapping)
	resp, err := svc.GetOperationDetail("m1", "op1")
	s.Require().NoError(err)
	s.Assert().Equal("op1", resp.Name)
	s.Assert().Len(resp.Schema, 1)
	s.Assert().Equal("f1", resp.Schema[0].Name)
}

func (s *ServiceTestSuite) TestGetOperationDetail_ModuleNotFound() {
	svc := newTestService(map[string]*Module{}, map[string]*Schema{}, s.mockActionMapping)
	_, err := svc.GetOperationDetail("m1", "op1")
	s.Assert().Equal(types.NewNotFoundError("module not found", MODULE_NOT_FOUND, "The requested module could not be found"), err)
}

func (s *ServiceTestSuite) TestGetOperationDetail_OperationNotFound() {
	modules := map[string]*Module{
		"m1": {ID: "m1", Name: "mod1", Operations: nil},
	}
	svc := newTestService(modules, map[string]*Schema{}, s.mockActionMapping)
	_, err := svc.GetOperationDetail("m1", "op1")
	s.Assert().Equal(types.NewNotFoundError("operation not found", OPERATION_NOT_FOUND, "The requested operation could not be found"), err)
}

func (s *ServiceTestSuite) TestGetOperationDetail_SchemaNotFound() {
	modules := map[string]*Module{
		"m1": {ID: "m1", Name: "mod1", Operations: []Operations{{ID: "op1", Name: "op1", Description: "desc", SchemaId: "s1"}}},
	}
	svc := newTestService(modules, map[string]*Schema{}, s.mockActionMapping)
	_, err := svc.GetOperationDetail("m1", "op1")
	s.Assert().Equal(types.NewNotFoundError("schema not found", SCHEMA_NOT_FOUND, "The requested schema could not be found"), err)
}

func (s *ServiceTestSuite) TestServiceImpl_LoadModules_Success() {
	mockFileReader := filereader.NewMockFileReader(s.ctrl)

	mockFactory := func(storeType string, filePath string) (filereader.FileReader, error) {
		s.Assert().Equal("yaml", storeType)
		s.Assert().Equal(DatabaseYAML_FILE_PATH, filePath)
		return mockFileReader, nil
	}

	expectedDB := DatabaseYAML{
		Modules: []Module{
			{ID: "m1", Name: "mod1", Description: "desc"},
			{ID: "m2", Name: "mod2", Description: "desc2"},
		},
	}

	mockFileReader.EXPECT().
		ReadFile(gomock.Any()).
		DoAndReturn(func(data interface{}) error {
			db := data.(*DatabaseYAML)
			*db = expectedDB
			return nil
		})

	svc := &serviceImpl{
		modules:           make(map[string]*Module),
		schemas:           make(map[string]*Schema),
		fileReaderFactory: mockFactory,
	}

	err := svc.loadModules()
	s.Assert().NoError(err)
	s.Assert().Len(svc.modules, 2)
	s.Assert().Contains(svc.modules, "m1")
	s.Assert().Contains(svc.modules, "m2")
	s.Assert().Equal("mod1", svc.modules["m1"].Name)
}

func (s *ServiceTestSuite) TestServiceImpl_LoadModules_FactoryError() {
	mockFactory := func(storeType string, filePath string) (filereader.FileReader, error) {
		return nil, errors.New("failed to create file reader")
	}

	svc := &serviceImpl{
		modules:           make(map[string]*Module),
		schemas:           make(map[string]*Schema),
		fileReaderFactory: mockFactory,
	}

	err := svc.loadModules()
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "failed to create file reader")
	s.Assert().Len(svc.modules, 0)
}

func (s *ServiceTestSuite) TestServiceImpl_LoadModules_ReadError() {
	mockFileReader := filereader.NewMockFileReader(s.ctrl)

	mockFactory := func(storeType string, filePath string) (filereader.FileReader, error) {
		return mockFileReader, nil
	}

	mockFileReader.EXPECT().
		ReadFile(gomock.Any()).
		Return(errors.New("failed to read file"))

	svc := &serviceImpl{
		modules:           make(map[string]*Module),
		schemas:           make(map[string]*Schema),
		fileReaderFactory: mockFactory,
	}

	err := svc.loadModules()
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "failed to read file")
	s.Assert().Len(svc.modules, 0)
}

func (s *ServiceTestSuite) TestServiceImpl_LoadSchemas_Success() {
	mockFileReader := filereader.NewMockFileReader(s.ctrl)

	mockFactory := func(storeType string, filePath string) (filereader.FileReader, error) {
		s.Assert().Equal("yaml", storeType)
		s.Assert().Equal(SchemaYAML_FILE_PATH, filePath)
		return mockFileReader, nil
	}

	expectedSchemas := SchemaYAML{
		Schemas: []Schema{
			{ID: "s1", Fields: []Field{{Name: "f1", Type: "string"}}},
			{ID: "s2", Fields: []Field{{Name: "f2", Type: "number"}}},
		},
	}

	mockFileReader.EXPECT().
		ReadFile(gomock.Any()).
		DoAndReturn(func(data interface{}) error {
			schemas := data.(*SchemaYAML)
			*schemas = expectedSchemas
			return nil
		})

	svc := &serviceImpl{
		modules:           make(map[string]*Module),
		schemas:           make(map[string]*Schema),
		fileReaderFactory: mockFactory,
	}

	err := svc.loadSchemas()
	s.Assert().NoError(err)
	s.Assert().Len(svc.schemas, 2)
	s.Assert().Contains(svc.schemas, "s1")
	s.Assert().Contains(svc.schemas, "s2")
}

func (s *ServiceTestSuite) TestServiceImpl_LoadSchemas_ReadError() {
	mockFileReader := filereader.NewMockFileReader(s.ctrl)

	mockFactory := func(storeType string, filePath string) (filereader.FileReader, error) {
		return mockFileReader, nil
	}

	mockFileReader.EXPECT().
		ReadFile(gomock.Any()).
		Return(errors.New("failed to parse YAML"))

	svc := &serviceImpl{
		modules:           make(map[string]*Module),
		schemas:           make(map[string]*Schema),
		fileReaderFactory: mockFactory,
	}

	err := svc.loadSchemas()
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "failed to parse YAML")
	s.Assert().Len(svc.schemas, 0)
}

func (s *ServiceTestSuite) TestGetOperationDetail_FieldNoResolver() {
	modules := map[string]*Module{
		"m1": {
			ID: "m1", Name: "mod1", Operations: []Operations{{ID: "op1", Name: "op1", Description: "desc", SchemaId: "s1"}},
			Resources: []Resources{{ID: "r1", Name: "res1"}},
		},
	}
	schemas := map[string]*Schema{
		"s1": {ID: "s1", Fields: []Field{{Name: "f1", Type: "string", Resolver: nil}}},
	}
	svc := newTestService(modules, schemas, s.mockActionMapping)
	resp, err := svc.GetOperationDetail("m1", "op1")
	s.Require().NoError(err)
	s.Assert().Len(resp.Schema, 1)
	s.Assert().Equal("f1", resp.Schema[0].Name)
	s.Assert().Equal("string", resp.Schema[0].Type)
}

func (s *ServiceTestSuite) TestServiceImpl_GetFilteredOperations_ModuleNotFound() {
	moduleID := "non-existent-module"
	actions := []string{"read"}

	svc := &serviceImpl{
		modules:      map[string]*Module{},
		actionMapper: s.mockActionMapping,
	}

	result, err := svc.GetFilteredOperations(moduleID, actions)

	s.Nil(result)
	s.Require().Error(err)

	expected := types.NewNotFoundError("module not found", MODULE_NOT_FOUND, "The requested module could not be found")
	s.Equal(expected, err)
}

func (s *ServiceTestSuite) TestServiceImpl_GetFilteredOperations_ActionMappingError() {
	moduleID := "m1"
	moduleName := "mod1"
	actions := []string{"read"}

	mod := &Module{
		ID:         moduleID,
		Name:       moduleName,
		Operations: []Operations{{ID: "op1", Name: "Operation 1"}},
	}

	s.mockActionMapping.EXPECT().
		GetValidOpIds(moduleName, actions).
		Return(nil, errors.New("action mapping failed"))

	svc := &serviceImpl{
		modules:      map[string]*Module{moduleID: mod},
		actionMapper: s.mockActionMapping,
	}

	result, err := svc.GetFilteredOperations(moduleID, actions)

	s.Nil(result)
	s.Require().Error(err)

	expected := types.NewNotFoundError("action mapping failed", ACTIONMAP_NOT_FOUND, "The requested module, action could not be found in the ActionMap")
	s.Equal(expected, err)
}

func (s *ServiceTestSuite) TestServiceImpl_GetFilteredOperations_Success() {
	moduleID := "m1"
	moduleName := "mod1"
	actions := []string{"read", "write"}

	module := &Module{
		ID:   moduleID,
		Name: moduleName,
		Operations: []Operations{
			{ID: "op1", Name: "Operation-1"},
			{ID: "op2", Name: "Operation-2"},
			{ID: "op3", Name: "Operation-3"},
			{ID: "op4", Name: "Operation-4"},
		},
	}

	s.mockActionMapping.EXPECT().
		GetValidOpIds(moduleName, actions).
		Return(map[string]bool{
			"op1": true,
			"op2": true,
		}, nil)

	svc := &serviceImpl{
		modules:      map[string]*Module{moduleID: module},
		actionMapper: s.mockActionMapping,
	}

	result, err := svc.GetFilteredOperations(moduleID, actions)

	s.Require().NoError(err)
	s.Assert().Len(result, 2)

	var ids []string
	for _, op := range result {
		ids = append(ids, op.ID)
	}
	s.Assert().ElementsMatch([]string{"op1", "op2"}, ids)
}

func (s *ServiceTestSuite) TestServiceImpl_GetFilteredOperations_NoAccess() {
	moduleID := "m1"
	moduleName := "mod1"
	actions := []string{"read"}

	mod := &Module{
		ID:   moduleID,
		Name: moduleName,
		Operations: []Operations{
			{ID: "op1", Name: "Operation 1"},
			{ID: "op2", Name: "Operation 2"},
		},
	}

	s.mockActionMapping.EXPECT().
		GetValidOpIds(moduleName, actions).
		Return(map[string]bool{}, nil)

	svc := &serviceImpl{
		modules:      map[string]*Module{moduleID: mod},
		actionMapper: s.mockActionMapping,
	}

	result, err := svc.GetFilteredOperations(moduleID, actions)

	s.Require().NoError(err)
	s.Assert().Empty(result)
}

func (s *ServiceTestSuite) TestGetOperationDetail_Validations_TopicName() {
	modules := map[string]*Module{
		"m1": {
			ID:   "m1",
			Name: "mod1",
			Operations: []Operations{
				{ID: "op1", Name: "op1", Description: "desc", SchemaId: "schema-2"},
			},
			Resources: []Resources{{ID: "r1", Name: "res1", Type: "clusterId"}},
		},
	}
	schemas := map[string]*Schema{
		"schema-2": {
			ID: "schema-2",
			Fields: []Field{
				{
					Name:        "clusterId",
					DisplayName: "Cluster ID",
					Type:        "dropdown",
					Resolver:    []Resolver{{Type: "resources", ID: "rid1"}},
					Validations: Validations{Required: true},
				},
				{
					Name:        "topicName",
					DisplayName: "Topic Name",
					Type:        "string",
					Validations: Validations{
						Required:     true,
						Min:          5,
						Max:          15,
						Pattern:      "^[a-zA-Z][a-zA-Z0-9_-]*$",
						ErrorMessage: "The input must start with a letter and can only contain letters, numbers, underscores, and hyphens.",
					},
				},
			},
		},
	}

	svc := newTestService(modules, schemas, s.mockActionMapping)
	resp, err := svc.GetOperationDetail("m1", "op1")
	s.NoError(err)
	s.NotNil(resp)
	s.Equal("op1", resp.Name)
	s.Len(resp.Schema, 2)
	topicNameField := resp.Schema[1]
	s.Equal("topicName", topicNameField.Name)
	s.Equal("string", topicNameField.Type)
	s.True(topicNameField.Validation.Required)
	s.Equal(5, topicNameField.Validation.Min)
	s.Equal(15, topicNameField.Validation.Max)
	s.Equal("^[a-zA-Z][a-zA-Z0-9_-]*$", topicNameField.Validation.Pattern)
	s.Equal("The input must start with a letter and can only contain letters, numbers, underscores, and hyphens.", topicNameField.Validation.ErrorMessage)
}

func (s *ServiceTestSuite) TestGetOperationDetail_Validations_NumPartitions() {
	modules := map[string]*Module{
		"m1": {
			ID:   "m1",
			Name: "mod1",
			Operations: []Operations{
				{ID: "op1", Name: "op1", Description: "desc", SchemaId: "schema-2"},
			},
			Resources: []Resources{{ID: "r1", Name: "res1", Type: "clusterId"}},
		},
	}
	schemas := map[string]*Schema{
		"schema-2": {
			ID: "schema-2",
			Fields: []Field{
				{
					Name:        "numPartitions",
					DisplayName: "Number of Partitions",
					Type:        "number",
					Validations: Validations{
						Required: true,
						Min:      1,
						Max:      10,
					},
				},
			},
		},
	}

	svc := newTestService(modules, schemas, s.mockActionMapping)
	resp, err := svc.GetOperationDetail("m1", "op1")
	s.NoError(err)
	s.NotNil(resp)
	s.Equal("op1", resp.Name)
	s.Len(resp.Schema, 1)

	// Validate numPartitions field validations
	numPartitionsField := resp.Schema[0]
	s.Equal("numPartitions", numPartitionsField.Name)
	s.Equal("number", numPartitionsField.Type)
	s.True(numPartitionsField.Validation.Required)
	s.Equal(1, numPartitionsField.Validation.Min)
	s.Equal(10, numPartitionsField.Validation.Max)
}

func (s *ServiceTestSuite) TestGetOperationDetail_Validations_RetentionMs() {
	modules := map[string]*Module{
		"m1": {
			ID:   "m1",
			Name: "mod1",
			Operations: []Operations{
				{ID: "op1", Name: "op1", Description: "desc", SchemaId: "schema-2"},
			},
			Resources: []Resources{{ID: "r1", Name: "res1", Type: "clusterId"}},
		},
	}
	schemas := map[string]*Schema{
		"schema-2": {
			ID: "schema-2",
			Fields: []Field{
				{
					Name:        "retentionMs",
					DisplayName: "Retention MS",
					Type:        "number",
					Validations: Validations{
						Required: true,
						Min:      6000,
						Max:      8000,
					},
				},
			},
		},
	}

	svc := newTestService(modules, schemas, s.mockActionMapping)
	resp, err := svc.GetOperationDetail("m1", "op1")
	s.NoError(err)
	s.NotNil(resp)
	s.Equal("op1", resp.Name)
	s.Len(resp.Schema, 1)

	// Validate retentionMs field validations
	retentionMsField := resp.Schema[0]
	s.Equal("retentionMs", retentionMsField.Name)
	s.Equal("number", retentionMsField.Type)
	s.True(retentionMsField.Validation.Required)
	s.Equal(6000, retentionMsField.Validation.Min)
	s.Equal(8000, retentionMsField.Validation.Max)
}
