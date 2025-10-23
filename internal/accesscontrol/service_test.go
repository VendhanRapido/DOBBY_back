package accesscontrol

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type AccessControlServiceTestSuite struct {
	suite.Suite
	ctrl                     *gomock.Controller
	mockAccessControlAdapter *MockAccessControlAdapter
	service                  AccessControlService
}

func TestAccessControlServiceTestSuite(t *testing.T) {
	suite.Run(t, &AccessControlServiceTestSuite{})
}

func (e *AccessControlServiceTestSuite) SetupTest() {
	e.ctrl = gomock.NewController(e.T())
	e.mockAccessControlAdapter = NewMockAccessControlAdapter(e.ctrl)
	e.service = NewAccessControlService(e.mockAccessControlAdapter)
}

func (e *AccessControlServiceTestSuite) TearDownTest() {
	e.ctrl.Finish()
}

func (e *AccessControlServiceTestSuite) TestGetPermittedModulesForRoles_Success() {
	roles := []string{"dev", "admin"}
	e.mockAccessControlAdapter.EXPECT().
		GetImplicitModulesForRole("dev").
		Return([]string{"kafka", "mongo"}, nil).
		Times(1)
	e.mockAccessControlAdapter.EXPECT().
		GetImplicitModulesForRole("admin").
		Return([]string{"kafka", "redis"}, nil).
		Times(1)

	modules := e.service.GetPermittedModulesForRoles(roles)

	assert.ElementsMatch(e.T(), []string{"kafka", "mongo", "redis"}, modules)
}

func (e *AccessControlServiceTestSuite) TestGetPermittedModulesForRoles_EmptyRoles() {
	roles := []string{}

	modules := e.service.GetPermittedModulesForRoles(roles)

	assert.Empty(e.T(), modules)
}

func (e *AccessControlServiceTestSuite) TestGetPermittedModulesForRoles_ErrorFromAdapter() {
	roles := []string{"dev"}
	e.mockAccessControlAdapter.EXPECT().
		GetImplicitModulesForRole("dev").
		Return(nil, errors.New("dummy error")).
		Times(1)

	modules := e.service.GetPermittedModulesForRoles(roles)

	assert.Empty(e.T(), modules)
}

func (e *AccessControlServiceTestSuite) TestGetPermittedActionsForRoles_Success() {
	roles := []string{"dev", "admin"}
	module := "kafka"

	e.mockAccessControlAdapter.EXPECT().
		GetImplicitActionsForRole("dev", module).
		Return([]string{"read", "write"}, nil).
		Times(1)
	e.mockAccessControlAdapter.EXPECT().
		GetImplicitActionsForRole("admin", module).
		Return([]string{"write", "delete"}, nil).
		Times(1)

	actions := e.service.GetPermittedActionsForRoles(roles, module)

	assert.ElementsMatch(e.T(), []string{"read", "write", "delete"}, actions)
}

func (e *AccessControlServiceTestSuite) TestGetPermittedActionsForRoles_NoActionsForRole() {
	roles := []string{"intern"}
	module := "kafka"
	e.mockAccessControlAdapter.EXPECT().
		GetImplicitActionsForRole("intern", module).
		Return([]string{}, nil).
		Times(1)

	actions := e.service.GetPermittedActionsForRoles(roles, module)

	assert.Empty(e.T(), actions)
}

func (e *AccessControlServiceTestSuite) TestGetPermittedActionsForRoles_ErrorFromAdapter() {
	roles := []string{"dev"}
	module := "kafka"
	e.mockAccessControlAdapter.EXPECT().
		GetImplicitActionsForRole("dev", module).
		Return(nil, errors.New("fetch error")).
		Times(1)

	actions := e.service.GetPermittedActionsForRoles(roles, module)

	assert.Empty(e.T(), actions)
}

func (e *AccessControlServiceTestSuite) TestGetPermittedActionsForRoles_EmptyRoles() {
	roles := []string{}
	module := "kafka"

	actions := e.service.GetPermittedActionsForRoles(roles, module)

	assert.Empty(e.T(), actions)
}

func (e *AccessControlServiceTestSuite) TestVerifyActionForRoles_TrueIfAnyRoleHasAccess() {
	roles := []string{"dev", "admin"}
	module := "kafka"
	action := "delete"
	e.mockAccessControlAdapter.EXPECT().
		Enforce("dev", module, action).
		Return(false, nil).
		Times(1)
	e.mockAccessControlAdapter.EXPECT().
		Enforce("admin", module, action).
		Return(true, nil).
		Times(1)

	allowed := e.service.VerifyActionForRoles(roles, module, action)

	assert.True(e.T(), allowed)
}

func (e *AccessControlServiceTestSuite) TestVerifyActionForRoles_FalseIfNoRoleHasAccess() {
	roles := []string{"dev", "admin"}
	module := "kafka"
	action := "delete"
	e.mockAccessControlAdapter.EXPECT().
		Enforce("dev", module, action).
		Return(false, nil).
		Times(1)
	e.mockAccessControlAdapter.EXPECT().
		Enforce("admin", module, action).
		Return(false, nil).
		Times(1)

	allowed := e.service.VerifyActionForRoles(roles, module, action)

	assert.False(e.T(), allowed)
}

func (e *AccessControlServiceTestSuite) TestVerifyActionForRoles_ErrorIsIgnored() {
	roles := []string{"dev", "admin"}
	module := "kafka"
	action := "write"
	e.mockAccessControlAdapter.EXPECT().
		Enforce("dev", module, action).
		Return(false, errors.New("err")).
		Times(1)
	e.mockAccessControlAdapter.EXPECT().
		Enforce("admin", module, action).
		Return(true, nil).
		Times(1)

	allowed := e.service.VerifyActionForRoles(roles, module, action)

	assert.True(e.T(), allowed)
}

func (e *AccessControlServiceTestSuite) TestVerifyActionForRoles_EmptyRoles() {
	roles := []string{}
	module := "kafka"
	action := "read"

	allowed := e.service.VerifyActionForRoles(roles, module, action)

	assert.False(e.T(), allowed)
}
