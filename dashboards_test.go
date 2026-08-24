package centreon

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

// dashboardListElement is the captured LIST element shape (25.10.16): it omits
// panels and refresh but carries shares, thumbnail, and is_favorite.
func dashboardListElement() map[string]any {
	return map[string]any{
		"id": 2, "name": "probe", "description": "d",
		"created_by": map[string]any{"id": 1, "name": "Admin Centreon"},
		"updated_by": map[string]any{"id": 1, "name": "Admin Centreon"},
		"created_at": "2026-08-24T12:17:18+00:00",
		"updated_at": "2026-08-24T12:17:18+00:00",
		"own_role":   "editor",
		"shares": map[string]any{
			"contacts": []map[string]any{
				{"id": 1, "name": "Admin Centreon", "email": "admin@centreon.local", "role": "editor"},
			},
			"contact_groups": []any{},
		},
		"thumbnail": nil, "is_favorite": false,
	}
}

func TestDashboardService_List(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("GET /centreon/api/latest/configuration/dashboards", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{dashboardListElement()},
			"meta":   map[string]any{"page": 1, "limit": 10, "total": 1},
		})
	})

	resp, err := c.Dashboards.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 1 {
		t.Fatalf("len(Result) = %d, want 1", len(resp.Result))
	}
	d := resp.Result[0]
	wantInt(t, "ID", d.ID, 2)
	wantStr(t, "Name", d.Name, "probe")
	wantStr(t, "OwnRole", d.OwnRole, "editor")
	wantInt(t, "CreatedBy.ID", d.CreatedBy.ID, 1)
	wantBool(t, "IsFavorite", d.IsFavorite, false)
	// List omits panels and refresh.
	if d.Panels != nil {
		t.Errorf("Panels = %v, want nil (list omits panels)", d.Panels)
	}
	if d.Refresh != nil {
		t.Errorf("Refresh = %v, want nil (list omits refresh)", d.Refresh)
	}
	// Shares present; nested contact fully decoded.
	if d.Shares == nil {
		t.Fatal("Shares = nil, want present in list")
	}
	if len(d.Shares.Contacts) != 1 {
		t.Fatalf("len(Shares.Contacts) = %d, want 1", len(d.Shares.Contacts))
	}
	wantStr(t, "Shares.Contacts[0].Email", d.Shares.Contacts[0].Email, "admin@centreon.local")
	wantStr(t, "Shares.Contacts[0].Role", d.Shares.Contacts[0].Role, "editor")
	if d.Thumbnail != nil {
		t.Errorf("Thumbnail = %v, want nil for JSON null", d.Thumbnail)
	}
}

func TestDashboardService_Get(t *testing.T) {
	tests := []struct {
		name        string
		interval    any
		wantHasIntv bool
		wantIntv    int
	}{
		{"global interval null", nil, false, 0},
		{"manual interval 30", 30, true, 30},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mux, c := newTestMux(t)
			mux.HandleFunc("GET /centreon/api/latest/configuration/dashboards/1", func(w http.ResponseWriter, _ *http.Request) {
				writeJSON(w, http.StatusOK, map[string]any{
					"id": 1, "name": "probe", "description": "d",
					"created_by": map[string]any{"id": 1, "name": "Admin Centreon"},
					"updated_by": map[string]any{"id": 1, "name": "Admin Centreon"},
					"created_at": "2026-08-24T12:17:16+00:00",
					"updated_at": "2026-08-24T12:17:16+00:00",
					"own_role":   "editor",
					"panels": []map[string]any{
						{
							"id": 1, "name": "p1",
							"layout":          map[string]any{"x": 0, "y": 0, "width": 6, "height": 4, "min_width": 2, "min_height": 2},
							"widget_type":     "centreon-widget-clock",
							"widget_settings": map[string]any{"foo": "bar"},
						},
					},
					"refresh":   map[string]any{"type": "global", "interval": tc.interval},
					"shares":    map[string]any{"contacts": []any{}, "contact_groups": []any{}},
					"thumbnail": nil, "is_favorite": false,
				})
			})

			d, err := c.Dashboards.Get(t.Context(), 1)
			if err != nil {
				t.Fatalf("Get: %v", err)
			}
			if len(d.Panels) != 1 {
				t.Fatalf("len(Panels) = %d, want 1", len(d.Panels))
			}
			p := d.Panels[0]
			wantInt(t, "Panels[0].ID", p.ID, 1)
			wantStr(t, "Panels[0].WidgetType", p.WidgetType, "centreon-widget-clock")
			// Layout: all six ints pinned (a mistagged field goes red here).
			wantInt(t, "Layout.X", p.Layout.X, 0)
			wantInt(t, "Layout.Y", p.Layout.Y, 0)
			wantInt(t, "Layout.Width", p.Layout.Width, 6)
			wantInt(t, "Layout.Height", p.Layout.Height, 4)
			wantInt(t, "Layout.MinWidth", p.Layout.MinWidth, 2)
			wantInt(t, "Layout.MinHeight", p.Layout.MinHeight, 2)
			// widget_settings round-trips as raw JSON.
			var ws map[string]string
			if err := json.Unmarshal(p.WidgetSettings, &ws); err != nil {
				t.Fatalf("unmarshal WidgetSettings: %v", err)
			}
			wantStr(t, "WidgetSettings[foo]", ws["foo"], "bar")

			if d.Refresh == nil {
				t.Fatal("Refresh = nil, want present in detail")
			}
			wantStr(t, "Refresh.Type", d.Refresh.Type, "global")
			if tc.wantHasIntv {
				wantIntPtr(t, "Refresh.Interval", d.Refresh.Interval, tc.wantIntv)
			} else {
				wantNilIntPtr(t, "Refresh.Interval", d.Refresh.Interval)
			}
		})
	}
}

func TestDashboardService_Create(t *testing.T) {
	mux, c := newTestMux(t)
	var gotBody map[string]any
	mux.HandleFunc("POST /centreon/api/latest/configuration/dashboards", func(w http.ResponseWriter, r *http.Request) {
		gotBody = decodeBody(t, r)
		writeJSON(w, http.StatusCreated, map[string]any{
			"id": 5, "name": "d", "description": "d",
			"created_by": map[string]any{"id": 1, "name": "Admin Centreon"},
			"updated_by": map[string]any{"id": 1, "name": "Admin Centreon"},
			"created_at": "2026-08-24T12:17:47+00:00",
			"updated_at": "2026-08-24T12:17:47+00:00",
			"own_role":   "editor",
			"panels":     []any{},
			"refresh":    map[string]any{"type": "global", "interval": nil},
		})
	})

	req := &CreateDashboardRequest{
		Name:        "d",
		Description: "desc",
		Panels: []DashboardPanelRequest{{
			Name:           "p1",
			Layout:         PanelLayout{X: 0, Y: 0, Width: 6, Height: 4, MinWidth: 2, MinHeight: 2},
			WidgetType:     "centreon-widget-clock",
			WidgetSettings: json.RawMessage(`{"foo":"bar"}`),
		}},
		Refresh: DashboardRefresh{Type: "global"},
	}
	id, err := c.Dashboards.Create(t.Context(), req)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	wantInt(t, "created id", id, 5)

	// Request body: exact top-level key set, and refresh is an object.
	wantJSONKeys(t, "create body", gotBody, "name", "description", "panels", "refresh")
	if _, ok := gotBody["refresh"].(map[string]any); !ok {
		t.Errorf("body.refresh = %v, want an object", gotBody["refresh"])
	}
	// Panel element: {name,layout,widget_type,widget_settings} and NO id.
	panels, ok := gotBody["panels"].([]any)
	if !ok || len(panels) != 1 {
		t.Fatalf("body.panels = %v, want 1-element array", gotBody["panels"])
	}
	panel, ok := panels[0].(map[string]any)
	if !ok {
		t.Fatalf("panels[0] is not an object: %v", panels[0])
	}
	wantJSONKeys(t, "create panel", panel, "name", "layout", "widget_type", "widget_settings")
	// On CREATE, widget_settings is a JSON OBJECT (unlike UPDATE, where it is a
	// JSON-encoded string). Pin the object form so a regression to the string form
	// is caught here rather than only by a live 500.
	if _, ok := panel["widget_settings"].(map[string]any); !ok {
		t.Errorf("create widget_settings = %v (type %T), want a JSON object", panel["widget_settings"], panel["widget_settings"])
	}
}

// TestDashboardService_Create_EmptyWidgetSettings pins that a nil panel
// WidgetSettings is normalized to the empty object {} on create (not null, which
// the API rejects), matching Update. Sabotage: remove the normalizeCreatePanels
// widget_settings default and this goes red.
func TestDashboardService_Create_EmptyWidgetSettings(t *testing.T) {
	mux, c := newTestMux(t)
	var gotBody map[string]any
	mux.HandleFunc("POST /centreon/api/latest/configuration/dashboards", func(w http.ResponseWriter, r *http.Request) {
		gotBody = decodeBody(t, r)
		writeJSON(w, http.StatusCreated, map[string]any{"id": 7})
	})

	req := &CreateDashboardRequest{
		Name: "d", Description: "x",
		Panels: []DashboardPanelRequest{{
			Name:       "p1",
			Layout:     PanelLayout{Width: 6, Height: 4, MinWidth: 2, MinHeight: 2},
			WidgetType: "centreon-widget-clock",
		}},
		Refresh: DashboardRefresh{Type: "global"},
	}
	if _, err := c.Dashboards.Create(t.Context(), req); err != nil {
		t.Fatalf("Create: %v", err)
	}
	panels, ok := gotBody["panels"].([]any)
	if !ok || len(panels) != 1 {
		t.Fatalf("body.panels = %v, want 1-element array", gotBody["panels"])
	}
	panel, ok := panels[0].(map[string]any)
	if !ok {
		t.Fatalf("panels[0] is not an object: %v", panels[0])
	}
	ws, ok := panel["widget_settings"].(map[string]any)
	if !ok || len(ws) != 0 {
		t.Errorf("create empty widget_settings = %v (type %T), want an empty object {}", panel["widget_settings"], panel["widget_settings"])
	}
	// Caller's request struct must not be mutated by the normalization.
	if req.Panels[0].WidgetSettings != nil {
		t.Errorf("caller req.Panels[0].WidgetSettings mutated to %s, want still nil", req.Panels[0].WidgetSettings)
	}
}

// TestDashboardService_Create_Error pins that a server error on create is
// surfaced (not swallowed) and returns a zero id.
func TestDashboardService_Create_Error(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("POST /centreon/api/latest/configuration/dashboards", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "boom"})
	})

	id, err := c.Dashboards.Create(t.Context(), &CreateDashboardRequest{Name: "d", Refresh: DashboardRefresh{Type: "global"}})
	if err == nil {
		t.Fatal("expected an error on HTTP 500, got nil")
	}
	if id != 0 {
		t.Errorf("id = %d, want 0 on error", id)
	}
	if _, ok := errors.AsType[*APIError](err); !ok {
		t.Errorf("expected *APIError, got %T", err)
	}
}

// TestDashboardService_Create_NilPanels pins that a nil Panels serializes as an
// empty array (not null), which the API requires. Sabotage: make
// normalizeCreatePanels return the input slice unchanged and this goes red.
func TestDashboardService_Create_NilPanels(t *testing.T) {
	mux, c := newTestMux(t)
	var gotBody map[string]any
	mux.HandleFunc("POST /centreon/api/latest/configuration/dashboards", func(w http.ResponseWriter, r *http.Request) {
		gotBody = decodeBody(t, r)
		writeJSON(w, http.StatusCreated, map[string]any{"id": 6})
	})

	req := &CreateDashboardRequest{Name: "d", Description: "x", Refresh: DashboardRefresh{Type: "global"}}
	if _, err := c.Dashboards.Create(t.Context(), req); err != nil {
		t.Fatalf("Create: %v", err)
	}
	panels, ok := gotBody["panels"].([]any)
	if !ok {
		t.Fatalf("body.panels = %v (type %T), want an empty array, not null", gotBody["panels"], gotBody["panels"])
	}
	if len(panels) != 0 {
		t.Errorf("body.panels = %v, want empty", panels)
	}
	// The caller's request struct must be left unmodified (normalized on a copy).
	if req.Panels != nil {
		t.Errorf("caller req.Panels mutated to %v, want still nil", req.Panels)
	}
}

// TestDashboardService_Update pins that update uses POST on the per-id path, NOT
// PUT. A PUT handler is registered alongside and fails the test if Update ever
// switches verbs (PUT /configuration/dashboards/{id} is 405 on the real server).
func TestDashboardService_Update(t *testing.T) {
	mux, c := newTestMux(t)
	var gotBody map[string]any
	mux.HandleFunc("POST /centreon/api/latest/configuration/dashboards/5", func(w http.ResponseWriter, r *http.Request) {
		gotBody = decodeBody(t, r)
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("PUT /centreon/api/latest/configuration/dashboards/5", func(w http.ResponseWriter, _ *http.Request) {
		t.Errorf("Update used PUT, want POST on the per-id path")
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	req := &UpdateDashboardRequest{
		Name: "d2", Description: "upd",
		Panels: []DashboardPanelRequest{{
			Name:           "p1",
			Layout:         PanelLayout{X: 0, Y: 0, Width: 6, Height: 4, MinWidth: 2, MinHeight: 2},
			WidgetType:     "centreon-widget-clock",
			WidgetSettings: json.RawMessage(`{"foo":"bar"}`),
		}},
		Refresh: DashboardRefresh{Type: "manual", Interval: ptr(15)},
	}
	if err := c.Dashboards.Update(t.Context(), 5, req); err != nil {
		t.Fatalf("Update: %v", err)
	}
	wantJSONKeys(t, "update body", gotBody, "name", "description", "panels", "refresh")
	wantStr(t, "body.name", asStr(gotBody["name"]), "d2")

	panels, ok := gotBody["panels"].([]any)
	if !ok || len(panels) != 1 {
		t.Fatalf("body.panels = %v, want 1-element array", gotBody["panels"])
	}
	panel, ok := panels[0].(map[string]any)
	if !ok {
		t.Fatalf("panels[0] is not an object: %v", panels[0])
	}
	// widget_settings on the UPDATE route must be a JSON-encoded STRING, not an
	// object (sending an object is HTTP 500 on 25.10.16). decodeBody yields a Go
	// string whose content is the JSON text.
	ws, ok := panel["widget_settings"].(string)
	if !ok {
		t.Fatalf("panels[0].widget_settings = %v (type %T), want a JSON string", panel["widget_settings"], panel["widget_settings"])
	}
	wantStr(t, "widget_settings string", ws, `{"foo":"bar"}`)

	// The caller's request struct must not be mutated: Update builds a fresh
	// dashboardUpdateBody and its panel WidgetSettings stays the original object.
	if got := string(req.Panels[0].WidgetSettings); got != `{"foo":"bar"}` {
		t.Errorf("caller req.Panels[0].WidgetSettings mutated to %q, want %q", got, `{"foo":"bar"}`)
	}
}

// TestDashboardService_Update_EmptyWidgetSettings pins that a nil/empty
// WidgetSettings becomes the string "{}" (not "" which the server's json_decode
// rejects).
func TestDashboardService_Update_EmptyWidgetSettings(t *testing.T) {
	mux, c := newTestMux(t)
	var gotBody map[string]any
	mux.HandleFunc("POST /centreon/api/latest/configuration/dashboards/5", func(w http.ResponseWriter, r *http.Request) {
		gotBody = decodeBody(t, r)
		w.WriteHeader(http.StatusNoContent)
	})

	req := &UpdateDashboardRequest{
		Name: "d", Description: "x",
		Panels: []DashboardPanelRequest{{
			Name:       "p1",
			Layout:     PanelLayout{Width: 6, Height: 4, MinWidth: 2, MinHeight: 2},
			WidgetType: "centreon-widget-clock",
		}},
		Refresh: DashboardRefresh{Type: "global"},
	}
	if err := c.Dashboards.Update(t.Context(), 5, req); err != nil {
		t.Fatalf("Update: %v", err)
	}
	panels, ok := gotBody["panels"].([]any)
	if !ok || len(panels) != 1 {
		t.Fatalf("body.panels = %v, want 1-element array", gotBody["panels"])
	}
	panel, ok := panels[0].(map[string]any)
	if !ok {
		t.Fatalf("panels[0] is not an object: %v", panels[0])
	}
	wantStr(t, "empty widget_settings", asStr(panel["widget_settings"]), "{}")
}

func TestDashboardService_Delete(t *testing.T) {
	mux, c := newTestMux(t)
	hit := false
	mux.HandleFunc("DELETE /centreon/api/latest/configuration/dashboards/5", func(w http.ResponseWriter, _ *http.Request) {
		hit = true
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.Dashboards.Delete(t.Context(), 5); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !hit {
		t.Error("Delete did not hit DELETE /configuration/dashboards/5")
	}
}

func TestDashboardService_UpdateShares(t *testing.T) {
	mux, c := newTestMux(t)
	var gotBody map[string]any
	mux.HandleFunc("PUT /centreon/api/latest/configuration/dashboards/5/shares", func(w http.ResponseWriter, r *http.Request) {
		gotBody = decodeBody(t, r)
		w.WriteHeader(http.StatusNoContent)
	})

	req := &UpdateDashboardSharesRequest{
		Contacts: []DashboardShareContactRequest{{ID: 1, Role: "editor"}},
	}
	if err := c.Dashboards.UpdateShares(t.Context(), 5, req); err != nil {
		t.Fatalf("UpdateShares: %v", err)
	}
	// Body: {contacts, contact_groups}; contact_groups normalized to [].
	wantJSONKeys(t, "shares body", gotBody, "contacts", "contact_groups")
	if cg, ok := gotBody["contact_groups"].([]any); !ok || len(cg) != 0 {
		t.Errorf("body.contact_groups = %v, want empty array", gotBody["contact_groups"])
	}
	contacts, ok := gotBody["contacts"].([]any)
	if !ok || len(contacts) != 1 {
		t.Fatalf("body.contacts = %v, want 1-element array", gotBody["contacts"])
	}
	// Write-side contact carries only {id, role}: no name/email.
	wantJSONKeys(t, "share contact", contacts[0], "id", "role")
}

// TestDashboardService_UpdateShares_Empty pins that an empty request clears
// shares by sending both arrays as [] (not null).
func TestDashboardService_UpdateShares_Empty(t *testing.T) {
	mux, c := newTestMux(t)
	var gotBody map[string]any
	mux.HandleFunc("PUT /centreon/api/latest/configuration/dashboards/5/shares", func(w http.ResponseWriter, r *http.Request) {
		gotBody = decodeBody(t, r)
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.Dashboards.UpdateShares(t.Context(), 5, &UpdateDashboardSharesRequest{}); err != nil {
		t.Fatalf("UpdateShares: %v", err)
	}
	for _, k := range []string{"contacts", "contact_groups"} {
		arr, ok := gotBody[k].([]any)
		if !ok || len(arr) != 0 {
			t.Errorf("body.%s = %v (type %T), want empty array", k, gotBody[k], gotBody[k])
		}
	}
}

func TestDashboardService_All(t *testing.T) {
	mux, c := newTestMux(t)
	// Two pages of one element each (limit 1, total 2), keyed on ?page=.
	mux.HandleFunc("GET /centreon/api/latest/configuration/dashboards", func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		id, name := 1, "first"
		if page == "2" {
			id, name = 2, "second"
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{{"id": id, "name": name}},
			"meta":   map[string]any{"page": 1, "limit": 1, "total": 2},
		})
	})

	var names []string
	for d, err := range c.Dashboards.All(t.Context()) {
		if err != nil {
			t.Fatalf("All: %v", err)
		}
		names = append(names, d.Name)
	}
	wantStrSlice(t, "All names", names, []string{"first", "second"})
}

func TestDashboardService_Get_NotFound(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("GET /centreon/api/latest/configuration/dashboards/99", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "Dashboard not found"})
	})

	got, err := c.Dashboards.Get(t.Context(), 99)
	if err == nil {
		t.Fatal("expected an error on HTTP 404, got nil")
	}
	if got != nil {
		t.Errorf("result = %v, want nil on error", got)
	}
	apiErr, ok := errors.AsType[*APIError](err)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	// A resource 404 ("Dashboard not found") is not a routing 404.
	if apiErr.IsRouteNotFound() {
		t.Error("IsRouteNotFound() = true, want false for a resource 404")
	}
}

// TestDashboard_InferredShapesDecode guards the two field types that could not be
// observed populated on the live box: a populated thumbnail (always null there)
// and a contact_groups share entry (unreachable without a dashboard ACL). It
// feeds synthetic JSON so a wrong tag on either would fail here rather than
// silently at runtime.
func TestDashboard_InferredShapesDecode(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("GET /centreon/api/latest/configuration/dashboards/1", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"id": 1, "name": "d", "description": "",
			"created_by": map[string]any{"id": 1, "name": "a"},
			"updated_by": map[string]any{"id": 1, "name": "a"},
			"created_at": "2026-08-24T12:17:16+00:00",
			"updated_at": "2026-08-24T12:17:16+00:00",
			"own_role":   "editor",
			"panels":     []any{},
			"refresh":    map[string]any{"type": "global", "interval": nil},
			"shares": map[string]any{
				"contacts": []any{},
				"contact_groups": []map[string]any{
					{"id": 9, "name": "grp", "role": "viewer"},
				},
			},
			"thumbnail": map[string]any{"directory": "thumbnails", "name": "dash-1.png"},
		})
	})

	d, err := c.Dashboards.Get(t.Context(), 1)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if d.Thumbnail == nil {
		t.Fatal("Thumbnail = nil, want the populated raw JSON")
	}
	var thumb map[string]string
	if err := json.Unmarshal(*d.Thumbnail, &thumb); err != nil {
		t.Fatalf("unmarshal Thumbnail: %v", err)
	}
	wantStr(t, "Thumbnail[name]", thumb["name"], "dash-1.png")

	if d.Shares == nil || len(d.Shares.ContactGroups) != 1 {
		t.Fatalf("Shares.ContactGroups = %v, want 1 element", d.Shares)
	}
	g := d.Shares.ContactGroups[0]
	wantInt(t, "ContactGroups[0].ID", g.ID, 9)
	wantStr(t, "ContactGroups[0].Name", g.Name, "grp")
	wantStr(t, "ContactGroups[0].Role", g.Role, "viewer")
}

// test-local helpers

func ptr[T any](v T) *T { return &v }

func asStr(v any) string {
	s, _ := v.(string)
	return s
}
