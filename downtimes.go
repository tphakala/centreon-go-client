package centreon

import (
	"context"
	"fmt"
	"iter"
	"time"
)

// Downtime represents a scheduled downtime for a host or service.
type Downtime struct {
	ID             int        `json:"id"`
	HostID         int        `json:"host_id"`
	ServiceID      *int       `json:"service_id"`
	AuthorID       int        `json:"author_id"`
	AuthorName     string     `json:"author_name"`
	Comment        string     `json:"comment"`
	IsFixed        bool       `json:"is_fixed"`
	StartTime      time.Time  `json:"start_time"`
	EndTime        time.Time  `json:"end_time"`
	Duration       int        `json:"duration"`
	ActivationTime time.Time  `json:"activation_time,omitzero"`
	EntryTime      time.Time  `json:"entry_time,omitzero"`
	DeletionTime   *time.Time `json:"deletion_time"`
	PollerID       int        `json:"poller_id"`
}

// CreateDowntimeRequest is the request body for scheduling a downtime on a host or service.
type CreateDowntimeRequest struct {
	Comment      string    `json:"comment"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	IsFixed      bool      `json:"is_fixed"`
	Duration     int       `json:"duration,omitzero"`
	WithServices bool      `json:"with_services,omitzero"`
}

// DowntimeService provides access to the downtime endpoints.
type DowntimeService struct {
	client *Client
}

// List returns a paginated list of downtimes.
func (s *DowntimeService) List(ctx context.Context, opts ...ListOption) (*ListResponse[Downtime], error) {
	var resp ListResponse[Downtime]
	err := s.client.list(ctx, "/monitoring/downtimes", opts, &resp)
	return &resp, err
}

// All returns an iterator over all downtimes.
func (s *DowntimeService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*Downtime, error] {
	return all(ctx, s.List, opts)
}

// Get returns the downtime with the given ID.
func (s *DowntimeService) Get(ctx context.Context, id int) (*Downtime, error) {
	var result Downtime
	if err := s.client.get(ctx, fmt.Sprintf("/monitoring/downtimes/%d", id), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Cancel cancels the downtime with the given ID.
func (s *DowntimeService) Cancel(ctx context.Context, id int) error {
	return s.client.delete(ctx, fmt.Sprintf("/monitoring/downtimes/%d", id))
}

// ListForHost returns a paginated list of downtimes for the given host.
func (s *DowntimeService) ListForHost(ctx context.Context, hostID int, opts ...ListOption) (*ListResponse[Downtime], error) {
	var resp ListResponse[Downtime]
	err := s.client.list(ctx, fmt.Sprintf("/monitoring/hosts/%d/downtimes", hostID), opts, &resp)
	return &resp, err
}

// ListForService returns a paginated list of downtimes for the given service on a host.
func (s *DowntimeService) ListForService(ctx context.Context, hostID, serviceID int, opts ...ListOption) (*ListResponse[Downtime], error) {
	var resp ListResponse[Downtime]
	err := s.client.list(ctx, fmt.Sprintf("/monitoring/hosts/%d/services/%d/downtimes", hostID, serviceID), opts, &resp)
	return &resp, err
}

// CreateForHost schedules a downtime for the given host.
func (s *DowntimeService) CreateForHost(ctx context.Context, hostID int, req *CreateDowntimeRequest) error {
	return s.client.post(ctx, fmt.Sprintf("/monitoring/hosts/%d/downtimes", hostID), req, nil)
}

// CreateForService schedules a downtime for the given service on a host.
func (s *DowntimeService) CreateForService(ctx context.Context, hostID, serviceID int, req *CreateDowntimeRequest) error {
	return s.client.post(ctx, fmt.Sprintf("/monitoring/hosts/%d/services/%d/downtimes", hostID, serviceID), req, nil)
}

// CancelForHost cancels all downtimes for the given host.
func (s *DowntimeService) CancelForHost(ctx context.Context, hostID int) error {
	return s.client.delete(ctx, fmt.Sprintf("/monitoring/hosts/%d/downtimes", hostID))
}

// CancelForService cancels all downtimes for the given service on a host.
func (s *DowntimeService) CancelForService(ctx context.Context, hostID, serviceID int) error {
	return s.client.delete(ctx, fmt.Sprintf("/monitoring/hosts/%d/services/%d/downtimes", hostID, serviceID))
}
