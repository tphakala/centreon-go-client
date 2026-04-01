package centreon

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestTimePeriodService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/time-periods", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[TimePeriod]{
			Result: []TimePeriod{
				{ID: 1, Name: "24x7", Alias: "Always"},
				{ID: 2, Name: "workhours", Alias: "Work Hours"},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 2},
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
}

func TestTimePeriodService_Get(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/time-periods/1", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, TimePeriod{
			ID:    1,
			Name:  "24x7",
			Alias: "Always",
			Days: []TimePeriodDay{
				{
					Day: "monday",
					TimeRanges: []TimeRange{
						{Start: "00:00", End: "24:00"},
					},
				},
			},
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
	if len(tp.Days) != 1 {
		t.Fatalf("len(Days) = %d, want 1", len(tp.Days))
	}
	if tp.Days[0].Day != "monday" {
		t.Errorf("Days[0].Day = %q, want %q", tp.Days[0].Day, "monday")
	}
	if len(tp.Days[0].TimeRanges) != 1 {
		t.Fatalf("len(Days[0].TimeRanges) = %d, want 1", len(tp.Days[0].TimeRanges))
	}
	if tp.Days[0].TimeRanges[0].Start != "00:00" {
		t.Errorf("TimeRanges[0].Start = %q, want %q", tp.Days[0].TimeRanges[0].Start, "00:00")
	}
}

func TestTimePeriodService_Create(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("POST /centreon/api/latest/configuration/time-periods", func(w http.ResponseWriter, r *http.Request) {
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
		if req.Days[0].Day != "monday" {
			t.Errorf("Days[0].Day = %q, want %q", req.Days[0].Day, "monday")
		}
		if len(req.Days[0].TimeRanges) != 1 {
			t.Fatalf("len(Days[0].TimeRanges) = %d, want 1", len(req.Days[0].TimeRanges))
		}
		if req.Days[0].TimeRanges[0].Start != "08:00" {
			t.Errorf("TimeRanges[0].Start = %q, want %q", req.Days[0].TimeRanges[0].Start, "08:00")
		}
		if req.Days[0].TimeRanges[0].End != "18:00" {
			t.Errorf("TimeRanges[0].End = %q, want %q", req.Days[0].TimeRanges[0].End, "18:00")
		}
		writeJSON(w, http.StatusCreated, map[string]int{"id": 5})
	})

	id, err := c.TimePeriods.Create(t.Context(), CreateTimePeriodRequest{
		Name:  "business-hours",
		Alias: "Business Hours",
		Days: []TimePeriodDay{
			{
				Day: "monday",
				TimeRanges: []TimeRange{
					{Start: "08:00", End: "18:00"},
				},
			},
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

	mux.HandleFunc("PUT /centreon/api/latest/configuration/time-periods/5", func(w http.ResponseWriter, r *http.Request) {
		var req UpdateTimePeriodRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if req.Name != "updated-hours" {
			t.Errorf("Name = %q, want %q", req.Name, "updated-hours")
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
	mux.HandleFunc("DELETE /centreon/api/latest/configuration/time-periods/5", func(w http.ResponseWriter, r *http.Request) {
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
