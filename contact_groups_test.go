package centreon

import (
	"net/http"
	"testing"
)

func TestContactGroupService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/users/contact-groups", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[ContactGroup]{
			Result: []ContactGroup{
				{ID: 1, Name: "admins", Alias: "Administrators", IsActivated: true},
				{ID: 2, Name: "operators", Alias: "Operators", IsActivated: true},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 2},
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
	if resp.Result[1].Name != "operators" {
		t.Errorf("Result[1].Name = %q, want %q", resp.Result[1].Name, "operators")
	}
}
