package centreon

import (
	"context"
	"fmt"
	"iter"
)

// Macro represents a custom host or service macro.
type Macro struct {
	Name        string `json:"name"`
	Value       string `json:"value,omitzero"`
	IsPassword  bool   `json:"is_password"`
	Description string `json:"description,omitzero"`
}

// NamedRef is a lightweight reference to a named Centreon resource.
// The Centreon API uses {"id": N, "name": "..."} objects for relationships
// such as templates, categories, groups, and monitoring servers.
type NamedRef struct {
	ID   int    `json:"id"`
	Name string `json:"name,omitzero"`
}

// Host represents a Centreon host configuration resource.
type Host struct {
	ID                  int        `json:"id"`
	MonitoringServer    NamedRef   `json:"monitoring_server"`
	Name                string     `json:"name"`
	Address             string     `json:"address"`
	Alias               string     `json:"alias,omitzero"`
	NormalCheckInterval *int       `json:"normal_check_interval"`
	RetryCheckInterval  *int       `json:"retry_check_interval"`
	CheckTimeperiod     *NamedRef  `json:"check_timeperiod"`
	NotifTimeperiod     *NamedRef  `json:"notification_timeperiod"`
	Severity            *NamedRef  `json:"severity"`
	Templates           []NamedRef `json:"templates,omitzero"`
	Categories          []NamedRef `json:"categories,omitzero"`
	Groups              []NamedRef `json:"groups,omitzero"`
	IsActivated         bool       `json:"is_activated"`
}

// CreateHostRequest is the request body for creating a host.
type CreateHostRequest struct {
	// Required
	MonitoringServerID int    `json:"monitoring_server_id"`
	Name               string `json:"name"`
	Address            string `json:"address"`

	// Optional basic
	Alias       string `json:"alias,omitzero"`
	Comment     string `json:"comment,omitzero"`
	GeoCoords   string `json:"geo_coords,omitzero"`
	IsActivated *bool  `json:"is_activated,omitempty"`

	// Monitoring config
	CheckCommandID      int      `json:"check_command_id,omitzero"`
	CheckCommandArgs    []string `json:"check_command_args,omitzero"`
	CheckTimeperiodID   int      `json:"check_timeperiod_id,omitzero"`
	MaxCheckAttempts    int      `json:"max_check_attempts,omitzero"`
	NormalCheckInterval int      `json:"normal_check_interval,omitzero"`
	RetryCheckInterval  int      `json:"retry_check_interval,omitzero"`

	// SNMP
	SNMPCommunity string `json:"snmp_community,omitzero"`
	SNMPVersion   string `json:"snmp_version,omitzero"`

	// Check toggles (0=Disabled, 1=Enabled, 2=Inherit)
	ActiveCheckEnabled  int `json:"active_check_enabled,omitzero"`
	PassiveCheckEnabled int `json:"passive_check_enabled,omitzero"`

	// Freshness
	FreshnessChecked   int `json:"freshness_checked,omitzero"`
	FreshnessThreshold int `json:"freshness_threshold,omitzero"`

	// Flap detection
	FlapDetectionEnabled int `json:"flap_detection_enabled,omitzero"`
	LowFlapThreshold     int `json:"low_flap_threshold,omitzero"`
	HighFlapThreshold    int `json:"high_flap_threshold,omitzero"`

	// Event handler
	EventHandlerEnabled     int    `json:"event_handler_enabled,omitzero"`
	EventHandlerCommandID   int    `json:"event_handler_command_id,omitzero"`
	EventHandlerCommandArgs string `json:"event_handler_command_args,omitzero"`

	// Notifications
	NotificationEnabled       int `json:"notification_enabled,omitzero"`
	NotificationOptions       int `json:"notification_options,omitzero"`
	NotificationInterval      int `json:"notification_interval,omitzero"`
	NotificationTimeperiodID  int `json:"notification_timeperiod_id,omitzero"`
	FirstNotificationDelay    int `json:"first_notification_delay,omitzero"`
	RecoveryNotificationDelay int `json:"recovery_notification_delay,omitzero"`
	AcknowledgementTimeout    int `json:"acknowledgement_timeout,omitzero"`

	// References
	TimezoneID int `json:"timezone_id,omitzero"`
	SeverityID int `json:"severity_id,omitzero"`
	IconID     int `json:"icon_id,omitzero"`

	// Descriptive
	Note            string `json:"note,omitzero"`
	NoteURL         string `json:"note_url,omitzero"`
	ActionURL       string `json:"action_url,omitzero"`
	IconAlternative string `json:"icon_alternative,omitzero"`

	// Relationships
	Templates  []int   `json:"templates,omitzero"`
	Groups     []int   `json:"groups,omitzero"`
	Categories []int   `json:"categories,omitzero"`
	Macros     []Macro `json:"macros,omitzero"`
}

// UpdateHostRequest is the request body for updating a host (PATCH).
type UpdateHostRequest struct {
	Name                      *string   `json:"name,omitempty"`
	Alias                     *string   `json:"alias,omitempty"`
	Address                   *string   `json:"address,omitempty"`
	Comment                   *string   `json:"comment,omitempty"`
	GeoCoords                 *string   `json:"geo_coords,omitempty"`
	IsActivated               *bool     `json:"is_activated,omitempty"`
	CheckCommandID            *int      `json:"check_command_id,omitempty"`
	CheckCommandArgs          *[]string `json:"check_command_args,omitempty"`
	CheckTimeperiodID         *int      `json:"check_timeperiod_id,omitempty"`
	MaxCheckAttempts          *int      `json:"max_check_attempts,omitempty"`
	NormalCheckInterval       *int      `json:"normal_check_interval,omitempty"`
	RetryCheckInterval        *int      `json:"retry_check_interval,omitempty"`
	ActiveCheckEnabled        *int      `json:"active_check_enabled,omitempty"`
	PassiveCheckEnabled       *int      `json:"passive_check_enabled,omitempty"`
	FreshnessChecked          *int      `json:"freshness_checked,omitempty"`
	FreshnessThreshold        *int      `json:"freshness_threshold,omitempty"`
	FlapDetectionEnabled      *int      `json:"flap_detection_enabled,omitempty"`
	LowFlapThreshold          *int      `json:"low_flap_threshold,omitempty"`
	HighFlapThreshold         *int      `json:"high_flap_threshold,omitempty"`
	EventHandlerEnabled       *int      `json:"event_handler_enabled,omitempty"`
	EventHandlerCommandID     *int      `json:"event_handler_command_id,omitempty"`
	SNMPCommunity             *string   `json:"snmp_community,omitempty"`
	SNMPVersion               *string   `json:"snmp_version,omitempty"`
	NotificationEnabled       *int      `json:"notification_enabled,omitempty"`
	NotificationOptions       *int      `json:"notification_options,omitempty"`
	NotificationInterval      *int      `json:"notification_interval,omitempty"`
	NotificationTimeperiodID  *int      `json:"notification_timeperiod_id,omitempty"`
	FirstNotificationDelay    *int      `json:"first_notification_delay,omitempty"`
	RecoveryNotificationDelay *int      `json:"recovery_notification_delay,omitempty"`
	AcknowledgementTimeout    *int      `json:"acknowledgement_timeout,omitempty"`
	TimezoneID                *int      `json:"timezone_id,omitempty"`
	SeverityID                *int      `json:"severity_id,omitempty"`
	IconID                    *int      `json:"icon_id,omitempty"`
	Note                      *string   `json:"note,omitempty"`
	NoteURL                   *string   `json:"note_url,omitempty"`
	ActionURL                 *string   `json:"action_url,omitempty"`
	IconAlternative           *string   `json:"icon_alternative,omitempty"`
	Templates                 *[]int    `json:"templates,omitempty"`
	Groups                    *[]int    `json:"groups,omitempty"`
	Categories                *[]int    `json:"categories,omitempty"`
	Macros                    *[]Macro  `json:"macros,omitempty"`
}

// HostService provides host configuration operations.
type HostService struct {
	client *Client
}

// List returns a paginated list of hosts.
func (s *HostService) List(ctx context.Context, opts ...ListOption) (*ListResponse[Host], error) {
	var resp ListResponse[Host]
	err := s.client.list(ctx, "/configuration/hosts", opts, &resp)
	return &resp, err
}

// All returns an iterator over all hosts.
func (s *HostService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*Host, error] {
	return all(ctx, s.List, opts)
}

// GetByID returns the host with the given ID using a filtered list lookup.
// Returns *NotFoundError if not found.
func (s *HostService) GetByID(ctx context.Context, id int) (*Host, error) {
	return getByID(ctx, s.List, "host", id)
}

// Create creates a new host and returns its ID.
func (s *HostService) Create(ctx context.Context, req *CreateHostRequest) (int, error) {
	var result struct {
		ID int `json:"id"`
	}
	if err := s.client.post(ctx, "/configuration/hosts", req, &result); err != nil {
		return 0, err
	}
	return result.ID, nil
}

// Update updates an existing host using PATCH.
func (s *HostService) Update(ctx context.Context, id int, req *UpdateHostRequest) error {
	return s.client.patch(ctx, fmt.Sprintf("/configuration/hosts/%d", id), req)
}

// Delete deletes a host by ID.
func (s *HostService) Delete(ctx context.Context, id int) error {
	return s.client.delete(ctx, fmt.Sprintf("/configuration/hosts/%d", id))
}
