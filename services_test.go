package centreon

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

func TestServiceService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/services", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{
					"id":   2755,
					"name": "Broker-Stats",
					"hosts": []map[string]any{
						{"id": 10246, "name": "AAAAAAProbe-template"},
					},
					"service_template":        map[string]any{"id": 666, "name": "App-Monitoring-Centreon-Broker-Stats-Poller-custom"},
					"check_timeperiod":        nil,
					"notification_timeperiod": nil,
					"severity":                nil,
					"categories":              []map[string]any{},
					"groups":                  []map[string]any{},
					"normal_check_interval":   nil,
					"retry_check_interval":    nil,
					"is_activated":            true,
				},
				{
					"id":   2800,
					"name": "CPU",
					"hosts": []map[string]any{
						{"id": 10246, "name": "AAAAAAProbe-template"},
					},
					"service_template":        nil,
					"check_timeperiod":        nil,
					"notification_timeperiod": nil,
					"severity":                nil,
					"categories":              []map[string]any{},
					"groups":                  []map[string]any{},
					"normal_check_interval":   5,
					"retry_check_interval":    1,
					"is_activated":            true,
				},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 2},
		})
	})

	resp, err := c.Services.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}

	svc := resp.Result[0]
	if svc.Name != "Broker-Stats" {
		t.Errorf("Result[0].Name = %q, want %q", svc.Name, "Broker-Stats")
	}
	if len(svc.Hosts) != 1 || svc.Hosts[0].ID != 10246 {
		t.Errorf("Result[0].Hosts = %+v, want [{ID:10246 Name:AAAAAAProbe-template}]", svc.Hosts)
	}
	if svc.ServiceTemplate == nil || svc.ServiceTemplate.ID != 666 {
		t.Errorf("Result[0].ServiceTemplate = %+v, want &{ID:666}", svc.ServiceTemplate)
	}
	if !svc.IsActivated {
		t.Error("Result[0].IsActivated = false, want true")
	}

	svc2 := resp.Result[1]
	if svc2.NormalCheckInterval == nil || *svc2.NormalCheckInterval != 5 {
		t.Errorf("Result[1].NormalCheckInterval = %v, want 5", svc2.NormalCheckInterval)
	}
	if svc2.RetryCheckInterval == nil || *svc2.RetryCheckInterval != 1 {
		t.Errorf("Result[1].RetryCheckInterval = %v, want 1", svc2.RetryCheckInterval)
	}
}

func TestServiceService_ListByHost(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/services", func(w http.ResponseWriter, r *http.Request) {
		search := r.URL.Query().Get("search")
		if search == "" {
			t.Error("expected search parameter, got empty")
		}
		// Verify the search parameter contains host.id filter
		if search != `{"host.id":{"$eq":10246}}` {
			t.Errorf("search = %q, want host.id eq filter", search)
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{
					"id":   2755,
					"name": "Broker-Stats",
					"hosts": []map[string]any{
						{"id": 10246, "name": "AAAAAAProbe-template"},
					},
					"service_template":        nil,
					"check_timeperiod":        nil,
					"notification_timeperiod": nil,
					"severity":                nil,
					"categories":              []map[string]any{},
					"groups":                  []map[string]any{},
					"normal_check_interval":   nil,
					"retry_check_interval":    nil,
					"is_activated":            true,
				},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 1},
		})
	})

	resp, err := c.Services.ListByHost(t.Context(), 10246)
	if err != nil {
		t.Fatalf("ListByHost: %v", err)
	}
	if len(resp.Result) != 1 {
		t.Fatalf("len(Result) = %d, want 1", len(resp.Result))
	}
	if len(resp.Result[0].Hosts) != 1 || resp.Result[0].Hosts[0].ID != 10246 {
		t.Errorf("Hosts = %+v, want [{ID:10246 ...}]", resp.Result[0].Hosts)
	}
}

func TestServiceService_Create(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("POST /centreon/api/latest/configuration/services", func(w http.ResponseWriter, r *http.Request) {
		var req CreateServiceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if req.Name != "new-service" {
			t.Errorf("Name = %q, want %q", req.Name, "new-service")
		}
		if req.HostID != 10 {
			t.Errorf("HostID = %d, want 10", req.HostID)
		}
		writeJSON(w, http.StatusCreated, map[string]int{"id": 99})
	})

	id, err := c.Services.Create(t.Context(), &CreateServiceRequest{
		HostID: 10,
		Name:   "new-service",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 99 {
		t.Errorf("id = %d, want 99", id)
	}
}

func TestServiceService_Update(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("PATCH /centreon/api/latest/configuration/services/42", func(w http.ResponseWriter, r *http.Request) {
		var req UpdateServiceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if req.Name == nil || *req.Name != "updated-service" {
			t.Errorf("Name = %v, want %q", req.Name, "updated-service")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	name := "updated-service"
	err := c.Services.Update(t.Context(), 42, &UpdateServiceRequest{Name: &name})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestServiceService_Delete(t *testing.T) {
	mux, c := newTestMux(t)

	var called bool
	mux.HandleFunc("DELETE /centreon/api/latest/configuration/services/42", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.Services.Delete(t.Context(), 42)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func checkServiceCreateRelations(t *testing.T, req *CreateServiceRequest) {
	t.Helper()
	if req.Name != "svc-with-template" {
		t.Errorf("Name = %q, want %q", req.Name, "svc-with-template")
	}
	if req.HostID != 5 {
		t.Errorf("HostID = %d, want 5", req.HostID)
	}
	if req.ServiceTemplateID != 100 {
		t.Errorf("ServiceTemplateID = %d, want 100", req.ServiceTemplateID)
	}
	if len(req.ServiceCategories) != 2 || req.ServiceCategories[0] != 7 || req.ServiceCategories[1] != 8 {
		t.Errorf("ServiceCategories = %v, want [7 8]", req.ServiceCategories)
	}
	if len(req.ServiceGroups) != 1 || req.ServiceGroups[0] != 4 {
		t.Errorf("ServiceGroups = %v, want [4]", req.ServiceGroups)
	}
	if len(req.Macros) != 1 || req.Macros[0].Name != "WARNING" || req.Macros[0].Value != "80" {
		t.Errorf("Macros = %+v, want [{Name:WARNING Value:80}]", req.Macros)
	}
}

func TestServiceService_Create_WithTemplateAndCategories(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("POST /centreon/api/latest/configuration/services", func(w http.ResponseWriter, r *http.Request) {
		var req CreateServiceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		checkServiceCreateRelations(t, &req)
		writeJSON(w, http.StatusCreated, map[string]int{"id": 55})
	})

	id, err := c.Services.Create(t.Context(), &CreateServiceRequest{
		HostID:            5,
		Name:              "svc-with-template",
		ServiceTemplateID: 100,
		ServiceCategories: []int{7, 8},
		ServiceGroups:     []int{4},
		Macros:            []Macro{{Name: "WARNING", Value: "80"}},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 55 {
		t.Errorf("id = %d, want 55", id)
	}
}

func TestServiceService_Update_WithRelationshipFields(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("PATCH /centreon/api/latest/configuration/services/20", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		for _, key := range []string{"service_categories", "service_template_id", "macros"} {
			if _, ok := body[key]; !ok {
				t.Errorf("expected %q key in PATCH body", key)
			}
		}
		if _, ok := body["name"]; ok {
			t.Error("unexpected 'name' key in PATCH body")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	tmplID := 200
	err := c.Services.Update(t.Context(), 20, &UpdateServiceRequest{
		ServiceTemplateID: &tmplID,
		ServiceCategories: &[]int{9, 10},
		Macros:            &[]Macro{{Name: "CRITICAL", Value: "95"}},
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
}

// realisticServiceDetailJSON returns a full GET /configuration/services/{id}
// detail payload (Centreon 25.10+) for service 7 on host 42, exercising
// nullable fields and both a plain and a password (null-value) macro. Note the
// service macro shape has no "id" key, unlike the host macro shape.
func realisticServiceDetailJSON() map[string]any {
	return map[string]any{
		"id":                                    7,
		"name":                                  "service-07",
		"host_id":                               42,
		"geo_coords":                            "",
		"comment":                               "",
		"service_template_id":                   12,
		"check_command_id":                      nil,
		"check_command_args":                    []string{"80", "90"},
		"check_timeperiod_id":                   nil,
		"max_check_attempts":                    3,
		"normal_check_interval":                 nil,
		"retry_check_interval":                  nil,
		"active_check_enabled":                  2,
		"passive_check_enabled":                 2,
		"volatility_enabled":                    2,
		"notification_enabled":                  2,
		"is_contact_additive_inheritance":       false,
		"is_contact_group_additive_inheritance": false,
		"notification_interval":                 nil,
		"notification_timeperiod_id":            1,
		"notification_type":                     nil,
		"first_notification_delay":              nil,
		"recovery_notification_delay":           nil,
		"acknowledgement_timeout":               nil,
		"freshness_checked":                     2,
		"freshness_threshold":                   nil,
		"flap_detection_enabled":                2,
		"low_flap_threshold":                    nil,
		"high_flap_threshold":                   nil,
		"event_handler_enabled":                 2,
		"event_handler_command_id":              nil,
		"event_handler_command_args":            []string{},
		"graph_template_id":                     nil,
		"note":                                  "",
		"note_url":                              "",
		"action_url":                            "",
		"icon_id":                               nil,
		"icon_alternative":                      "",
		"severity_id":                           nil,
		"is_activated":                          true,
		"categories":                            []map[string]any{{"id": 1, "name": "svc-cat"}},
		"groups":                                []map[string]any{{"id": 2, "name": "svc-grp"}},
		"macros": []map[string]any{
			{"name": "USER1", "value": "/usr/lib/nagios/plugins", "is_password": false, "description": "plugin path"},
			{"name": "SERVICEPASSWORD", "value": nil, "is_password": true, "description": nil},
		},
	}
}

func TestServiceService_Get(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/services/7", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, realisticServiceDetailJSON())
	})

	svc, err := c.Services.Get(t.Context(), 7)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if svc.ID != 7 {
		t.Errorf("ID = %d, want 7", svc.ID)
	}
	if svc.HostID != 42 {
		t.Errorf("HostID = %d, want 42", svc.HostID)
	}
	if svc.ServiceTemplateID == nil || *svc.ServiceTemplateID != 12 {
		t.Errorf("ServiceTemplateID = %v, want 12", svc.ServiceTemplateID)
	}
	if svc.CheckCommandID != nil {
		t.Errorf("CheckCommandID = %v, want nil", svc.CheckCommandID)
	}

	t.Run("macros", func(t *testing.T) {
		if len(svc.Macros) != 2 {
			t.Fatalf("len(Macros) = %d, want 2", len(svc.Macros))
		}
		if svc.Macros[0].Name != "USER1" {
			t.Errorf("Macros[0].Name = %q, want USER1", svc.Macros[0].Name)
		}
		if svc.Macros[0].Value == nil || *svc.Macros[0].Value != "/usr/lib/nagios/plugins" {
			t.Errorf("Macros[0].Value = %v, want plugin path", svc.Macros[0].Value)
		}
	})

	t.Run("password_macro_null_value", func(t *testing.T) {
		if len(svc.Macros) != 2 {
			t.Fatalf("len(Macros) = %d, want 2", len(svc.Macros))
		}
		pw := svc.Macros[1]
		if !pw.IsPassword {
			t.Errorf("Macros[1].IsPassword = false, want true")
		}
		if pw.Value != nil {
			t.Errorf("Macros[1].Value = %q, want nil (password masked)", *pw.Value)
		}
	})
}

func TestServiceService_Get_NotFound(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/services/999", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotFound, map[string]any{"code": 404, "message": "Service not found"})
	})

	svc, err := c.Services.Get(t.Context(), 999)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if svc != nil {
		t.Errorf("expected nil service, got %+v", svc)
	}
	apiErr, ok := errors.AsType[*APIError](err)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.HTTPStatus != 404 {
		t.Errorf("APIError.HTTPStatus = %d, want 404", apiErr.HTTPStatus)
	}
}
