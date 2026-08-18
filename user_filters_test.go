package centreon

import (
	"net/http"
	"testing"
)

func TestUserFilterService_List(t *testing.T) {
	mux, c := newTestMux(t)

	// Map-based body pins the wire keys independently of the struct: the server
	// sends the plural "criterias" and an "order" field, and both must decode.
	mux.HandleFunc("GET /centreon/api/latest/users/filters/events-view", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{"id": 1, "name": "my-filter", "order": 0, "criterias": []any{}},
				{"id": 2, "name": "another-filter", "order": 1, "criterias": []any{}},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 2},
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
	if resp.Result[1].Order != 1 {
		t.Errorf("Result[1].Order = %d, want 1", resp.Result[1].Order)
	}
}

func TestUserFilterService_Get(t *testing.T) {
	mux, c := newTestMux(t)

	// The live server sends the plural "criterias" key and an "order" field.
	// Using a map body (not a UserFilter literal) pins those wire keys so a tag
	// regression back to the singular "criteria" would fail the decode below.
	mux.HandleFunc("GET /centreon/api/latest/users/filters/events-view/3", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"id":    3,
			"name":  "test-filter",
			"order": 2,
			"criterias": []map[string]any{
				{"name": "status", "type": "multi_select", "value": "OK", "object_type": ""},
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
	if uf.Order != 2 {
		t.Errorf("Order = %d, want 2", uf.Order)
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

	mux.HandleFunc("POST /centreon/api/latest/users/filters/events-view", func(w http.ResponseWriter, r *http.Request) {
		body := decodeBody(t, r)
		if body["name"] != "new-filter" {
			t.Errorf("name = %q, want %q", body["name"], "new-filter")
		}
		// The server requires the plural "criterias" array (additionalProperties:
		// false) and never the singular "criteria"; a nil slice is normalized to
		// an empty array so the required field is always present.
		if _, ok := body["criterias"].([]any); !ok {
			t.Errorf("criterias must be an array, got %T (%v)", body["criterias"], body["criterias"])
		}
		if _, exists := body["criteria"]; exists {
			t.Error(`legacy singular key "criteria" must not be sent`)
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

	mux.HandleFunc("PUT /centreon/api/latest/users/filters/events-view/3", func(w http.ResponseWriter, r *http.Request) {
		body := decodeBody(t, r)
		if body["name"] != "updated-filter" {
			t.Errorf("name = %q, want %q", body["name"], "updated-filter")
		}
		if _, ok := body["criterias"].([]any); !ok {
			t.Errorf("criterias must be an array, got %T (%v)", body["criterias"], body["criterias"])
		}
		if _, exists := body["criteria"]; exists {
			t.Error(`legacy singular key "criteria" must not be sent`)
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

	// PATCH on this route reorders: it requires "order" and rejects "name".
	mux.HandleFunc("PATCH /centreon/api/latest/users/filters/events-view/3", func(w http.ResponseWriter, r *http.Request) {
		body := decodeBody(t, r)
		order, ok := body["order"].(float64)
		if !ok {
			t.Fatalf("order must be a JSON number, got %T (%v)", body["order"], body["order"])
		}
		if order != 2 {
			t.Errorf("order = %v, want 2", order)
		}
		if _, exists := body["name"]; exists {
			t.Error(`"name" must not be sent on the reorder PATCH route`)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.UserFilters.Patch(t.Context(), 3, PatchUserFilterRequest{Order: 2})
	if err != nil {
		t.Fatalf("Patch: %v", err)
	}
}

func TestUserFilterService_Delete(t *testing.T) {
	mux, c := newTestMux(t)

	var called bool
	mux.HandleFunc("DELETE /centreon/api/latest/users/filters/events-view/3", func(w http.ResponseWriter, r *http.Request) {
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
