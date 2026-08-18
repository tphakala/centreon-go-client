package centreon

import (
	"net/http"
	"testing"
)

func TestConnectorService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/connectors", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{
					"id":           1,
					"name":         "Perl Connector",
					"command_line": "centreon_connector_perl --log-file=/var/log/centreon-engine/connector-perl.log",
					"description":  nil,
					"commands":     []any{},
					"is_activated": true,
				},
				{
					"id":           4,
					"name":         "Telegraf",
					"command_line": "opentelemetry --processor=nagios_telegraf",
					"description":  "Telegraf",
					"commands": []map[string]any{
						{"id": 93, "type": 2, "name": "check_something"},
					},
					"is_activated": true,
				},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 2},
		})
	})

	resp, err := c.Connectors.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}

	// Pin the plain scalar tags so a retag of id/name/is_activated cannot hide.
	wantInt(t, "Result[0].ID", resp.Result[0].ID, 1)
	wantStr(t, "Result[0].Name", resp.Result[0].Name, "Perl Connector")
	wantBool(t, "Result[0].IsActivated", resp.Result[0].IsActivated, true)
	wantInt(t, "Result[1].ID", resp.Result[1].ID, 4)
	wantStr(t, "Result[1].Name", resp.Result[1].Name, "Telegraf")

	// A null description decodes to a nil pointer, distinct from an empty string.
	if resp.Result[0].Description != nil {
		t.Errorf("Result[0].Description = %v, want nil", *resp.Result[0].Description)
	}
	if resp.Result[0].CommandLine == "" {
		t.Error("Result[0].CommandLine = empty, want the Perl connector command line")
	}
	if len(resp.Result[0].Commands) != 0 {
		t.Errorf("len(Result[0].Commands) = %d, want 0", len(resp.Result[0].Commands))
	}

	// A present description decodes to a non-nil pointer.
	if resp.Result[1].Description == nil || *resp.Result[1].Description != "Telegraf" {
		t.Errorf("Result[1].Description = %v, want %q", resp.Result[1].Description, "Telegraf")
	}
	if len(resp.Result[1].Commands) != 1 {
		t.Fatalf("len(Result[1].Commands) = %d, want 1", len(resp.Result[1].Commands))
	}
	if got := resp.Result[1].Commands[0]; got.ID != 93 || got.Type != 2 || got.Name != "check_something" {
		t.Errorf("Result[1].Commands[0] = %+v, want {ID:93 Type:2 Name:check_something}", got)
	}
}
