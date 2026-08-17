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
	CheckCommandID      *int   `json:"check_command_id"`
	CheckTimeperiodID   *int   `json:"check_timeperiod_id"`
	MaxCheckAttempts    *int   `json:"max_check_attempts"`
	NormalCheckInterval *int   `json:"normal_check_interval"`
	RetryCheckInterval  *int   `json:"retry_check_interval"`

	CheckCommandArgs []string `json:"check_command_args,omitzero"`

	ServiceTemplateID *int `json:"service_template_id"`
	SeverityID        *int `json:"severity_id"`

	// HostTemplates holds the IDs of the host templates this service template
	// is attached to, as returned by /configuration/services/templates. The
	// wire shape is a JSON array of bare integer IDs (e.g. [87, 88, 1053]),
	// verified live against Centreon Web 24.10.29; an empty relationship
	// arrives as [] and decodes to an empty slice. See issue #22.
	HostTemplates []int `json:"host_templates,omitzero"`

	ActiveCheckEnabled  int `json:"active_check_enabled"`
	PassiveCheckEnabled int `json:"passive_check_enabled"`
	VolatilityEnabled   int `json:"volatility_enabled"`

	NotificationEnabled               int  `json:"notification_enabled"`
	IsContactAdditiveInheritance      bool `json:"is_contact_additive_inheritance"`
	IsContactGroupAdditiveInheritance bool `json:"is_contact_group_additive_inheritance"`
	NotificationInterval              *int `json:"notification_interval"`
	NotificationTimeperiodID          *int `json:"notification_timeperiod_id"`
	NotificationType                  *int `json:"notification_type"`
	FirstNotificationDelay            *int `json:"first_notification_delay"`
	RecoveryNotificationDelay         *int `json:"recovery_notification_delay"`
	AcknowledgementTimeout            *int `json:"acknowledgement_timeout"`

	FreshnessChecked   int  `json:"freshness_checked"`
	FreshnessThreshold *int `json:"freshness_threshold"`

	FlapDetectionEnabled int  `json:"flap_detection_enabled"`
	LowFlapThreshold     *int `json:"low_flap_threshold"`
	HighFlapThreshold    *int `json:"high_flap_threshold"`

	EventHandlerEnabled     int      `json:"event_handler_enabled"`
	EventHandlerCommandID   *int     `json:"event_handler_command_id"`
	EventHandlerCommandArgs []string `json:"event_handler_command_args,omitzero"`

	GraphTemplateID *int   `json:"graph_template_id"`
	NoteURL         string `json:"note_url,omitzero"`
	Note            string `json:"note,omitzero"`
	ActionURL       string `json:"action_url,omitzero"`
	IconID          *int   `json:"icon_id"`
	IconAlternative string `json:"icon_alternative,omitzero"`
	Comment         string `json:"comment,omitzero"`

	IsLocked bool `json:"is_locked"`
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
	CheckTimeperiodID   *int    `json:"check_timeperiod_id,omitempty"`
	MaxCheckAttempts    *int    `json:"max_check_attempts,omitempty"`
	NormalCheckInterval *int    `json:"normal_check_interval,omitempty"`
	RetryCheckInterval  *int    `json:"retry_check_interval,omitempty"`
	IsLocked            *bool   `json:"is_locked,omitempty"`
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
	return s.client.patch(ctx, fmt.Sprintf("/configuration/services/templates/%d", id), req)
}

// Delete deletes a service template by ID.
func (s *ServiceTemplateService) Delete(ctx context.Context, id int) error {
	return s.client.delete(ctx, fmt.Sprintf("/configuration/services/templates/%d", id))
}
