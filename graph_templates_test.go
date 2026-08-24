package centreon

import (
	"errors"
	"net/http"
	"testing"
)

func TestGraphTemplateService_List(t *testing.T) {
	mux, c := newTestMux(t)
	// Two elements: the first has upper_limit:null (must decode to a nil
	// *float64, distinct from 0), the second has a real 200.5. Distinct base
	// (1000 vs 1024) and asymmetric is_default_centreon_template vs is_default so
	// a retyped or swapped tag fails. Retyping UpperLimit to float64 would decode
	// null to 0 and fail the nil assertion below.
	mux.HandleFunc("GET /centreon/api/latest/configuration/graphs/templates", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{
					"id": 1, "name": "Default_Graph", "vertical_axis_label": "Value",
					"width": 550, "height": 140,
					"grid": map[string]any{"lower_limit": 0, "upper_limit": nil, "is_upper_limit_sized_to_max": false},
					"base": 1000, "is_graph_scaled": true,
					"is_default_centreon_template": true, "is_default": false,
				},
				{
					"id": 3, "name": "Storage", "vertical_axis_label": "Storage",
					"width": 400, "height": 200,
					"grid": map[string]any{"lower_limit": 10, "upper_limit": 200.5, "is_upper_limit_sized_to_max": true},
					"base": 1024, "is_graph_scaled": false,
					"is_default_centreon_template": false, "is_default": true,
				},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 2},
		})
	})

	resp, err := c.GraphTemplates.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}

	g0 := resp.Result[0]
	wantInt(t, "g0.ID", g0.ID, 1)
	wantStr(t, "g0.Name", g0.Name, "Default_Graph")
	wantStr(t, "g0.VerticalAxisLabel", g0.VerticalAxisLabel, "Value")
	wantInt(t, "g0.Width", g0.Width, 550)
	wantInt(t, "g0.Height", g0.Height, 140)
	wantInt(t, "g0.Base", g0.Base, 1000)
	wantBool(t, "g0.IsGraphScaled", g0.IsGraphScaled, true)
	wantBool(t, "g0.IsDefaultCentreonTemplate", g0.IsDefaultCentreonTemplate, true)
	wantBool(t, "g0.IsDefault", g0.IsDefault, false)
	if g0.Grid.LowerLimit != 0 {
		t.Errorf("g0.Grid.LowerLimit = %v, want 0", g0.Grid.LowerLimit)
	}
	if g0.Grid.UpperLimit != nil {
		t.Errorf("g0.Grid.UpperLimit = %v, want nil (null on wire)", *g0.Grid.UpperLimit)
	}
	wantBool(t, "g0.Grid.IsUpperLimitSizedToMax", g0.Grid.IsUpperLimitSizedToMax, false)

	g1 := resp.Result[1]
	wantInt(t, "g1.Base", g1.Base, 1024)
	wantBool(t, "g1.IsDefault", g1.IsDefault, true)
	if g1.Grid.UpperLimit == nil {
		t.Fatalf("g1.Grid.UpperLimit = nil, want 200.5")
	}
	if *g1.Grid.UpperLimit != 200.5 {
		t.Errorf("g1.Grid.UpperLimit = %v, want 200.5", *g1.Grid.UpperLimit)
	}
	if g1.Grid.LowerLimit != 10 {
		t.Errorf("g1.Grid.LowerLimit = %v, want 10", g1.Grid.LowerLimit)
	}
}

func TestGraphTemplateService_All(t *testing.T) {
	mux, c := newTestMux(t)
	// Single page (total 2, default limit): All must yield both elements once.
	mux.HandleFunc("GET /centreon/api/latest/configuration/graphs/templates", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{"id": 1, "name": "Default_Graph", "grid": map[string]any{"upper_limit": nil}},
				{"id": 3, "name": "Storage", "grid": map[string]any{"upper_limit": 200.5}},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 2},
		})
	})

	var names []string
	for gt, err := range c.GraphTemplates.All(t.Context()) {
		if err != nil {
			t.Fatalf("All: %v", err)
		}
		names = append(names, gt.Name)
	}
	wantStrSlice(t, "All names", names, []string{"Default_Graph", "Storage"})
}

func TestGraphTemplateService_List_Error(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("GET /centreon/api/latest/configuration/graphs/templates", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "boom"})
	})

	resp, err := c.GraphTemplates.List(t.Context())
	if err == nil {
		t.Fatal("expected an error on HTTP 500, got nil")
	}
	if _, ok := errors.AsType[*APIError](err); !ok {
		t.Errorf("expected *APIError, got %T", err)
	}
	// resp is non-nil (the service returns &resp, err) but must carry no results.
	if resp != nil && len(resp.Result) != 0 {
		t.Errorf("Result = %v, want empty on error", resp.Result)
	}
}
