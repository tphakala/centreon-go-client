package centreon

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"testing"
)

func TestServiceTemplateService_List_AllFields(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/services/templates", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{
					"id": 1, "name": "tpl-full", "is_locked": false,
					"check_command_args":                    []string{"-c", "check_http"},
					"host_templates":                        []int{87, 88, 1053},
					"service_template_id":                   8,
					"severity_id":                           7,
					"active_check_enabled":                  2,
					"passive_check_enabled":                 1,
					"volatility_enabled":                    2,
					"notification_enabled":                  2,
					"is_contact_additive_inheritance":       true,
					"is_contact_group_additive_inheritance": true,
					"notification_interval":                 10,
					"notification_timeperiod_id":            4,
					"notification_type":                     3,
					"first_notification_delay":              30,
					"recovery_notification_delay":           15,
					"acknowledgement_timeout":               60,
					"freshness_checked":                     1,
					"freshness_threshold":                   120,
					"flap_detection_enabled":                2,
					"low_flap_threshold":                    20,
					"high_flap_threshold":                   80,
					"event_handler_enabled":                 1,
					"event_handler_command_id":              9,
					"event_handler_command_args":            []string{"-w", "5"},
					"graph_template_id":                     12,
					"note_url":                              "http://note",
					"note":                                  "a note",
					"action_url":                            "http://action",
					"icon_id":                               11,
					"icon_alternative":                      "alt",
					"comment":                               "a comment",
				},
				{
					"id": 2, "name": "tpl-nulls", "is_locked": true,
					"host_templates":      []int{},
					"service_template_id": nil,
					"severity_id":         nil,
					"notification_type":   nil,
					"graph_template_id":   nil,
					"icon_id":             nil,
					// A JSON null for a plain-int toggle decodes to 0 with no
					// error (encoding/json treats null into a non-pointer as a
					// no-op), so it does not break the List decode.
					"active_check_enabled": nil,
					"volatility_enabled":   nil,
				},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 2},
		})
	})

	resp, err := c.ServiceTemplates.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}

	tpl := resp.Result[0]
	wantStrSlice(t, "CheckCommandArgs", tpl.CheckCommandArgs, []string{"-c", "check_http"})
	if !slices.Equal(tpl.HostTemplates, []int{87, 88, 1053}) {
		t.Errorf("HostTemplates = %v, want [87 88 1053]", tpl.HostTemplates)
	}
	wantIntPtr(t, "ServiceTemplateID", tpl.ServiceTemplateID, 8)
	wantIntPtr(t, "SeverityID", tpl.SeverityID, 7)
	wantInt(t, "ActiveCheckEnabled", tpl.ActiveCheckEnabled, 2)
	wantInt(t, "PassiveCheckEnabled", tpl.PassiveCheckEnabled, 1)
	wantInt(t, "VolatilityEnabled", tpl.VolatilityEnabled, 2)
	wantInt(t, "NotificationEnabled", tpl.NotificationEnabled, 2)
	wantBool(t, "IsContactAdditiveInheritance", tpl.IsContactAdditiveInheritance, true)
	wantBool(t, "IsContactGroupAdditiveInheritance", tpl.IsContactGroupAdditiveInheritance, true)
	wantIntPtr(t, "NotificationInterval", tpl.NotificationInterval, 10)
	wantIntPtr(t, "NotificationTimeperiodID", tpl.NotificationTimeperiodID, 4)
	wantIntPtr(t, "NotificationType", tpl.NotificationType, 3)
	wantIntPtr(t, "FirstNotificationDelay", tpl.FirstNotificationDelay, 30)
	wantIntPtr(t, "RecoveryNotificationDelay", tpl.RecoveryNotificationDelay, 15)
	wantIntPtr(t, "AcknowledgementTimeout", tpl.AcknowledgementTimeout, 60)
	wantInt(t, "FreshnessChecked", tpl.FreshnessChecked, 1)
	wantIntPtr(t, "FreshnessThreshold", tpl.FreshnessThreshold, 120)
	wantInt(t, "FlapDetectionEnabled", tpl.FlapDetectionEnabled, 2)
	wantIntPtr(t, "LowFlapThreshold", tpl.LowFlapThreshold, 20)
	wantIntPtr(t, "HighFlapThreshold", tpl.HighFlapThreshold, 80)
	wantInt(t, "EventHandlerEnabled", tpl.EventHandlerEnabled, 1)
	wantIntPtr(t, "EventHandlerCommandID", tpl.EventHandlerCommandID, 9)
	wantStrSlice(t, "EventHandlerCommandArgs", tpl.EventHandlerCommandArgs, []string{"-w", "5"})
	wantIntPtr(t, "GraphTemplateID", tpl.GraphTemplateID, 12)
	wantStr(t, "NoteURL", tpl.NoteURL, "http://note")
	wantStr(t, "Note", tpl.Note, "a note")
	wantStr(t, "ActionURL", tpl.ActionURL, "http://action")
	wantIntPtr(t, "IconID", tpl.IconID, 11)
	wantStr(t, "IconAlternative", tpl.IconAlternative, "alt")
	wantStr(t, "Comment", tpl.Comment, "a comment")

	nulls := resp.Result[1]
	wantNilIntPtr(t, "Result[1].ServiceTemplateID", nulls.ServiceTemplateID)
	wantNilIntPtr(t, "Result[1].SeverityID", nulls.SeverityID)
	wantNilIntPtr(t, "Result[1].NotificationType", nulls.NotificationType)
	wantNilIntPtr(t, "Result[1].GraphTemplateID", nulls.GraphTemplateID)
	wantNilIntPtr(t, "Result[1].IconID", nulls.IconID)
	wantInt(t, "Result[1].ActiveCheckEnabled", nulls.ActiveCheckEnabled, 0)
	wantInt(t, "Result[1].VolatilityEnabled", nulls.VolatilityEnabled, 0)
	if nulls.HostTemplates == nil || len(nulls.HostTemplates) != 0 {
		t.Errorf("Result[1].HostTemplates = %v, want non-nil empty slice", nulls.HostTemplates)
	}
}

// TestServiceTemplateService_List_HostTemplatesNullAndAbsent pins the two wire
// shapes the populated/empty fixtures do not cover: a JSON null value and an
// absent host_templates key must both decode to a nil slice without breaking
// the List decode. The bare integer-array shape is verified live against
// Centreon Web 24.10.29; these two cases are defensive.
func TestServiceTemplateService_List_HostTemplatesNullAndAbsent(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/services/templates", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{"id": 1, "name": "null-value", "host_templates": nil},
				{"id": 2, "name": "absent-key"},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 2},
		})
	})

	resp, err := c.ServiceTemplates.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}
	if got := resp.Result[0].HostTemplates; got != nil {
		t.Errorf("null-value HostTemplates = %v, want nil", got)
	}
	if got := resp.Result[1].HostTemplates; got != nil {
		t.Errorf("absent-key HostTemplates = %v, want nil", got)
	}
}

func TestServiceTemplateService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/services/templates", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{"id": 1, "name": "generic-service", "alias": "Generic", "check_command_id": nil, "check_timeperiod_id": 1, "max_check_attempts": nil, "normal_check_interval": nil, "retry_check_interval": nil, "is_locked": false},
				{"id": 2, "name": "ping-template", "alias": "Ping", "check_command_id": 3, "check_timeperiod_id": nil, "max_check_attempts": 5, "normal_check_interval": 10, "retry_check_interval": 2, "is_locked": true},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 2},
		})
	})

	resp, err := c.ServiceTemplates.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}

	tpl0 := resp.Result[0]
	if tpl0.Name != "generic-service" {
		t.Errorf("Result[0].Name = %q, want %q", tpl0.Name, "generic-service")
	}
	if tpl0.CheckCommandID != nil {
		t.Errorf("Result[0].CheckCommandID = %v, want nil", tpl0.CheckCommandID)
	}
	if tpl0.CheckTimeperiodID == nil || *tpl0.CheckTimeperiodID != 1 {
		t.Errorf("Result[0].CheckTimeperiodID = %v, want 1", tpl0.CheckTimeperiodID)
	}
	if tpl0.IsLocked {
		t.Error("Result[0].IsLocked = true, want false")
	}

	tpl1 := resp.Result[1]
	if tpl1.CheckCommandID == nil || *tpl1.CheckCommandID != 3 {
		t.Errorf("Result[1].CheckCommandID = %v, want 3", tpl1.CheckCommandID)
	}
	if tpl1.MaxCheckAttempts == nil || *tpl1.MaxCheckAttempts != 5 {
		t.Errorf("Result[1].MaxCheckAttempts = %v, want 5", tpl1.MaxCheckAttempts)
	}
	if !tpl1.IsLocked {
		t.Error("Result[1].IsLocked = false, want true")
	}
}

func TestServiceTemplateService_GetByID_Found(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/services/templates", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{"id": 42, "name": "generic-service", "alias": "Generic", "check_command_id": nil, "check_timeperiod_id": nil, "max_check_attempts": nil, "normal_check_interval": nil, "retry_check_interval": nil, "is_locked": false},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 1},
		})
	})

	tmpl, err := c.ServiceTemplates.GetByID(t.Context(), 42)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if tmpl.ID != 42 {
		t.Errorf("ID = %d, want 42", tmpl.ID)
	}
	if tmpl.Name != "generic-service" {
		t.Errorf("Name = %q, want %q", tmpl.Name, "generic-service")
	}
}

func TestServiceTemplateService_GetByID_NotFound(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/services/templates", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[ServiceTemplate]{
			Result: []ServiceTemplate{},
			Meta:   Meta{Page: 1, Limit: 10, Total: 0},
		})
	})

	_, err := c.ServiceTemplates.GetByID(t.Context(), 999)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	nfe, ok := errors.AsType[*NotFoundError](err)
	if !ok {
		t.Fatalf("expected *NotFoundError, got %T: %v", err, err)
	}
	if nfe.ID != 999 {
		t.Errorf("NotFoundError.ID = %d, want 999", nfe.ID)
	}
}

func TestServiceTemplateService_Get(t *testing.T) {
	mux, c := newTestMux(t)

	// A live service-template detail returns [] for categories and groups on a
	// bare template; the fixture deliberately populates them so the decode is
	// asserted and the field lines are sabotage-verifiable. Service-template
	// macros carry no id (unlike host-template macros).
	// The detail struct redeclares every ServiceTemplate base field flat, so this
	// fixture populates and asserts the full base surface, not just the extra
	// detail relationships; a json-tag typo on any redeclared field would
	// otherwise ship green. graph_template_id is left null to pin the nullable
	// path. Service-template macros carry no id (unlike host-template macros).
	mux.HandleFunc("GET /centreon/api/latest/configuration/services/templates/42", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"id":                                    42,
			"name":                                  "generic-service",
			"alias":                                 "Generic",
			"check_command_id":                      5,
			"check_timeperiod_id":                   6,
			"max_check_attempts":                    3,
			"normal_check_interval":                 5,
			"retry_check_interval":                  1,
			"service_template_id":                   7,
			"severity_id":                           8,
			"host_templates":                        []int{88, 91},
			"active_check_enabled":                  2,
			"passive_check_enabled":                 1,
			"volatility_enabled":                    2,
			"notification_enabled":                  2,
			"is_contact_additive_inheritance":       true,
			"is_contact_group_additive_inheritance": true,
			"notification_interval":                 10,
			"notification_timeperiod_id":            4,
			"notification_type":                     3,
			"first_notification_delay":              30,
			"recovery_notification_delay":           15,
			"acknowledgement_timeout":               60,
			"freshness_checked":                     1,
			"freshness_threshold":                   120,
			"flap_detection_enabled":                2,
			"low_flap_threshold":                    20,
			"high_flap_threshold":                   80,
			"event_handler_enabled":                 1,
			"event_handler_command_id":              9,
			"graph_template_id":                     nil,
			"note_url":                              "http://note",
			"note":                                  "a note",
			"action_url":                            "http://action",
			"icon_id":                               11,
			"icon_alternative":                      "alt",
			"comment":                               "a comment",
			"is_locked":                             false,
			"categories": []map[string]any{
				{"id": 3, "name": "ping"},
			},
			"groups": []map[string]any{
				{"id": 4, "name": "web-services"},
			},
			"macros": []map[string]any{
				{"name": "WARN", "value": "80", "is_password": false, "description": "warn"},
				{"name": "SECRET", "value": nil, "is_password": true, "description": ""},
			},
		})
	})

	tmpl, err := c.ServiceTemplates.Get(t.Context(), 42)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	wantInt(t, "ID", tmpl.ID, 42)
	wantStr(t, "Name", tmpl.Name, "generic-service")

	// Redeclared base fields: pin every tag so a Detail-only typo cannot hide.
	wantStr(t, "Alias", tmpl.Alias, "Generic")
	wantIntPtr(t, "CheckCommandID", tmpl.CheckCommandID, 5)
	wantIntPtr(t, "CheckTimeperiodID", tmpl.CheckTimeperiodID, 6)
	wantIntPtr(t, "MaxCheckAttempts", tmpl.MaxCheckAttempts, 3)
	wantIntPtr(t, "NormalCheckInterval", tmpl.NormalCheckInterval, 5)
	wantIntPtr(t, "RetryCheckInterval", tmpl.RetryCheckInterval, 1)
	wantIntPtr(t, "ServiceTemplateID", tmpl.ServiceTemplateID, 7)
	wantIntPtr(t, "SeverityID", tmpl.SeverityID, 8)
	if !slices.Equal(tmpl.HostTemplates, []int{88, 91}) {
		t.Errorf("HostTemplates = %v, want [88 91]", tmpl.HostTemplates)
	}
	wantInt(t, "ActiveCheckEnabled", tmpl.ActiveCheckEnabled, 2)
	wantInt(t, "PassiveCheckEnabled", tmpl.PassiveCheckEnabled, 1)
	wantInt(t, "VolatilityEnabled", tmpl.VolatilityEnabled, 2)
	wantInt(t, "NotificationEnabled", tmpl.NotificationEnabled, 2)
	wantBool(t, "IsContactAdditiveInheritance", tmpl.IsContactAdditiveInheritance, true)
	wantBool(t, "IsContactGroupAdditiveInheritance", tmpl.IsContactGroupAdditiveInheritance, true)
	wantIntPtr(t, "NotificationInterval", tmpl.NotificationInterval, 10)
	wantIntPtr(t, "NotificationTimeperiodID", tmpl.NotificationTimeperiodID, 4)
	wantIntPtr(t, "NotificationType", tmpl.NotificationType, 3)
	wantIntPtr(t, "FirstNotificationDelay", tmpl.FirstNotificationDelay, 30)
	wantIntPtr(t, "RecoveryNotificationDelay", tmpl.RecoveryNotificationDelay, 15)
	wantIntPtr(t, "AcknowledgementTimeout", tmpl.AcknowledgementTimeout, 60)
	wantInt(t, "FreshnessChecked", tmpl.FreshnessChecked, 1)
	wantIntPtr(t, "FreshnessThreshold", tmpl.FreshnessThreshold, 120)
	wantInt(t, "FlapDetectionEnabled", tmpl.FlapDetectionEnabled, 2)
	wantIntPtr(t, "LowFlapThreshold", tmpl.LowFlapThreshold, 20)
	wantIntPtr(t, "HighFlapThreshold", tmpl.HighFlapThreshold, 80)
	wantInt(t, "EventHandlerEnabled", tmpl.EventHandlerEnabled, 1)
	wantIntPtr(t, "EventHandlerCommandID", tmpl.EventHandlerCommandID, 9)
	wantNilIntPtr(t, "GraphTemplateID", tmpl.GraphTemplateID)
	wantStr(t, "NoteURL", tmpl.NoteURL, "http://note")
	wantStr(t, "Note", tmpl.Note, "a note")
	wantStr(t, "ActionURL", tmpl.ActionURL, "http://action")
	wantIntPtr(t, "IconID", tmpl.IconID, 11)
	wantStr(t, "IconAlternative", tmpl.IconAlternative, "alt")
	wantStr(t, "Comment", tmpl.Comment, "a comment")
	wantBool(t, "IsLocked", tmpl.IsLocked, false)

	if len(tmpl.Categories) != 1 {
		t.Fatalf("len(Categories) = %d, want 1", len(tmpl.Categories))
	}
	wantInt(t, "Categories[0].ID", tmpl.Categories[0].ID, 3)
	wantStr(t, "Categories[0].Name", tmpl.Categories[0].Name, "ping")

	if len(tmpl.Groups) != 1 {
		t.Fatalf("len(Groups) = %d, want 1", len(tmpl.Groups))
	}
	wantInt(t, "Groups[0].ID", tmpl.Groups[0].ID, 4)
	wantStr(t, "Groups[0].Name", tmpl.Groups[0].Name, "web-services")

	if len(tmpl.Macros) != 2 {
		t.Fatalf("len(Macros) = %d, want 2", len(tmpl.Macros))
	}
	wantStr(t, "Macros[0].Name", tmpl.Macros[0].Name, "WARN")
	if tmpl.Macros[0].Value == nil || *tmpl.Macros[0].Value != "80" {
		t.Errorf("Macros[0].Value = %v, want 80", tmpl.Macros[0].Value)
	}
	// A masked password macro decodes to a nil Value pointer.
	wantBool(t, "Macros[1].IsPassword", tmpl.Macros[1].IsPassword, true)
	if tmpl.Macros[1].Value != nil {
		t.Errorf("Macros[1].Value = %q, want nil (password masked)", *tmpl.Macros[1].Value)
	}
}

func TestServiceTemplateService_Get_NotFound(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/services/templates/999", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "ServiceTemplate not found"})
	})

	tmpl, err := c.ServiceTemplates.Get(t.Context(), 999)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if tmpl != nil {
		t.Errorf("expected nil template, got %+v", tmpl)
	}
	apiErr, ok := errors.AsType[*APIError](err)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.HTTPStatus != 404 {
		t.Errorf("APIError.HTTPStatus = %d, want 404", apiErr.HTTPStatus)
	}
}

func TestServiceTemplateService_Create(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("POST /centreon/api/latest/configuration/services/templates", func(w http.ResponseWriter, r *http.Request) {
		var req CreateServiceTemplateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if req.Name != "new-template" {
			t.Errorf("Name = %q, want %q", req.Name, "new-template")
		}
		writeJSON(w, http.StatusCreated, map[string]int{"id": 50})
	})

	id, err := c.ServiceTemplates.Create(t.Context(), CreateServiceTemplateRequest{
		Name:  "new-template",
		Alias: "New Template",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 50 {
		t.Errorf("id = %d, want 50", id)
	}
}

func TestServiceTemplateService_Update(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("PATCH /centreon/api/latest/configuration/services/templates/42", func(w http.ResponseWriter, r *http.Request) {
		var req UpdateServiceTemplateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if req.Name == nil || *req.Name != "updated-template" {
			t.Errorf("Name = %v, want %q", req.Name, "updated-template")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	name := "updated-template"
	err := c.ServiceTemplates.Update(t.Context(), 42, UpdateServiceTemplateRequest{Name: &name})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestServiceTemplateService_Delete(t *testing.T) {
	mux, c := newTestMux(t)

	var called bool
	mux.HandleFunc("DELETE /centreon/api/latest/configuration/services/templates/42", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.ServiceTemplates.Delete(t.Context(), 42)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}
