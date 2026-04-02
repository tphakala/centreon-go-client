package centreon

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestAcknowledgementService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/monitoring/acknowledgements", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{
					"id":                    1,
					"host_id":               10,
					"author_id":             413,
					"author_name":           "admin",
					"comment":               "Host ack",
					"is_sticky":             true,
					"is_persistent_comment": true,
					"is_notify_contacts":    false,
					"state":                 1,
					"type":                  1,
					"entry_time":            "2024-01-15T09:00:00Z",
					"deletion_time":         nil,
				},
				{
					"id":                    2,
					"host_id":               10,
					"service_id":            5,
					"author_id":             413,
					"author_name":           "admin",
					"comment":               "Service ack",
					"is_sticky":             false,
					"is_persistent_comment": false,
					"is_notify_contacts":    true,
					"state":                 2,
					"type":                  1,
					"entry_time":            "2024-01-15T09:00:00Z",
					"deletion_time":         nil,
				},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 2},
		})
	})

	resp, err := c.Acknowledgements.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}
	if resp.Result[0].Comment != "Host ack" {
		t.Errorf("Result[0].Comment = %q, want %q", resp.Result[0].Comment, "Host ack")
	}
	if resp.Result[0].Type != 1 {
		t.Errorf("Result[0].Type = %d, want 1", resp.Result[0].Type)
	}
	if resp.Result[1].ServiceID == nil {
		t.Fatal("Result[1].ServiceID is nil, want non-nil")
	}
	if *resp.Result[1].ServiceID != 5 {
		t.Errorf("Result[1].ServiceID = %d, want 5", *resp.Result[1].ServiceID)
	}
	if resp.Result[1].Type != 1 {
		t.Errorf("Result[1].Type = %d, want 1", resp.Result[1].Type)
	}
}

func TestAcknowledgementService_Get(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/monitoring/acknowledgements/42", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"id":                    42,
			"host_id":               10,
			"author_id":             1,
			"author_name":           "admin",
			"comment":               "Acknowledged for investigation",
			"is_sticky":             true,
			"is_persistent_comment": true,
			"is_notify_contacts":    false,
			"state":                 1,
			"type":                  1,
			"entry_time":            "2024-01-15T09:00:00Z",
			"deletion_time":         nil,
		})
	})

	ack, err := c.Acknowledgements.Get(t.Context(), 42)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if ack.ID != 42 {
		t.Errorf("ID = %d, want 42", ack.ID)
	}
	if ack.Comment != "Acknowledged for investigation" {
		t.Errorf("Comment = %q, want %q", ack.Comment, "Acknowledged for investigation")
	}
	if ack.Type != 1 {
		t.Errorf("Type = %d, want 1", ack.Type)
	}
}

func TestAcknowledgementService_ListForHost(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/monitoring/hosts/10/acknowledgements", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{
					"id":          1,
					"host_id":     10,
					"author_id":   1,
					"author_name": "admin",
					"comment":     "Under investigation",
					"is_sticky":   true,
					"state":       1,
					"type":        1,
					"entry_time":  "2024-01-15T09:00:00Z",
				},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 1},
		})
	})

	resp, err := c.Acknowledgements.ListForHost(t.Context(), 10)
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

func TestAcknowledgementService_ListForService(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/monitoring/hosts/10/services/5/acknowledgements", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[Acknowledgement]{
			Result: []Acknowledgement{{ID: 1, HostID: 10, ServiceID: new(5), Comment: "svc ack"}},
			Meta:   Meta{Page: 1, Limit: 10, Total: 1},
		})
	})

	resp, err := c.Acknowledgements.ListForService(t.Context(), 10, 5)
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

func TestAcknowledgementService_CreateForHost(t *testing.T) {
	mux, c := newTestMux(t)

	var called bool
	mux.HandleFunc("POST /centreon/api/latest/monitoring/hosts/10/acknowledgements", func(w http.ResponseWriter, r *http.Request) {
		called = true
		var req CreateAcknowledgementRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if req.Comment != "Acknowledged by operator" {
			t.Errorf("Comment = %q, want %q", req.Comment, "Acknowledged by operator")
		}
		if !req.IsSticky {
			t.Error("IsSticky should be true")
		}
		if !req.IsPersistentComment {
			t.Error("IsPersistentComment should be true")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.Acknowledgements.CreateForHost(t.Context(), 10, &CreateAcknowledgementRequest{
		Comment:             "Acknowledged by operator",
		IsSticky:            true,
		IsPersistentComment: true,
	})
	if err != nil {
		t.Fatalf("CreateForHost: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestAcknowledgementService_CreateForService(t *testing.T) {
	mux, c := newTestMux(t)

	var called bool
	mux.HandleFunc("POST /centreon/api/latest/monitoring/hosts/10/services/5/acknowledgements", func(w http.ResponseWriter, r *http.Request) {
		called = true
		var req CreateAcknowledgementRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if req.Comment != "Service acknowledged" {
			t.Errorf("Comment = %q, want %q", req.Comment, "Service acknowledged")
		}
		if !req.IsNotifyContacts {
			t.Error("IsNotifyContacts should be true")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.Acknowledgements.CreateForService(t.Context(), 10, 5, &CreateAcknowledgementRequest{
		Comment:          "Service acknowledged",
		IsNotifyContacts: true,
	})
	if err != nil {
		t.Fatalf("CreateForService: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestAcknowledgementService_CreateForHost_WithServices(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("POST /centreon/api/latest/monitoring/hosts/10/acknowledgements", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["with_services"] != true {
			t.Errorf("with_services = %v, want true", body["with_services"])
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.Acknowledgements.CreateForHost(t.Context(), 10, &CreateAcknowledgementRequest{
		Comment:      "test ack",
		WithServices: true,
	})
	if err != nil {
		t.Fatalf("CreateForHost: %v", err)
	}
}

func TestAcknowledgementService_CancelForHost(t *testing.T) {
	mux, c := newTestMux(t)

	var called bool
	mux.HandleFunc("DELETE /centreon/api/latest/monitoring/hosts/10/acknowledgements", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	if err := c.Acknowledgements.CancelForHost(t.Context(), 10); err != nil {
		t.Fatalf("CancelForHost: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestAcknowledgementService_CancelForService(t *testing.T) {
	mux, c := newTestMux(t)

	var called bool
	mux.HandleFunc("DELETE /centreon/api/latest/monitoring/hosts/10/services/5/acknowledgements", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	if err := c.Acknowledgements.CancelForService(t.Context(), 10, 5); err != nil {
		t.Fatalf("CancelForService: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}
