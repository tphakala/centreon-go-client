package centreon

import (
	"net/http"
	"testing"
)

func TestTimePeriodService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/timeperiods", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{
					"id": 1, "name": "24x7", "alias": "Always",
					"days":       []any{},
					"templates":  []any{},
					"exceptions": []any{},
					"in_period":  true,
				},
				{
					"id": 2, "name": "workhours", "alias": "Work Hours",
					"days":       []any{},
					"templates":  []any{},
					"exceptions": []any{},
					"in_period":  false,
				},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 2},
		})
	})

	resp, err := c.TimePeriods.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}
	if resp.Result[0].Name != "24x7" {
		t.Errorf("Result[0].Name = %q, want %q", resp.Result[0].Name, "24x7")
	}
	if resp.Result[1].Name != "workhours" {
		t.Errorf("Result[1].Name = %q, want %q", resp.Result[1].Name, "workhours")
	}
	if resp.Result[0].InPeriod != true {
		t.Errorf("Result[0].InPeriod = %v, want true", resp.Result[0].InPeriod)
	}
}

func TestTimePeriodService_Get(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/timeperiods/1", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"id":    1,
			"name":  "24x7",
			"alias": "24_Hours_A_Day,_7_Days_A_Week",
			"days": []map[string]any{
				{"day": 1, "time_range": "00:00-24:00"},
				{"day": 2, "time_range": "00:00-24:00"},
				{"day": 3, "time_range": "00:00-24:00"},
				{"day": 4, "time_range": "00:00-24:00"},
				{"day": 5, "time_range": "00:00-24:00"},
				{"day": 6, "time_range": "00:00-24:00"},
				{"day": 7, "time_range": "00:00-24:00"},
			},
			"templates":  []any{},
			"exceptions": []any{},
			"in_period":  true,
		})
	})

	tp, err := c.TimePeriods.Get(t.Context(), 1)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if tp.ID != 1 {
		t.Errorf("ID = %d, want 1", tp.ID)
	}
	if tp.Name != "24x7" {
		t.Errorf("Name = %q, want %q", tp.Name, "24x7")
	}
	if len(tp.Days) != 7 {
		t.Fatalf("len(Days) = %d, want 7", len(tp.Days))
	}
	if tp.Days[0].Day != 1 {
		t.Errorf("Days[0].Day = %d, want 1", tp.Days[0].Day)
	}
	if tp.Days[0].TimeRange != "00:00-24:00" {
		t.Errorf("Days[0].TimeRange = %q, want %q", tp.Days[0].TimeRange, "00:00-24:00")
	}
	if tp.Days[6].Day != 7 {
		t.Errorf("Days[6].Day = %d, want 7", tp.Days[6].Day)
	}
	if !tp.InPeriod {
		t.Errorf("InPeriod = false, want true")
	}
}

func TestTimePeriodService_Get_TypedExceptions(t *testing.T) {
	mux, c := newTestMux(t)

	// The live 25.10.16 wire shape for an exception: a server-assigned id plus
	// day_range and time_range. Pins that the typed TimePeriodException decodes
	// all three, including the read-only id.
	mux.HandleFunc("GET /centreon/api/latest/configuration/timeperiods/2", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"id": 2, "name": "holidays", "alias": "Holidays",
			"days":      []any{},
			"templates": []any{},
			"exceptions": []map[string]any{
				{"id": 2, "day_range": "january 1", "time_range": "00:00-00:00"},
			},
			"in_period": false,
		})
	})

	tp, err := c.TimePeriods.Get(t.Context(), 2)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(tp.Exceptions) != 1 {
		t.Fatalf("len(Exceptions) = %d, want 1", len(tp.Exceptions))
	}
	wantInt(t, "Exceptions[0].ID", tp.Exceptions[0].ID, 2)
	wantStr(t, "Exceptions[0].DayRange", tp.Exceptions[0].DayRange, "january 1")
	wantStr(t, "Exceptions[0].TimeRange", tp.Exceptions[0].TimeRange, "00:00-00:00")
}

func TestTimePeriodService_Create(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("POST /centreon/api/latest/configuration/timeperiods", func(w http.ResponseWriter, r *http.Request) {
		body := decodeBody(t, r)
		if body["name"] != "business-hours" {
			t.Errorf("name = %q, want %q", body["name"], "business-hours")
		}
		if body["alias"] != "Business Hours" {
			t.Errorf("alias = %q, want %q", body["alias"], "Business Hours")
		}
		days, ok := body["days"].([]any)
		if !ok {
			t.Fatalf("days must be an array, got %T (%v)", body["days"], body["days"])
		}
		if len(days) != 1 {
			t.Fatalf("len(days) = %d, want 1", len(days))
		}
		templates, ok := body["templates"].([]any)
		if !ok {
			t.Fatalf("templates must be an array, got %T (%v)", body["templates"], body["templates"])
		}
		if len(templates) != 0 {
			t.Errorf("len(templates) = %d, want 0", len(templates))
		}
		writeJSON(w, http.StatusCreated, map[string]int{"id": 5})
	})

	id, err := c.TimePeriods.Create(t.Context(), &CreateTimePeriodRequest{
		Name:  "business-hours",
		Alias: "Business Hours",
		Days: []TimePeriodDay{
			{Day: 1, TimeRange: "08:00-18:00"},
		},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 5 {
		t.Errorf("id = %d, want 5", id)
	}
}

func TestTimePeriodService_Create_NilSlices(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("POST /centreon/api/latest/configuration/timeperiods", func(w http.ResponseWriter, r *http.Request) {
		body := decodeBody(t, r)
		// API rejects null: both must be arrays even when the caller omits them.
		if _, ok := body["days"].([]any); !ok {
			t.Errorf("days must be an array, got %T (%v)", body["days"], body["days"])
		}
		if _, ok := body["templates"].([]any); !ok {
			t.Errorf("templates must be an array, got %T (%v)", body["templates"], body["templates"])
		}
		writeJSON(w, http.StatusCreated, map[string]int{"id": 6})
	})

	_, err := c.TimePeriods.Create(t.Context(), &CreateTimePeriodRequest{
		Name:  "minimal",
		Alias: "Minimal",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
}

func TestTimePeriodService_Update(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("PUT /centreon/api/latest/configuration/timeperiods/5", func(w http.ResponseWriter, r *http.Request) {
		body := decodeBody(t, r)
		if body["name"] != "updated-hours" {
			t.Errorf("name = %q, want %q", body["name"], "updated-hours")
		}
		if body["alias"] != "Updated Hours" {
			t.Errorf("alias = %q, want %q", body["alias"], "Updated Hours")
		}
		if _, ok := body["days"].([]any); !ok {
			t.Errorf("days must be an array, got %T (%v)", body["days"], body["days"])
		}
		if _, ok := body["templates"].([]any); !ok {
			t.Errorf("templates must be an array, got %T (%v)", body["templates"], body["templates"])
		}
		// On 25.10.x the API rejects a missing exceptions field on update; a nil
		// Exceptions slice must be normalized to an empty array, not null.
		if _, ok := body["exceptions"].([]any); !ok {
			t.Errorf("exceptions must be an array, got %T (%v)", body["exceptions"], body["exceptions"])
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.TimePeriods.Update(t.Context(), 5, &UpdateTimePeriodRequest{
		Name:  "updated-hours",
		Alias: "Updated Hours",
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestTimePeriodService_Update_TypedExceptions(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("PUT /centreon/api/latest/configuration/timeperiods/5", func(w http.ResponseWriter, r *http.Request) {
		body := decodeBody(t, r)
		exc, ok := body["exceptions"].([]any)
		if !ok {
			t.Fatalf("exceptions must be an array, got %T (%v)", body["exceptions"], body["exceptions"])
		}
		if len(exc) != 1 {
			t.Fatalf("len(exceptions) = %d, want 1", len(exc))
		}
		e0, ok := exc[0].(map[string]any)
		if !ok {
			t.Fatalf("exceptions[0] must be an object, got %T", exc[0])
		}
		if e0["day_range"] != "january 1" {
			t.Errorf("exceptions[0].day_range = %v, want %q", e0["day_range"], "january 1")
		}
		if e0["time_range"] != "00:00-00:00" {
			t.Errorf("exceptions[0].time_range = %v, want %q", e0["time_range"], "00:00-00:00")
		}
		// The read-only id must never be sent on a write. TimePeriodExceptionRequest
		// has no id field, so the key must be absent; this pins that against a
		// future regression that adds one.
		if _, present := e0["id"]; present {
			t.Errorf("exceptions[0] must not contain an id key, got %v", e0)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.TimePeriods.Update(t.Context(), 5, &UpdateTimePeriodRequest{
		Name:  "holidays",
		Alias: "Holidays",
		Exceptions: []TimePeriodExceptionRequest{
			{DayRange: "january 1", TimeRange: "00:00-00:00"},
		},
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
}

// TestTimePeriodService_UpdateDoesNotMutateRequest pins the guarantee that
// Update normalizes nil slices on a copy, leaving the caller's request struct
// untouched. Without the copy (the pre-fix form that mutated req directly) the
// caller's nil slices would come back as empty slices.
func TestTimePeriodService_UpdateDoesNotMutateRequest(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("PUT /centreon/api/latest/configuration/timeperiods/5", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	req := &UpdateTimePeriodRequest{Name: "x", Alias: "y"}
	if err := c.TimePeriods.Update(t.Context(), 5, req); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if req.Days != nil {
		t.Errorf("Update mutated caller's req.Days to %v, want nil", req.Days)
	}
	if req.Templates != nil {
		t.Errorf("Update mutated caller's req.Templates to %v, want nil", req.Templates)
	}
	if req.Exceptions != nil {
		t.Errorf("Update mutated caller's req.Exceptions to %v, want nil", req.Exceptions)
	}
}

// TestTimePeriodService_CreateDoesNotMutateRequest is the Create-side sibling of
// TestTimePeriodService_UpdateDoesNotMutateRequest.
func TestTimePeriodService_CreateDoesNotMutateRequest(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("POST /centreon/api/latest/configuration/timeperiods", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusCreated, map[string]int{"id": 8})
	})

	req := &CreateTimePeriodRequest{Name: "x", Alias: "y"}
	if _, err := c.TimePeriods.Create(t.Context(), req); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if req.Days != nil {
		t.Errorf("Create mutated caller's req.Days to %v, want nil", req.Days)
	}
	if req.Templates != nil {
		t.Errorf("Create mutated caller's req.Templates to %v, want nil", req.Templates)
	}
}

func TestTimePeriodService_Delete(t *testing.T) {
	mux, c := newTestMux(t)

	var called bool
	mux.HandleFunc("DELETE /centreon/api/latest/configuration/timeperiods/5", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.TimePeriods.Delete(t.Context(), 5)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}
