package modules

import (
	"github.com/roppenlabs/dobby-service/internal/accesscontrol"
	"github.com/roppenlabs/dobby-service/internal/utils/filereader"
)

type Module struct {
	ID          string       `yaml:"id" json:"id"`
	Name        string       `yaml:"name" json:"name"`
	Description string       `yaml:"description" json:"description"`
	Resources   []Resources  `yaml:"resources" json:"resources"`
	Operations  []Operations `yaml:"operations" json:"operations"`
}

type ModuleResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Resources struct {
	Type string `yaml:"type" json:"type"`
	ID   string `yaml:"id" json:"id"`
	Name string `yaml:"name" json:"name"`
}

type Operations struct {
	ID          string `yaml:"id" json:"id"`
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"desc" json:"description"`
	SchemaId    string `yaml:"schemaId" json:"schemaId"`
}

type Schema struct {
	ID     string  `yaml:"id" json:"id"`
	Fields []Field `yaml:"fields" json:"fields"`
}

type Field struct {
	Name        string      `yaml:"name" json:"name"`
	DisplayName string      `yaml:"displayName" json:"displayName"`
	Type        string      `yaml:"type" json:"type"`
	Description string      `yaml:"description" json:"description"`
	Resolver    []Resolver  `yaml:"resolver" json:"resolver"`
	Validations Validations `yaml:"validations" json:"validations"`
}

type Resolver struct {
	Type string `yaml:"type" json:"type"`
	ID   string `yaml:"id" json:"id"`
}

type OperationsResponse struct {
	Module      string       `json:"module"`
	Description string       `json:"description"`
	Operations  []Operations `json:"operations"`
}

type OperationDetailResponse struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Schema      []FormField `json:"formFields"`
}

type FormField struct {
	Name        string      `json:"name"`
	DisplayName string      `json:"displayName"`
	Type        string      `json:"type"`
	Options     []Option    `json:"options,omitempty"`
	Validation  Validations `json:"validations,omitempty"`
}

type Validations struct {
	Required     bool   `json:"required"`
	Pattern      string `json:"pattern,omitempty"`
	ErrorMessage string `yaml:"errorMessage,omitempty" json:"errorMessage,omitempty"`
	Min          int    `json:"min,omitempty"`
	Max          int    `json:"max,omitempty"`
}

type Option struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type DatabaseYAML struct {
	Modules []Module `yaml:"modules"`
}

type SchemaYAML struct {
	Schemas []Schema `yaml:"schemas"`
}

type serviceImpl struct {
	modules           map[string]*Module
	schemas           map[string]*Schema
	actionMapper      accesscontrol.ActionMapping
	fileReaderFactory filereader.FileReaderFactory
}
