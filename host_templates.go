package centreon

import (
	"context"
	"fmt"
	"iter"
)

// HostTemplate represents a Centreon host template configuration resource.
type HostTemplate struct {
	ID                  int    `json:"id"`
	Name                string `json:"name"`
	Alias               string `json:"alias,omitzero"`
	CheckCommandID      *int   `json:"check_command_id"`
	CheckTimeperiodID   *int   `json:"check_timeperiod_id"`
	MaxCheckAttempts    *int   `json:"max_check_attempts"`
	NormalCheckInterval *int   `json:"normal_check_interval"`
	RetryCheckInterval  *int   `json:"retry_check_interval"`

	CheckCommandArgs []string `json:"check_command_args,omitzero"`

	SNMPVersion   string `json:"snmp_version,omitzero"`
	SNMPCommunity string `json:"snmp_community,omitzero"`

	TimezoneID *int `json:"timezone_id"`
	SeverityID *int `json:"severity_id"`

	ActiveCheckEnabled  int `json:"active_check_enabled"`
	PassiveCheckEnabled int `json:"passive_check_enabled"`

	NotificationEnabled       int  `json:"notification_enabled"`
	NotificationOptions       *int `json:"notification_options"`
	NotificationInterval      *int `json:"notification_interval"`
	NotificationTimeperiodID  *int `json:"notification_timeperiod_id"`
	AddInheritedContactGroup  bool `json:"add_inherited_contact_group"`
	AddInheritedContact       bool `json:"add_inherited_contact"`
	FirstNotificationDelay    *int `json:"first_notification_delay"`
	RecoveryNotificationDelay *int `json:"recovery_notification_delay"`
	AcknowledgementTimeout    *int `json:"acknowledgement_timeout"`

	FreshnessChecked   int  `json:"freshness_checked"`
	FreshnessThreshold *int `json:"freshness_threshold"`

	FlapDetectionEnabled int  `json:"flap_detection_enabled"`
	LowFlapThreshold     *int `json:"low_flap_threshold"`
	HighFlapThreshold    *int `json:"high_flap_threshold"`

	EventHandlerEnabled     int      `json:"event_handler_enabled"`
	EventHandlerCommandID   *int     `json:"event_handler_command_id"`
	EventHandlerCommandArgs []string `json:"event_handler_command_args,omitzero"`

	NoteURL         string `json:"note_url,omitzero"`
	Note            string `json:"note,omitzero"`
	ActionURL       string `json:"action_url,omitzero"`
	IconID          *int   `json:"icon_id"`
	IconAlternative string `json:"icon_alternative,omitzero"`
	Comment         string `json:"comment,omitzero"`

	IsLocked bool `json:"is_locked"`
}

// HostTemplateDetail is the full host template configuration returned by
// HostTemplateService.Get (GET /configuration/hosts/templates/{id}) on
// Centreon 25.10+. It carries every field of the list representation
// (HostTemplate) plus the categories, parent templates, and custom macros that
// the list endpoint omits. The fields are redeclared flat (rather than
// embedding HostTemplate) to match HostDetail/ServiceDetail and to insulate the
// detail model from any future drift in the list endpoint.
type HostTemplateDetail struct {
	ID                  int    `json:"id"`
	Name                string `json:"name"`
	Alias               string `json:"alias,omitzero"`
	CheckCommandID      *int   `json:"check_command_id"`
	CheckTimeperiodID   *int   `json:"check_timeperiod_id"`
	MaxCheckAttempts    *int   `json:"max_check_attempts"`
	NormalCheckInterval *int   `json:"normal_check_interval"`
	RetryCheckInterval  *int   `json:"retry_check_interval"`

	CheckCommandArgs []string `json:"check_command_args,omitzero"`

	SNMPVersion   string `json:"snmp_version,omitzero"`
	SNMPCommunity string `json:"snmp_community,omitzero"`

	TimezoneID *int `json:"timezone_id"`
	SeverityID *int `json:"severity_id"`

	ActiveCheckEnabled  int `json:"active_check_enabled"`
	PassiveCheckEnabled int `json:"passive_check_enabled"`

	NotificationEnabled       int  `json:"notification_enabled"`
	NotificationOptions       *int `json:"notification_options"`
	NotificationInterval      *int `json:"notification_interval"`
	NotificationTimeperiodID  *int `json:"notification_timeperiod_id"`
	AddInheritedContactGroup  bool `json:"add_inherited_contact_group"`
	AddInheritedContact       bool `json:"add_inherited_contact"`
	FirstNotificationDelay    *int `json:"first_notification_delay"`
	RecoveryNotificationDelay *int `json:"recovery_notification_delay"`
	AcknowledgementTimeout    *int `json:"acknowledgement_timeout"`

	FreshnessChecked   int  `json:"freshness_checked"`
	FreshnessThreshold *int `json:"freshness_threshold"`

	FlapDetectionEnabled int  `json:"flap_detection_enabled"`
	LowFlapThreshold     *int `json:"low_flap_threshold"`
	HighFlapThreshold    *int `json:"high_flap_threshold"`

	EventHandlerEnabled     int      `json:"event_handler_enabled"`
	EventHandlerCommandID   *int     `json:"event_handler_command_id"`
	EventHandlerCommandArgs []string `json:"event_handler_command_args,omitzero"`

	NoteURL         string `json:"note_url,omitzero"`
	Note            string `json:"note,omitzero"`
	ActionURL       string `json:"action_url,omitzero"`
	IconID          *int   `json:"icon_id"`
	IconAlternative string `json:"icon_alternative,omitzero"`
	Comment         string `json:"comment,omitzero"`

	IsLocked bool `json:"is_locked"`

	Categories []NamedRef  `json:"categories,omitzero"`
	Templates  []NamedRef  `json:"templates,omitzero"`
	Macros     []HostMacro `json:"macros,omitzero"`
}

// CreateHostTemplateRequest is the request body for creating a host template.
type CreateHostTemplateRequest struct {
	Name           string `json:"name"`
	Alias          string `json:"alias,omitzero"`
	CheckCommandID int    `json:"check_command_id,omitzero"`
}

// UpdateHostTemplateRequest is the request body for updating a host template (PATCH).
type UpdateHostTemplateRequest struct {
	Name                *string `json:"name,omitempty"`
	Alias               *string `json:"alias,omitempty"`
	CheckCommandID      *int    `json:"check_command_id,omitempty"`
	CheckTimeperiodID   *int    `json:"check_timeperiod_id,omitempty"`
	MaxCheckAttempts    *int    `json:"max_check_attempts,omitempty"`
	NormalCheckInterval *int    `json:"normal_check_interval,omitempty"`
	RetryCheckInterval  *int    `json:"retry_check_interval,omitempty"`
	IsLocked            *bool   `json:"is_locked,omitempty"`
}

// HostTemplateService provides host template configuration operations.
type HostTemplateService struct {
	client *Client
}

// List returns a paginated list of host templates.
func (s *HostTemplateService) List(ctx context.Context, opts ...ListOption) (*ListResponse[HostTemplate], error) {
	var resp ListResponse[HostTemplate]
	err := s.client.list(ctx, "/configuration/hosts/templates", opts, &resp)
	return &resp, err
}

// All returns an iterator over all host templates.
func (s *HostTemplateService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*HostTemplate, error] {
	return all(ctx, s.List, opts)
}

// GetByID returns the host template with the given ID using a filtered list lookup.
// Returns *NotFoundError if not found.
func (s *HostTemplateService) GetByID(ctx context.Context, id int) (*HostTemplate, error) {
	return getByID(ctx, s.List, "host template", id)
}

// Get returns the full host template configuration for the given ID using a
// direct GET request, including categories, parent templates, and custom
// macros. A macro's Value is nil when its IsPassword is true because the API
// masks password values on read.
//
// This requires Centreon 25.10 or later; earlier versions have no
// single-resource GET route for templates and return an *APIError with HTTP
// status 404. For a version-independent lookup of the list representation
// (without the extra detail fields), use GetByID.
func (s *HostTemplateService) Get(ctx context.Context, id int) (*HostTemplateDetail, error) {
	var t HostTemplateDetail
	if err := s.client.get(ctx, fmt.Sprintf("/configuration/hosts/templates/%d", id), &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// Create creates a new host template and returns its ID.
func (s *HostTemplateService) Create(ctx context.Context, req CreateHostTemplateRequest) (int, error) {
	var result struct {
		ID int `json:"id"`
	}
	if err := s.client.post(ctx, "/configuration/hosts/templates", req, &result); err != nil {
		return 0, err
	}
	return result.ID, nil
}

// Update updates an existing host template using PATCH.
func (s *HostTemplateService) Update(ctx context.Context, id int, req UpdateHostTemplateRequest) error {
	return s.client.patch(ctx, fmt.Sprintf("/configuration/hosts/templates/%d", id), req)
}

// Delete deletes a host template by ID.
func (s *HostTemplateService) Delete(ctx context.Context, id int) error {
	return s.client.delete(ctx, fmt.Sprintf("/configuration/hosts/templates/%d", id))
}
