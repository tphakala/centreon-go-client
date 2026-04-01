package centreon

import (
	"net/http"
	"testing"
	"time"
)

func TestMonitoringHostService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/monitoring/hosts", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[MonitoringHost]{
			Result: []MonitoringHost{
				{
					ID:      1,
					Name:    "host-01",
					Address: "10.0.0.1",
					Alias:   "Main host",
					Status:  ResourceStatus{Code: 0, Name: "UP", SeverityCode: 5},
				},
				{
					ID:      2,
					Name:    "host-02",
					Address: "10.0.0.2",
					Status:  ResourceStatus{Code: 1, Name: "DOWN", SeverityCode: 1},
				},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 2},
		})
	})

	resp, err := c.MonitoringHosts.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}
	if resp.Result[0].Name != "host-01" {
		t.Errorf("Result[0].Name = %q, want %q", resp.Result[0].Name, "host-01")
	}
	if resp.Result[1].Status.Name != "DOWN" {
		t.Errorf("Result[1].Status.Name = %q, want %q", resp.Result[1].Status.Name, "DOWN")
	}
}

func TestMonitoringHostService_Get(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/monitoring/hosts/42", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, MonitoringHost{
			ID:      42,
			Name:    "host-42",
			Address: "10.0.0.42",
			Status:  ResourceStatus{Code: 0, Name: "UP", SeverityCode: 5},
		})
	})

	host, err := c.MonitoringHosts.Get(t.Context(), 42)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if host.ID != 42 {
		t.Errorf("ID = %d, want 42", host.ID)
	}
	if host.Name != "host-42" {
		t.Errorf("Name = %q, want %q", host.Name, "host-42")
	}
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
