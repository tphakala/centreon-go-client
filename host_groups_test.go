package centreon

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestHostGroupService_List(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/hosts/groups", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, ListResponse[HostGroup]{
			Result: []HostGroup{
				{ID: 1, Name: "linux-servers", Alias: "Linux Servers", IsActivated: true},
				{ID: 2, Name: "windows-servers", Alias: "Windows Servers", IsActivated: true},
			},
			Meta: Meta{Page: 1, Limit: 10, Total: 2},
		})
	})

	resp, err := c.HostGroups.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(resp.Result) != 2 {
		t.Fatalf("len(Result) = %d, want 2", len(resp.Result))
	}
	if resp.Result[0].Name != "linux-servers" {
		t.Errorf("Result[0].Name = %q, want %q", resp.Result[0].Name, "linux-servers")
	}
}

func TestHostGroupService_Get(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/hosts/groups/5", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, HostGroup{
			ID: 5, Name: "network-devices", Alias: "Network Devices", IsActivated: true,
		})
	})

	hg, err := c.HostGroups.Get(t.Context(), 5)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if hg.ID != 5 {
		t.Errorf("ID = %d, want 5", hg.ID)
	}
	if hg.Name != "network-devices" {
		t.Errorf("Name = %q, want %q", hg.Name, "network-devices")
	}
}

func TestHostGroupService_Create(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("POST /centreon/api/latest/configuration/hosts/groups", func(w http.ResponseWriter, r *http.Request) {
		var req CreateHostGroupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if req.Name != "new-group" {
			t.Errorf("Name = %q, want %q", req.Name, "new-group")
		}
		writeJSON(w, http.StatusCreated, map[string]int{"id": 10})
	})

	id, err := c.HostGroups.Create(t.Context(), CreateHostGroupRequest{
		Name:  "new-group",
		Alias: "New Group",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 10 {
		t.Errorf("id = %d, want 10", id)
	}
}

func TestHostGroupService_Update(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("PUT /centreon/api/latest/configuration/hosts/groups/5", func(w http.ResponseWriter, r *http.Request) {
		var req UpdateHostGroupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if req.Name != "updated-group" {
			t.Errorf("Name = %q, want %q", req.Name, "updated-group")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.HostGroups.Update(t.Context(), 5, UpdateHostGroupRequest{Name: "updated-group", Alias: "Updated Group"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestHostGroupService_Delete(t *testing.T) {
	mux, c := newTestMux(t)

	var called bool
	mux.HandleFunc("DELETE /centreon/api/latest/configuration/hosts/groups/5", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.HostGroups.Delete(t.Context(), 5)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}
