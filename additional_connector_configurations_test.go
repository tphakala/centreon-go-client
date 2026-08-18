package centreon

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestAdditionalConnectorConfigurationService_List(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("GET /centreon/api/latest/configuration/additional-connector-configurations", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{{
				"id": 1, "name": "acc-1", "type": "vmware_v6", "description": "d",
				"created_at": "2026-08-18T07:42:30+00:00",
				"created_by": map[string]any{"id": 1, "name": "Creator Centreon"},
				"updated_at": "2026-08-18T09:15:45+00:00",
				"updated_by": map[string]any{"id": 2, "name": "Updater Centreon"},
			}},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 1},
		})
	})
	resp, err := c.AdditionalConnectorConfigurations.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 1 {
		t.Fatalf("len(Result) = %d, want 1", len(resp.Result))
	}
	got := resp.Result[0]
	wantInt(t, "ID", got.ID, 1)
	wantStr(t, "Type", got.Type, "vmware_v6")
	wantStr(t, "Description", got.Description, "d")
	// created_* and updated_* carry distinct values so a created/updated tag
	// swap on the struct would change the decoded result and fail this test.
	wantInt(t, "CreatedBy.ID", got.CreatedBy.ID, 1)
	wantStr(t, "CreatedBy.Name", got.CreatedBy.Name, "Creator Centreon")
	wantInt(t, "UpdatedBy.ID", got.UpdatedBy.ID, 2)
	wantStr(t, "UpdatedBy.Name", got.UpdatedBy.Name, "Updater Centreon")
	wantCreatedAt := time.Date(2026, 8, 18, 7, 42, 30, 0, time.UTC)
	if !got.CreatedAt.Equal(wantCreatedAt) {
		t.Errorf("CreatedAt = %s, want %s", got.CreatedAt, wantCreatedAt)
	}
	wantUpdatedAt := time.Date(2026, 8, 18, 9, 15, 45, 0, time.UTC)
	if !got.UpdatedAt.Equal(wantUpdatedAt) {
		t.Errorf("UpdatedAt = %s, want %s", got.UpdatedAt, wantUpdatedAt)
	}
}

func TestAdditionalConnectorConfigurationService_Get(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("GET /centreon/api/latest/configuration/additional-connector-configurations/5", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"id": 5, "name": "acc-5", "type": "vmware_v6", "description": "d",
			"parameters": map[string]any{
				"port": 5700,
				"vcenters": []map[string]any{{
					"id": 1, "name": "vc1", "url": "https://vc1/sdk", "username": "user", "password": nil,
				}},
			},
			"pollers":    []map[string]any{{"id": 1, "name": "Central"}},
			"created_at": "2026-08-18T07:42:30+00:00",
			"created_by": map[string]any{"id": 1, "name": "Admin Centreon"},
			"updated_at": "2026-08-18T07:43:23+00:00",
			"updated_by": map[string]any{"id": 1, "name": "Admin Centreon"},
		})
	})
	acc, err := c.AdditionalConnectorConfigurations.Get(t.Context(), 5)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	wantInt(t, "ID", acc.ID, 5)
	wantStr(t, "Type", acc.Type, "vmware_v6")
	if len(acc.Pollers) != 1 || acc.Pollers[0].ID != 1 {
		t.Fatalf("Pollers = %+v, want one ref to id 1", acc.Pollers)
	}
	var params struct {
		Port     int `json:"port"`
		Vcenters []struct {
			Name     string  `json:"name"`
			Password *string `json:"password"`
		} `json:"vcenters"`
	}
	if err := json.Unmarshal(acc.Parameters, &params); err != nil {
		t.Fatalf("decode parameters: %v", err)
	}
	wantInt(t, "port", params.Port, 5700)
	if len(params.Vcenters) != 1 {
		t.Fatalf("len(vcenters) = %d, want 1", len(params.Vcenters))
	}
	wantStr(t, "vcenters[0].name", params.Vcenters[0].Name, "vc1")
	if params.Vcenters[0].Password != nil {
		t.Errorf("vcenters[0].password = %q, want nil (masked)", *params.Vcenters[0].Password)
	}
}

func TestAdditionalConnectorConfigurationService_Create(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("POST /centreon/api/latest/configuration/additional-connector-configurations", func(w http.ResponseWriter, r *http.Request) {
		var req CreateAdditionalConnectorConfigurationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		wantStr(t, "Type", req.Type, "vmware_v6")
		wantStr(t, "Description", req.Description, "desc")
		if len(req.Pollers) != 1 || req.Pollers[0] != 1 {
			t.Errorf("Pollers = %v, want [1]", req.Pollers)
		}
		writeJSON(w, http.StatusCreated, map[string]int{"id": 30})
	})
	id, err := c.AdditionalConnectorConfigurations.Create(t.Context(), &CreateAdditionalConnectorConfigurationRequest{
		Type: "vmware_v6", Name: "acc-new", Description: "desc", Pollers: []int{1},
		Parameters: json.RawMessage(`{"port":5700,"vcenters":[{"name":"vc1","url":"https://vc1/sdk","username":"u","password":"p"}]}`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 30 {
		t.Errorf("id = %d, want 30", id)
	}
}

func TestAdditionalConnectorConfigurationService_CreateNormalizesNilPollers(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("POST /centreon/api/latest/configuration/additional-connector-configurations", func(w http.ResponseWriter, r *http.Request) {
		var raw map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got := string(raw["pollers"]); got != "[]" {
			t.Errorf("pollers = %s, want []", got)
		}
		writeJSON(w, http.StatusCreated, map[string]int{"id": 1})
	})
	if _, err := c.AdditionalConnectorConfigurations.Create(t.Context(), &CreateAdditionalConnectorConfigurationRequest{
		Type: "vmware_v6", Name: "acc", Description: "d", Parameters: json.RawMessage(`{}`),
	}); err != nil {
		t.Fatalf("Create: %v", err)
	}
}

func TestAdditionalConnectorConfigurationService_UpdateNormalizesNilPollers(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("PUT /centreon/api/latest/configuration/additional-connector-configurations/5", func(w http.ResponseWriter, r *http.Request) {
		var raw map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got := string(raw["pollers"]); got != "[]" {
			t.Errorf("pollers = %s, want []", got)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.AdditionalConnectorConfigurations.Update(t.Context(), 5, &UpdateAdditionalConnectorConfigurationRequest{
		Type: "vmware_v6", Name: "acc", Description: "d", Parameters: json.RawMessage(`{}`),
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestAdditionalConnectorConfigurationService_Update(t *testing.T) {
	mux, c := newTestMux(t)
	var called bool
	mux.HandleFunc("PUT /centreon/api/latest/configuration/additional-connector-configurations/5", func(w http.ResponseWriter, r *http.Request) {
		called = true
		var req UpdateAdditionalConnectorConfigurationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		wantStr(t, "Name", req.Name, "acc-upd")
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.AdditionalConnectorConfigurations.Update(t.Context(), 5, &UpdateAdditionalConnectorConfigurationRequest{
		Type: "vmware_v6", Name: "acc-upd", Description: "d2", Pollers: []int{1},
		Parameters: json.RawMessage(`{"port":5701,"vcenters":[{"id":1,"name":"vc1","url":"https://vc1/sdk","username":"u","password":"p"}]}`),
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestAdditionalConnectorConfigurationService_Delete(t *testing.T) {
	mux, c := newTestMux(t)
	var called bool
	mux.HandleFunc("DELETE /centreon/api/latest/configuration/additional-connector-configurations/5", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.AdditionalConnectorConfigurations.Delete(t.Context(), 5); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}
