package centreon

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestAcknowledgementService_List(t *testing.T) {
	mux, c := newTestMux(t)

	entry := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	svcID := 5

	mux.HandleFunc("GET /centreon/api/latest/monitoring/acknowledgements", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[Acknowledgement]{
			Result: []Acknowledgement{
				{
					ID:                  1,
					HostID:              10,
					AuthorID:            1,
					AuthorName:          "admin",
					Comment:             "Host ack",
					IsSticky:            true,
					IsPersistentComment: true,
					IsNotifyContacts:    false,
					State:               1,
					EntryTime:           entry,
					PollerID:            1,
				},
				{
					ID:                  2,
					HostID:              10,
					ServiceID:           &svcID,
					AuthorID:            1,
					AuthorName:          "admin",
					Comment:             "Service ack",
					IsSticky:            false,
					IsPersistentComment: false,
					IsNotifyContacts:    true,
					State:               2,
					EntryTime:           entry,
					PollerID:            1,
				},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 2},
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
	if resp.Result[1].ServiceID == nil {
		t.Fatal("Result[1].ServiceID is nil, want non-nil")
	}
	if *resp.Result[1].ServiceID != 5 {
		t.Errorf("Result[1].ServiceID = %d, want 5", *resp.Result[1].ServiceID)
	}
}

func TestAcknowledgementService_Get(t *testing.T) {
	mux, c := newTestMux(t)

	entry := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)

	mux.HandleFunc("GET /centreon/api/latest/monitoring/acknowledgements/42", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, Acknowledgement{
			ID:                  42,
			HostID:              10,
			AuthorID:            1,
			AuthorName:          "admin",
			Comment:             "Acknowledged for investigation",
			IsSticky:            true,
			IsPersistentComment: true,
			IsNotifyContacts:    false,
			State:               1,
			EntryTime:           entry,
			PollerID:            1,
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
}

func TestAcknowledgementService_ListForHost(t *testing.T) {
	mux, c := newTestMux(t)

	entry := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)

	mux.HandleFunc("GET /centreon/api/latest/monitoring/hosts/10/acknowledgements", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[Acknowledgement]{
			Result: []Acknowledgement{
				{
					ID:         1,
					HostID:     10,
					AuthorID:   1,
					AuthorName: "admin",
					Comment:    "Under investigation",
					IsSticky:   true,
					State:      1,
					EntryTime:  entry,
					PollerID:   1,
				},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 1},
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

	err := c.Acknowledgements.CreateForHost(t.Context(), 10, CreateAcknowledgementRequest{
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

	err := c.Acknowledgements.CreateForService(t.Context(), 10, 5, CreateAcknowledgementRequest{
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
