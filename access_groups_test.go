package centreon

import (
	"net/http"
	"testing"
)

func TestAccessGroupService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/access-groups", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{"id": 1, "name": "ALL", "alias": "ALL", "has_changed": false, "is_activated": true},
				{"id": 2, "name": "ops", "alias": "Operations", "has_changed": true, "is_activated": false},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 2},
		})
	})

	resp, err := c.AccessGroups.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}
	if resp.Result[0].Name != "ALL" || resp.Result[0].Alias != "ALL" {
		t.Errorf("Result[0] name/alias = %q/%q, want ALL/ALL", resp.Result[0].Name, resp.Result[0].Alias)
	}
	if resp.Result[0].HasChanged {
		t.Error("Result[0].HasChanged = true, want false")
	}
	if !resp.Result[0].IsActivated {
		t.Error("Result[0].IsActivated = false, want true")
	}
	if !resp.Result[1].HasChanged {
		t.Error("Result[1].HasChanged = false, want true")
	}
	if resp.Result[1].IsActivated {
		t.Error("Result[1].IsActivated = true, want false")
	}
}
