package centreon

import (
	"context"
	"iter"
)

// MonitoringService represents a service as seen from the monitoring engine.
type MonitoringService struct {
	ID     int            `json:"id"`
	Name   string         `json:"name"`
	Status ResourceStatus `json:"status"`
}

// ServiceStatusCount holds status counts for services.
type ServiceStatusCount struct {
	OK       int `json:"ok"`
	Warning  int `json:"warning"`
	Critical int `json:"critical"`
	Unknown  int `json:"unknown"`
	Pending  int `json:"pending"`
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
