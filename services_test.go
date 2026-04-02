package centreon

import (
	"encoding/json"
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
	if resp.Result[0].Hosts[0].ID != 10246 {
		t.Errorf("Hosts[0].ID = %d, want 10246", resp.Result[0].Hosts[0].ID)
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
