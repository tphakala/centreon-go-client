package centreon

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"testing"
	"time"
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

func TestMonitoringServiceService_Timeline(t *testing.T) {
	mux, c := newTestMux(t)

	eventTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	var gotLimit string

	mux.HandleFunc("GET /centreon/api/latest/monitoring/hosts/10/services/20/timeline", func(w http.ResponseWriter, r *http.Request) {
		gotLimit = r.URL.Query().Get("limit")
		// Map-based body pins the JSON tags independently of TimelineEvent, so a
		// tag typo cannot round-trip invisibly. The first event carries the extra
		// fields the live API sends (status, tries, start_date, end_date, contact)
		// to confirm the struct correctly ignores them.
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{
					"id":         1,
					"type":       "event",
					"date":       "2024-01-15T10:00:00Z",
					"content":    "OK - 10.0.0.1: rta 0.207ms lost 0%",
					"start_date": nil,
					"end_date":   nil,
					"contact":    nil,
					"status":     map[string]any{"code": 0, "name": "OK", "severity_code": 5},
					"tries":      1,
				},
				{
					"id":      2,
					"type":    "notification",
					"date":    "2024-01-15T10:00:00Z",
					"content": "Notification sent",
				},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 2},
		})
	})

	resp, err := c.MonitoringServices.Timeline(t.Context(), 10, 20, WithLimit(5))
	if err != nil {
		t.Fatalf("Timeline: %v", err)
	}
	// Confirms the ListOption variadic is forwarded to the query string.
	if gotLimit != "5" {
		t.Errorf("forwarded limit = %q, want %q", gotLimit, "5")
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}
	if resp.Result[0].Type != "event" {
		t.Errorf("Result[0].Type = %q, want %q", resp.Result[0].Type, "event")
	}
	if !resp.Result[0].Date.Equal(eventTime) {
		t.Errorf("Result[0].Date = %v, want %v", resp.Result[0].Date, eventTime)
	}
	if resp.Result[1].Content != "Notification sent" {
		t.Errorf("Result[1].Content = %q, want %q", resp.Result[1].Content, "Notification sent")
	}
}

func TestMonitoringServiceService_Metrics(t *testing.T) {
	mux, c := newTestMux(t)

	// The payload is []map[string]any on purpose: it pins the JSON tags
	// independently of the Metric struct, so a tag typo cannot cancel out.
	// nil map values marshal to JSON null, exercising the pointer fields.
	mux.HandleFunc("GET /centreon/api/latest/monitoring/hosts/10/services/20/metrics", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []map[string]any{
			{
				"id":                      5053,
				"name":                    "pl",
				"unit":                    "%",
				"current_value":           0,
				"warning_high_threshold":  75,
				"warning_low_threshold":   0,
				"critical_high_threshold": 100,
				"critical_low_threshold":  0,
			},
			{
				"id":                      5054,
				"name":                    "rtmax",
				"unit":                    "ms",
				"current_value":           0.474,
				"warning_high_threshold":  nil,
				"warning_low_threshold":   nil,
				"critical_high_threshold": nil,
				"critical_low_threshold":  nil,
			},
		})
	})

	metrics, err := c.MonitoringServices.Metrics(t.Context(), 10, 20)
	if err != nil {
		t.Fatalf("Metrics: %v", err)
	}
	if len(metrics) != 2 {
		t.Fatalf("len(metrics) = %d, want 2", len(metrics))
	}

	pl := metrics[0]
	if pl.ID != 5053 {
		t.Errorf("metrics[0].ID = %d, want 5053", pl.ID)
	}
	if pl.Name != "pl" {
		t.Errorf("metrics[0].Name = %q, want %q", pl.Name, "pl")
	}
	if pl.Unit != "%" {
		t.Errorf("metrics[0].Unit = %q, want %q", pl.Unit, "%")
	}
	if pl.CurrentValue == nil {
		t.Fatal("metrics[0].CurrentValue = nil, want 0")
	}
	if *pl.CurrentValue != 0 {
		t.Errorf("*metrics[0].CurrentValue = %v, want 0", *pl.CurrentValue)
	}
	if pl.WarningHighThreshold == nil {
		t.Fatal("metrics[0].WarningHighThreshold = nil, want 75")
	}
	if *pl.WarningHighThreshold != 75 {
		t.Errorf("*metrics[0].WarningHighThreshold = %v, want 75", *pl.WarningHighThreshold)
	}

	rtmax := metrics[1]
	if rtmax.Unit != "ms" {
		t.Errorf("metrics[1].Unit = %q, want %q", rtmax.Unit, "ms")
	}
	if rtmax.CurrentValue == nil {
		t.Fatal("metrics[1].CurrentValue = nil, want 0.474")
	}
	if *rtmax.CurrentValue != 0.474 {
		t.Errorf("*metrics[1].CurrentValue = %v, want 0.474", *rtmax.CurrentValue)
	}
	if rtmax.WarningHighThreshold != nil {
		t.Errorf("metrics[1].WarningHighThreshold = %v, want nil", *rtmax.WarningHighThreshold)
	}
	if rtmax.CriticalHighThreshold != nil {
		t.Errorf("metrics[1].CriticalHighThreshold = %v, want nil", *rtmax.CriticalHighThreshold)
	}
}

// TestMonitoringServiceService_Metrics_RoundTrip verifies that Metric
// deserializes correctly from the exact JSON the live Centreon API returns.
func TestMonitoringServiceService_Metrics_RoundTrip(t *testing.T) {
	apiJSON := `[
		{"id":5053,"name":"pl","unit":"%","current_value":0,"warning_high_threshold":75,"warning_low_threshold":0,"critical_high_threshold":100,"critical_low_threshold":0},
		{"id":5054,"name":"rtmax","unit":"ms","current_value":0.474,"warning_high_threshold":null,"warning_low_threshold":null,"critical_high_threshold":null,"critical_low_threshold":null}
	]`

	var metrics []Metric
	if err := json.Unmarshal([]byte(apiJSON), &metrics); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	// reflect.DeepEqual dereferences the *float64 fields, so it distinguishes
	// a set threshold from a null one (nil pointer) across the whole slice.
	f := func(v float64) *float64 { return &v }
	want := []Metric{
		{
			ID: 5053, Name: "pl", Unit: "%",
			CurrentValue:          f(0),
			WarningHighThreshold:  f(75),
			WarningLowThreshold:   f(0),
			CriticalHighThreshold: f(100),
			CriticalLowThreshold:  f(0),
		},
		{
			ID: 5054, Name: "rtmax", Unit: "ms",
			CurrentValue: f(0.474),
			// warning/critical thresholds stay nil: the live payload sends null.
		},
	}
	if !reflect.DeepEqual(metrics, want) {
		// Marshal back to JSON for a readable diff: pointer fields render as
		// their numeric value, and a nil threshold renders as null.
		gotJSON, _ := json.Marshal(metrics) //nolint:errchkjson // Metric marshals without error
		wantJSON, _ := json.Marshal(want)   //nolint:errchkjson // Metric marshals without error
		t.Errorf("round-trip mismatch:\n got  %s\n want %s", gotJSON, wantJSON)
	}
}

func TestMonitoringServiceService_Metrics_Error(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/monitoring/hosts/10/services/20/metrics", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "boom"})
	})

	metrics, err := c.MonitoringServices.Metrics(t.Context(), 10, 20)
	if err == nil {
		t.Fatal("Metrics: want error on HTTP 500, got nil")
	}
	if metrics != nil {
		t.Errorf("Metrics result = %v, want nil on error", metrics)
	}
}

func TestMonitoringServiceService_Metrics_NoMetrics(t *testing.T) {
	mux, c := newTestMux(t)

	// On 25.10.x a service with no perfdata returns 404 "metrics not found".
	// Metrics maps that specific response to an empty result, not an error.
	mux.HandleFunc("GET /centreon/api/latest/monitoring/hosts/10/services/20/metrics", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "metrics not found"})
	})

	metrics, err := c.MonitoringServices.Metrics(t.Context(), 10, 20)
	if err != nil {
		t.Fatalf("Metrics: want nil error for a no-perfdata 404, got %v", err)
	}
	if len(metrics) != 0 {
		t.Errorf("Metrics = %v, want empty", metrics)
	}
}

func TestMonitoringServiceService_Metrics_NotFoundError(t *testing.T) {
	mux, c := newTestMux(t)

	// A 404 whose message is not "metrics not found" must surface as an *APIError.
	// (On live 25.10.16 a nonexistent host or service actually returns "metrics
	// not found" too and is mapped to empty; this synthetic different-message
	// case pins the message gate for any other 404 shape the server might send.)
	mux.HandleFunc("GET /centreon/api/latest/monitoring/hosts/10/services/20/metrics", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "Resource not found"})
	})

	metrics, err := c.MonitoringServices.Metrics(t.Context(), 10, 20)
	if err == nil {
		t.Fatal("Metrics: want error on a non-perfdata 404, got nil")
	}
	if metrics != nil {
		t.Errorf("Metrics result = %v, want nil on error", metrics)
	}
	apiErr, ok := errors.AsType[*APIError](err)
	if !ok {
		t.Fatalf("error is %T, want *APIError", err)
	}
	if apiErr.HTTPStatus != http.StatusNotFound {
		t.Errorf("HTTPStatus = %d, want 404", apiErr.HTTPStatus)
	}
}

func TestMonitoringServiceService_Metrics_NotFoundMessageOn500(t *testing.T) {
	mux, c := newTestMux(t)

	// The no-metrics mapping is gated on HTTP 404 specifically. A non-404 status
	// carrying the same "metrics not found" message (here a 500) must still
	// surface as an error, not be mapped to an empty result. This pins the
	// HTTPStatus == 404 condition of the guard.
	mux.HandleFunc("GET /centreon/api/latest/monitoring/hosts/10/services/20/metrics", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "metrics not found"})
	})

	metrics, err := c.MonitoringServices.Metrics(t.Context(), 10, 20)
	if err == nil {
		t.Fatal("Metrics: want error for a 500 even with a 'metrics not found' message, got nil")
	}
	if metrics != nil {
		t.Errorf("Metrics result = %v, want nil on error", metrics)
	}
}

func TestMonitoringServiceService_Metrics_NotFoundMessageSuperstring(t *testing.T) {
	mux, c := newTestMux(t)

	// The no-metrics mapping matches the message exactly. A 404 whose message
	// merely contains "metrics not found" as a substring (here with trailing
	// context) is a different response and must surface as an *APIError, not be
	// mapped to an empty result. This pins the exact-message gate against a
	// substring match.
	mux.HandleFunc("GET /centreon/api/latest/monitoring/hosts/10/services/20/metrics", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "metrics not found for service 20"})
	})

	metrics, err := c.MonitoringServices.Metrics(t.Context(), 10, 20)
	if err == nil {
		t.Fatal("Metrics: want error for a 404 whose message only contains 'metrics not found', got nil")
	}
	if metrics != nil {
		t.Errorf("Metrics result = %v, want nil on error", metrics)
	}
	apiErr, ok := errors.AsType[*APIError](err)
	if !ok {
		t.Fatalf("error is %T, want *APIError", err)
	}
	if apiErr.HTTPStatus != http.StatusNotFound {
		t.Errorf("HTTPStatus = %d, want 404", apiErr.HTTPStatus)
	}
}
