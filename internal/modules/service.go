package modules

import (
	"log"
	"strings"

	"github.com/roppenlabs/dobby-service/internal/accesscontrol"
	"github.com/roppenlabs/dobby-service/internal/types"
	"github.com/roppenlabs/dobby-service/internal/utils/filereader"
)

var RESOURCES = "resources"

type Service interface {
	GetModules(userModules []string) ([]ModuleResponse, error)
	GetOperationDetail(moduleID, operationID string) (*OperationDetailResponse, error)
	GetFilteredOperations(
		moduleID string,
		allowedActions []string,
	) ([]Operations, error)
}

func NewService(actionMapper accesscontrol.ActionMapping, fileReaderFactory filereader.FileReaderFactory) Service {
	service := &serviceImpl{
		modules:           make(map[string]*Module),
		schemas:           make(map[string]*Schema),
		actionMapper:      actionMapper,
		fileReaderFactory: fileReaderFactory,
	}

	// Load data from YAML files
	if err := service.loadModules(); err != nil {
		log.Printf("Error loading modules: %v", err)
	}

	if err := service.loadSchemas(); err != nil {
		log.Printf("Error loading schemas: %v", err)
	}

	return service
}

func (s *serviceImpl) loadModules() error {
	fileReader, err := s.fileReaderFactory("yaml", DatabaseYAML_FILE_PATH)
	if err != nil {
		log.Printf("Failed to create file reader: %v", err)
		return err
	}

	var database DatabaseYAML
	err = fileReader.ReadFile(&database)
	if err != nil {
		log.Printf("Failed to read database.yaml: %v", err)
		return err
	}

	for _, module := range database.Modules {
		s.modules[module.ID] = &module
	}
	return nil
}

func (s *serviceImpl) loadSchemas() error {
	fileReader, err := s.fileReaderFactory("yaml", SchemaYAML_FILE_PATH)
	if err != nil {
		log.Printf("Failed to create file reader: %v", err)
		return err
	}

	var schemaData SchemaYAML
	err = fileReader.ReadFile(&schemaData)
	if err != nil {
		log.Printf("Failed to read schema.yaml: %v", err)
		return err
	}

	for _, schema := range schemaData.Schemas {
		s.schemas[schema.ID] = &schema
	}
	return nil
}

func (s *serviceImpl) GetModules(userModules []string) ([]ModuleResponse, error) {
	modules := make([]ModuleResponse, 0, len(s.modules))
	for _, module := range s.modules {
		modules = append(modules, ModuleResponse{
			ID:          module.ID,
			Name:        module.Name,
			Description: module.Description,
		})
	}
	allowed := make(map[string]struct{})
	for _, name := range userModules {
		allowed[strings.ToLower(name)] = struct{}{}
	}

	var filtered []ModuleResponse
	for _, m := range modules {
		if _, ok := allowed[strings.ToLower(m.Name)]; ok {
			filtered = append(filtered, m)
		}
	}

	return filtered, nil
}

func (s *serviceImpl) GetFilteredOperations(
	moduleID string,
	allowedActions []string,
) ([]Operations, error) {
	module, exists := s.modules[moduleID]
	if !exists {
		return nil, types.NewNotFoundError("module not found", MODULE_NOT_FOUND, "The requested module could not be found")
	}
	moduleName := module.Name

	allowedOpIDs, err := s.actionMapper.GetValidOpIds(moduleName, allowedActions)
	if err != nil {
		return nil, types.NewNotFoundError(err.Error(), ACTIONMAP_NOT_FOUND, "The requested module, action could not be found in the ActionMap")
	}

	filteredOps := []Operations{}
	for _, op := range module.Operations {
		if allowedOpIDs[op.ID] {
			filteredOps = append(filteredOps, op)
		}
	}

	return filteredOps, nil
}

type SchemaResponse struct {
	Fields []Field `yaml:"fields"`
}

func (s *serviceImpl) ResolveSchema(resolver *Resolver, module *Module, field *Field) []Option {
	var resources []Option
	for _, resource := range module.Resources {
		if resource.Type == field.Name {
			resources = append(resources, Option{
				Name: resource.Name,
				ID:   resource.ID,
			})
		}
	}
	return resources
}

func (s *serviceImpl) GetOperationDetail(moduleID, operationID string) (*OperationDetailResponse, error) {
	var formFields []FormField
	module, exists := s.modules[moduleID]
	if !exists {
		return nil, types.NewNotFoundError("module not found", MODULE_NOT_FOUND, "The requested module could not be found")
	}
	var foundOperation *Operations
	for _, operation := range module.Operations {
		if operation.ID == operationID {
			foundOperation = &operation
			break
		}
	}

	if foundOperation == nil {
		return nil, types.NewNotFoundError("operation not found", OPERATION_NOT_FOUND, "The requested operation could not be found")
	}
	schema, exists := s.schemas[foundOperation.SchemaId]
	if !exists {
		return nil, types.NewNotFoundError("schema not found", SCHEMA_NOT_FOUND, "The requested schema could not be found")
	}

	for _, field := range schema.Fields {
		if field.Resolver != nil {
			for _, resolver := range field.Resolver {
				formField := FormField{
					Name:        field.Name,
					DisplayName: field.DisplayName,
					Type:        field.Type,
					Validation:  field.Validations,
				}
				if resolver.Type == RESOURCES {
					options := s.ResolveSchema(&resolver, module, &field)
					formField.Options = options
				}
				formFields = append(formFields, formField)
			}
		} else {
			formField := FormField{
				Name:        field.Name,
				DisplayName: field.DisplayName,
				Type:        field.Type,
				Validation:  field.Validations,
			}
			formFields = append(formFields, formField)
		}
	}

	return &OperationDetailResponse{
		Name:        foundOperation.Name,
		Description: foundOperation.Description,
		Schema:      formFields,
	}, nil
}
