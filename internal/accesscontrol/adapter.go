package accesscontrol

import (
	"embed"

	"github.com/casbin/casbin/v2"
	casbin_fs_adapter "github.com/naucon/casbin-fs-adapter"
)

//go:embed rbac_model.conf policy.csv
var configFiles embed.FS

const (
	RoleIndex   = 0
	ModuleIndex = 1
	ActionIndex = 2
)

//go:generate mockgen -destination=adapter_mock.go -package=accesscontrol . AccessControlAdapter

type AccessControlAdapter interface {
	GetImplicitModulesForRole(role string) ([]string, error)
	GetImplicitActionsForRole(role string, module string) ([]string, error)
	Enforce(params ...interface{}) (bool, error)
}

type casbinAdapter struct {
	enforcer *casbin.Enforcer
}

func NewAccessControlAdapter() (AccessControlAdapter, error) {
	// Create model from embedded config file
	model, err := casbin_fs_adapter.NewModel(configFiles, "rbac_model.conf")
	if err != nil {
		return nil, err
	}

	// Create adapter from embedded policy file
	policies := casbin_fs_adapter.NewAdapter(configFiles, "policy.csv")

	// Create enforcer with embedded model and adapter
	enforcer, err := casbin.NewEnforcer(model, policies)
	if err != nil {
		return nil, err
	}

	return &casbinAdapter{
		enforcer: enforcer,
	}, nil
}

// func AccessAdapter(model string, policy string) (AccessControlAdapter, error) {
// 	enforcer, err := casbin.NewEnforcer(model, policy)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &casbinAdapter{
// 		enforcer: enforcer,
// 	}, nil
// }

func (c *casbinAdapter) GetImplicitModulesForRole(role string) ([]string, error) {
	permissions, err := c.enforcer.GetImplicitPermissionsForUser(role)
	if err != nil {
		return nil, err
	}

	var modules []string
	moduleExists := make(map[string]bool)
	for _, permission := range permissions {
		if _, exists := moduleExists[permission[ModuleIndex]]; !exists {
			modules = append(modules, permission[ModuleIndex])
			moduleExists[permission[ModuleIndex]] = true
		}
	}
	return modules, nil
}

func (c *casbinAdapter) GetImplicitActionsForRole(role string, module string) ([]string, error) {
	permissions, err := c.enforcer.GetImplicitPermissionsForUser(role)
	if err != nil {
		return nil, err
	}

	var actions []string
	actionExists := make(map[string]bool)
	for _, permission := range permissions {
		if !actionExists[permission[ActionIndex]] && permission[ModuleIndex] == module {
			actions = append(actions, permission[ActionIndex])
			actionExists[permission[ActionIndex]] = true
		}
	}

	return actions, nil
}

func (c *casbinAdapter) Enforce(params ...interface{}) (bool, error) {
	return c.enforcer.Enforce(params...)
}
