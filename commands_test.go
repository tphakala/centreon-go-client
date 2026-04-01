package centreon

import (
	"net/http"
	"testing"
)

func TestCommandService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/commands", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[Command]{
			Result: []Command{
				{ID: 1, Name: "check_ping", Type: 2, CommandLine: "/usr/lib/nagios/plugins/check_ping -H $HOSTADDRESS$"},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 1},
		})
	})

	resp, err := c.Commands.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 1 {
		t.Fatalf("len(Result) = %d, want 1", len(resp.Result))
	}
	cmd := resp.Result[0]
	if cmd.Name != "check_ping" {
		t.Errorf("Name = %q, want %q", cmd.Name, "check_ping")
	}
	if cmd.Type != 2 {
		t.Errorf("Type = %d, want 2", cmd.Type)
	}
	if cmd.CommandLine != "/usr/lib/nagios/plugins/check_ping -H $HOSTADDRESS$" {
		t.Errorf("CommandLine = %q, unexpected value", cmd.CommandLine)
	}
}
