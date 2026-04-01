package centreon

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestHostSeverityService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/hosts/severities", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[HostSeverity]{
			Result: []HostSeverity{
				{ID: 1, Name: "critical", Alias: "Critical", Level: 1, IconID: 10, IsActivated: true},
				{ID: 2, Name: "warning", Alias: "Warning", Level: 2, IconID: 11, IsActivated: true},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 2},
		})
	})

	resp, err := c.HostSeverities.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}
	if resp.Result[0].Name != "critical" {
		t.Errorf("Result[0].Name = %q, want %q", resp.Result[0].Name, "critical")
	}
	if resp.Result[0].Level != 1 {
		t.Errorf("Result[0].Level = %d, want 1", resp.Result[0].Level)
	}
}

func TestHostSeverityService_Get(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/hosts/severities/4", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, HostSeverity{
			ID: 4, Name: "info", Alias: "Informational", Level: 4, IconID: 14, IsActivated: true,
		})
	})

	sev, err := c.HostSeverities.Get(t.Context(), 4)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if sev.ID != 4 {
		t.Errorf("ID = %d, want 4", sev.ID)
	}
	if sev.Level != 4 {
		t.Errorf("Level = %d, want 4", sev.Level)
	}
}

func TestHostSeverityService_Create(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("POST /centreon/api/latest/configuration/hosts/severities", func(w http.ResponseWriter, r *http.Request) {
		var req CreateHostSeverityRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if req.Name != "new-severity" {
			t.Errorf("Name = %q, want %q", req.Name, "new-severity")
		}
		if req.Level != 3 {
			t.Errorf("Level = %d, want 3", req.Level)
		}
		if req.IconID != 20 {
			t.Errorf("IconID = %d, want 20", req.IconID)
		}
		writeJSON(w, http.StatusCreated, map[string]int{"id": 15})
	})

	id, err := c.HostSeverities.Create(t.Context(), CreateHostSeverityRequest{
		Name:   "new-severity",
		Alias:  "New Severity",
		Level:  3,
		IconID: 20,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 15 {
		t.Errorf("id = %d, want 15", id)
	}
}

func TestHostSeverityService_Update(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("PUT /centreon/api/latest/configuration/hosts/severities/4", func(w http.ResponseWriter, r *http.Request) {
		var req UpdateHostSeverityRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if req.Name != "updated-severity" {
			t.Errorf("Name = %q, want %q", req.Name, "updated-severity")
		}
		if req.Level != 5 {
			t.Errorf("Level = %d, want 5", req.Level)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.HostSeverities.Update(t.Context(), 4, UpdateHostSeverityRequest{
		Name:   "updated-severity",
		Alias:  "Updated Severity",
		Level:  5,
		IconID: 14,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestHostSeverityService_Delete(t *testing.T) {
	mux, c := newTestMux(t)

	var called bool
	mux.HandleFunc("DELETE /centreon/api/latest/configuration/hosts/severities/4", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.HostSeverities.Delete(t.Context(), 4)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}
