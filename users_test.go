package centreon

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestUserService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/users", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{"id": 1, "name": "admin", "alias": "Administrator", "email": "admin@example.com", "is_admin": true},
				{"id": 2, "name": "user1", "alias": "User One", "email": "user1@example.com", "is_admin": false},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 2},
		})
	})

	resp, err := c.Users.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}
	if resp.Result[0].Name != "admin" {
		t.Errorf("Result[0].Name = %q, want %q", resp.Result[0].Name, "admin")
	}
	if !resp.Result[0].IsAdmin {
		t.Error("Result[0].IsAdmin = false, want true")
	}
	if resp.Result[1].Name != "user1" {
		t.Errorf("Result[1].Name = %q, want %q", resp.Result[1].Name, "user1")
	}
	if resp.Result[1].IsAdmin {
		t.Error("Result[1].IsAdmin = true, want false")
	}
}

func TestUserService_Update(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("PATCH /centreon/api/latest/configuration/users/5", func(w http.ResponseWriter, r *http.Request) {
		var req UpdateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if req.Name == nil || *req.Name != "updated-user" {
			t.Errorf("Name = %v, want %q", req.Name, "updated-user")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	name := "updated-user"
	err := c.Users.Update(t.Context(), 5, UpdateUserRequest{Name: &name})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
}
