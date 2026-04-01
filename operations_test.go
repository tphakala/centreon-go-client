package centreon

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestOperationsService_Acknowledge(t *testing.T) {
	mux, c := newTestMux(t)

	var called bool
	mux.HandleFunc("POST /centreon/api/latest/monitoring/resources/acknowledge", func(w http.ResponseWriter, r *http.Request) {
		called = true
		var req AcknowledgeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if len(req.Resources) != 1 {
			t.Errorf("len(Resources) = %d, want 1", len(req.Resources))
		}
		if req.Resources[0].Type != "host" {
			t.Errorf("Resources[0].Type = %q, want %q", req.Resources[0].Type, "host")
		}
		if req.Resources[0].ID != 42 {
			t.Errorf("Resources[0].ID = %d, want 42", req.Resources[0].ID)
		}
		if req.Comment != "Acknowledged by operator" {
			t.Errorf("Comment = %q, want %q", req.Comment, "Acknowledged by operator")
		}
		if !req.IsSticky {
			t.Error("IsSticky should be true")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.Operations.Acknowledge(t.Context(), &AcknowledgeRequest{
		Resources: []ResourceRef{
			{Type: "host", ID: 42},
		},
		Comment:  "Acknowledged by operator",
		IsSticky: true,
	})
	if err != nil {
		t.Fatalf("Acknowledge: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestOperationsService_Downtime(t *testing.T) {
	mux, c := newTestMux(t)

	start := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	var called bool
	mux.HandleFunc("POST /centreon/api/latest/monitoring/resources/downtime", func(w http.ResponseWriter, r *http.Request) {
		called = true
		var req DowntimeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if len(req.Resources) != 2 {
			t.Errorf("len(Resources) = %d, want 2", len(req.Resources))
		}
		if req.Comment != "Maintenance window" {
			t.Errorf("Comment = %q, want %q", req.Comment, "Maintenance window")
		}
		if !req.Fixed {
			t.Error("Fixed should be true")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.Operations.Downtime(t.Context(), &DowntimeRequest{
		Resources: []ResourceRef{
			{Type: "host", ID: 1},
			{Type: "host", ID: 2},
		},
		Comment:   "Maintenance window",
		StartTime: start,
		EndTime:   end,
		Fixed:     true,
	})
	if err != nil {
		t.Fatalf("Downtime: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestOperationsService_Check(t *testing.T) {
	mux, c := newTestMux(t)

	var called bool
	mux.HandleFunc("POST /centreon/api/latest/monitoring/resources/check", func(w http.ResponseWriter, r *http.Request) {
		called = true
		var req CheckRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if len(req.Resources) != 1 {
			t.Errorf("len(Resources) = %d, want 1", len(req.Resources))
		}
		if req.Resources[0].Type != "service" {
			t.Errorf("Resources[0].Type = %q, want %q", req.Resources[0].Type, "service")
		}
		if req.Resources[0].ID != 7 {
			t.Errorf("Resources[0].ID = %d, want 7", req.Resources[0].ID)
		}
		if req.Resources[0].Parent == nil || req.Resources[0].Parent.ID != 3 {
			t.Errorf("Resources[0].Parent.ID should be 3")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.Operations.Check(t.Context(), &CheckRequest{
		Resources: []ResourceRef{
			{Type: "service", ID: 7, Parent: &ResourceRef{Type: "host", ID: 3}},
		},
	})
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestOperationsService_Submit(t *testing.T) {
	mux, c := newTestMux(t)

	var called bool
	mux.HandleFunc("POST /centreon/api/latest/monitoring/resources/submit", func(w http.ResponseWriter, r *http.Request) {
		called = true
		var req SubmitResultRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if len(req.Resources) != 1 {
			t.Errorf("len(Resources) = %d, want 1", len(req.Resources))
		}
		res := req.Resources[0]
		if res.Type != "service" {
			t.Errorf("Resources[0].Type = %q, want %q", res.Type, "service")
		}
		if res.Status != 0 {
			t.Errorf("Resources[0].Status = %d, want 0", res.Status)
		}
		if res.Output != "All systems nominal" {
			t.Errorf("Resources[0].Output = %q, want %q", res.Output, "All systems nominal")
		}
		if res.PerfData != "rta=1ms" {
			t.Errorf("Resources[0].PerfData = %q, want %q", res.PerfData, "rta=1ms")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.Operations.Submit(t.Context(), &SubmitResultRequest{
		Resources: []SubmitResource{
			{
				Type:     "service",
				ID:       5,
				Parent:   &ResourceRef{Type: "host", ID: 1},
				Status:   0,
				Output:   "All systems nominal",
				PerfData: "rta=1ms",
			},
		},
	})
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestOperationsService_Comment(t *testing.T) {
	mux, c := newTestMux(t)

	var called bool
	mux.HandleFunc("POST /centreon/api/latest/monitoring/resources/comments", func(w http.ResponseWriter, r *http.Request) {
		called = true
		var req CommentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if len(req.Resources) != 1 {
			t.Errorf("len(Resources) = %d, want 1", len(req.Resources))
		}
		if req.Resources[0].Type != "host" {
			t.Errorf("Resources[0].Type = %q, want %q", req.Resources[0].Type, "host")
		}
		if req.Resources[0].ID != 10 {
			t.Errorf("Resources[0].ID = %d, want 10", req.Resources[0].ID)
		}
		if req.Comment != "Under investigation" {
			t.Errorf("Comment = %q, want %q", req.Comment, "Under investigation")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.Operations.Comment(t.Context(), &CommentRequest{
		Resources: []ResourceRef{
			{Type: "host", ID: 10},
		},
		Comment: "Under investigation",
	})
	if err != nil {
		t.Fatalf("Comment: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}
