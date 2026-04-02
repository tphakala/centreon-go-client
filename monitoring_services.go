package centreon

import (
	"context"
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
