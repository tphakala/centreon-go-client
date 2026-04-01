package centreon

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestHostTemplateService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/hosts/templates", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[HostTemplate]{
			Result: []HostTemplate{
				{ID: 1, Name: "tpl-linux", Alias: "Linux Template", Address: "0.0.0.0", IsActivated: true},
				{ID: 2, Name: "tpl-windows", Alias: "Windows Template", Address: "0.0.0.0", IsActivated: true},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 2},
		})
	})

	resp, err := c.HostTemplates.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}
	if resp.Result[0].Name != "tpl-linux" {
		t.Errorf("Result[0].Name = %q, want %q", resp.Result[0].Name, "tpl-linux")
	}
}

func TestHostTemplateService_GetByID_Found(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/hosts/templates", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[HostTemplate]{
			Result: []HostTemplate{
				{ID: 10, Name: "tpl-snmp", Alias: "SNMP Template", Address: "0.0.0.0", IsActivated: true},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 1},
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
	nfe, ok := err.(*NotFoundError)
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
		if req.Address != "0.0.0.0" {
			t.Errorf("Address = %q, want %q", req.Address, "0.0.0.0")
		}
		writeJSON(w, http.StatusCreated, map[string]int{"id": 20})
	})

	id, err := c.HostTemplates.Create(t.Context(), CreateHostTemplateRequest{
		Name:    "tpl-new",
		Alias:   "New Template",
		Address: "0.0.0.0",
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
