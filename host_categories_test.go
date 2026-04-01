package centreon

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestHostCategoryService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/hosts/categories", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[HostCategory]{
			Result: []HostCategory{
				{ID: 1, Name: "database", Alias: "Database Servers", IsActivated: true},
				{ID: 2, Name: "web", Alias: "Web Servers", IsActivated: true},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 2},
		})
	})

	resp, err := c.HostCategories.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}
	if resp.Result[0].Name != "database" {
		t.Errorf("Result[0].Name = %q, want %q", resp.Result[0].Name, "database")
	}
}

func TestHostCategoryService_Get(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/hosts/categories/3", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, HostCategory{
			ID: 3, Name: "storage", Alias: "Storage Servers", IsActivated: true,
		})
	})

	cat, err := c.HostCategories.Get(t.Context(), 3)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if cat.ID != 3 {
		t.Errorf("ID = %d, want 3", cat.ID)
	}
	if cat.Name != "storage" {
		t.Errorf("Name = %q, want %q", cat.Name, "storage")
	}
}

func TestHostCategoryService_Create(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("POST /centreon/api/latest/configuration/hosts/categories", func(w http.ResponseWriter, r *http.Request) {
		var req CreateHostCategoryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if req.Name != "new-category" {
			t.Errorf("Name = %q, want %q", req.Name, "new-category")
		}
		writeJSON(w, http.StatusCreated, map[string]int{"id": 7})
	})

	id, err := c.HostCategories.Create(t.Context(), CreateHostCategoryRequest{
		Name:  "new-category",
		Alias: "New Category",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 7 {
		t.Errorf("id = %d, want 7", id)
	}
}

func TestHostCategoryService_Update(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("PUT /centreon/api/latest/configuration/hosts/categories/3", func(w http.ResponseWriter, r *http.Request) {
		var req UpdateHostCategoryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if req.Name != "updated-category" {
			t.Errorf("Name = %q, want %q", req.Name, "updated-category")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.HostCategories.Update(t.Context(), 3, UpdateHostCategoryRequest{Name: "updated-category", Alias: "Updated Category"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestHostCategoryService_Delete(t *testing.T) {
	mux, c := newTestMux(t)

	var called bool
	mux.HandleFunc("DELETE /centreon/api/latest/configuration/hosts/categories/3", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.HostCategories.Delete(t.Context(), 3)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}
