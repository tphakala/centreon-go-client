package centreon

import (
	"context"
	"fmt"
	"iter"
	"time"
)

// MonitoringHost represents a host as seen from the monitoring engine.
type MonitoringHost struct {
	ID      int            `json:"id"`
	Name    string         `json:"name"`
	Address string         `json:"address,omitzero"`
	Alias   string         `json:"alias,omitzero"`
	Status  ResourceStatus `json:"status"`
}

// HostStatusCount holds status counts for hosts.
type HostStatusCount struct {
	Up          int `json:"up"`
	Down        int `json:"down"`
	Unreachable int `json:"unreachable"`
	Pending     int `json:"pending"`
}

// TimelineEvent represents an event in a resource's timeline.
type TimelineEvent struct {
	ID      int       `json:"id"`
	Type    string    `json:"type"`
	Content string    `json:"content"`
	Date    time.Time `json:"date,omitzero"`
}

// MonitoringHostService provides access to the monitoring hosts endpoints.
type MonitoringHostService struct {
	client *Client
}

// List returns a paginated list of monitoring hosts.
func (s *MonitoringHostService) List(ctx context.Context, opts ...ListOption) (*ListResponse[MonitoringHost], error) {
	var resp ListResponse[MonitoringHost]
	err := s.client.list(ctx, "/monitoring/hosts", opts, &resp)
	return &resp, err
}

// All returns an iterator over all monitoring hosts.
func (s *MonitoringHostService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*MonitoringHost, error] {
	return all(ctx, s.List, opts)
}

// Get returns the monitoring host with the given ID.
func (s *MonitoringHostService) Get(ctx context.Context, id int) (*MonitoringHost, error) {
	var result MonitoringHost
	if err := s.client.get(ctx, fmt.Sprintf("/monitoring/hosts/%d", id), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// StatusCounts returns counts of hosts grouped by status.
func (s *MonitoringHostService) StatusCounts(ctx context.Context) (*HostStatusCount, error) {
	var result HostStatusCount
	if err := s.client.get(ctx, "/monitoring/hosts/status", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Services returns a paginated list of services for a given host.
func (s *MonitoringHostService) Services(ctx context.Context, hostID int, opts ...ListOption) (*ListResponse[MonitoringService], error) {
	var resp ListResponse[MonitoringService]
	err := s.client.list(ctx, fmt.Sprintf("/monitoring/hosts/%d/services", hostID), opts, &resp)
	return &resp, err
}

// Timeline returns a paginated list of timeline events for a given host.
func (s *MonitoringHostService) Timeline(ctx context.Context, hostID int, opts ...ListOption) (*ListResponse[TimelineEvent], error) {
	var resp ListResponse[TimelineEvent]
	err := s.client.list(ctx, fmt.Sprintf("/monitoring/hosts/%d/timeline", hostID), opts, &resp)
	return &resp, err
}
