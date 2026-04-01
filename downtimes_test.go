package centreon

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestDowntimeService_List(t *testing.T) {
	mux, c := newTestMux(t)

	start := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	svcID := 5

	mux.HandleFunc("GET /centreon/api/latest/monitoring/downtimes", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[Downtime]{
			Result: []Downtime{
				{
					ID:         1,
					HostID:     10,
					AuthorID:   1,
					AuthorName: "admin",
					Comment:    "Maintenance window",
					IsFixed:    true,
					StartTime:  start,
					EndTime:    end,
					Duration:   7200,
					PollerID:   1,
				},
				{
					ID:         2,
					HostID:     10,
					ServiceID:  &svcID,
					AuthorID:   1,
					AuthorName: "admin",
					Comment:    "Service downtime",
					IsFixed:    false,
					StartTime:  start,
					EndTime:    end,
					Duration:   3600,
					PollerID:   1,
				},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 2},
		})
	})

	resp, err := c.Downtimes.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}
	if resp.Result[0].Comment != "Maintenance window" {
		t.Errorf("Result[0].Comment = %q, want %q", resp.Result[0].Comment, "Maintenance window")
	}
	if resp.Result[1].ServiceID == nil {
		t.Fatal("Result[1].ServiceID is nil, want non-nil")
	}
	if *resp.Result[1].ServiceID != 5 {
		t.Errorf("Result[1].ServiceID = %d, want 5", *resp.Result[1].ServiceID)
	}
}

func TestDowntimeService_Get(t *testing.T) {
	mux, c := newTestMux(t)

	start := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	mux.HandleFunc("GET /centreon/api/latest/monitoring/downtimes/42", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, Downtime{
			ID:         42,
			HostID:     10,
			AuthorID:   1,
			AuthorName: "admin",
			Comment:    "Planned maintenance",
			IsFixed:    true,
			StartTime:  start,
			EndTime:    end,
			Duration:   7200,
			PollerID:   1,
		})
	})

	dt, err := c.Downtimes.Get(t.Context(), 42)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if dt.ID != 42 {
		t.Errorf("ID = %d, want 42", dt.ID)
	}
	if dt.Comment != "Planned maintenance" {
		t.Errorf("Comment = %q, want %q", dt.Comment, "Planned maintenance")
	}
}

func TestDowntimeService_Cancel(t *testing.T) {
	mux, c := newTestMux(t)

	var called bool
	mux.HandleFunc("DELETE /centreon/api/latest/monitoring/downtimes/7", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	if err := c.Downtimes.Cancel(t.Context(), 7); err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestDowntimeService_ListForHost(t *testing.T) {
	mux, c := newTestMux(t)

	start := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	mux.HandleFunc("GET /centreon/api/latest/monitoring/hosts/10/downtimes", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[Downtime]{
			Result: []Downtime{
				{
					ID:         1,
					HostID:     10,
					AuthorID:   1,
					AuthorName: "admin",
					Comment:    "Host maintenance",
					IsFixed:    true,
					StartTime:  start,
					EndTime:    end,
					Duration:   7200,
					PollerID:   1,
				},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 1},
		})
	})

	resp, err := c.Downtimes.ListForHost(t.Context(), 10)
	if err != nil {
		t.Fatalf("ListForHost: %v", err)
	}
	if len(resp.Result) != 1 {
		t.Fatalf("len(Result) = %d, want 1", len(resp.Result))
	}
	if resp.Result[0].HostID != 10 {
		t.Errorf("Result[0].HostID = %d, want 10", resp.Result[0].HostID)
	}
}

func TestDowntimeService_ListForService(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/monitoring/hosts/10/services/5/downtimes", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[Downtime]{
			Result: []Downtime{{ID: 1, HostID: 10, ServiceID: new(5), Comment: "svc downtime"}},
			Meta:   Meta{Page: 1, Limit: 10, Total: 1},
		})
	})

	resp, err := c.Downtimes.ListForService(t.Context(), 10, 5)
	if err != nil {
		t.Fatalf("ListForService: %v", err)
	}
	if len(resp.Result) != 1 {
		t.Fatalf("len(Result) = %d, want 1", len(resp.Result))
	}
	if resp.Result[0].ServiceID == nil || *resp.Result[0].ServiceID != 5 {
		t.Error("expected ServiceID 5")
	}
}

func TestDowntimeService_CreateForHost(t *testing.T) {
	mux, c := newTestMux(t)

	start := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	var called bool

	mux.HandleFunc("POST /centreon/api/latest/monitoring/hosts/10/downtimes", func(w http.ResponseWriter, r *http.Request) {
		called = true
		var req CreateDowntimeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if req.Comment != "Host maintenance" {
			t.Errorf("Comment = %q, want %q", req.Comment, "Host maintenance")
		}
		if !req.IsFixed {
			t.Error("IsFixed should be true")
		}
		if !req.StartTime.Equal(start) {
			t.Errorf("StartTime = %v, want %v", req.StartTime, start)
		}
		if !req.EndTime.Equal(end) {
			t.Errorf("EndTime = %v, want %v", req.EndTime, end)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.Downtimes.CreateForHost(t.Context(), 10, &CreateDowntimeRequest{
		Comment:   "Host maintenance",
		StartTime: start,
		EndTime:   end,
		IsFixed:   true,
		Duration:  7200,
	})
	if err != nil {
		t.Fatalf("CreateForHost: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestDowntimeService_CreateForService(t *testing.T) {
	mux, c := newTestMux(t)

	start := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	var called bool

	mux.HandleFunc("POST /centreon/api/latest/monitoring/hosts/10/services/5/downtimes", func(w http.ResponseWriter, r *http.Request) {
		called = true
		var req CreateDowntimeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if req.Comment != "Service maintenance" {
			t.Errorf("Comment = %q, want %q", req.Comment, "Service maintenance")
		}
		if req.IsFixed {
			t.Error("IsFixed should be false")
		}
		if req.Duration != 3600 {
			t.Errorf("Duration = %d, want 3600", req.Duration)
		}
		if !req.StartTime.Equal(start) {
			t.Errorf("StartTime = %v, want %v", req.StartTime, start)
		}
		if !req.EndTime.Equal(end) {
			t.Errorf("EndTime = %v, want %v", req.EndTime, end)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.Downtimes.CreateForService(t.Context(), 10, 5, &CreateDowntimeRequest{
		Comment:   "Service maintenance",
		StartTime: start,
		EndTime:   end,
		IsFixed:   false,
		Duration:  3600,
	})
	if err != nil {
		t.Fatalf("CreateForService: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestDowntimeService_CreateForHost_WithServices(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("POST /centreon/api/latest/monitoring/hosts/10/downtimes", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["with_services"] != true {
			t.Errorf("with_services = %v, want true", body["with_services"])
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.Downtimes.CreateForHost(t.Context(), 10, &CreateDowntimeRequest{
		Comment:      "test",
		StartTime:    time.Now(),
		EndTime:      time.Now().Add(time.Hour),
		IsFixed:      true,
		WithServices: true,
	})
	if err != nil {
		t.Fatalf("CreateForHost: %v", err)
	}
}

func TestDowntimeService_CancelForHost(t *testing.T) {
	mux, c := newTestMux(t)

	var called bool
	mux.HandleFunc("DELETE /centreon/api/latest/monitoring/hosts/10/downtimes", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	if err := c.Downtimes.CancelForHost(t.Context(), 10); err != nil {
		t.Fatalf("CancelForHost: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestDowntimeService_CancelForService(t *testing.T) {
	mux, c := newTestMux(t)

	var called bool
	mux.HandleFunc("DELETE /centreon/api/latest/monitoring/hosts/10/services/5/downtimes", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	if err := c.Downtimes.CancelForService(t.Context(), 10, 5); err != nil {
		t.Fatalf("CancelForService: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}
