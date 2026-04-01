package centreon

import (
	"context"
	"fmt"
	"iter"
)

// ServiceTemplate represents a Centreon service template configuration resource.
type ServiceTemplate struct {
	ID                  int    `json:"id"`
	Name                string `json:"name"`
	Alias               string `json:"alias,omitzero"`
	CheckCommandID      int    `json:"check_command_id,omitzero"`
	MaxCheckAttempts    int    `json:"max_check_attempts,omitzero"`
	NormalCheckInterval int    `json:"normal_check_interval,omitzero"`
	RetryCheckInterval  int    `json:"retry_check_interval,omitzero"`
	IsActivated         bool   `json:"is_activated"`
}

// CreateServiceTemplateRequest is the request body for creating a service template.
type CreateServiceTemplateRequest struct {
	Name           string `json:"name"`
	Alias          string `json:"alias,omitzero"`
	CheckCommandID int    `json:"check_command_id,omitzero"`
}

// UpdateServiceTemplateRequest is the request body for updating a service template (PATCH).
type UpdateServiceTemplateRequest struct {
	Name                *string `json:"name,omitempty"`
	Alias               *string `json:"alias,omitempty"`
	CheckCommandID      *int    `json:"check_command_id,omitempty"`
	MaxCheckAttempts    *int    `json:"max_check_attempts,omitempty"`
	NormalCheckInterval *int    `json:"normal_check_interval,omitempty"`
	RetryCheckInterval  *int    `json:"retry_check_interval,omitempty"`
	IsActivated         *bool   `json:"is_activated,omitempty"`
}

// ServiceTemplateService provides service template configuration operations.
type ServiceTemplateService struct {
	client *Client
}

// List returns a paginated list of service templates.
func (s *ServiceTemplateService) List(ctx context.Context, opts ...ListOption) (*ListResponse[ServiceTemplate], error) {
	var resp ListResponse[ServiceTemplate]
	err := s.client.list(ctx, "/configuration/services/templates", opts, &resp)
	return &resp, err
}

// All returns an iterator over all service templates.
func (s *ServiceTemplateService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*ServiceTemplate, error] {
	return all(ctx, s.List, opts)
}

// GetByID returns the service template with the given ID using a filtered list lookup.
// Returns *NotFoundError if not found.
func (s *ServiceTemplateService) GetByID(ctx context.Context, id int) (*ServiceTemplate, error) {
	return getByID(ctx, s.List, "service template", id)
}

// Create creates a new service template and returns its ID.
func (s *ServiceTemplateService) Create(ctx context.Context, req CreateServiceTemplateRequest) (int, error) {
	var result struct {
		ID int `json:"id"`
	}
	if err := s.client.post(ctx, "/configuration/services/templates", req, &result); err != nil {
		return 0, err
	}
	return result.ID, nil
}

// Update updates an existing service template using PATCH.
func (s *ServiceTemplateService) Update(ctx context.Context, id int, req UpdateServiceTemplateRequest) error {
	return s.client.patch(ctx, fmt.Sprintf("/configuration/services/templates/%d", id), req, nil)
}

// Delete deletes a service template by ID.
func (s *ServiceTemplateService) Delete(ctx context.Context, id int) error {
	return s.client.delete(ctx, fmt.Sprintf("/configuration/services/templates/%d", id))
}
