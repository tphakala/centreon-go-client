package centreon

import (
	"context"
	"fmt"
	"iter"
	"time"
)

// Acknowledgement represents an acknowledgement for a host or service.
type Acknowledgement struct {
	ID                  int        `json:"id"`
	HostID              int        `json:"host_id"`
	ServiceID           *int       `json:"service_id"`
	AuthorID            int        `json:"author_id"`
	AuthorName          string     `json:"author_name"`
	Comment             string     `json:"comment"`
	IsSticky            bool       `json:"is_sticky"`
	IsPersistentComment bool       `json:"is_persistent_comment"`
	IsNotifyContacts    bool       `json:"is_notify_contacts"`
	State               int        `json:"state"`
	Type                int        `json:"type"`
	EntryTime           time.Time  `json:"entry_time,omitzero"`
	DeletionTime        *time.Time `json:"deletion_time"`
}

// CreateAcknowledgementRequest is the request body for acknowledging a host or service.
type CreateAcknowledgementRequest struct {
	Comment             string `json:"comment"`
	IsNotifyContacts    bool   `json:"is_notify_contacts"`
	IsPersistentComment bool   `json:"is_persistent_comment"`
	IsSticky            bool   `json:"is_sticky"`
	WithServices        bool   `json:"with_services"`
}

// AcknowledgementService provides access to the acknowledgement endpoints.
type AcknowledgementService struct {
	client *Client
}

// List returns a paginated list of acknowledgements.
func (s *AcknowledgementService) List(ctx context.Context, opts ...ListOption) (*ListResponse[Acknowledgement], error) {
	var resp ListResponse[Acknowledgement]
	err := s.client.list(ctx, "/monitoring/acknowledgements", opts, &resp)
	return &resp, err
}

// All returns an iterator over all acknowledgements.
func (s *AcknowledgementService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*Acknowledgement, error] {
	return all(ctx, s.List, opts)
}

// Get returns the acknowledgement with the given ID.
func (s *AcknowledgementService) Get(ctx context.Context, id int) (*Acknowledgement, error) {
	var result Acknowledgement
	if err := s.client.get(ctx, fmt.Sprintf("/monitoring/acknowledgements/%d", id), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListForHost returns a paginated list of acknowledgements for the given host.
func (s *AcknowledgementService) ListForHost(ctx context.Context, hostID int, opts ...ListOption) (*ListResponse[Acknowledgement], error) {
	var resp ListResponse[Acknowledgement]
	err := s.client.list(ctx, fmt.Sprintf("/monitoring/hosts/%d/acknowledgements", hostID), opts, &resp)
	return &resp, err
}

// ListForService returns a paginated list of acknowledgements for the given service on a host.
func (s *AcknowledgementService) ListForService(ctx context.Context, hostID, serviceID int, opts ...ListOption) (*ListResponse[Acknowledgement], error) {
	var resp ListResponse[Acknowledgement]
	err := s.client.list(ctx, fmt.Sprintf("/monitoring/hosts/%d/services/%d/acknowledgements", hostID, serviceID), opts, &resp)
	return &resp, err
}

// CreateForHost acknowledges the given host.
func (s *AcknowledgementService) CreateForHost(ctx context.Context, hostID int, req *CreateAcknowledgementRequest) error {
	return s.client.post(ctx, fmt.Sprintf("/monitoring/hosts/%d/acknowledgements", hostID), req, nil)
}

// CreateForService acknowledges the given service on a host.
func (s *AcknowledgementService) CreateForService(ctx context.Context, hostID, serviceID int, req *CreateAcknowledgementRequest) error {
	return s.client.post(ctx, fmt.Sprintf("/monitoring/hosts/%d/services/%d/acknowledgements", hostID, serviceID), req, nil)
}

// CancelForHost cancels the acknowledgement for the given host.
func (s *AcknowledgementService) CancelForHost(ctx context.Context, hostID int) error {
	return s.client.delete(ctx, fmt.Sprintf("/monitoring/hosts/%d/acknowledgements", hostID))
}

// CancelForService cancels the acknowledgement for the given service on a host.
func (s *AcknowledgementService) CancelForService(ctx context.Context, hostID, serviceID int) error {
	return s.client.delete(ctx, fmt.Sprintf("/monitoring/hosts/%d/services/%d/acknowledgements", hostID, serviceID))
}
