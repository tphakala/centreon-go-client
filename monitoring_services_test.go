package centreon

import (
	"net/http"
	"testing"
)

func TestMonitoringServiceService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/monitoring/services", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[MonitoringService]{
			Result: []MonitoringService{
				{ID: 1, Name: "Ping", Status: ResourceStatus{Code: 0, Name: "OK", SeverityCode: 5}},
				{ID: 2, Name: "CPU", Status: ResourceStatus{Code: 1, Name: "WARNING", SeverityCode: 3}},
				{ID: 3, Name: "Memory", Status: ResourceStatus{Code: 2, Name: "CRITICAL", SeverityCode: 1}},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 3},
		})
	})

	resp, err := c.MonitoringServices.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 3 {
		t.Fatalf("len(Result) = %d, want 3", len(resp.Result))
	}
	if resp.Result[0].Name != "Ping" {
		t.Errorf("Result[0].Name = %q, want %q", resp.Result[0].Name, "Ping")
	}
	if resp.Result[2].Status.Name != "CRITICAL" {
		t.Errorf("Result[2].Status.Name = %q, want %q", resp.Result[2].Status.Name, "CRITICAL")
	}
}

func TestMonitoringServiceService_StatusCounts(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/monitoring/services/status", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ServiceStatusCount{
			OK:       50,
			Warning:  5,
			Critical: 2,
			Unknown:  1,
			Pending:  3,
		})
	})

	counts, err := c.MonitoringServices.StatusCounts(t.Context())
	if err != nil {
		t.Fatalf("StatusCounts: %v", err)
	}
	if counts.OK != 50 {
		t.Errorf("OK = %d, want 50", counts.OK)
	}
	if counts.Warning != 5 {
		t.Errorf("Warning = %d, want 5", counts.Warning)
	}
	if counts.Critical != 2 {
		t.Errorf("Critical = %d, want 2", counts.Critical)
	}
}
