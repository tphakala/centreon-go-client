package centreon

import (
	"context"
	"fmt"
	"iter"
)

// MonitoringServer represents a Centreon monitoring server (poller).
type MonitoringServer struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Address     string `json:"address,omitzero"`
	IsActivated bool   `json:"is_activate"`
	IsDefault   bool   `json:"is_default"`
}

type MonitoringServerService struct {
	client *Client
}

func (s *MonitoringServerService) List(ctx context.Context, opts ...ListOption) (*ListResponse[MonitoringServer], error) {
	var resp ListResponse[MonitoringServer]
	err := s.client.list(ctx, "/configuration/monitoring-servers", opts, &resp)
	return &resp, err
}

func (s *MonitoringServerService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*MonitoringServer, error] {
	return all(ctx, s.List, opts)
}

func (s *MonitoringServerService) GenerateAndReload(ctx context.Context, serverID int) error {
	return s.client.get(ctx, fmt.Sprintf("/configuration/monitoring-servers/%d/generate-and-reload", serverID), nil)
}

func (s *MonitoringServerService) GenerateAndReloadAll(ctx context.Context) error {
	return s.client.get(ctx, "/configuration/monitoring-servers/generate-and-reload", nil)
}
