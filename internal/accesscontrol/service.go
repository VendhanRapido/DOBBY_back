package accesscontrol

//go:generate mockgen -destination=service_mock.go -package=accesscontrol . AccessControlService

type AccessControlService interface {
	VerifyActionForRoles(roles []string, module, action string) bool
	GetPermittedActionsForRoles(roles []string, module string) []string
	GetPermittedModulesForRoles(roles []string) []string
}

type accessControlServiceImpl struct {
	enforcer AccessControlAdapter
}

func NewAccessControlService(adapter AccessControlAdapter) AccessControlService {
	return &accessControlServiceImpl{
		enforcer: adapter,
	}
}

func (s *accessControlServiceImpl) verifyAction(role string, module string, action string) bool {
	hasAccess, _ := s.enforcer.Enforce(role, module, action)
	return hasAccess
}

func (s *accessControlServiceImpl) VerifyActionForRoles(roles []string, module, action string) bool {
	for _, role := range roles {
		hasAccess := s.verifyAction(role, module, action)
		if hasAccess {
			return true
		}
	}

	return false
}

func (s *accessControlServiceImpl) getPermittedModules(role string) []string {
	modules, _ := s.enforcer.GetImplicitModulesForRole(role)
	return modules
}

func (s *accessControlServiceImpl) GetPermittedModulesForRoles(roles []string) []string {
	var modules []string
	moduleExists := make(map[string]bool)

	for _, role := range roles {
		for _, module := range s.getPermittedModules(role) {
			if !moduleExists[module] {
				modules = append(modules, module)
				moduleExists[module] = true
			}
		}
	}

	return modules
}

func (s *accessControlServiceImpl) getPermittedActions(role string, module string) []string {
	actions, _ := s.enforcer.GetImplicitActionsForRole(role, module)
	return actions
}

func (s *accessControlServiceImpl) GetPermittedActionsForRoles(roles []string, module string) []string {
	var actions []string
	actionExists := make(map[string]bool)

	for _, role := range roles {
		for _, action := range s.getPermittedActions(role, module) {
			if !actionExists[action] {
				actions = append(actions, action)
				actionExists[action] = true
			}
		}
	}

	return actions
}
