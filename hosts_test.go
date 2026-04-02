package centreon

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

func TestNamedRef_JSON(t *testing.T) {
	orig := NamedRef{ID: 4, Name: "probe-05"}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got NamedRef
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got != orig {
		t.Errorf("round-trip mismatch: got %+v, want %+v", got, orig)
	}
}

func TestHostService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/hosts", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[Host]{
			Result: []Host{
				{ID: 1, Name: "host-01", Address: "10.0.0.1", MonitoringServer: NamedRef{ID: 1}, IsActivated: true},
				{ID: 2, Name: "host-02", Address: "10.0.0.2", MonitoringServer: NamedRef{ID: 1}, IsActivated: true},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 2},
		})
	})

	resp, err := c.Hosts.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}
	if resp.Result[0].MonitoringServer.ID != 1 {
		t.Errorf("Result[0].MonitoringServer.ID = %d, want 1", resp.Result[0].MonitoringServer.ID)
	}
	if resp.Result[0].Name != "host-01" {
		t.Errorf("Result[0].Name = %q, want %q", resp.Result[0].Name, "host-01")
	}
}

func TestHostService_List_WithSearch(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/hosts", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("search") == "" {
			t.Error("expected search query param")
		}
		writeJSON(w, http.StatusOK, ListResponse[Host]{
			Result: []Host{
				{ID: 1, Name: "host-01", Address: "10.0.0.1", MonitoringServer: NamedRef{ID: 1}, IsActivated: true},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 1},
		})
	})

	resp, err := c.Hosts.List(t.Context(), WithSearch(Eq("name", "host-01")))
	if err != nil {
		t.Fatalf("List with search: %v", err)
	}
	if len(resp.Result) != 1 {
		t.Fatalf("len(Result) = %d, want 1", len(resp.Result))
	}
}

func TestHostService_GetByID_Found(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/hosts", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[Host]{
			Result: []Host{
				{ID: 42, Name: "host-42", Address: "10.0.0.42", MonitoringServer: NamedRef{ID: 1}, IsActivated: true},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 1},
		})
	})

	host, err := c.Hosts.GetByID(t.Context(), 42)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if host.ID != 42 {
		t.Errorf("ID = %d, want 42", host.ID)
	}
	if host.Name != "host-42" {
		t.Errorf("Name = %q, want %q", host.Name, "host-42")
	}
}

func TestHostService_GetByID_NotFound(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/hosts", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[Host]{
			Result: []Host{},
			Meta:   Meta{Page: 1, Limit: 10, Total: 0},
		})
	})

	_, err := c.Hosts.GetByID(t.Context(), 999)
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

func TestHostService_Create(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("POST /centreon/api/latest/configuration/hosts", func(w http.ResponseWriter, r *http.Request) {
		var req CreateHostRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if req.Name != "new-host" {
			t.Errorf("Name = %q, want %q", req.Name, "new-host")
		}
		if req.MonitoringServerID != 1 {
			t.Errorf("MonitoringServerID = %d, want 1", req.MonitoringServerID)
		}
		if req.Address != "10.0.0.99" {
			t.Errorf("Address = %q, want %q", req.Address, "10.0.0.99")
		}
		writeJSON(w, http.StatusCreated, map[string]int{"id": 99})
	})

	id, err := c.Hosts.Create(t.Context(), &CreateHostRequest{
		Name:               "new-host",
		MonitoringServerID: 1,
		Address:            "10.0.0.99",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 99 {
		t.Errorf("id = %d, want 99", id)
	}
}

func TestHostService_Update(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("PATCH /centreon/api/latest/configuration/hosts/42", func(w http.ResponseWriter, r *http.Request) {
		var req UpdateHostRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if req.Name == nil || *req.Name != "updated-host" {
			t.Errorf("Name = %v, want %q", req.Name, "updated-host")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	name := "updated-host"
	err := c.Hosts.Update(t.Context(), 42, &UpdateHostRequest{Name: &name})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestHostService_Delete(t *testing.T) {
	mux, c := newTestMux(t)

	var called bool
	mux.HandleFunc("DELETE /centreon/api/latest/configuration/hosts/42", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.Hosts.Delete(t.Context(), 42)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func checkHostCreateFields(t *testing.T, req *CreateHostRequest) {
	t.Helper()
	if req.Name != "host-with-relations" {
		t.Errorf("Name = %q, want %q", req.Name, "host-with-relations")
	}
	if req.MonitoringServerID != 2 {
		t.Errorf("MonitoringServerID = %d, want 2", req.MonitoringServerID)
	}
	if req.SNMPCommunity != "public" || req.SNMPVersion != "2c" {
		t.Errorf("SNMP = (%q, %q), want (public, 2c)", req.SNMPCommunity, req.SNMPVersion)
	}
}

func checkHostCreateSlices(t *testing.T, req *CreateHostRequest) {
	t.Helper()
	if len(req.Templates) != 2 || req.Templates[0] != 10 || req.Templates[1] != 20 {
		t.Errorf("Templates = %v, want [10 20]", req.Templates)
	}
	if len(req.Groups) != 1 || req.Groups[0] != 5 {
		t.Errorf("Groups = %v, want [5]", req.Groups)
	}
	if len(req.Categories) != 1 || req.Categories[0] != 3 {
		t.Errorf("Categories = %v, want [3]", req.Categories)
	}
	if len(req.Macros) != 2 {
		t.Errorf("len(Macros) = %d, want 2", len(req.Macros))
		return
	}
	if req.Macros[0].Name != "COMMUNITY" || req.Macros[0].Value != "public" {
		t.Errorf("Macros[0] = %+v, want {Name:COMMUNITY Value:public}", req.Macros[0])
	}
	if req.Macros[1].Name != "PASSWORD" || !req.Macros[1].IsPassword {
		t.Errorf("Macros[1] = %+v, want {Name:PASSWORD IsPassword:true}", req.Macros[1])
	}
}

func TestHostService_Create_WithTemplatesGroupsMacros(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("POST /centreon/api/latest/configuration/hosts", func(w http.ResponseWriter, r *http.Request) {
		var req CreateHostRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		checkHostCreateFields(t, &req)
		checkHostCreateSlices(t, &req)
		writeJSON(w, http.StatusCreated, map[string]int{"id": 101})
	})

	id, err := c.Hosts.Create(t.Context(), &CreateHostRequest{
		Name:               "host-with-relations",
		MonitoringServerID: 2,
		Address:            "192.168.1.10",
		SNMPCommunity:      "public",
		SNMPVersion:        "2c",
		Templates:          []int{10, 20},
		Groups:             []int{5},
		Categories:         []int{3},
		Macros: []Macro{
			{Name: "COMMUNITY", Value: "public"},
			{Name: "PASSWORD", IsPassword: true},
		},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 101 {
		t.Errorf("id = %d, want 101", id)
	}
}

func TestHostService_Update_WithRelationshipFields(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("PATCH /centreon/api/latest/configuration/hosts/10", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		for _, key := range []string{"templates", "macros", "snmp_community"} {
			if _, ok := body[key]; !ok {
				t.Errorf("expected %q key in PATCH body", key)
			}
		}
		if _, ok := body["name"]; ok {
			t.Error("unexpected 'name' key in PATCH body")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	community := "private"
	err := c.Hosts.Update(t.Context(), 10, &UpdateHostRequest{
		SNMPCommunity: &community,
		Templates:     &[]int{100, 200},
		Macros:        &[]Macro{{Name: "ENV", Value: "prod"}},
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
}
