package centreon

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

func TestHostTemplateService_List_AllFields(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/hosts/templates", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{
					"id": 1, "name": "tpl-full", "is_locked": false,
					"check_command_args":          []string{"-c", "check_ping"},
					"snmp_version":                "2c",
					"snmp_community":              "public",
					"timezone_id":                 3,
					"severity_id":                 7,
					"active_check_enabled":        2,
					"passive_check_enabled":       1,
					"notification_enabled":        2,
					"notification_options":        5,
					"notification_interval":       10,
					"notification_timeperiod_id":  4,
					"add_inherited_contact_group": true,
					"add_inherited_contact":       true,
					"first_notification_delay":    30,
					"recovery_notification_delay": 15,
					"acknowledgement_timeout":     60,
					"freshness_checked":           1,
					"freshness_threshold":         120,
					"flap_detection_enabled":      2,
					"low_flap_threshold":          20,
					"high_flap_threshold":         80,
					"event_handler_enabled":       1,
					"event_handler_command_id":    9,
					"event_handler_command_args":  []string{"-w", "5"},
					"note_url":                    "http://note",
					"note":                        "a note",
					"action_url":                  "http://action",
					"icon_id":                     11,
					"icon_alternative":            "alt",
					"comment":                     "a comment",
				},
				{
					"id": 2, "name": "tpl-nulls", "is_locked": true,
					"timezone_id":              nil,
					"severity_id":              nil,
					"icon_id":                  nil,
					"event_handler_command_id": nil,
					// A JSON null for a plain-int toggle decodes to 0 with no
					// error (encoding/json treats null into a non-pointer as a
					// no-op), so it does not break the List decode.
					"active_check_enabled": nil,
					"notification_enabled": nil,
				},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 2},
		})
	})

	resp, err := c.HostTemplates.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}

	tpl := resp.Result[0]
	wantStrSlice(t, "CheckCommandArgs", tpl.CheckCommandArgs, []string{"-c", "check_ping"})
	wantStr(t, "SNMPVersion", tpl.SNMPVersion, "2c")
	wantStr(t, "SNMPCommunity", tpl.SNMPCommunity, "public")
	wantIntPtr(t, "TimezoneID", tpl.TimezoneID, 3)
	wantIntPtr(t, "SeverityID", tpl.SeverityID, 7)
	wantInt(t, "ActiveCheckEnabled", tpl.ActiveCheckEnabled, 2)
	wantInt(t, "PassiveCheckEnabled", tpl.PassiveCheckEnabled, 1)
	wantInt(t, "NotificationEnabled", tpl.NotificationEnabled, 2)
	wantIntPtr(t, "NotificationOptions", tpl.NotificationOptions, 5)
	wantIntPtr(t, "NotificationInterval", tpl.NotificationInterval, 10)
	wantIntPtr(t, "NotificationTimeperiodID", tpl.NotificationTimeperiodID, 4)
	wantBool(t, "AddInheritedContactGroup", tpl.AddInheritedContactGroup, true)
	wantBool(t, "AddInheritedContact", tpl.AddInheritedContact, true)
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
	wantStr(t, "NoteURL", tpl.NoteURL, "http://note")
	wantStr(t, "Note", tpl.Note, "a note")
	wantStr(t, "ActionURL", tpl.ActionURL, "http://action")
	wantIntPtr(t, "IconID", tpl.IconID, 11)
	wantStr(t, "IconAlternative", tpl.IconAlternative, "alt")
	wantStr(t, "Comment", tpl.Comment, "a comment")

	nulls := resp.Result[1]
	wantNilIntPtr(t, "Result[1].TimezoneID", nulls.TimezoneID)
	wantNilIntPtr(t, "Result[1].SeverityID", nulls.SeverityID)
	wantNilIntPtr(t, "Result[1].IconID", nulls.IconID)
	wantNilIntPtr(t, "Result[1].EventHandlerCommandID", nulls.EventHandlerCommandID)
	wantInt(t, "Result[1].ActiveCheckEnabled", nulls.ActiveCheckEnabled, 0)
	wantInt(t, "Result[1].NotificationEnabled", nulls.NotificationEnabled, 0)
}

func TestHostTemplateService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/hosts/templates", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{"id": 1, "name": "tpl-linux", "alias": "Linux Template", "check_command_id": nil, "check_timeperiod_id": 1, "max_check_attempts": nil, "normal_check_interval": nil, "retry_check_interval": nil, "is_locked": false},
				{"id": 2, "name": "tpl-windows", "alias": "Windows Template", "check_command_id": 5, "check_timeperiod_id": nil, "max_check_attempts": 3, "normal_check_interval": 5, "retry_check_interval": 1, "is_locked": true},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 2},
		})
	})

	resp, err := c.HostTemplates.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}

	tpl0 := resp.Result[0]
	if tpl0.Name != "tpl-linux" {
		t.Errorf("Result[0].Name = %q, want %q", tpl0.Name, "tpl-linux")
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
	if tpl1.CheckCommandID == nil || *tpl1.CheckCommandID != 5 {
		t.Errorf("Result[1].CheckCommandID = %v, want 5", tpl1.CheckCommandID)
	}
	if tpl1.MaxCheckAttempts == nil || *tpl1.MaxCheckAttempts != 3 {
		t.Errorf("Result[1].MaxCheckAttempts = %v, want 3", tpl1.MaxCheckAttempts)
	}
	if !tpl1.IsLocked {
		t.Error("Result[1].IsLocked = false, want true")
	}
}

func TestHostTemplateService_GetByID_Found(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/hosts/templates", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{"id": 10, "name": "tpl-snmp", "alias": "SNMP Template", "check_command_id": nil, "check_timeperiod_id": nil, "max_check_attempts": nil, "normal_check_interval": nil, "retry_check_interval": nil, "is_locked": false},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 1},
		})
	})

	tpl, err := c.HostTemplates.GetByID(t.Context(), 10)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if tpl.ID != 10 {
		t.Errorf("ID = %d, want 10", tpl.ID)
	}
	if tpl.Name != "tpl-snmp" {
		t.Errorf("Name = %q, want %q", tpl.Name, "tpl-snmp")
	}
}

func TestHostTemplateService_GetByID_NotFound(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/hosts/templates", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[HostTemplate]{
			Result: []HostTemplate{},
			Meta:   Meta{Page: 1, Limit: 10, Total: 0},
		})
	})

	_, err := c.HostTemplates.GetByID(t.Context(), 999)
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

func TestHostTemplateService_Create(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("POST /centreon/api/latest/configuration/hosts/templates", func(w http.ResponseWriter, r *http.Request) {
		var req CreateHostTemplateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if req.Name != "tpl-new" {
			t.Errorf("Name = %q, want %q", req.Name, "tpl-new")
		}
		writeJSON(w, http.StatusCreated, map[string]int{"id": 20})
	})

	id, err := c.HostTemplates.Create(t.Context(), CreateHostTemplateRequest{
		Name:  "tpl-new",
		Alias: "New Template",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 20 {
		t.Errorf("id = %d, want 20", id)
	}
}

func TestHostTemplateService_Update(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("PATCH /centreon/api/latest/configuration/hosts/templates/10", func(w http.ResponseWriter, r *http.Request) {
		var req UpdateHostTemplateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if req.Name == nil || *req.Name != "tpl-updated" {
			t.Errorf("Name = %v, want %q", req.Name, "tpl-updated")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	name := "tpl-updated"
	err := c.HostTemplates.Update(t.Context(), 10, UpdateHostTemplateRequest{Name: &name})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestHostTemplateService_Delete(t *testing.T) {
	mux, c := newTestMux(t)

	var called bool
	mux.HandleFunc("DELETE /centreon/api/latest/configuration/hosts/templates/10", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.HostTemplates.Delete(t.Context(), 10)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}
