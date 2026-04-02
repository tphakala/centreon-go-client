package centreon

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

func TestServiceTemplateService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/services/templates", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{"id": 1, "name": "generic-service", "alias": "Generic", "check_command_id": nil, "check_timeperiod_id": 1, "max_check_attempts": nil, "normal_check_interval": nil, "retry_check_interval": nil, "is_locked": false},
				{"id": 2, "name": "ping-template", "alias": "Ping", "check_command_id": 3, "check_timeperiod_id": nil, "max_check_attempts": 5, "normal_check_interval": 10, "retry_check_interval": 2, "is_locked": true},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 2},
		})
	})

	resp, err := c.ServiceTemplates.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}

	tpl0 := resp.Result[0]
	if tpl0.Name != "generic-service" {
		t.Errorf("Result[0].Name = %q, want %q", tpl0.Name, "generic-service")
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
	if tpl1.CheckCommandID == nil || *tpl1.CheckCommandID != 3 {
		t.Errorf("Result[1].CheckCommandID = %v, want 3", tpl1.CheckCommandID)
	}
	if tpl1.MaxCheckAttempts == nil || *tpl1.MaxCheckAttempts != 5 {
		t.Errorf("Result[1].MaxCheckAttempts = %v, want 5", tpl1.MaxCheckAttempts)
	}
	if !tpl1.IsLocked {
		t.Error("Result[1].IsLocked = false, want true")
	}
}

func TestServiceTemplateService_GetByID_Found(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/services/templates", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{"id": 42, "name": "generic-service", "alias": "Generic", "check_command_id": nil, "check_timeperiod_id": nil, "max_check_attempts": nil, "normal_check_interval": nil, "retry_check_interval": nil, "is_locked": false},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 1},
		})
	})

	tmpl, err := c.ServiceTemplates.GetByID(t.Context(), 42)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if tmpl.ID != 42 {
		t.Errorf("ID = %d, want 42", tmpl.ID)
	}
	if tmpl.Name != "generic-service" {
		t.Errorf("Name = %q, want %q", tmpl.Name, "generic-service")
	}
}

func TestServiceTemplateService_GetByID_NotFound(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/services/templates", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[ServiceTemplate]{
			Result: []ServiceTemplate{},
			Meta:   Meta{Page: 1, Limit: 10, Total: 0},
		})
	})

	_, err := c.ServiceTemplates.GetByID(t.Context(), 999)
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

func TestServiceTemplateService_Create(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("POST /centreon/api/latest/configuration/services/templates", func(w http.ResponseWriter, r *http.Request) {
		var req CreateServiceTemplateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if req.Name != "new-template" {
			t.Errorf("Name = %q, want %q", req.Name, "new-template")
		}
		writeJSON(w, http.StatusCreated, map[string]int{"id": 50})
	})

	id, err := c.ServiceTemplates.Create(t.Context(), CreateServiceTemplateRequest{
		Name:  "new-template",
		Alias: "New Template",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 50 {
		t.Errorf("id = %d, want 50", id)
	}
}

func TestServiceTemplateService_Update(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("PATCH /centreon/api/latest/configuration/services/templates/42", func(w http.ResponseWriter, r *http.Request) {
		var req UpdateServiceTemplateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if req.Name == nil || *req.Name != "updated-template" {
			t.Errorf("Name = %v, want %q", req.Name, "updated-template")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	name := "updated-template"
	err := c.ServiceTemplates.Update(t.Context(), 42, UpdateServiceTemplateRequest{Name: &name})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestServiceTemplateService_Delete(t *testing.T) {
	mux, c := newTestMux(t)

	var called bool
	mux.HandleFunc("DELETE /centreon/api/latest/configuration/services/templates/42", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.ServiceTemplates.Delete(t.Context(), 42)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}
