package centreon

import (
	"context"
	"fmt"
	"iter"
)

// ResourceStatus represents the monitoring status of a resource.
type ResourceStatus struct {
	Code         int    `json:"code"`
	Name         string `json:"name"`
	SeverityCode int    `json:"severity_code"`
}

// MonitoringResourceParent represents the parent resource (typically a host for a service).
type MonitoringResourceParent struct {
	ID     int            `json:"id"`
	Name   string         `json:"name"`
	Type   string         `json:"type"`
	Status ResourceStatus `json:"status"`
}

// MonitoringResource represents a unified monitoring resource (host or service).
type MonitoringResource struct {
	ID                   int                       `json:"id"`
	Name                 string                    `json:"name"`
	Type                 string                    `json:"type"` // "host" or "service"
	Alias                string                    `json:"alias,omitzero"`
	FQDN                 string                    `json:"fqdn,omitzero"`
	HostID               int                       `json:"host_id,omitzero"`
	ServiceID            int                       `json:"service_id,omitzero"`
	MonitoringServerName string                    `json:"monitoring_server_name,omitzero"`
	Parent               *MonitoringResourceParent `json:"parent"`
	Status               ResourceStatus            `json:"status"`
	IsInDowntime         bool                      `json:"is_in_downtime"`
	IsAcknowledged       bool                      `json:"is_acknowledged"`
	Information          string                    `json:"information,omitzero"`
	Tries                string                    `json:"tries,omitzero"`
	LastStatusChange     string                    `json:"last_status_change,omitzero"`
	NotificationEnabled  bool                      `json:"is_notification_enabled"`
}

// MonitoringResourceService provides access to the unified monitoring resources endpoint.
type MonitoringResourceService struct {
	client *Client
}

// List returns a paginated list of monitoring resources.
func (s *MonitoringResourceService) List(ctx context.Context, opts ...ListOption) (*ListResponse[MonitoringResource], error) {
	var resp ListResponse[MonitoringResource]
	err := s.client.list(ctx, "/monitoring/resources", opts, &resp)
	return &resp, err
}

// All returns an iterator over all monitoring resources.
func (s *MonitoringResourceService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*MonitoringResource, error] {
	return all(ctx, s.List, opts)
}

// GetHost returns the monitoring resource for a specific host.
func (s *MonitoringResourceService) GetHost(ctx context.Context, hostID int) (*MonitoringResource, error) {
	var result MonitoringResource
	if err := s.client.get(ctx, fmt.Sprintf("/monitoring/resources/hosts/%d", hostID), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetService returns the monitoring resource for a specific service on a host.
func (s *MonitoringResourceService) GetService(ctx context.Context, hostID, serviceID int) (*MonitoringResource, error) {
	var result MonitoringResource
	if err := s.client.get(ctx, fmt.Sprintf("/monitoring/resources/hosts/%d/services/%d", hostID, serviceID), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// WithResourceTypes filters the /monitoring/resources listing by resource type.
// It emits the "types" query parameter. Allowed values: host, service,
// metaservice. Verified against centreon-web FindResourcesRequestValidator.php
// source, not yet against a live instance.
func WithResourceTypes(types ...string) ListOption {
	return WithArrayFilter("types", types...)
}

// WithStatuses filters the /monitoring/resources listing by status. It emits the
// "statuses" query parameter. Allowed values: OK, UP, WARNING, DOWN, CRITICAL,
// UNREACHABLE, UNKNOWN, PENDING. Verified against centreon-web
// FindResourcesRequestValidator.php source, not yet against a live instance.
func WithStatuses(statuses ...string) ListOption {
	return WithArrayFilter("statuses", statuses...)
}

// WithStatusTypes filters the /monitoring/resources listing by status type. It
// emits the "status_types" query parameter. Allowed values: hard, soft.
// Verified against centreon-web FindResourcesRequestValidator.php source, not
// yet against a live instance.
func WithStatusTypes(statusTypes ...string) ListOption {
	return WithArrayFilter("status_types", statusTypes...)
}

// WithStates filters the /monitoring/resources listing by state. It emits the
// "states" query parameter. Allowed values: unhandled_problems,
// resources_problems, in_downtime, acknowledged, in_flapping, all. Verified
// against centreon-web FindResourcesRequestValidator.php source, not yet against
// a live instance.
func WithStates(states ...string) ListOption {
	return WithArrayFilter("states", states...)
}

// WithHostGroupNames filters the /monitoring/resources listing by host group
// name. It emits the "hostgroup_names" query parameter. Verified against
// centreon-web FindResourcesRequestValidator.php source, not yet against a live
// instance.
func WithHostGroupNames(names ...string) ListOption {
	return WithArrayFilter("hostgroup_names", names...)
}

// WithServiceGroupNames filters the /monitoring/resources listing by service
// group name. It emits the "servicegroup_names" query parameter. Verified
// against centreon-web FindResourcesRequestValidator.php source, not yet against
// a live instance.
func WithServiceGroupNames(names ...string) ListOption {
	return WithArrayFilter("servicegroup_names", names...)
}

// WithHostCategoryNames filters the /monitoring/resources listing by host
// category name. It emits the "host_category_names" query parameter. Verified
// against centreon-web FindResourcesRequestValidator.php source, not yet against
// a live instance.
func WithHostCategoryNames(names ...string) ListOption {
	return WithArrayFilter("host_category_names", names...)
}

// WithServiceCategoryNames filters the /monitoring/resources listing by service
// category name. It emits the "service_category_names" query parameter. Verified
// against centreon-web FindResourcesRequestValidator.php source, not yet against
// a live instance.
func WithServiceCategoryNames(names ...string) ListOption {
	return WithArrayFilter("service_category_names", names...)
}

// WithMonitoringServerNames filters the /monitoring/resources listing by
// monitoring server (poller) name. It emits the "monitoring_server_names" query
// parameter. Verified against centreon-web FindResourcesRequestValidator.php
// source, not yet against a live instance.
func WithMonitoringServerNames(names ...string) ListOption {
	return WithArrayFilter("monitoring_server_names", names...)
}
