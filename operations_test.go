package centreon

import (
	"encoding/json"
	"net/http"
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

	start := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

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
		if _, hasDate := res["date"]; !hasDate {
			t.Error("date is missing, want timestamp")
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
