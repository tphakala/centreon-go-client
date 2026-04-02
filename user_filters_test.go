package centreon

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestUserFilterService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/users/filters", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[UserFilter]{
			Result: []UserFilter{
				{ID: 1, Name: "my-filter"},
				{ID: 2, Name: "another-filter"},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 2},
		})
	})

	resp, err := c.UserFilters.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}
	if resp.Result[0].Name != "my-filter" {
		t.Errorf("Result[0].Name = %q, want %q", resp.Result[0].Name, "my-filter")
	}
}

func TestUserFilterService_Get(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/users/filters/3", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, UserFilter{
			ID:   3,
			Name: "test-filter",
			Criteria: []FilterCriteria{
				{Name: "status", Type: "string", Value: "OK"},
			},
		})
	})

	uf, err := c.UserFilters.Get(t.Context(), 3)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if uf.ID != 3 {
		t.Errorf("ID = %d, want 3", uf.ID)
	}
	if uf.Name != "test-filter" {
		t.Errorf("Name = %q, want %q", uf.Name, "test-filter")
	}
	if len(uf.Criteria) != 1 {
		t.Fatalf("len(Criteria) = %d, want 1", len(uf.Criteria))
	}
	if uf.Criteria[0].Name != "status" {
		t.Errorf("Criteria[0].Name = %q, want %q", uf.Criteria[0].Name, "status")
	}
}

func TestUserFilterService_Create(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("POST /centreon/api/latest/configuration/users/filters", func(w http.ResponseWriter, r *http.Request) {
		var req CreateUserFilterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if req.Name != "new-filter" {
			t.Errorf("Name = %q, want %q", req.Name, "new-filter")
		}
		writeJSON(w, http.StatusCreated, map[string]int{"id": 7})
	})

	id, err := c.UserFilters.Create(t.Context(), CreateUserFilterRequest{
		Name: "new-filter",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 7 {
		t.Errorf("id = %d, want 7", id)
	}
}

func TestUserFilterService_Update(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("PUT /centreon/api/latest/configuration/users/filters/3", func(w http.ResponseWriter, r *http.Request) {
		var req UpdateUserFilterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if req.Name != "updated-filter" {
			t.Errorf("Name = %q, want %q", req.Name, "updated-filter")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.UserFilters.Update(t.Context(), 3, UpdateUserFilterRequest{Name: "updated-filter"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestUserFilterService_Patch(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("PATCH /centreon/api/latest/configuration/users/filters/3", func(w http.ResponseWriter, r *http.Request) {
		var req PatchUserFilterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if req.Name == nil || *req.Name != "patched-filter" {
			t.Errorf("Name = %v, want %q", req.Name, "patched-filter")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	name := "patched-filter"
	err := c.UserFilters.Patch(t.Context(), 3, PatchUserFilterRequest{Name: &name})
	if err != nil {
		t.Fatalf("Patch: %v", err)
	}
}

func TestUserFilterService_Delete(t *testing.T) {
	mux, c := newTestMux(t)

	var called bool
	mux.HandleFunc("DELETE /centreon/api/latest/configuration/users/filters/3", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.UserFilters.Delete(t.Context(), 3)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}
