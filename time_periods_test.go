package centreon

import (
	"encoding/json"
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

func TestTimePeriodService_Create(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("POST /centreon/api/latest/configuration/timeperiods", func(w http.ResponseWriter, r *http.Request) {
		var req CreateTimePeriodRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if req.Name != "business-hours" {
			t.Errorf("Name = %q, want %q", req.Name, "business-hours")
		}
		if req.Alias != "Business Hours" {
			t.Errorf("Alias = %q, want %q", req.Alias, "Business Hours")
		}
		if len(req.Days) != 1 {
			t.Fatalf("len(Days) = %d, want 1", len(req.Days))
		}
		if req.Days[0].Day != 1 {
			t.Errorf("Days[0].Day = %d, want 1", req.Days[0].Day)
		}
		if req.Days[0].TimeRange != "08:00-18:00" {
			t.Errorf("Days[0].TimeRange = %q, want %q", req.Days[0].TimeRange, "08:00-18:00")
		}
		writeJSON(w, http.StatusCreated, map[string]int{"id": 5})
	})

	id, err := c.TimePeriods.Create(t.Context(), CreateTimePeriodRequest{
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

func TestTimePeriodService_Update(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("PUT /centreon/api/latest/configuration/timeperiods/5", func(w http.ResponseWriter, r *http.Request) {
		var req UpdateTimePeriodRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if req.Name != "updated-hours" {
			t.Errorf("Name = %q, want %q", req.Name, "updated-hours")
		}
		if req.Alias != "Updated Hours" {
			t.Errorf("Alias = %q, want %q", req.Alias, "Updated Hours")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.TimePeriods.Update(t.Context(), 5, UpdateTimePeriodRequest{
		Name:  "updated-hours",
		Alias: "Updated Hours",
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
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
