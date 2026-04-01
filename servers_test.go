package centreon

import (
	"net/http"
	"testing"
)

func TestMonitoringServerService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/monitoring-servers", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[MonitoringServer]{
			Result: []MonitoringServer{
				{ID: 1, Name: "Central", Address: "192.168.1.1", IsActivated: true, IsDefault: true},
				{ID: 2, Name: "Poller", Address: "192.168.1.2", IsActivated: true, IsDefault: false},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 2},
		})
	})

	resp, err := c.MonitoringServers.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}
	if resp.Result[0].Name != "Central" {
		t.Errorf("Result[0].Name = %q, want %q", resp.Result[0].Name, "Central")
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
