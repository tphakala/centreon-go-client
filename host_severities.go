package centreon

import (
	"context"
	"fmt"
	"iter"
)

// HostSeverity represents a Centreon host severity configuration resource.
type HostSeverity struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Alias       string `json:"alias,omitzero"`
	Level       int    `json:"level"`
	IconID      int    `json:"icon_id"`
	IsActivated bool   `json:"is_activated"`
}

// CreateHostSeverityRequest is the request body for creating a host severity.
type CreateHostSeverityRequest struct {
	Name   string `json:"name"`
	Alias  string `json:"alias,omitzero"`
	Level  int    `json:"level"`
	IconID int    `json:"icon_id"`
}

// UpdateHostSeverityRequest is the request body for updating a host severity (PUT).
type UpdateHostSeverityRequest struct {
	Name   string `json:"name"`
	Alias  string `json:"alias,omitzero"`
	Level  int    `json:"level"`
	IconID int    `json:"icon_id"`
}

// HostSeverityService provides host severity configuration operations.
type HostSeverityService struct {
	client *Client
}

// List returns a paginated list of host severities.
func (s *HostSeverityService) List(ctx context.Context, opts ...ListOption) (*ListResponse[HostSeverity], error) {
	var resp ListResponse[HostSeverity]
	err := s.client.list(ctx, "/configuration/hosts/severities", opts, &resp)
	return &resp, err
}

// All returns an iterator over all host severities.
func (s *HostSeverityService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*HostSeverity, error] {
	return all(ctx, s.List, opts)
}

// Get returns the host severity with the given ID using a direct GET request.
func (s *HostSeverityService) Get(ctx context.Context, id int) (*HostSeverity, error) {
	var sev HostSeverity
	if err := s.client.get(ctx, fmt.Sprintf("/configuration/hosts/severities/%d", id), &sev); err != nil {
		return nil, err
	}
	return &sev, nil
}

// Create creates a new host severity and returns its ID.
func (s *HostSeverityService) Create(ctx context.Context, req CreateHostSeverityRequest) (int, error) {
	var result struct {
		ID int `json:"id"`
	}
	if err := s.client.post(ctx, "/configuration/hosts/severities", req, &result); err != nil {
		return 0, err
	}
	return result.ID, nil
}

// Update replaces an existing host severity using PUT.
func (s *HostSeverityService) Update(ctx context.Context, id int, req UpdateHostSeverityRequest) error {
	return s.client.put(ctx, fmt.Sprintf("/configuration/hosts/severities/%d", id), req)
}

// Delete deletes a host severity by ID.
func (s *HostSeverityService) Delete(ctx context.Context, id int) error {
	return s.client.delete(ctx, fmt.Sprintf("/configuration/hosts/severities/%d", id))
}
