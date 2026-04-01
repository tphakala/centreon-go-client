package centreon

import (
	"context"
	"fmt"
	"iter"
)

// ServiceSeverity represents a Centreon service severity configuration resource.
type ServiceSeverity struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Alias       string `json:"alias,omitzero"`
	Level       int    `json:"level"`
	IconID      int    `json:"icon_id"`
	IsActivated bool   `json:"is_activated"`
}

// CreateServiceSeverityRequest is the request body for creating a service severity.
type CreateServiceSeverityRequest struct {
	Name   string `json:"name"`
	Alias  string `json:"alias,omitzero"`
	Level  int    `json:"level"`
	IconID int    `json:"icon_id"`
}

// UpdateServiceSeverityRequest is the request body for updating a service severity (PUT).
type UpdateServiceSeverityRequest struct {
	Name   string `json:"name"`
	Alias  string `json:"alias,omitzero"`
	Level  int    `json:"level"`
	IconID int    `json:"icon_id"`
}

// ServiceSeverityService provides service severity configuration operations.
type ServiceSeverityService struct {
	client *Client
}

// List returns a paginated list of service severities.
func (s *ServiceSeverityService) List(ctx context.Context, opts ...ListOption) (*ListResponse[ServiceSeverity], error) {
	var resp ListResponse[ServiceSeverity]
	err := s.client.list(ctx, "/configuration/services/severities", opts, &resp)
	return &resp, err
}

// All returns an iterator over all service severities.
func (s *ServiceSeverityService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*ServiceSeverity, error] {
	return all(ctx, s.List, opts)
}

// Create creates a new service severity and returns its ID.
func (s *ServiceSeverityService) Create(ctx context.Context, req CreateServiceSeverityRequest) (int, error) {
	var result struct {
		ID int `json:"id"`
	}
	if err := s.client.post(ctx, "/configuration/services/severities", req, &result); err != nil {
		return 0, err
	}
	return result.ID, nil
}

// Update replaces an existing service severity using PUT.
func (s *ServiceSeverityService) Update(ctx context.Context, id int, req UpdateServiceSeverityRequest) error {
	return s.client.put(ctx, fmt.Sprintf("/configuration/services/severities/%d", id), req, nil)
}

// Delete deletes a service severity by ID.
func (s *ServiceSeverityService) Delete(ctx context.Context, id int) error {
	return s.client.delete(ctx, fmt.Sprintf("/configuration/services/severities/%d", id))
}
