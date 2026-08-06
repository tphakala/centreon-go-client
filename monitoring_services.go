package centreon

import (
	"context"
	"fmt"
	"iter"
)

// MonitoringServiceHost is the nested host reference in a monitoring service response.
type MonitoringServiceHost struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Alias string `json:"alias,omitzero"`
	State int    `json:"state"`
}

// MonitoringService represents a service as seen from the monitoring engine.
type MonitoringService struct {
	ID               int                   `json:"id"`
	Description      string                `json:"description"`
	DisplayName      string                `json:"display_name,omitzero"`
	Host             MonitoringServiceHost `json:"host"`
	State            int                   `json:"state"`
	StateType        int                   `json:"state_type"`
	Output           string                `json:"output,omitzero"`
	Status           ResourceStatus        `json:"status"`
	Acknowledged     bool                  `json:"is_acknowledged"`
	DowntimeDepth    int                   `json:"scheduled_downtime_depth"`
	LastCheck        string                `json:"last_check,omitzero"`
	LastStateChange  string                `json:"last_state_change,omitzero"`
	MaxCheckAttempts int                   `json:"max_check_attempts"`
}

// ServiceStatusCount holds status counts for services.
type ServiceStatusCount struct {
	OK       StatusValue `json:"ok"`
	Warning  StatusValue `json:"warning"`
	Critical StatusValue `json:"critical"`
	Unknown  StatusValue `json:"unknown"`
	Pending  StatusValue `json:"pending"`
	Total    int         `json:"total"`
}

// Metric is a performance metric of a service, as returned by
// GET /monitoring/hosts/{host_id}/services/{service_id}/metrics.
// Pointer fields are nil when the API returns JSON null, which is
// common for thresholds that are not configured on the metric.
type Metric struct {
	ID                    int      `json:"id"`
	Name                  string   `json:"name"`
	Unit                  string   `json:"unit"`
	CurrentValue          *float64 `json:"current_value"`
	WarningHighThreshold  *float64 `json:"warning_high_threshold"`
	WarningLowThreshold   *float64 `json:"warning_low_threshold"`
	CriticalHighThreshold *float64 `json:"critical_high_threshold"`
	CriticalLowThreshold  *float64 `json:"critical_low_threshold"`
}

// MonitoringServiceService provides access to the monitoring services endpoints.
type MonitoringServiceService struct {
	client *Client
}

// List returns a paginated list of monitoring services.
func (s *MonitoringServiceService) List(ctx context.Context, opts ...ListOption) (*ListResponse[MonitoringService], error) {
	var resp ListResponse[MonitoringService]
	err := s.client.list(ctx, "/monitoring/services", opts, &resp)
	return &resp, err
}

// All returns an iterator over all monitoring services.
func (s *MonitoringServiceService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*MonitoringService, error] {
	return all(ctx, s.List, opts)
}

// StatusCounts returns counts of services grouped by status.
func (s *MonitoringServiceService) StatusCounts(ctx context.Context) (*ServiceStatusCount, error) {
	var result ServiceStatusCount
	if err := s.client.get(ctx, "/monitoring/services/status", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Timeline returns a paginated list of timeline events for a given service on a host.
func (s *MonitoringServiceService) Timeline(ctx context.Context, hostID, serviceID int, opts ...ListOption) (*ListResponse[TimelineEvent], error) {
	var resp ListResponse[TimelineEvent]
	err := s.client.list(ctx, fmt.Sprintf("/monitoring/hosts/%d/services/%d/timeline", hostID, serviceID), opts, &resp)
	return &resp, err
}

// Metrics returns the performance metrics for a given service on a host.
// The endpoint returns a plain JSON array, so there is no pagination and
// no ListOption support.
func (s *MonitoringServiceService) Metrics(ctx context.Context, hostID, serviceID int) ([]Metric, error) {
	var result []Metric
	if err := s.client.get(ctx, fmt.Sprintf("/monitoring/hosts/%d/services/%d/metrics", hostID, serviceID), &result); err != nil {
		return nil, err
	}
	return result, nil
}
