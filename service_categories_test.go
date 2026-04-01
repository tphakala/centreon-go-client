package centreon

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestServiceCategoryService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/services/categories", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[ServiceCategory]{
			Result: []ServiceCategory{
				{ID: 1, Name: "Web", IsActivated: true},
				{ID: 2, Name: "Database", IsActivated: true},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 2},
		})
	})

	resp, err := c.ServiceCategories.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}
	if resp.Result[0].Name != "Web" {
		t.Errorf("Result[0].Name = %q, want %q", resp.Result[0].Name, "Web")
	}
}

func TestServiceCategoryService_Create(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("POST /centreon/api/latest/configuration/services/categories", func(w http.ResponseWriter, r *http.Request) {
		var req CreateServiceCategoryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if req.Name != "new-category" {
			t.Errorf("Name = %q, want %q", req.Name, "new-category")
		}
		writeJSON(w, http.StatusCreated, map[string]int{"id": 20})
	})

	id, err := c.ServiceCategories.Create(t.Context(), CreateServiceCategoryRequest{Name: "new-category"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 20 {
		t.Errorf("id = %d, want 20", id)
	}
}

func TestServiceCategoryService_Delete(t *testing.T) {
	mux, c := newTestMux(t)

	var called bool
	mux.HandleFunc("DELETE /centreon/api/latest/configuration/services/categories/20", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.ServiceCategories.Delete(t.Context(), 20)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}
