package centreon

import (
	"net/http"
	"testing"
)

func TestContactTemplateService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/contacts/templates", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[ContactTemplate]{
			Result: []ContactTemplate{
				{ID: 1, Name: "generic-contact", Alias: "Generic Contact", IsActivated: true},
				{ID: 2, Name: "admin-contact", Alias: "Admin Contact", IsActivated: true},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 2},
		})
	})

	resp, err := c.ContactTemplates.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}
	if resp.Result[0].Name != "generic-contact" {
		t.Errorf("Result[0].Name = %q, want %q", resp.Result[0].Name, "generic-contact")
	}
	if resp.Result[1].Name != "admin-contact" {
		t.Errorf("Result[1].Name = %q, want %q", resp.Result[1].Name, "admin-contact")
	}
}
