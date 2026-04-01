package centreon

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestServiceSeverityService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/services/severities", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[ServiceSeverity]{
			Result: []ServiceSeverity{
				{ID: 1, Name: "Critical", Level: 1, IconID: 5, IsActivated: true},
				{ID: 2, Name: "Warning", Level: 2, IconID: 6, IsActivated: true},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 2},
		})
	})

	resp, err := c.ServiceSeverities.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}
	if resp.Result[0].Name != "Critical" {
		t.Errorf("Result[0].Name = %q, want %q", resp.Result[0].Name, "Critical")
	}
}

func TestServiceSeverityService_Create(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("POST /centreon/api/latest/configuration/services/severities", func(w http.ResponseWriter, r *http.Request) {
		var req CreateServiceSeverityRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if req.Name != "new-severity" {
			t.Errorf("Name = %q, want %q", req.Name, "new-severity")
		}
		if req.Level != 3 {
			t.Errorf("Level = %d, want 3", req.Level)
		}
		if req.IconID != 7 {
			t.Errorf("IconID = %d, want 7", req.IconID)
		}
		writeJSON(w, http.StatusCreated, map[string]int{"id": 30})
	})

	id, err := c.ServiceSeverities.Create(t.Context(), CreateServiceSeverityRequest{
		Name:   "new-severity",
		Level:  3,
		IconID: 7,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 30 {
		t.Errorf("id = %d, want 30", id)
	}
}

func TestServiceSeverityService_Update(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("PUT /centreon/api/latest/configuration/services/severities/30", func(w http.ResponseWriter, r *http.Request) {
		var req UpdateServiceSeverityRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if req.Name != "updated-severity" {
			t.Errorf("Name = %q, want %q", req.Name, "updated-severity")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.ServiceSeverities.Update(t.Context(), 30, UpdateServiceSeverityRequest{
		Name:   "updated-severity",
		Level:  3,
		IconID: 7,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestServiceSeverityService_Delete(t *testing.T) {
	mux, c := newTestMux(t)

	var called bool
	mux.HandleFunc("DELETE /centreon/api/latest/configuration/services/severities/30", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.ServiceSeverities.Delete(t.Context(), 30)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}
