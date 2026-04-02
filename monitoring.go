package centreon

import (
	"context"
	"fmt"
	"iter"
)

// ResourceStatus represents the monitoring status of a resource.
type ResourceStatus struct {
	Code         int    `json:"code"`
	Name         string `json:"name"`
	SeverityCode int    `json:"severity_code"`
}

// MonitoringResourceParent represents the parent resource (typically a host for a service).
type MonitoringResourceParent struct {
	ID     int            `json:"id"`
	Name   string         `json:"name"`
	Type   string         `json:"type"`
	Status ResourceStatus `json:"status"`
}

// MonitoringResource represents a unified monitoring resource (host or service).
type MonitoringResource struct {
	ID                   int                       `json:"id"`
	Name                 string                    `json:"name"`
	Type                 string                    `json:"type"` // "host" or "service"
	Alias                string                    `json:"alias,omitzero"`
	FQDN                 string                    `json:"fqdn,omitzero"`
	HostID               int                       `json:"host_id,omitzero"`
	ServiceID            int                       `json:"service_id,omitzero"`
	MonitoringServerName string                    `json:"monitoring_server_name,omitzero"`
	Parent               *MonitoringResourceParent `json:"parent"`
	Status               ResourceStatus            `json:"status"`
	IsInDowntime         bool                      `json:"is_in_downtime"`
	IsAcknowledged       bool                      `json:"is_acknowledged"`
	Information          string                    `json:"information,omitzero"`
	Tries                string                    `json:"tries,omitzero"`
	LastStatusChange     string                    `json:"last_status_change,omitzero"`
	NotificationEnabled  bool                      `json:"is_notification_enabled"`
}

// MonitoringResourceService provides access to the unified monitoring resources endpoint.
type MonitoringResourceService struct {
	client *Client
}

// List returns a paginated list of monitoring resources.
func (s *MonitoringResourceService) List(ctx context.Context, opts ...ListOption) (*ListResponse[MonitoringResource], error) {
	var resp ListResponse[MonitoringResource]
	err := s.client.list(ctx, "/monitoring/resources", opts, &resp)
	return &resp, err
}

// All returns an iterator over all monitoring resources.
func (s *MonitoringResourceService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*MonitoringResource, error] {
	return all(ctx, s.List, opts)
}

// GetHost returns the monitoring resource for a specific host.
func (s *MonitoringResourceService) GetHost(ctx context.Context, hostID int) (*MonitoringResource, error) {
	var result MonitoringResource
	if err := s.client.get(ctx, fmt.Sprintf("/monitoring/resources/hosts/%d", hostID), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetService returns the monitoring resource for a specific service on a host.
func (s *MonitoringResourceService) GetService(ctx context.Context, hostID, serviceID int) (*MonitoringResource, error) {
	var result MonitoringResource
	if err := s.client.get(ctx, fmt.Sprintf("/monitoring/resources/hosts/%d/services/%d", hostID, serviceID), &result); err != nil {
		return nil, err
	}
	return &result, nil
}
