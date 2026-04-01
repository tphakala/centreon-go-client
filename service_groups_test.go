package centreon

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestServiceGroupService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/services/groups", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[ServiceGroup]{
			Result: []ServiceGroup{
				{ID: 1, Name: "Web Services", IsActivated: true},
				{ID: 2, Name: "DB Services", IsActivated: true},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 2},
		})
	})

	resp, err := c.ServiceGroups.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}
	if resp.Result[0].Name != "Web Services" {
		t.Errorf("Result[0].Name = %q, want %q", resp.Result[0].Name, "Web Services")
	}
}

func TestServiceGroupService_Create(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("POST /centreon/api/latest/configuration/services/groups", func(w http.ResponseWriter, r *http.Request) {
		var req CreateServiceGroupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if req.Name != "new-group" {
			t.Errorf("Name = %q, want %q", req.Name, "new-group")
		}
		writeJSON(w, http.StatusCreated, map[string]int{"id": 10})
	})

	id, err := c.ServiceGroups.Create(t.Context(), CreateServiceGroupRequest{Name: "new-group"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 10 {
		t.Errorf("id = %d, want 10", id)
	}
}

func TestServiceGroupService_Delete(t *testing.T) {
	mux, c := newTestMux(t)

	var called bool
	mux.HandleFunc("DELETE /centreon/api/latest/configuration/services/groups/10", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.ServiceGroups.Delete(t.Context(), 10)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}
