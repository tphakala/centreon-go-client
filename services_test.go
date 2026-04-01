package centreon

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestServiceService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/services", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[Service]{
			Result: []Service{
				{ID: 1, HostID: 10, Name: "Ping", IsActivated: true},
				{ID: 2, HostID: 10, Name: "CPU", IsActivated: true},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 2},
		})
	})

	resp, err := c.Services.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}
	if resp.Result[0].Name != "Ping" {
		t.Errorf("Result[0].Name = %q, want %q", resp.Result[0].Name, "Ping")
	}
}

func TestServiceService_GetByID_Found(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/services", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[Service]{
			Result: []Service{
				{ID: 42, HostID: 10, Name: "Ping", IsActivated: true},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 1},
		})
	})

	svc, err := c.Services.GetByID(t.Context(), 42)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if svc.ID != 42 {
		t.Errorf("ID = %d, want 42", svc.ID)
	}
	if svc.Name != "Ping" {
		t.Errorf("Name = %q, want %q", svc.Name, "Ping")
	}
}

func TestServiceService_GetByID_NotFound(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/services", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[Service]{
			Result: []Service{},
			Meta:   Meta{Page: 1, Limit: 10, Total: 0},
		})
	})

	_, err := c.Services.GetByID(t.Context(), 999)
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

func TestServiceService_Create(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("POST /centreon/api/latest/configuration/services", func(w http.ResponseWriter, r *http.Request) {
		var req CreateServiceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if req.Name != "new-service" {
			t.Errorf("Name = %q, want %q", req.Name, "new-service")
		}
		if req.HostID != 10 {
			t.Errorf("HostID = %d, want 10", req.HostID)
		}
		writeJSON(w, http.StatusCreated, map[string]int{"id": 99})
	})

	id, err := c.Services.Create(t.Context(), CreateServiceRequest{
		HostID: 10,
		Name:   "new-service",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 99 {
		t.Errorf("id = %d, want 99", id)
	}
}

func TestServiceService_Update(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("PATCH /centreon/api/latest/configuration/services/42", func(w http.ResponseWriter, r *http.Request) {
		var req UpdateServiceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if req.Name == nil || *req.Name != "updated-service" {
			t.Errorf("Name = %v, want %q", req.Name, "updated-service")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	name := "updated-service"
	err := c.Services.Update(t.Context(), 42, UpdateServiceRequest{Name: &name})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestServiceService_Delete(t *testing.T) {
	mux, c := newTestMux(t)

	var called bool
	mux.HandleFunc("DELETE /centreon/api/latest/configuration/services/42", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.Services.Delete(t.Context(), 42)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}
