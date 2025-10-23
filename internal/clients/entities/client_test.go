package entitiesclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/roppenlabs/dobby-service/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type EntitiesClientTestSuite struct {
	suite.Suite
	server *httptest.Server
	cfg    *config.Config
	client EntitiesClient
}

func (suite *EntitiesClientTestSuite) SetupTest() {
	suite.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"roles": []string{"admin", "editor"},
			},
		})
	}))

	suite.cfg = &config.Config{
		Clients: config.Clients{
			Entities: config.Entities{
				Svc: config.SvcConfig{
					Host: suite.server.URL,
					Port: "",
				},
				TimeoutInMS: config.EntitiesTimeouts{
					GetUser: 5000,
				},
			},
		},
	}

	suite.client = NewEntitiesClient(suite.cfg)
}

func (suite *EntitiesClientTestSuite) TearDownTest() {
	suite.server.Close()

}

func (suite *EntitiesClientTestSuite) TestGetRolesFromEmail_Success() {
	roles, err := suite.client.GetRolesFromEmail("test@example.com")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), []string{"admin", "editor"}, roles)
}

func (suite *EntitiesClientTestSuite) TestGetRolesFromEmail_InvalidJSON() {
	suite.server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("invalid json"))
	})

	roles, err := suite.client.GetRolesFromEmail("test@example.com")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), roles)
}

func (suite *EntitiesClientTestSuite) TestGetRolesFromEmail_HTTPFailure() {
	suite.server.Close()
	suite.client = NewEntitiesClient(suite.cfg)

	roles, err := suite.client.GetRolesFromEmail("test@example.com")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), roles)
}

func (suite *EntitiesClientTestSuite) TestGetRolesFromEmail_MissingRolesKey() {
	suite.server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{},
		})
	})

	roles, err := suite.client.GetRolesFromEmail("test@example.com")
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), roles)
}

func (suite *EntitiesClientTestSuite) TestGetRolesFromEmail_InvalidURL() {
	suite.cfg.Clients.Entities.Svc.Host = "http://[::1"
	suite.client = NewEntitiesClient(suite.cfg)

	roles, err := suite.client.GetRolesFromEmail("test@example.com")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to create request")
	assert.Nil(suite.T(), roles)
}

func TestEntitiesClientTestSuite(t *testing.T) {
	suite.Run(t, new(EntitiesClientTestSuite))
}
