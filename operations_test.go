package centreon

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

// decodeBody decodes a JSON request body into a map and returns it.
func decodeBody(t *testing.T, r *http.Request) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	return body
}

// resourceAt extracts resource at index i from the body's "resources" array.
func resourceAt(t *testing.T, body map[string]any, i int) map[string]any {
	t.Helper()
	resources, ok := body["resources"].([]any)
	if !ok || len(resources) <= i {
		t.Fatalf("resources = %v, want array with at least %d elements", body["resources"], i+1)
	}
	res, ok := resources[i].(map[string]any)
	if !ok {
		t.Fatalf("resources[%d] is not an object", i)
	}
	return res
}

// requireNullParent checks that a resource has "parent": null (present but nil).
func requireNullParent(t *testing.T, res map[string]any, label string) {
	t.Helper()
	if _, hasParent := res["parent"]; !hasParent {
		t.Errorf("%s.parent is missing, want explicit null", label)
	}
	if res["parent"] != nil {
		t.Errorf("%s.parent = %v, want null", label, res["parent"])
	}
}

func TestOperationsService_Acknowledge(t *testing.T) {
	mux, c := newTestMux(t)

	var called bool
	mux.HandleFunc("POST /centreon/api/latest/monitoring/resources/acknowledge", func(w http.ResponseWriter, r *http.Request) {
		called = true
		body := decodeBody(t, r)
		res := resourceAt(t, body, 0)
		if res["type"] != "host" {
			t.Errorf("resources[0].type = %v, want host", res["type"])
		}
		requireNullParent(t, res, "resources[0]")

		ack, ok := body["acknowledgement"].(map[string]any)
		if !ok {
			t.Fatalf("acknowledgement wrapper missing, got body: %v", body)
		}
		if ack["comment"] != "Acknowledged by operator" {
			t.Errorf("acknowledgement.comment = %v, want %q", ack["comment"], "Acknowledged by operator")
		}
		if ack["is_sticky"] != true {
			t.Error("acknowledgement.is_sticky should be true")
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

	// Sub-second precision so the whole-second normalization is observable on the
	// wire (this endpoint tolerates fractional, but the client normalizes anyway
	// to stay consistent with the stricter per-host/service downtime endpoints).
	start := time.Date(2024, 1, 15, 8, 0, 0, 500000000, time.UTC)
	end := time.Date(2024, 1, 15, 10, 0, 0, 500000000, time.UTC)

	var called bool
	mux.HandleFunc("POST /centreon/api/latest/monitoring/resources/downtime", func(w http.ResponseWriter, r *http.Request) {
		called = true
		body := decodeBody(t, r)
		for i := range 2 {
			res := resourceAt(t, body, i)
			requireNullParent(t, res, "resources[0]")
		}

		dt, ok := body["downtime"].(map[string]any)
		if !ok {
			t.Fatalf("downtime wrapper missing, got body: %v", body)
		}
		if dt["comment"] != "Maintenance window" {
			t.Errorf("downtime.comment = %v, want %q", dt["comment"], "Maintenance window")
		}
		if dt["is_fixed"] != true {
			t.Error("downtime.is_fixed should be true")
		}
		wantTimes := map[string]string{
			"start_time": "2024-01-15T08:00:00Z",
			"end_time":   "2024-01-15T10:00:00Z",
		}
		for k, want := range wantTimes {
			s, ok := dt[k].(string)
			if !ok {
				t.Errorf("downtime.%s = %v, want a whole-second timestamp string", k, dt[k])
				continue
			}
			if s != want {
				t.Errorf("downtime.%s = %q, want %q (truncated to whole seconds)", k, s, want)
			}
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
		body := decodeBody(t, r)
		res := resourceAt(t, body, 0)
		if res["type"] != "service" {
			t.Errorf("resources[0].type = %v, want service", res["type"])
		}
		parent, ok := res["parent"].(map[string]any)
		if !ok {
			t.Fatal("resources[0].parent is not an object")
		}
		if parent["id"] != float64(3) {
			t.Errorf("resources[0].parent.id = %v, want 3", parent["id"])
		}
		if _, hasType := parent["type"]; hasType {
			t.Error("resources[0].parent.type must not be present; API rejects it")
		}
		if _, hasParent := parent["parent"]; hasParent {
			t.Error("resources[0].parent.parent must not be present; API rejects it")
		}

		check, ok := body["check"].(map[string]any)
		if !ok {
			t.Fatal("check wrapper missing")
		}
		if check["is_forced"] != true {
			t.Errorf("check.is_forced = %v, want true", check["is_forced"])
		}

		w.WriteHeader(http.StatusNoContent)
	})

	err := c.Operations.Check(t.Context(), &CheckRequest{
		Resources: []ResourceRef{
			{Type: "service", ID: 7, Parent: &ParentRef{ID: 3}},
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
		body := decodeBody(t, r)
		res := resourceAt(t, body, 0)
		if res["type"] != "service" {
			t.Errorf("type = %v, want service", res["type"])
		}
		if res["output"] != "All systems nominal" {
			t.Errorf("output = %v, want %q", res["output"], "All systems nominal")
		}
		if res["performance_data"] != "rta=1ms" {
			t.Errorf("performance_data = %v, want %q", res["performance_data"], "rta=1ms")
		}
		parent, ok := res["parent"].(map[string]any)
		if !ok {
			t.Fatal("resources[0].parent is not an object")
		}
		if parent["id"] != float64(1) {
			t.Errorf("parent.id = %v, want 1", parent["id"])
		}
		if _, hasType := parent["type"]; hasType {
			t.Error("resources[0].parent.type must not be present; API rejects it")
		}

		w.WriteHeader(http.StatusNoContent)
	})

	err := c.Operations.Submit(t.Context(), &SubmitResultRequest{
		Resources: []SubmitResource{
			{
				Type:     "service",
				ID:       5,
				Parent:   &ParentRef{ID: 1},
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
		body := decodeBody(t, r)
		res := resourceAt(t, body, 0)
		if res["type"] != "host" {
			t.Errorf("type = %v, want host", res["type"])
		}
		if res["comment"] != "Under investigation" {
			t.Errorf("comment = %v, want %q", res["comment"], "Under investigation")
		}
		requireNullParent(t, res, "resources[0]")
		s, hasDate := res["date"].(string)
		if !hasDate {
			t.Error("date is missing, want timestamp")
		}
		// Comment stamps time.Now(), truncated to whole seconds.
		if strings.Contains(s, ".") {
			t.Errorf("date = %q, want whole seconds (no fractional)", s)
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

// All five resource-posting endpoints reject a null resources with HTTP 500
// ("[resources] NULL value found, but an array is required") and accept an empty
// array as a no-op, so a nil Resources must marshal to [] rather than null.
// Acknowledge/Downtime/Check/Submit normalize via nilToEmpty; Comment is already
// safe because it builds the slice with make(), and its test guards that a future
// refactor cannot reintroduce the null.

func TestOperationsService_AcknowledgeNormalizesNilResources(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("POST /centreon/api/latest/monitoring/resources/acknowledge", func(w http.ResponseWriter, r *http.Request) {
		var raw map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got := string(raw["resources"]); got != "[]" {
			t.Errorf("resources = %s, want [] (nil must normalize to an empty array, not null)", got)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	if err := c.Operations.Acknowledge(t.Context(), &AcknowledgeRequest{Comment: "x"}); err != nil {
		t.Fatalf("Acknowledge: %v", err)
	}
}

func TestOperationsService_DowntimeNormalizesNilResources(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("POST /centreon/api/latest/monitoring/resources/downtime", func(w http.ResponseWriter, r *http.Request) {
		var raw map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got := string(raw["resources"]); got != "[]" {
			t.Errorf("resources = %s, want [] (nil must normalize to an empty array, not null)", got)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	if err := c.Operations.Downtime(t.Context(), &DowntimeRequest{Comment: "x"}); err != nil {
		t.Fatalf("Downtime: %v", err)
	}
}

func TestOperationsService_CheckNormalizesNilResources(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("POST /centreon/api/latest/monitoring/resources/check", func(w http.ResponseWriter, r *http.Request) {
		var raw map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got := string(raw["resources"]); got != "[]" {
			t.Errorf("resources = %s, want [] (nil must normalize to an empty array, not null)", got)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	if err := c.Operations.Check(t.Context(), &CheckRequest{}); err != nil {
		t.Fatalf("Check: %v", err)
	}
}

func TestOperationsService_SubmitNormalizesNilResources(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("POST /centreon/api/latest/monitoring/resources/submit", func(w http.ResponseWriter, r *http.Request) {
		var raw map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got := string(raw["resources"]); got != "[]" {
			t.Errorf("resources = %s, want [] (nil must normalize to an empty array, not null)", got)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	req := &SubmitResultRequest{} // Resources is nil.
	if err := c.Operations.Submit(t.Context(), req); err != nil {
		t.Fatalf("Submit: %v", err)
	}
	// Submit normalizes on a copy, so the caller's request must be untouched.
	if req.Resources != nil {
		t.Errorf("Submit mutated caller's req.Resources = %v, want nil", req.Resources)
	}
}

func TestOperationsService_CommentNormalizesNilResources(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("POST /centreon/api/latest/monitoring/resources/comments", func(w http.ResponseWriter, r *http.Request) {
		var raw map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got := string(raw["resources"]); got != "[]" {
			t.Errorf("resources = %s, want [] (nil must normalize to an empty array, not null)", got)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	if err := c.Operations.Comment(t.Context(), &CommentRequest{Comment: "x"}); err != nil {
		t.Fatalf("Comment: %v", err)
	}
}
