package centreon

import (
	"context"
	"fmt"
	"iter"
)

// MonitoringServer represents a Centreon monitoring server (poller).
//
// Unlike the template structs (whose fields mirror the live-verified HostDetail
// and ServiceDetail types), the scalar poller fields below are typed by
// inference from the Centreon /configuration/monitoring-servers response: the
// booleans follow the proven is_activate/is_default typing already used on this
// same endpoint, and every field is exercised by a round-trip decode test.
//
// The last_restart field returned by /configuration/monitoring-servers is not
// modeled yet: its wire type is unconfirmed (RFC3339 string vs unix integer),
// and mistyping it would fail the whole List decode. See issue #23.
type MonitoringServer struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Address     string `json:"address,omitzero"`
	IsActivated bool   `json:"is_activate"`
	IsDefault   bool   `json:"is_default"`
	IsLocalhost bool   `json:"is_localhost"`
	IsUpdated   bool   `json:"is_updated"`

	EngineStartCommand   string `json:"engine_start_command,omitzero"`
	EngineStopCommand    string `json:"engine_stop_command,omitzero"`
	EngineRestartCommand string `json:"engine_restart_command,omitzero"`
	EngineReloadCommand  string `json:"engine_reload_command,omitzero"`
	NagiosBin            string `json:"nagios_bin,omitzero"`
	NagiostatsBin        string `json:"nagiostats_bin,omitzero"`

	BrokerReloadCommand      string `json:"broker_reload_command,omitzero"`
	CentreonBrokerCfgPath    string `json:"centreonbroker_cfg_path,omitzero"`
	CentreonBrokerModulePath string `json:"centreonbroker_module_path,omitzero"`
	CentreonBrokerLogsPath   string `json:"centreonbroker_logs_path,omitzero"`
	CentreonConnectorPath    string `json:"centreonconnector_path,omitzero"`

	SSHPort                 int    `json:"ssh_port"`
	InitScriptCentreontrapd string `json:"init_script_centreontrapd,omitzero"`
	SNMPTrapdPathConf       string `json:"snmp_trapd_path_conf,omitzero"`

	RemoteID               *int `json:"remote_id"`
	RemoteServerUseAsProxy bool `json:"remote_server_use_as_proxy"`
}

type MonitoringServerService struct {
	client *Client
}

func (s *MonitoringServerService) List(ctx context.Context, opts ...ListOption) (*ListResponse[MonitoringServer], error) {
	var resp ListResponse[MonitoringServer]
	err := s.client.list(ctx, "/configuration/monitoring-servers", opts, &resp)
	return &resp, err
}

func (s *MonitoringServerService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*MonitoringServer, error] {
	return all(ctx, s.List, opts)
}

func (s *MonitoringServerService) GenerateAndReload(ctx context.Context, serverID int) error {
	return s.client.get(ctx, fmt.Sprintf("/configuration/monitoring-servers/%d/generate-and-reload", serverID), nil)
}

func (s *MonitoringServerService) GenerateAndReloadAll(ctx context.Context) error {
	return s.client.get(ctx, "/configuration/monitoring-servers/generate-and-reload", nil)
}
