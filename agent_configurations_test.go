package centreon

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestAgentConfigurationService_List(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("GET /centreon/api/latest/configuration/agent-configurations", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[AgentConfiguration]{
			Result: []AgentConfiguration{
				{ID: 1, Name: "tg-1", Type: "telegraf", IsAgentInitiated: true,
					Pollers: []AgentPoller{{ID: 1, Name: "Central", IsCentral: true}}},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 1},
		})
	})
	resp, err := c.AgentConfigurations.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 1 {
		t.Fatalf("len(Result) = %d, want 1", len(resp.Result))
	}
	got := resp.Result[0]
	wantInt(t, "ID", got.ID, 1)
	wantStr(t, "Type", got.Type, "telegraf")
	wantBool(t, "IsAgentInitiated", got.IsAgentInitiated, true)
	if len(got.Pollers) != 1 {
		t.Fatalf("len(Pollers) = %d, want 1", len(got.Pollers))
	}
	wantInt(t, "Pollers[0].ID", got.Pollers[0].ID, 1)
	wantBool(t, "Pollers[0].IsCentral", got.Pollers[0].IsCentral, true)
}

func TestAgentConfigurationService_Get(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("GET /centreon/api/latest/configuration/agent-configurations/7", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"id": 7, "name": "tg-7", "type": "telegraf", "connection_mode": "no-tls",
			"configuration": map[string]any{"conf_server_port": 1443, "otel_public_certificate": "/etc/pki/pub.crt"},
			"pollers":       []map[string]any{{"id": 1, "name": "Central"}},
		})
	})
	ac, err := c.AgentConfigurations.Get(t.Context(), 7)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	wantInt(t, "ID", ac.ID, 7)
	wantStr(t, "Type", ac.Type, "telegraf")
	wantStr(t, "ConnectionMode", ac.ConnectionMode, "no-tls")
	if len(ac.Pollers) != 1 || ac.Pollers[0].ID != 1 {
		t.Fatalf("Pollers = %+v, want one ref to id 1", ac.Pollers)
	}
	var conf struct {
		ConfServerPort        int    `json:"conf_server_port"`
		OtelPublicCertificate string `json:"otel_public_certificate"`
	}
	if err := json.Unmarshal(ac.Configuration, &conf); err != nil {
		t.Fatalf("decode configuration: %v", err)
	}
	wantInt(t, "conf_server_port", conf.ConfServerPort, 1443)
	wantStr(t, "otel_public_certificate", conf.OtelPublicCertificate, "/etc/pki/pub.crt")
}

func TestAgentConfigurationService_Create(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("POST /centreon/api/latest/configuration/agent-configurations", func(w http.ResponseWriter, r *http.Request) {
		var req CreateAgentConfigurationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		wantStr(t, "Type", req.Type, "telegraf")
		wantStr(t, "ConnectionMode", req.ConnectionMode, "secure")
		if len(req.PollerIDs) != 1 || req.PollerIDs[0] != 1 {
			t.Errorf("PollerIDs = %v, want [1]", req.PollerIDs)
		}
		var conf map[string]any
		if err := json.Unmarshal(req.Configuration, &conf); err != nil {
			t.Errorf("configuration not valid JSON: %v", err)
		}
		writeJSON(w, http.StatusCreated, map[string]int{"id": 42})
	})
	id, err := c.AgentConfigurations.Create(t.Context(), &CreateAgentConfigurationRequest{
		Type: "telegraf", Name: "tg-new", ConnectionMode: "secure",
		PollerIDs: []int{1}, Configuration: json.RawMessage(`{"conf_server_port":1443}`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 42 {
		t.Errorf("id = %d, want 42", id)
	}
}

func TestAgentConfigurationService_CreateNormalizesNilPollerIDs(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("POST /centreon/api/latest/configuration/agent-configurations", func(w http.ResponseWriter, r *http.Request) {
		var raw map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got := string(raw["poller_ids"]); got != "[]" {
			t.Errorf("poller_ids = %s, want []", got)
		}
		writeJSON(w, http.StatusCreated, map[string]int{"id": 1})
	})
	if _, err := c.AgentConfigurations.Create(t.Context(), &CreateAgentConfigurationRequest{
		Type: "telegraf", Name: "tg", ConnectionMode: "secure",
		Configuration: json.RawMessage(`{}`),
	}); err != nil {
		t.Fatalf("Create: %v", err)
	}
}

func TestAgentConfigurationService_UpdateNormalizesNilPollerIDs(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("PUT /centreon/api/latest/configuration/agent-configurations/7", func(w http.ResponseWriter, r *http.Request) {
		var raw map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got := string(raw["poller_ids"]); got != "[]" {
			t.Errorf("poller_ids = %s, want []", got)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.AgentConfigurations.Update(t.Context(), 7, &UpdateAgentConfigurationRequest{
		Name: "tg", Configuration: json.RawMessage(`{}`),
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestAgentConfigurationService_Update(t *testing.T) {
	mux, c := newTestMux(t)
	var called bool
	mux.HandleFunc("PUT /centreon/api/latest/configuration/agent-configurations/7", func(w http.ResponseWriter, r *http.Request) {
		called = true
		var req UpdateAgentConfigurationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		wantStr(t, "Name", req.Name, "tg-upd")
		wantStr(t, "ConnectionMode", req.ConnectionMode, "no-tls")
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.AgentConfigurations.Update(t.Context(), 7, &UpdateAgentConfigurationRequest{
		Type: "telegraf", Name: "tg-upd", ConnectionMode: "no-tls",
		PollerIDs: []int{1}, Configuration: json.RawMessage(`{"conf_server_port":1443}`),
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestAgentConfigurationService_Delete(t *testing.T) {
	mux, c := newTestMux(t)
	var called bool
	mux.HandleFunc("DELETE /centreon/api/latest/configuration/agent-configurations/7", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.AgentConfigurations.Delete(t.Context(), 7); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}
