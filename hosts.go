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

	// Notifications
	NotificationEnabled      int `json:"notification_enabled,omitzero"`
	NotificationOptions      int `json:"notification_options,omitzero"`
	NotificationInterval     int `json:"notification_interval,omitzero"`
	NotificationTimeperiodID int `json:"notification_timeperiod_id,omitzero"`

	// References
	TimezoneID int `json:"timezone_id,omitzero"`
	SeverityID int `json:"severity_id,omitzero"`
	IconID     int `json:"icon_id,omitzero"`

	// Relationships
	Templates  []int   `json:"templates,omitzero"`
	Groups     []int   `json:"groups,omitzero"`
	Categories []int   `json:"categories,omitzero"`
	Macros     []Macro `json:"macros,omitzero"`
}

// UpdateHostRequest is the request body for updating a host (PATCH).
type UpdateHostRequest struct {
	Name                *string   `json:"name,omitempty"`
	Alias               *string   `json:"alias,omitempty"`
	Address             *string   `json:"address,omitempty"`
	CheckCommandID      *int      `json:"check_command_id,omitempty"`
	CheckCommandArgs    *[]string `json:"check_command_args,omitempty"`
	CheckTimeperiodID   *int      `json:"check_timeperiod_id,omitempty"`
	MaxCheckAttempts    *int      `json:"max_check_attempts,omitempty"`
	NormalCheckInterval *int      `json:"normal_check_interval,omitempty"`
	RetryCheckInterval  *int      `json:"retry_check_interval,omitempty"`
	ActiveCheckEnabled  *int      `json:"active_check_enabled,omitempty"`
	PassiveCheckEnabled *int      `json:"passive_check_enabled,omitempty"`
	IsActivated         *bool     `json:"is_activated,omitempty"`
	SNMPCommunity       *string   `json:"snmp_community,omitempty"`
	SNMPVersion         *string   `json:"snmp_version,omitempty"`
	NotificationEnabled *int      `json:"notification_enabled,omitempty"`
	TimezoneID          *int      `json:"timezone_id,omitempty"`
	SeverityID          *int      `json:"severity_id,omitempty"`
	Templates           *[]int    `json:"templates,omitempty"`
	Groups              *[]int    `json:"groups,omitempty"`
	Categories          *[]int    `json:"categories,omitempty"`
	Macros              *[]Macro  `json:"macros,omitempty"`
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
