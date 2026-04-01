package centreon

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestServiceTemplateService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/services/templates", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[ServiceTemplate]{
			Result: []ServiceTemplate{
				{ID: 1, Name: "generic-service", IsActivated: true},
				{ID: 2, Name: "ping-template", IsActivated: true},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 2},
		})
	})

	resp, err := c.ServiceTemplates.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}
	if resp.Result[0].Name != "generic-service" {
		t.Errorf("Result[0].Name = %q, want %q", resp.Result[0].Name, "generic-service")
	}
}

func TestServiceTemplateService_GetByID_Found(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/services/templates", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[ServiceTemplate]{
			Result: []ServiceTemplate{
				{ID: 42, Name: "generic-service", IsActivated: true},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 1},
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
	nfe, ok := err.(*NotFoundError)
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
