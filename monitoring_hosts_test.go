package centreon

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

// checkMonitoringHost asserts all fields of a MonitoringHost against expected values.
func checkMonitoringHost(t *testing.T, label string, got, want *MonitoringHost) {
	t.Helper()
	if *got != *want {
		t.Errorf("%s mismatch:\n got  %+v\n want %+v", label, *got, *want)
	}
}

func TestMonitoringHostService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/monitoring/hosts", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{
					"id":                       1,
					"poller_id":                1,
					"name":                     "host-01",
					"acknowledged":             false,
					"address_ip":               "10.0.0.1",
					"alias":                    "Main host",
					"check_attempt":            1,
					"max_check_attempts":       3,
					"state":                    0,
					"state_type":               1,
					"output":                   "OK - 10.0.0.1 rta 0.5ms lost 0%\n",
					"execution_time":           0.099,
					"last_check":               "2026-04-02T09:33:56+03:00",
					"last_state_change":        "2026-04-01T20:10:45+03:00",
					"scheduled_downtime_depth": 0,
				},
				{
					"id":                       2,
					"poller_id":                2,
					"name":                     "host-02",
					"acknowledged":             true,
					"address_ip":               "10.0.0.2",
					"alias":                    "",
					"check_attempt":            2,
					"max_check_attempts":       3,
					"state":                    1,
					"state_type":               1,
					"output":                   "CRITICAL - Host unreachable\n",
					"execution_time":           0.012,
					"last_check":               "2026-04-02T09:34:00+03:00",
					"last_state_change":        "2026-04-02T08:00:00+03:00",
					"scheduled_downtime_depth": 1,
				},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 2},
		})
	})

	resp, err := c.MonitoringHosts.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}

	checkMonitoringHost(t, "Result[0]", &resp.Result[0], &MonitoringHost{
		ID: 1, PollerID: 1, Name: "host-01", AddressIP: "10.0.0.1",
		Alias: "Main host", State: 0, StateType: 1,
		Output: "OK - 10.0.0.1 rta 0.5ms lost 0%\n", Acknowledged: false,
		CheckAttempt: 1, MaxCheckAttempts: 3, ExecutionTime: 0.099,
		LastCheck: "2026-04-02T09:33:56+03:00", LastStateChange: "2026-04-01T20:10:45+03:00",
		DowntimeDepth: 0,
	})

	checkMonitoringHost(t, "Result[1]", &resp.Result[1], &MonitoringHost{
		ID: 2, PollerID: 2, Name: "host-02", AddressIP: "10.0.0.2",
		State: 1, StateType: 1, Output: "CRITICAL - Host unreachable\n",
		Acknowledged: true, CheckAttempt: 2, MaxCheckAttempts: 3,
		ExecutionTime: 0.012, LastCheck: "2026-04-02T09:34:00+03:00",
		LastStateChange: "2026-04-02T08:00:00+03:00", DowntimeDepth: 1,
	})
}

func TestMonitoringHostService_Get(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/monitoring/hosts/42", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"id":                       42,
			"poller_id":                1,
			"name":                     "host-42",
			"acknowledged":             false,
			"address_ip":               "10.0.0.42",
			"alias":                    "test-host",
			"check_attempt":            1,
			"max_check_attempts":       2,
			"state":                    0,
			"state_type":               1,
			"output":                   "OK - 10.0.0.42 rta 0.2ms lost 0%\n",
			"execution_time":           0.055,
			"last_check":               "2026-04-02T10:00:00+03:00",
			"last_state_change":        "2026-04-01T12:00:00+03:00",
			"scheduled_downtime_depth": 0,
		})
	})

	host, err := c.MonitoringHosts.Get(t.Context(), 42)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	checkMonitoringHost(t, "host", host, &MonitoringHost{
		ID: 42, PollerID: 1, Name: "host-42", AddressIP: "10.0.0.42",
		Alias: "test-host", State: 0, StateType: 1,
		Output:        "OK - 10.0.0.42 rta 0.2ms lost 0%\n",
		ExecutionTime: 0.055, CheckAttempt: 1, MaxCheckAttempts: 2,
		LastCheck: "2026-04-02T10:00:00+03:00", LastStateChange: "2026-04-01T12:00:00+03:00",
	})
}

// TestMonitoringHostService_Get_RoundTrip verifies that the struct deserializes
// correctly from the exact JSON the live Centreon API returns.
func TestMonitoringHostService_Get_RoundTrip(t *testing.T) {
	apiJSON := `{
		"id": 10300,
		"poller_id": 1,
		"name": "19-11-2025_Host_on_central",
		"acknowledged": false,
		"address_ip": "10.204.74.29",
		"alias": "testmonmap03",
		"check_attempt": 1,
		"max_check_attempts": 2,
		"state": 0,
		"state_type": 1,
		"output": "OK - 10.204.74.29 rta 0.223ms lost 0%\n",
		"execution_time": 0.099193,
		"last_check": "2026-04-02T09:33:56+03:00",
		"last_state_change": "2026-04-01T20:10:45+03:00",
		"scheduled_downtime_depth": 0
	}`

	var host MonitoringHost
	if err := json.Unmarshal([]byte(apiJSON), &host); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	checkMonitoringHost(t, "host", &host, &MonitoringHost{
		ID: 10300, PollerID: 1, Name: "19-11-2025_Host_on_central",
		AddressIP: "10.204.74.29", Alias: "testmonmap03",
		State: 0, StateType: 1,
		Output:        "OK - 10.204.74.29 rta 0.223ms lost 0%\n",
		ExecutionTime: 0.099193, CheckAttempt: 1, MaxCheckAttempts: 2,
		LastCheck: "2026-04-02T09:33:56+03:00", LastStateChange: "2026-04-01T20:10:45+03:00",
	})
}

func TestMonitoringHostService_StatusCounts(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/monitoring/hosts/status", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, HostStatusCount{
			Up:          StatusValue{Total: 10},
			Down:        StatusValue{Total: 2},
			Unreachable: StatusValue{Total: 1},
			Pending:     StatusValue{Total: 0},
			Total:       13,
		})
	})

	counts, err := c.MonitoringHosts.StatusCounts(t.Context())
	if err != nil {
		t.Fatalf("StatusCounts: %v", err)
	}
	if counts.Up.Total != 10 {
		t.Errorf("Up.Total = %d, want 10", counts.Up.Total)
	}
	if counts.Down.Total != 2 {
		t.Errorf("Down.Total = %d, want 2", counts.Down.Total)
	}
	if counts.Unreachable.Total != 1 {
		t.Errorf("Unreachable.Total = %d, want 1", counts.Unreachable.Total)
	}
	if counts.Total != 13 {
		t.Errorf("Total = %d, want 13", counts.Total)
	}
}

func TestMonitoringHostService_Services(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/monitoring/hosts/10/services", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[MonitoringService]{
			Result: []MonitoringService{
				{ID: 1, Name: "Ping", Status: ResourceStatus{Code: 0, Name: "OK", SeverityCode: 5}},
				{ID: 2, Name: "CPU", Status: ResourceStatus{Code: 1, Name: "WARNING", SeverityCode: 3}},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 2},
		})
	})

	resp, err := c.MonitoringHosts.Services(t.Context(), 10)
	if err != nil {
		t.Fatalf("Services: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}
	if resp.Result[0].Name != "Ping" {
		t.Errorf("Result[0].Name = %q, want %q", resp.Result[0].Name, "Ping")
	}
}

func TestMonitoringHostService_Timeline(t *testing.T) {
	mux, c := newTestMux(t)

	eventTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	mux.HandleFunc("GET /centreon/api/latest/monitoring/hosts/10/timeline", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[TimelineEvent]{
			Result: []TimelineEvent{
				{ID: 1, Type: "alert", Content: "Host went DOWN", Date: eventTime},
				{ID: 2, Type: "notification", Content: "Notification sent", Date: eventTime},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 2},
		})
	})

	resp, err := c.MonitoringHosts.Timeline(t.Context(), 10)
	if err != nil {
		t.Fatalf("Timeline: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}
	if resp.Result[0].Type != "alert" {
		t.Errorf("Result[0].Type = %q, want %q", resp.Result[0].Type, "alert")
	}
	if resp.Result[1].Content != "Notification sent" {
		t.Errorf("Result[1].Content = %q, want %q", resp.Result[1].Content, "Notification sent")
	}
}
