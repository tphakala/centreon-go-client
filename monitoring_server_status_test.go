package centreon

import (
	"encoding/json"
	"net/http"
	"testing"
)

// TestMonitoringServerStatusService_List decodes the exact wire shape captured
// live from GET /monitoring/servers on Centreon Web 25.10.16, field by field.
func TestMonitoringServerStatusService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/monitoring/servers", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{
					"id":         1,
					"name":       "Central",
					"address":    "127.0.0.1",
					"is_running": true,
					"last_alive": 1787562347,
					"version":    "25.10.11",
				},
				// A down poller. The endpoint sends 0 / false for a stopped
				// poller (it coerces its internal null heartbeat and running
				// flag to 0 / false server-side rather than emitting JSON null,
				// verified live on 25.10.16), so plain value fields decode the
				// wire shape with no pointer. Row 0 carries the tag/type
				// protection; this row pins that the stopped-poller wire shape
				// decodes cleanly.
				{
					"id":         2,
					"name":       "Poller",
					"address":    "10.0.0.9",
					"is_running": false,
					"last_alive": 0,
					"version":    "25.10.11",
				},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 2},
		})
	})

	resp, err := c.MonitoringServerStatus.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}

	want := []MonitoringServerStatus{
		{ID: 1, Name: "Central", Address: "127.0.0.1", IsRunning: true, LastAlive: 1787562347, Version: "25.10.11"},
		{ID: 2, Name: "Poller", Address: "10.0.0.9", IsRunning: false, LastAlive: 0, Version: "25.10.11"},
	}
	for i := range want {
		if resp.Result[i] != want[i] {
			t.Errorf("Result[%d] mismatch:\n got  %+v\n want %+v", i, resp.Result[i], want[i])
		}
	}
}

// TestMonitoringServerStatus_Decode pins the last_alive json tag and its int64
// width. The fixture uses a post-2038 Unix timestamp (4102444800, 2100-01-01)
// that exceeds math.MaxInt32, so the assertion fails two ways on a regression:
// renaming the json:"last_alive" tag leaves LastAlive at 0, and narrowing the
// field below int64 makes encoding/json reject the out-of-range number. The
// real captured value (1787562347) is exercised by
// TestMonitoringServerStatusService_List.
func TestMonitoringServerStatus_Decode(t *testing.T) {
	const wantAlive = int64(4102444800) // 2100-01-01T00:00:00Z, > math.MaxInt32
	const body = `{"id":1,"name":"Central","address":"127.0.0.1","is_running":true,"last_alive":4102444800,"version":"25.10.11"}`

	var got MonitoringServerStatus
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.LastAlive != wantAlive {
		t.Errorf("LastAlive = %d, want %d", got.LastAlive, wantAlive)
	}
	if !got.IsRunning {
		t.Errorf("IsRunning = false, want true")
	}
	if got.Version != "25.10.11" {
		t.Errorf("Version = %q, want %q", got.Version, "25.10.11")
	}
}

// TestMonitoringServerStatusService_All verifies the All iterator walks every
// page of /monitoring/servers.
func TestMonitoringServerStatusService_All(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/monitoring/servers", func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		switch page {
		case "", "1":
			writeJSON(w, http.StatusOK, map[string]any{
				"result": []map[string]any{
					{"id": 1, "name": "Central", "is_running": true, "last_alive": 1787562347},
				},
				"meta": map[string]any{"page": 1, "limit": 1, "total": 2},
			})
		case "2":
			writeJSON(w, http.StatusOK, map[string]any{
				"result": []map[string]any{
					{"id": 2, "name": "Poller", "is_running": false, "last_alive": 0},
				},
				"meta": map[string]any{"page": 2, "limit": 1, "total": 2},
			})
		default:
			t.Errorf("unexpected page %q", page)
			writeJSON(w, http.StatusOK, map[string]any{"result": []map[string]any{}, "meta": map[string]any{"page": 3, "limit": 1, "total": 2}})
		}
	})

	var names []string
	for s, err := range c.MonitoringServerStatus.All(t.Context(), WithLimit(1)) {
		if err != nil {
			t.Fatalf("All: %v", err)
		}
		names = append(names, s.Name)
	}
	if len(names) != 2 || names[0] != "Central" || names[1] != "Poller" {
		t.Errorf("All names = %v, want [Central Poller]", names)
	}
}
