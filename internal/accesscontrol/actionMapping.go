package accesscontrol

import (
	"fmt"

	"github.com/roppenlabs/dobby-service/internal/utils/filereader"
)

//go:generate mockgen -destination=actionMapping_mock.go -package=accesscontrol . ActionMapping

type ActionMapping interface {
	GetValidOpIds(module string, allowedActions []string) (map[string]bool, error)
	LoadRBACMap(path string) error
	GetValidAction(module, opid string) (string, error)
}

type actionMappingImpl struct {
	rbacMap                  map[string]map[string]string
	moduleActionOperationMap map[string]map[string][]string
	fileReaderFactory        filereader.FileReaderFactory
}

func NewActionMapping(fileReaderFactory filereader.FileReaderFactory) ActionMapping {
	return &actionMappingImpl{
		rbacMap:                  make(map[string]map[string]string),
		moduleActionOperationMap: make(map[string]map[string][]string),
		fileReaderFactory:        fileReaderFactory,
	}
}

func (a *actionMappingImpl) GetValidOpIds(module string, allowedActions []string) (map[string]bool, error) {
	moduleMap, ok := a.moduleActionOperationMap[module]
	if !ok {
		return nil, fmt.Errorf("module %q not found", module)
	}
	allowedOpIDs := map[string]bool{}
	for _, action := range allowedActions {
		opIDs, ok := moduleMap[action]
		if !ok {
			return nil, fmt.Errorf("action %q not found in module %q", action, module)
		}
		for _, opID := range opIDs {
			allowedOpIDs[opID] = true
		}
	}
	return allowedOpIDs, nil
}

func (a *actionMappingImpl) LoadRBACMap(path string) error {
	filereader, err := a.fileReaderFactory("yaml", path)
	if err != nil {
		return fmt.Errorf("failed to read RBAC file: %w", err)
	}

	var rbacFile RBACFile
	err = filereader.ReadFile(&rbacFile)

	if err != nil {
		return fmt.Errorf("failed to parse RBAC file: %w", err)
	}

	rbacMap := make(map[string]map[string]string)
	moduleActionOperationMap := make(map[string]map[string][]string)

	for _, mod := range rbacFile.Modules {
		opMap := make(map[string]string)
		actionMap := make(map[string][]string)

		for _, op := range mod.Operations {
			opMap[op.ID] = op.RBACAction
			actionMap[op.RBACAction] = append(actionMap[op.RBACAction], op.ID)
		}

		rbacMap[mod.Name] = opMap
		moduleActionOperationMap[mod.Name] = actionMap
	}

	a.rbacMap = rbacMap
	a.moduleActionOperationMap = moduleActionOperationMap

	return nil
}

func (a *actionMappingImpl) GetValidAction(module, opid string) (string, error) {
	moduleMap, ok := a.rbacMap[module]
	if !ok {
		return "", fmt.Errorf("module %q not found", module)
	}

	action, ok := moduleMap[opid]
	if !ok {
		return "", fmt.Errorf("operation ID %q not found in module %q", opid, module)
	}

	return action, nil
}
