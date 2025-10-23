package entitiesclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"time"

	"github.com/roppenlabs/dobby-service/internal/config"
)

//go:generate mockgen -destination client_mock.go -package=entitiesclient . EntitiesClient

type entitiesClient struct {
	baseURL    string
	httpClient *http.Client
	timeouts   config.EntitiesTimeouts
}

type EntitiesClient interface {
	GetRolesFromEmail(email string) ([]string, error)
}

func NewEntitiesClient(cfg *config.Config) EntitiesClient {
	return &entitiesClient{
		baseURL:    fmt.Sprintf("%s:%s", cfg.Clients.Entities.Svc.Host, cfg.Clients.Entities.Svc.Port),
		timeouts:   cfg.Clients.Entities.TimeoutInMS,
		httpClient: &http.Client{},
	}
}

func (c *entitiesClient) GetRolesFromEmail(email string) ([]string, error) {
	reqBody, _ := json.Marshal(map[string]string{"email": email})

	url := fmt.Sprintf("%s/api/dash/supply/user/get", c.baseURL)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.timeouts.GetUser)*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call rapido user service: %w", err)
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	var apiRes GetUserResponse
	if err := json.Unmarshal(body, &apiRes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	roles := apiRes.Data.Roles

	return roles, nil
}
