package centreon

import (
	"net/http"
	"testing"
)

func TestIconService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/icons", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{
					"id":        1,
					"directory": "operatingsystems",
					"name":      "linux",
					"url":       "/centreon/img/media/operatingsystems/linux.png",
				},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 1},
		})
	})

	resp, err := c.Icons.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 1 {
		t.Fatalf("len(Result) = %d, want 1", len(resp.Result))
	}
	got := resp.Result[0]
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
	if got.Directory != "operatingsystems" {
		t.Errorf("Directory = %q, want %q", got.Directory, "operatingsystems")
	}
	if got.Name != "linux" {
		t.Errorf("Name = %q, want %q", got.Name, "linux")
	}
	if got.URL != "/centreon/img/media/operatingsystems/linux.png" {
		t.Errorf("URL = %q, want the media path", got.URL)
	}
}
