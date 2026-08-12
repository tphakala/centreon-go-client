package centreon

import (
	"net/http"
	"testing"
)

func TestMonitoringServerService_List_AllFields(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/monitoring-servers", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{
					"id": 1, "name": "Central", "address": "127.0.0.1",
					"is_activate": true, "is_default": true,
					"is_localhost": true, "is_updated": false,
					"engine_start_command":       "/usr/sbin/centengine",
					"engine_stop_command":        "service centengine stop",
					"engine_restart_command":     "service centengine restart",
					"engine_reload_command":      "service centengine reload",
					"nagios_bin":                 "/usr/sbin/centengine",
					"nagiostats_bin":             "/usr/sbin/centenginestats",
					"broker_reload_command":      "service cbd reload",
					"centreonbroker_cfg_path":    "/etc/centreon-broker",
					"centreonbroker_module_path": "/usr/share/centreon/lib/centreon-broker",
					"centreonbroker_logs_path":   "/var/log/centreon-broker",
					"centreonconnector_path":     "/usr/lib64/centreon-connector",
					"ssh_port":                   22,
					"init_script_centreontrapd":  "centreontrapd",
					"snmp_trapd_path_conf":       "/etc/snmp/centreon_traps",
					"remote_id":                  nil,
					"remote_server_use_as_proxy": true,
				},
				{
					"id": 2, "name": "Remote", "address": "10.0.0.9",
					"is_activate": true, "is_default": false,
					"is_localhost": false, "is_updated": true,
					"ssh_port":                   2222,
					"remote_id":                  5,
					"remote_server_use_as_proxy": false,
				},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 2},
		})
	})

	resp, err := c.MonitoringServers.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}

	s := resp.Result[0]
	wantBool(t, "IsLocalhost", s.IsLocalhost, true)
	wantBool(t, "IsUpdated", s.IsUpdated, false)
	wantStr(t, "EngineStartCommand", s.EngineStartCommand, "/usr/sbin/centengine")
	wantStr(t, "EngineStopCommand", s.EngineStopCommand, "service centengine stop")
	wantStr(t, "EngineRestartCommand", s.EngineRestartCommand, "service centengine restart")
	wantStr(t, "EngineReloadCommand", s.EngineReloadCommand, "service centengine reload")
	wantStr(t, "NagiosBin", s.NagiosBin, "/usr/sbin/centengine")
	wantStr(t, "NagiostatsBin", s.NagiostatsBin, "/usr/sbin/centenginestats")
	wantStr(t, "BrokerReloadCommand", s.BrokerReloadCommand, "service cbd reload")
	wantStr(t, "CentreonBrokerCfgPath", s.CentreonBrokerCfgPath, "/etc/centreon-broker")
	wantStr(t, "CentreonBrokerModulePath", s.CentreonBrokerModulePath, "/usr/share/centreon/lib/centreon-broker")
	wantStr(t, "CentreonBrokerLogsPath", s.CentreonBrokerLogsPath, "/var/log/centreon-broker")
	wantStr(t, "CentreonConnectorPath", s.CentreonConnectorPath, "/usr/lib64/centreon-connector")
	wantInt(t, "SSHPort", s.SSHPort, 22)
	wantStr(t, "InitScriptCentreontrapd", s.InitScriptCentreontrapd, "centreontrapd")
	wantStr(t, "SNMPTrapdPathConf", s.SNMPTrapdPathConf, "/etc/snmp/centreon_traps")
	wantNilIntPtr(t, "RemoteID", s.RemoteID)
	wantBool(t, "RemoteServerUseAsProxy", s.RemoteServerUseAsProxy, true)

	s1 := resp.Result[1]
	wantIntPtr(t, "Result[1].RemoteID", s1.RemoteID, 5)
	wantInt(t, "Result[1].SSHPort", s1.SSHPort, 2222)
	wantBool(t, "Result[1].IsLocalhost", s1.IsLocalhost, false)
	wantBool(t, "Result[1].IsUpdated", s1.IsUpdated, true)
}

func TestMonitoringServerService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/monitoring-servers", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{"id": 1, "name": "Central", "address": "192.168.1.1", "is_activate": true, "is_default": true},
				{"id": 2, "name": "Poller", "address": "192.168.1.2", "is_activate": true, "is_default": false},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 2},
		})
	})

	resp, err := c.MonitoringServers.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}
	s := resp.Result[0]
	if s.Name != "Central" {
		t.Errorf("Result[0].Name = %q, want %q", s.Name, "Central")
	}
	if !s.IsActivated {
		t.Error("Result[0].IsActivated = false, want true")
	}
	if resp.Result[1].Name != "Poller" {
		t.Errorf("Result[1].Name = %q, want %q", resp.Result[1].Name, "Poller")
	}
}

func TestMonitoringServerService_GenerateAndReload(t *testing.T) {
	mux, c := newTestMux(t)

	var called bool
	mux.HandleFunc("GET /centreon/api/latest/configuration/monitoring-servers/42/generate-and-reload", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.MonitoringServers.GenerateAndReload(t.Context(), 42)
	if err != nil {
		t.Fatalf("GenerateAndReload: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestMonitoringServerService_GenerateAndReloadAll(t *testing.T) {
	mux, c := newTestMux(t)

	var called bool
	mux.HandleFunc("GET /centreon/api/latest/configuration/monitoring-servers/generate-and-reload", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.MonitoringServers.GenerateAndReloadAll(t.Context())
	if err != nil {
		t.Fatalf("GenerateAndReloadAll: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}
