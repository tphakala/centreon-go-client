package centreon

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestMonitoringServiceService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/monitoring/services", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{
					"id":                       1,
					"description":              "Ping",
					"host":                     map[string]any{"id": 100, "name": "web01", "alias": "web01.example.com", "state": 0},
					"state":                    0,
					"state_type":               1,
					"output":                   "OK - rta 0.5ms",
					"status":                   map[string]any{"code": 0, "name": "OK", "severity_code": 5},
					"is_acknowledged":          false,
					"scheduled_downtime_depth": 0,
					"last_check":               "2026-04-02T09:00:00+03:00",
					"max_check_attempts":       3,
				},
				{
					"id":                       2,
					"description":              "CPU",
					"host":                     map[string]any{"id": 101, "name": "db01", "state": 1},
					"state":                    1,
					"state_type":               1,
					"output":                   "WARNING: CPU 85%",
					"status":                   map[string]any{"code": 1, "name": "WARNING", "severity_code": 3},
					"is_acknowledged":          true,
					"scheduled_downtime_depth": 0,
					"max_check_attempts":       2,
				},
				{
					"id":                       3,
					"description":              "Memory",
					"host":                     map[string]any{"id": 102, "name": "app01", "state": 0},
					"state":                    2,
					"state_type":               1,
					"output":                   "CRITICAL: Memory 98%",
					"status":                   map[string]any{"code": 2, "name": "CRITICAL", "severity_code": 1},
					"is_acknowledged":          false,
					"scheduled_downtime_depth": 1,
					"max_check_attempts":       3,
				},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 3},
		})
	})

	resp, err := c.MonitoringServices.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 3 {
		t.Fatalf("len(Result) = %d, want 3", len(resp.Result))
	}

	svc := resp.Result[0]
	if svc.Description != "Ping" {
		t.Errorf("Result[0].Description = %q, want %q", svc.Description, "Ping")
	}
	if svc.Host.ID != 100 {
		t.Errorf("Result[0].Host.ID = %d, want 100", svc.Host.ID)
	}
	if svc.Host.Name != "web01" {
		t.Errorf("Result[0].Host.Name = %q, want %q", svc.Host.Name, "web01")
	}
	if svc.State != 0 {
		t.Errorf("Result[0].State = %d, want 0", svc.State)
	}
	if svc.Output != "OK - rta 0.5ms" {
		t.Errorf("Result[0].Output = %q, want %q", svc.Output, "OK - rta 0.5ms")
	}
	if svc.Status.Name != "OK" {
		t.Errorf("Result[0].Status.Name = %q, want %q", svc.Status.Name, "OK")
	}
	if svc.MaxCheckAttempts != 3 {
		t.Errorf("Result[0].MaxCheckAttempts = %d, want 3", svc.MaxCheckAttempts)
	}
	if svc.LastCheck != "2026-04-02T09:00:00+03:00" {
		t.Errorf("Result[0].LastCheck = %q, want %q", svc.LastCheck, "2026-04-02T09:00:00+03:00")
	}

	// Verify second service (acknowledged)
	if !resp.Result[1].Acknowledged {
		t.Errorf("Result[1].Acknowledged = false, want true")
	}
	if resp.Result[1].Description != "CPU" {
		t.Errorf("Result[1].Description = %q, want %q", resp.Result[1].Description, "CPU")
	}

	// Verify third service (critical with downtime)
	if resp.Result[2].Status.Name != "CRITICAL" {
		t.Errorf("Result[2].Status.Name = %q, want %q", resp.Result[2].Status.Name, "CRITICAL")
	}
	if resp.Result[2].DowntimeDepth != 1 {
		t.Errorf("Result[2].DowntimeDepth = %d, want 1", resp.Result[2].DowntimeDepth)
	}
}

func TestMonitoringService_JSON(t *testing.T) {
	raw := `{
		"id": 5568,
		"description": "CPU by average",
		"display_name": "CPU by average",
		"host": {"id": 10300, "name": "testhost", "alias": "testmonmap03", "state": 0},
		"is_acknowledged": false,
		"scheduled_downtime_depth": 0,
		"last_check": "2026-04-02T09:31:35+03:00",
		"last_state_change": "2026-04-02T07:15:35+03:00",
		"max_check_attempts": 2,
		"output": "OK: 2 CPU(s) average usage is 2.00 %",
		"state": 0,
		"state_type": 1,
		"status": {"code": 0, "name": "OK", "severity_code": 5}
	}`
	var svc MonitoringService
	if err := json.Unmarshal([]byte(raw), &svc); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if svc.ID != 5568 {
		t.Errorf("ID = %d, want 5568", svc.ID)
	}
	if svc.Description != "CPU by average" {
		t.Errorf("Description = %q, want %q", svc.Description, "CPU by average")
	}
	if svc.Host.ID != 10300 {
		t.Errorf("Host.ID = %d, want 10300", svc.Host.ID)
	}
	if svc.Host.Alias != "testmonmap03" {
		t.Errorf("Host.Alias = %q, want %q", svc.Host.Alias, "testmonmap03")
	}
	if svc.Output != "OK: 2 CPU(s) average usage is 2.00 %" {
		t.Errorf("Output = %q, want %q", svc.Output, "OK: 2 CPU(s) average usage is 2.00 %")
	}
	if svc.Status.Name != "OK" {
		t.Errorf("Status.Name = %q, want %q", svc.Status.Name, "OK")
	}
	if svc.MaxCheckAttempts != 2 {
		t.Errorf("MaxCheckAttempts = %d, want 2", svc.MaxCheckAttempts)
	}
	if svc.LastStateChange != "2026-04-02T07:15:35+03:00" {
		t.Errorf("LastStateChange = %q, want %q", svc.LastStateChange, "2026-04-02T07:15:35+03:00")
	}
}

func TestMonitoringServiceService_StatusCounts(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/monitoring/services/status", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ServiceStatusCount{
			OK:       StatusValue{Total: 50},
			Warning:  StatusValue{Total: 5},
			Critical: StatusValue{Total: 2},
			Unknown:  StatusValue{Total: 1},
			Pending:  StatusValue{Total: 3},
			Total:    61,
		})
	})

	counts, err := c.MonitoringServices.StatusCounts(t.Context())
	if err != nil {
		t.Fatalf("StatusCounts: %v", err)
	}
	if counts.OK.Total != 50 {
		t.Errorf("OK.Total = %d, want 50", counts.OK.Total)
	}
	if counts.Warning.Total != 5 {
		t.Errorf("Warning.Total = %d, want 5", counts.Warning.Total)
	}
	if counts.Critical.Total != 2 {
		t.Errorf("Critical.Total = %d, want 2", counts.Critical.Total)
	}
	if counts.Total != 61 {
		t.Errorf("Total = %d, want 61", counts.Total)
	}
}
