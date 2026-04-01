package centreon

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestUserService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/users", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[User]{
			Result: []User{
				{ID: 1, Name: "admin", Alias: "Administrator", Email: "admin@example.com", IsAdmin: true, IsActivated: true},
				{ID: 2, Name: "user1", Alias: "User One", Email: "user1@example.com", IsAdmin: false, IsActivated: true},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 2},
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
	if resp.Result[1].Name != "user1" {
		t.Errorf("Result[1].Name = %q, want %q", resp.Result[1].Name, "user1")
	}
}

func TestUserService_Update(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("PATCH /centreon/api/latest/users/5", func(w http.ResponseWriter, r *http.Request) {
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
