package centreon

import (
	"context"
	"fmt"
	"iter"
	"time"
)

// MonitoringHost represents a host as seen from the monitoring engine.
type MonitoringHost struct {
	ID               int     `json:"id"`
	Name             string  `json:"name"`
	AddressIP        string  `json:"address_ip,omitzero"`
	Alias            string  `json:"alias,omitzero"`
	State            int     `json:"state"`
	StateType        int     `json:"state_type"`
	Output           string  `json:"output,omitzero"`
	Acknowledged     bool    `json:"acknowledged"`
	PollerID         int     `json:"poller_id"`
	CheckAttempt     int     `json:"check_attempt"`
	MaxCheckAttempts int     `json:"max_check_attempts"`
	LastCheck        string  `json:"last_check,omitzero"`
	LastStateChange  string  `json:"last_state_change,omitzero"`
	ExecutionTime    float64 `json:"execution_time"`
	DowntimeDepth    int     `json:"scheduled_downtime_depth"`
	IconImage        string  `json:"icon_image,omitzero"`
	IconImageAlt     string  `json:"icon_image_alt,omitzero"`
}

// StatusValue holds a count with a total subfield, as returned by the Centreon API.
type StatusValue struct {
	Total int `json:"total"`
}

// HostStatusCount holds status counts for hosts.
type HostStatusCount struct {
	Up          StatusValue `json:"up"`
	Down        StatusValue `json:"down"`
	Unreachable StatusValue `json:"unreachable"`
	Pending     StatusValue `json:"pending"`
	Total       int         `json:"total"`
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
