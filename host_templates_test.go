package centreon

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

func TestHostTemplateService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/hosts/templates", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{"id": 1, "name": "tpl-linux", "alias": "Linux Template", "check_command_id": nil, "check_timeperiod_id": 1, "max_check_attempts": nil, "normal_check_interval": nil, "retry_check_interval": nil, "is_locked": false},
				{"id": 2, "name": "tpl-windows", "alias": "Windows Template", "check_command_id": 5, "check_timeperiod_id": nil, "max_check_attempts": 3, "normal_check_interval": 5, "retry_check_interval": 1, "is_locked": true},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 2},
		})
	})

	resp, err := c.HostTemplates.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}

	tpl0 := resp.Result[0]
	if tpl0.Name != "tpl-linux" {
		t.Errorf("Result[0].Name = %q, want %q", tpl0.Name, "tpl-linux")
	}
	if tpl0.CheckCommandID != nil {
		t.Errorf("Result[0].CheckCommandID = %v, want nil", tpl0.CheckCommandID)
	}
	if tpl0.CheckTimeperiodID == nil || *tpl0.CheckTimeperiodID != 1 {
		t.Errorf("Result[0].CheckTimeperiodID = %v, want 1", tpl0.CheckTimeperiodID)
	}
	if tpl0.IsLocked {
		t.Error("Result[0].IsLocked = true, want false")
	}

	tpl1 := resp.Result[1]
	if tpl1.CheckCommandID == nil || *tpl1.CheckCommandID != 5 {
		t.Errorf("Result[1].CheckCommandID = %v, want 5", tpl1.CheckCommandID)
	}
	if tpl1.MaxCheckAttempts == nil || *tpl1.MaxCheckAttempts != 3 {
		t.Errorf("Result[1].MaxCheckAttempts = %v, want 3", tpl1.MaxCheckAttempts)
	}
	if !tpl1.IsLocked {
		t.Error("Result[1].IsLocked = false, want true")
	}
}

func TestHostTemplateService_GetByID_Found(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/hosts/templates", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{"id": 10, "name": "tpl-snmp", "alias": "SNMP Template", "check_command_id": nil, "check_timeperiod_id": nil, "max_check_attempts": nil, "normal_check_interval": nil, "retry_check_interval": nil, "is_locked": false},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 1},
		})
	})

	tpl, err := c.HostTemplates.GetByID(t.Context(), 10)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if tpl.ID != 10 {
		t.Errorf("ID = %d, want 10", tpl.ID)
	}
	if tpl.Name != "tpl-snmp" {
		t.Errorf("Name = %q, want %q", tpl.Name, "tpl-snmp")
	}
}

func TestHostTemplateService_GetByID_NotFound(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/hosts/templates", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[HostTemplate]{
			Result: []HostTemplate{},
			Meta:   Meta{Page: 1, Limit: 10, Total: 0},
		})
	})

	_, err := c.HostTemplates.GetByID(t.Context(), 999)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	nfe, ok := errors.AsType[*NotFoundError](err)
	if !ok {
		t.Fatalf("expected *NotFoundError, got %T: %v", err, err)
	}
	if nfe.ID != 999 {
		t.Errorf("NotFoundError.ID = %d, want 999", nfe.ID)
	}
}

func TestHostTemplateService_Create(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("POST /centreon/api/latest/configuration/hosts/templates", func(w http.ResponseWriter, r *http.Request) {
		var req CreateHostTemplateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if req.Name != "tpl-new" {
			t.Errorf("Name = %q, want %q", req.Name, "tpl-new")
		}
		writeJSON(w, http.StatusCreated, map[string]int{"id": 20})
	})

	id, err := c.HostTemplates.Create(t.Context(), CreateHostTemplateRequest{
		Name:  "tpl-new",
		Alias: "New Template",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 20 {
		t.Errorf("id = %d, want 20", id)
	}
}

func TestHostTemplateService_Update(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("PATCH /centreon/api/latest/configuration/hosts/templates/10", func(w http.ResponseWriter, r *http.Request) {
		var req UpdateHostTemplateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if req.Name == nil || *req.Name != "tpl-updated" {
			t.Errorf("Name = %v, want %q", req.Name, "tpl-updated")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	name := "tpl-updated"
	err := c.HostTemplates.Update(t.Context(), 10, UpdateHostTemplateRequest{Name: &name})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestHostTemplateService_Delete(t *testing.T) {
	mux, c := newTestMux(t)

	var called bool
	mux.HandleFunc("DELETE /centreon/api/latest/configuration/hosts/templates/10", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.HostTemplates.Delete(t.Context(), 10)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}
