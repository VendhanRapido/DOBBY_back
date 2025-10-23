package accesscontrol

import (
	"testing"

	"github.com/stretchr/testify/suite"
	gomock "go.uber.org/mock/gomock"
)

type AdapterTestSuite struct {
	suite.Suite
	ctrl    *gomock.Controller
	adapter AccessControlAdapter
}

func TestAdapterTestSuite(t *testing.T) {
	suite.Run(t, &AdapterTestSuite{})
}

func (a *AdapterTestSuite) SetupTest() {
	var err error
	a.ctrl = gomock.NewController(a.T())

	a.adapter, err = NewAccessControlAdapter()
	a.Require().Nil(err)
}

func (a *AdapterTestSuite) TestGetImplicitModulesForRole() {
	role := "developer"
	var expectedResponse = []string{"kafka"}
	response, err := a.adapter.GetImplicitModulesForRole(role)
	a.Require().NoError(err)
	a.Equal(response, expectedResponse)
}

func (a *AdapterTestSuite) TestGetImplicitActionsForRole() {
	module := "kafka"
	role := "developer"
	var expectedResponse = []string{"write"}
	response, err := a.adapter.GetImplicitActionsForRole(role, module)
	a.Require().NoError(err)
	a.Equal(response, expectedResponse)
}

func (a *AdapterTestSuite) TestEnforce() {
	module := "kafka"
	role := "developer"
	action := "write"

	response, err := a.adapter.Enforce(role, module, action)
	a.Require().NoError(err)
	a.True(response)
}

func (a *AdapterTestSuite) TestGetImplicitModulesForRoleNoPermissions() {
	role := "undefinedRole"
	modules, err := a.adapter.GetImplicitModulesForRole(role)
	a.Require().NoError(err)
	a.Empty(modules)
}

func (a *AdapterTestSuite) TestGetImplicitActionsForRoleNoPermissions() {
	role := "undefinedRole"
	actions, err := a.adapter.GetImplicitActionsForRole(role, "kafka")
	a.Require().NoError(err)
	a.Empty(actions)
}
