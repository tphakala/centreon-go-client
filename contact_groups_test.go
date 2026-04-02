package centreon

import (
	"net/http"
	"testing"
)

func TestContactGroupService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/users/contact-groups", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{"id": 1, "name": "admins", "alias": "Administrators", "type": "local", "is_activated": true},
				{"id": 2, "name": "operators", "alias": "Operators", "type": "ldap", "is_activated": true},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 2},
		})
	})

	resp, err := c.ContactGroups.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}
	if resp.Result[0].Name != "admins" {
		t.Errorf("Result[0].Name = %q, want %q", resp.Result[0].Name, "admins")
	}
	if resp.Result[0].Type != "local" {
		t.Errorf("Result[0].Type = %q, want %q", resp.Result[0].Type, "local")
	}
	if !resp.Result[0].IsActivated {
		t.Error("Result[0].IsActivated = false, want true")
	}
	if resp.Result[1].Name != "operators" {
		t.Errorf("Result[1].Name = %q, want %q", resp.Result[1].Name, "operators")
	}
	if resp.Result[1].Type != "ldap" {
		t.Errorf("Result[1].Type = %q, want %q", resp.Result[1].Type, "ldap")
	}
}
