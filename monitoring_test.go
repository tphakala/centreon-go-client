package centreon

import (
	"net/http"
	"testing"
)

func TestMonitoringResourceService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/monitoring/resources", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[MonitoringResource]{
			Result: []MonitoringResource{
				{
					ID:   1,
					Name: "host-01",
					Type: "host",
					FQDN: "host-01.example.com",
					Status: ResourceStatus{
						Code:         0,
						Name:         "OK",
						SeverityCode: 5,
					},
				},
				{
					ID:   2,
					Name: "Ping",
					Type: "service",
					Status: ResourceStatus{
						Code:         0,
						Name:         "OK",
						SeverityCode: 5,
					},
				},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 2},
		})
	})

	resp, err := c.Monitoring.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}
	if resp.Result[0].Type != "host" {
		t.Errorf("Result[0].Type = %q, want %q", resp.Result[0].Type, "host")
	}
	if resp.Result[1].Type != "service" {
		t.Errorf("Result[1].Type = %q, want %q", resp.Result[1].Type, "service")
	}
}

func TestMonitoringResourceService_GetHost(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/monitoring/resources/hosts/42", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, MonitoringResource{
			ID:   42,
			Name: "host-42",
			Type: "host",
			FQDN: "host-42.example.com",
			Status: ResourceStatus{
				Code:         0,
				Name:         "UP",
				SeverityCode: 5,
			},
		})
	})

	resource, err := c.Monitoring.GetHost(t.Context(), 42)
	if err != nil {
		t.Fatalf("GetHost: %v", err)
	}
	if resource.ID != 42 {
		t.Errorf("ID = %d, want 42", resource.ID)
	}
	if resource.Type != "host" {
		t.Errorf("Type = %q, want %q", resource.Type, "host")
	}
	if resource.FQDN != "host-42.example.com" {
		t.Errorf("FQDN = %q, want %q", resource.FQDN, "host-42.example.com")
	}
}

func TestMonitoringResourceService_GetService(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/monitoring/resources/hosts/10/services/5", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, MonitoringResource{
			ID:   5,
			Name: "Ping",
			Type: "service",
			Status: ResourceStatus{
				Code:         0,
				Name:         "OK",
				SeverityCode: 5,
			},
		})
	})

	resource, err := c.Monitoring.GetService(t.Context(), 10, 5)
	if err != nil {
		t.Fatalf("GetService: %v", err)
	}
	if resource.ID != 5 {
		t.Errorf("ID = %d, want 5", resource.ID)
	}
	if resource.Type != "service" {
		t.Errorf("Type = %q, want %q", resource.Type, "service")
	}
}
