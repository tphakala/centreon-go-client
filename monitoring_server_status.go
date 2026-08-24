package centreon

import (
	"context"
	"iter"
)

// MonitoringServerStatus represents the real-time health of a Centreon
// monitoring server (poller), as returned by GET /monitoring/servers.
//
// This is distinct from the configuration-side MonitoringServer
// (/configuration/monitoring-servers): the realtime endpoint reports live
// poller health (is_running, last_alive, the running engine version) rather
// than the poller's configuration. It is list-only; GET /monitoring/servers/{id}
// returns HTTP 404 (verified live on Centreon Web 25.10.16), so there is no Get.
//
// LastAlive is the poller's last-heartbeat time as a Unix timestamp (epoch
// seconds), NOT the RFC3339 string that the config counterpart's last_restart
// uses (see MonitoringServer.LastRestart). It is typed int64, not int, so the
// second count is safe past 2038 on 32-bit platforms. It is a value, not a
// pointer: the endpoint coerces a null heartbeat to 0 rather than emitting JSON
// null. This was verified live on 25.10.16 by nulling the underlying
// instances.last_alive / running columns, after which the endpoint returned
// "last_alive": 0 and "is_running": false (not null) for that poller.
type MonitoringServerStatus struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Address   string `json:"address,omitzero"`
	IsRunning bool   `json:"is_running"`
	LastAlive int64  `json:"last_alive"`
	Version   string `json:"version,omitzero"`
}

// MonitoringServerStatusService provides access to the real-time monitoring
// servers endpoint (GET /monitoring/servers). See issue #86.
type MonitoringServerStatusService struct {
	client *Client
}

// List returns a paginated list of real-time monitoring server statuses.
func (s *MonitoringServerStatusService) List(ctx context.Context, opts ...ListOption) (*ListResponse[MonitoringServerStatus], error) {
	var resp ListResponse[MonitoringServerStatus]
	err := s.client.list(ctx, "/monitoring/servers", opts, &resp)
	return &resp, err
}

// All returns an iterator over all real-time monitoring server statuses.
func (s *MonitoringServerStatusService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*MonitoringServerStatus, error] {
	return all(ctx, s.List, opts)
}
