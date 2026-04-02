package centreon

import (
	"encoding/json"
	"net/http"
	"testing"
)

// checkMonitoringResource asserts all fields of a MonitoringResource against expected values.
func checkMonitoringResource(t *testing.T, label string, got, want *MonitoringResource) {
	t.Helper()
	checkMonitoringResourceIdentity(t, label, got, want)
	checkMonitoringResourceState(t, label, got, want)
	checkMonitoringResourceParent(t, label, got.Parent, want.Parent)
}

// checkMonitoringResourceIdentity asserts identity and metadata fields.
func checkMonitoringResourceIdentity(t *testing.T, label string, got, want *MonitoringResource) {
	t.Helper()
	if got.ID != want.ID {
		t.Errorf("%s.ID = %d, want %d", label, got.ID, want.ID)
	}
	if got.Name != want.Name {
		t.Errorf("%s.Name = %q, want %q", label, got.Name, want.Name)
	}
	if got.Type != want.Type {
		t.Errorf("%s.Type = %q, want %q", label, got.Type, want.Type)
	}
	if got.Alias != want.Alias {
		t.Errorf("%s.Alias = %q, want %q", label, got.Alias, want.Alias)
	}
	if got.FQDN != want.FQDN {
		t.Errorf("%s.FQDN = %q, want %q", label, got.FQDN, want.FQDN)
	}
	if got.HostID != want.HostID {
		t.Errorf("%s.HostID = %d, want %d", label, got.HostID, want.HostID)
	}
	if got.ServiceID != want.ServiceID {
		t.Errorf("%s.ServiceID = %d, want %d", label, got.ServiceID, want.ServiceID)
	}
	if got.MonitoringServerName != want.MonitoringServerName {
		t.Errorf("%s.MonitoringServerName = %q, want %q", label, got.MonitoringServerName, want.MonitoringServerName)
	}
}

// checkMonitoringResourceState asserts status and operational state fields.
func checkMonitoringResourceState(t *testing.T, label string, got, want *MonitoringResource) {
	t.Helper()
	if got.Status != want.Status {
		t.Errorf("%s.Status = %+v, want %+v", label, got.Status, want.Status)
	}
	if got.IsInDowntime != want.IsInDowntime {
		t.Errorf("%s.IsInDowntime = %v, want %v", label, got.IsInDowntime, want.IsInDowntime)
	}
	if got.IsAcknowledged != want.IsAcknowledged {
		t.Errorf("%s.IsAcknowledged = %v, want %v", label, got.IsAcknowledged, want.IsAcknowledged)
	}
	if got.Information != want.Information {
		t.Errorf("%s.Information = %q, want %q", label, got.Information, want.Information)
	}
	if got.Tries != want.Tries {
		t.Errorf("%s.Tries = %q, want %q", label, got.Tries, want.Tries)
	}
	if got.LastStatusChange != want.LastStatusChange {
		t.Errorf("%s.LastStatusChange = %q, want %q", label, got.LastStatusChange, want.LastStatusChange)
	}
	if got.NotificationEnabled != want.NotificationEnabled {
		t.Errorf("%s.NotificationEnabled = %v, want %v", label, got.NotificationEnabled, want.NotificationEnabled)
	}
}

// checkMonitoringResourceParent asserts parent fields or nil-ness.
func checkMonitoringResourceParent(t *testing.T, label string, got, want *MonitoringResourceParent) {
	t.Helper()
	if want == nil {
		if got != nil {
			t.Errorf("%s.Parent = %+v, want nil", label, got)
		}
		return
	}
	if got == nil {
		t.Fatalf("%s.Parent is nil, want %+v", label, want)
	}
	if *got != *want {
		t.Errorf("%s.Parent = %+v, want %+v", label, *got, *want)
	}
}

func TestMonitoringResourceService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/monitoring/resources", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{
					"id":                      1,
					"name":                    "host-01",
					"type":                    "host",
					"alias":                   "Primary host",
					"fqdn":                    "host-01.example.com",
					"host_id":                 1,
					"monitoring_server_name":  "Central",
					"parent":                  nil,
					"status":                  map[string]any{"code": 0, "name": "UP", "severity_code": 5},
					"is_in_downtime":          false,
					"is_acknowledged":         false,
					"information":             "OK - 10.0.0.1 rta 0.5ms lost 0%",
					"tries":                   "1/3 (H)",
					"last_status_change":      "2026-04-01T20:10:45+03:00",
					"is_notification_enabled": true,
				},
				{
					"id":                     5447,
					"name":                   "DB Content",
					"type":                   "service",
					"alias":                  nil,
					"fqdn":                   nil,
					"host_id":                5643,
					"service_id":             5447,
					"monitoring_server_name": "Central",
					"parent": map[string]any{
						"id":     5643,
						"name":   "Report_test_server",
						"type":   "host",
						"status": map[string]any{"code": 0, "name": "UP", "severity_code": 5},
					},
					"status":                  map[string]any{"code": 2, "name": "CRITICAL", "severity_code": 1},
					"is_in_downtime":          false,
					"is_acknowledged":         true,
					"information":             "CRITICAL - Unexpected output",
					"tries":                   "2/2 (H)",
					"last_status_change":      "2026-03-27T15:44:27+02:00",
					"is_notification_enabled": true,
				},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 2},
		})
	})

	resp, err := c.Monitoring.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}

	checkMonitoringResource(t, "Result[0]", &resp.Result[0], &MonitoringResource{
		ID: 1, Name: "host-01", Type: "host",
		Alias: "Primary host", FQDN: "host-01.example.com",
		HostID: 1, MonitoringServerName: "Central",
		Status:              ResourceStatus{Code: 0, Name: "UP", SeverityCode: 5},
		Information:         "OK - 10.0.0.1 rta 0.5ms lost 0%",
		Tries:               "1/3 (H)",
		LastStatusChange:    "2026-04-01T20:10:45+03:00",
		NotificationEnabled: true,
	})

	checkMonitoringResource(t, "Result[1]", &resp.Result[1], &MonitoringResource{
		ID: 5447, Name: "DB Content", Type: "service",
		HostID: 5643, ServiceID: 5447, MonitoringServerName: "Central",
		Parent: &MonitoringResourceParent{
			ID: 5643, Name: "Report_test_server", Type: "host",
			Status: ResourceStatus{Code: 0, Name: "UP", SeverityCode: 5},
		},
		Status:              ResourceStatus{Code: 2, Name: "CRITICAL", SeverityCode: 1},
		IsAcknowledged:      true,
		Information:         "CRITICAL - Unexpected output",
		Tries:               "2/2 (H)",
		LastStatusChange:    "2026-03-27T15:44:27+02:00",
		NotificationEnabled: true,
	})
}

func TestMonitoringResourceService_GetHost(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/monitoring/resources/hosts/42", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"id":                      42,
			"name":                    "host-42",
			"type":                    "host",
			"fqdn":                    "host-42.example.com",
			"host_id":                 42,
			"monitoring_server_name":  "Central",
			"parent":                  nil,
			"status":                  map[string]any{"code": 0, "name": "UP", "severity_code": 5},
			"is_in_downtime":          false,
			"is_acknowledged":         false,
			"information":             "OK - host-42 rta 0.2ms lost 0%",
			"tries":                   "1/2 (H)",
			"last_status_change":      "2026-04-01T12:00:00+03:00",
			"is_notification_enabled": true,
		})
	})

	resource, err := c.Monitoring.GetHost(t.Context(), 42)
	if err != nil {
		t.Fatalf("GetHost: %v", err)
	}

	checkMonitoringResource(t, "host", resource, &MonitoringResource{
		ID: 42, Name: "host-42", Type: "host",
		FQDN: "host-42.example.com", HostID: 42,
		MonitoringServerName: "Central",
		Status:               ResourceStatus{Code: 0, Name: "UP", SeverityCode: 5},
		Information:          "OK - host-42 rta 0.2ms lost 0%",
		Tries:                "1/2 (H)",
		LastStatusChange:     "2026-04-01T12:00:00+03:00",
		NotificationEnabled:  true,
	})
}

func TestMonitoringResourceService_GetService(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/monitoring/resources/hosts/10/services/5", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"id":                     5,
			"name":                   "Ping",
			"type":                   "service",
			"host_id":                10,
			"service_id":             5,
			"monitoring_server_name": "Central",
			"parent": map[string]any{
				"id":     10,
				"name":   "web01",
				"type":   "host",
				"status": map[string]any{"code": 0, "name": "UP", "severity_code": 5},
			},
			"status":                  map[string]any{"code": 0, "name": "OK", "severity_code": 5},
			"is_in_downtime":          false,
			"is_acknowledged":         false,
			"information":             "OK - rta 0.5ms lost 0%",
			"tries":                   "1/3 (H)",
			"last_status_change":      "2026-04-01T08:00:00+03:00",
			"is_notification_enabled": true,
		})
	})

	resource, err := c.Monitoring.GetService(t.Context(), 10, 5)
	if err != nil {
		t.Fatalf("GetService: %v", err)
	}

	checkMonitoringResource(t, "service", resource, &MonitoringResource{
		ID: 5, Name: "Ping", Type: "service",
		HostID: 10, ServiceID: 5, MonitoringServerName: "Central",
		Parent: &MonitoringResourceParent{
			ID: 10, Name: "web01", Type: "host",
			Status: ResourceStatus{Code: 0, Name: "UP", SeverityCode: 5},
		},
		Status:              ResourceStatus{Code: 0, Name: "OK", SeverityCode: 5},
		Information:         "OK - rta 0.5ms lost 0%",
		Tries:               "1/3 (H)",
		LastStatusChange:    "2026-04-01T08:00:00+03:00",
		NotificationEnabled: true,
	})
}

// TestMonitoringResource_RoundTrip verifies that the struct deserializes
// correctly from the exact JSON the live Centreon API returns.
func TestMonitoringResource_RoundTrip(t *testing.T) {
	apiJSON := `{
		"uuid": "h5643-s5447",
		"id": 5447,
		"type": "service",
		"name": "DB Content",
		"alias": null,
		"fqdn": null,
		"host_id": 5643,
		"service_id": 5447,
		"monitoring_server_name": "Central",
		"parent": {
			"uuid": "h5643",
			"id": 5643,
			"name": "Report_test_server",
			"type": "host",
			"short_type": "h",
			"status": {"code": 0, "name": "UP", "severity_code": 5},
			"alias": "test report server",
			"fqdn": "10.204.74.23"
		},
		"status": {"code": 2, "name": "CRITICAL", "severity_code": 1},
		"is_in_downtime": false,
		"is_acknowledged": true,
		"is_in_flapping": false,
		"has_active_checks_enabled": true,
		"has_passive_checks_enabled": true,
		"last_status_change": "2026-03-27T15:44:27+02:00",
		"tries": "2/2 (H)",
		"information": "CRITICAL - Unexpected output: ...",
		"is_notification_enabled": true,
		"severity": null
	}`

	var r MonitoringResource
	if err := json.Unmarshal([]byte(apiJSON), &r); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	checkMonitoringResource(t, "resource", &r, &MonitoringResource{
		ID: 5447, Name: "DB Content", Type: "service",
		HostID: 5643, ServiceID: 5447, MonitoringServerName: "Central",
		Parent: &MonitoringResourceParent{
			ID: 5643, Name: "Report_test_server", Type: "host",
			Status: ResourceStatus{Code: 0, Name: "UP", SeverityCode: 5},
		},
		Status:              ResourceStatus{Code: 2, Name: "CRITICAL", SeverityCode: 1},
		IsAcknowledged:      true,
		Information:         "CRITICAL - Unexpected output: ...",
		Tries:               "2/2 (H)",
		LastStatusChange:    "2026-03-27T15:44:27+02:00",
		NotificationEnabled: true,
	})
}
