//go:build integration

package centreon

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

// Integration tests require a live Centreon instance.
// Run with: CENTREON_URL=https://... CENTREON_TOKEN=... CENTREON_INSECURE=1 go test -tags integration -v ./...
//
// Required environment variables:
//   CENTREON_URL       - Base URL (e.g., https://centreon.example.com)
//   CENTREON_USERNAME  - Login username
//   CENTREON_PASSWORD  - Login password
//
// Optional:
//   CENTREON_VERSION   - API version (default: latest)
//   CENTREON_TOKEN     - Use API token instead of username/password
//   CENTREON_INSECURE  - Set to skip TLS certificate verification

func newIntegrationClient(t *testing.T) *Client {
	t.Helper()

	baseURL := os.Getenv("CENTREON_URL")
	if baseURL == "" {
		t.Skip("CENTREON_URL not set, skipping integration test")
	}

	var opts []Option
	if os.Getenv("CENTREON_INSECURE") != "" {
		opts = append(opts, WithInsecureTLS())
	}
	if token := os.Getenv("CENTREON_TOKEN"); token != "" {
		opts = append(opts, WithAPIToken(token))
	} else {
		username := os.Getenv("CENTREON_USERNAME")
		password := os.Getenv("CENTREON_PASSWORD")
		if username == "" || password == "" {
			t.Skip("CENTREON_USERNAME/CENTREON_PASSWORD not set, skipping integration test")
		}
		opts = append(opts, WithCredentials(username, password))
	}

	if v := os.Getenv("CENTREON_VERSION"); v != "" {
		opts = append(opts, WithVersion(v))
	}

	client, err := NewClient(baseURL, opts...)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	if client.username != "" {
		if err := client.Login(context.Background()); err != nil {
			t.Fatalf("login: %v", err)
		}
		t.Cleanup(func() { client.Logout(context.Background()) })
	}

	return client
}

// --- Configuration endpoints ---

func TestIntegration_ListHosts(t *testing.T) {
	client := newIntegrationClient(t)

	resp, err := client.Hosts.List(t.Context(), WithLimit(5))
	if err != nil {
		t.Fatalf("Hosts.List: %v", err)
	}
	t.Logf("Found %d hosts (total: %d)", len(resp.Result), resp.Meta.Total)
	for _, h := range resp.Result {
		t.Logf("  %d: %s (%s)", h.ID, h.Name, h.Address)
	}
}

func TestIntegration_ListServices(t *testing.T) {
	client := newIntegrationClient(t)

	resp, err := client.Services.List(t.Context(), WithLimit(5))
	if err != nil {
		t.Fatalf("Services.List: %v", err)
	}
	t.Logf("Found %d services (total: %d)", len(resp.Result), resp.Meta.Total)
}

func TestIntegration_ListTimePeriods(t *testing.T) {
	client := newIntegrationClient(t)

	resp, err := client.TimePeriods.List(t.Context())
	if err != nil {
		t.Fatalf("TimePeriods.List: %v", err)
	}
	t.Logf("Found %d time periods", len(resp.Result))
	for _, tp := range resp.Result {
		t.Logf("  %d: %s", tp.ID, tp.Name)
	}
}

func TestIntegration_ListHostGroups(t *testing.T) {
	client := newIntegrationClient(t)

	resp, err := client.HostGroups.List(t.Context())
	if err != nil {
		t.Fatalf("HostGroups.List: %v", err)
	}
	t.Logf("Found %d host groups (total: %d)", len(resp.Result), resp.Meta.Total)
}

func TestIntegration_ListHostCategories(t *testing.T) {
	client := newIntegrationClient(t)

	resp, err := client.HostCategories.List(t.Context())
	if err != nil {
		t.Fatalf("HostCategories.List: %v", err)
	}
	t.Logf("Found %d host categories (total: %d)", len(resp.Result), resp.Meta.Total)
}

func TestIntegration_ListHostTemplates(t *testing.T) {
	client := newIntegrationClient(t)

	resp, err := client.HostTemplates.List(t.Context())
	if err != nil {
		t.Fatalf("HostTemplates.List: %v", err)
	}
	t.Logf("Found %d host templates (total: %d)", len(resp.Result), resp.Meta.Total)
}

func TestIntegration_ListServiceTemplates(t *testing.T) {
	client := newIntegrationClient(t)

	resp, err := client.ServiceTemplates.List(t.Context())
	if err != nil {
		t.Fatalf("ServiceTemplates.List: %v", err)
	}
	t.Logf("Found %d service templates (total: %d)", len(resp.Result), resp.Meta.Total)
}

func TestIntegration_ListServiceGroups(t *testing.T) {
	client := newIntegrationClient(t)

	resp, err := client.ServiceGroups.List(t.Context())
	if err != nil {
		t.Fatalf("ServiceGroups.List: %v", err)
	}
	t.Logf("Found %d service groups (total: %d)", len(resp.Result), resp.Meta.Total)
}

func TestIntegration_ListCommands(t *testing.T) {
	client := newIntegrationClient(t)

	resp, err := client.Commands.List(t.Context(), WithLimit(5))
	if err != nil {
		t.Fatalf("Commands.List: %v", err)
	}
	t.Logf("Found %d commands (total: %d)", len(resp.Result), resp.Meta.Total)
}

func TestIntegration_ListMonitoringServers(t *testing.T) {
	client := newIntegrationClient(t)

	resp, err := client.MonitoringServers.List(t.Context())
	if err != nil {
		// Some API tokens may not have permission for this endpoint
		t.Skipf("MonitoringServers.List: %v (may require admin permissions)", err)
	}
	t.Logf("Found %d monitoring servers", len(resp.Result))
	for _, s := range resp.Result {
		t.Logf("  %d: %s (default=%v)", s.ID, s.Name, s.IsDefault)
	}
}

// --- User/contact endpoints (fixed in #28) ---

func TestIntegration_ListUsers(t *testing.T) {
	client := newIntegrationClient(t)

	resp, err := client.Users.List(t.Context(), WithLimit(5))
	if err != nil {
		t.Fatalf("Users.List: %v", err)
	}
	t.Logf("Found %d users (total: %d)", len(resp.Result), resp.Meta.Total)
	for _, u := range resp.Result {
		t.Logf("  %d: %s (admin=%v)", u.ID, u.Name, u.IsAdmin)
	}
}

func TestIntegration_ListContactGroups(t *testing.T) {
	client := newIntegrationClient(t)

	resp, err := client.ContactGroups.List(t.Context())
	if err != nil {
		t.Fatalf("ContactGroups.List: %v", err)
	}
	t.Logf("Found %d contact groups (total: %d)", len(resp.Result), resp.Meta.Total)
}

func TestIntegration_ListContactTemplates(t *testing.T) {
	client := newIntegrationClient(t)

	resp, err := client.ContactTemplates.List(t.Context())
	if err != nil {
		t.Fatalf("ContactTemplates.List: %v", err)
	}
	t.Logf("Found %d contact templates (total: %d)", len(resp.Result), resp.Meta.Total)
}

func TestIntegration_ListUserFilters(t *testing.T) {
	client := newIntegrationClient(t)

	resp, err := client.UserFilters.List(t.Context())
	if err != nil {
		t.Fatalf("UserFilters.List: %v", err)
	}
	t.Logf("Found %d user filters (total: %d)", len(resp.Result), resp.Meta.Total)
}

// --- Monitoring endpoints ---

func TestIntegration_MonitoringStatus(t *testing.T) {
	client := newIntegrationClient(t)

	hostCounts, err := client.MonitoringHosts.StatusCounts(t.Context())
	if err != nil {
		t.Fatalf("MonitoringHosts.StatusCounts: %v", err)
	}
	t.Logf("Host status: UP=%d DOWN=%d Unreachable=%d Pending=%d",
		hostCounts.Up.Total, hostCounts.Down.Total, hostCounts.Unreachable.Total, hostCounts.Pending.Total)

	svcCounts, err := client.MonitoringServices.StatusCounts(t.Context())
	if err != nil {
		t.Fatalf("MonitoringServices.StatusCounts: %v", err)
	}
	t.Logf("Service status: OK=%d Warning=%d Critical=%d Unknown=%d Pending=%d",
		svcCounts.OK.Total, svcCounts.Warning.Total, svcCounts.Critical.Total, svcCounts.Unknown.Total, svcCounts.Pending.Total)
}

func TestIntegration_ListMonitoringHosts(t *testing.T) {
	client := newIntegrationClient(t)

	resp, err := client.MonitoringHosts.List(t.Context(), WithLimit(3))
	if err != nil {
		t.Fatalf("MonitoringHosts.List: %v", err)
	}
	t.Logf("Found %d monitoring hosts (total: %d)", len(resp.Result), resp.Meta.Total)
}

func TestIntegration_ListMonitoringServices(t *testing.T) {
	client := newIntegrationClient(t)

	resp, err := client.MonitoringServices.List(t.Context(), WithLimit(3))
	if err != nil {
		t.Fatalf("MonitoringServices.List: %v", err)
	}
	t.Logf("Found %d monitoring services (total: %d)", len(resp.Result), resp.Meta.Total)
}

func TestIntegration_ListMonitoringResources(t *testing.T) {
	client := newIntegrationClient(t)

	resp, err := client.Monitoring.List(t.Context(), WithLimit(3))
	if err != nil {
		t.Fatalf("Monitoring.List: %v", err)
	}
	t.Logf("Found %d resources (total: %d)", len(resp.Result), resp.Meta.Total)
}

func TestIntegration_ListDowntimes(t *testing.T) {
	client := newIntegrationClient(t)

	resp, err := client.Downtimes.List(t.Context())
	if err != nil {
		t.Fatalf("Downtimes.List: %v", err)
	}
	t.Logf("Found %d downtimes (total: %d)", len(resp.Result), resp.Meta.Total)
}

func TestIntegration_ListAcknowledgements(t *testing.T) {
	client := newIntegrationClient(t)

	resp, err := client.Acknowledgements.List(t.Context())
	if err != nil {
		t.Fatalf("Acknowledgements.List: %v", err)
	}
	t.Logf("Found %d acknowledgements (total: %d)", len(resp.Result), resp.Meta.Total)
}

// --- Search/filter ---

func TestIntegration_SearchFilter(t *testing.T) {
	client := newIntegrationClient(t)

	// Configuration endpoint uses "name" not "host.name"
	resp, err := client.Hosts.List(t.Context(),
		WithSearch(Lk("name", "%")),
		WithLimit(3),
	)
	if err != nil {
		t.Fatalf("Hosts.List with search: %v", err)
	}
	t.Logf("Search returned %d hosts (total: %d)", len(resp.Result), resp.Meta.Total)
}

func TestIntegration_SearchMonitoringResources(t *testing.T) {
	client := newIntegrationClient(t)

	// Monitoring endpoint uses "host.name" prefix
	resp, err := client.Monitoring.List(t.Context(),
		WithSearch(Lk("host.name", "%")),
		WithLimit(3),
	)
	if err != nil {
		t.Fatalf("Monitoring.List with search: %v", err)
	}
	t.Logf("Search returned %d resources (total: %d)", len(resp.Result), resp.Meta.Total)
}

// --- Notification policies ---

func TestIntegration_NotificationPolicy(t *testing.T) {
	client := newIntegrationClient(t)

	// Get a host ID to query notification policy
	hosts, err := client.Hosts.List(t.Context(), WithLimit(1))
	if err != nil {
		t.Fatalf("Hosts.List: %v", err)
	}
	if len(hosts.Result) == 0 {
		t.Skip("no hosts found")
	}
	hostID := hosts.Result[0].ID

	np, err := client.NotificationPolicies.GetForHost(t.Context(), hostID)
	if err != nil {
		t.Skipf("NotificationPolicies.GetForHost(%d): %v", hostID, err)
	}
	t.Logf("Host %d notification policy: enabled=%v, contacts=%d, groups=%d",
		hostID, np.IsNotificationEnabled, len(np.Contacts), len(np.ContactGroups))
}

// --- Configuration write lifecycle: custom macro round-trip (#36, #13) ---
//
// These create real objects and clean them up with t.Cleanup (which runs even
// after a t.Fatalf), so they are safe to run repeatedly. They exercise the
// detail GET read path (HostService.Get / ServiceService.Get) that is the only
// way to read custom macros back; the list endpoints do not return macros.
//
// Cleanup calls use context.Background(), not t.Context(), because t.Context()
// is canceled before cleanup functions run, which would cancel the Delete.

// defaultMonitoringServerID returns the default poller ID, or the first one.
// It skips the test when the endpoint is unavailable or returns nothing.
func defaultMonitoringServerID(t *testing.T, client *Client) int {
	t.Helper()
	resp, err := client.MonitoringServers.List(t.Context())
	if err != nil {
		t.Skipf("MonitoringServers.List: %v (may require admin permissions)", err)
	}
	if len(resp.Result) == 0 {
		t.Skip("no monitoring servers available")
	}
	for i := range resp.Result {
		if resp.Result[i].IsDefault {
			return resp.Result[i].ID
		}
	}
	return resp.Result[0].ID
}

// serviceCreatePrereq returns either a check command ID or a service template
// ID for creating a service (exactly one is non-zero). A service requires one
// of the two, and the client cannot create a check command, so the test is
// skipped when the instance has neither.
func serviceCreatePrereq(t *testing.T, client *Client) (checkCommandID, serviceTemplateID int) {
	t.Helper()
	cmds, err := client.Commands.List(t.Context(), WithLimit(200))
	if err != nil {
		t.Skipf("Commands.List: %v", err)
	}
	for i := range cmds.Result {
		if cmds.Result[i].Type == 2 && cmds.Result[i].IsActivated { // type 2 = check command
			return cmds.Result[i].ID, 0
		}
	}
	tmpls, err := client.ServiceTemplates.List(t.Context(), WithLimit(1))
	if err == nil && len(tmpls.Result) > 0 {
		return 0, tmpls.Result[0].ID
	}
	t.Skip("no check command (type 2) or service template available to create a service")
	return 0, 0
}

func findHostMacro(macros []HostMacro, name string) *HostMacro {
	for i := range macros {
		if macros[i].Name == name {
			return &macros[i]
		}
	}
	return nil
}

func findServiceMacro(macros []ServiceMacro, name string) *ServiceMacro {
	for i := range macros {
		if macros[i].Name == name {
			return &macros[i]
		}
	}
	return nil
}

// wantMacroValue asserts a *string macro value equals want, printing the actual
// value rather than a pointer address on failure. Serves both HostMacro and
// ServiceMacro, whose Value fields are both *string.
func wantMacroValue(t *testing.T, macroName string, got *string, want string) {
	t.Helper()
	switch {
	case got == nil:
		t.Errorf("macro %s value = nil, want %q", macroName, want)
	case *got != want:
		t.Errorf("macro %s value = %q, want %q", macroName, *got, want)
	}
}

// wantMaskedMacro asserts a password macro reads back masked: IsPassword true
// and Value nil. Centreon masks secret macro values on read, so this pins the
// masking contract that a plain-macro assertion (is_password == false, the bool
// zero value) cannot.
func wantMaskedMacro(t *testing.T, macroName string, isPassword bool, value *string) {
	t.Helper()
	if !isPassword {
		t.Errorf("macro %s is_password = false, want true", macroName)
	}
	if value != nil {
		t.Errorf("macro %s value = %q, want nil (password masked on read)", macroName, *value)
	}
}

func TestIntegration_HostMacroLifecycle(t *testing.T) {
	client := newIntegrationClient(t)
	serverID := defaultMonitoringServerID(t, client)

	name := fmt.Sprintf("go-it-host-%d", time.Now().UnixNano())
	hostID, err := client.Hosts.Create(t.Context(), &CreateHostRequest{
		MonitoringServerID: serverID,
		Name:               name,
		Address:            "127.0.0.1",
		Macros: []Macro{
			{Name: "GOITMACRO", Value: "roundtrip", IsPassword: false, Description: "integration test"},
			{Name: "GOITSECRET", Value: "s3cret", IsPassword: true, Description: "secret"},
		},
	})
	if err != nil {
		t.Fatalf("Hosts.Create: %v", err)
	}
	t.Cleanup(func() {
		if err := client.Hosts.Delete(context.Background(), hostID); err != nil {
			t.Logf("cleanup: delete host %d: %v", hostID, err)
		}
	})
	t.Logf("created host %d (%s)", hostID, name)

	h, err := client.Hosts.Get(t.Context(), hostID)
	if err != nil {
		t.Fatalf("Hosts.Get(%d): %v", hostID, err)
	}

	m := findHostMacro(h.Macros, "GOITMACRO")
	if m == nil {
		t.Fatalf("macro GOITMACRO not found in %d returned macro(s)", len(h.Macros))
	}
	wantMacroValue(t, "GOITMACRO", m.Value, "roundtrip")
	if m.IsPassword {
		t.Errorf("macro GOITMACRO is_password = true, want false")
	}
	if m.Description != "integration test" {
		t.Errorf("macro GOITMACRO description = %q, want %q", m.Description, "integration test")
	}

	secret := findHostMacro(h.Macros, "GOITSECRET")
	if secret == nil {
		t.Fatalf("macro GOITSECRET not found in %d returned macro(s)", len(h.Macros))
	}
	wantMaskedMacro(t, "GOITSECRET", secret.IsPassword, secret.Value)
	t.Logf("host macro round-trip OK: %d macro(s) returned", len(h.Macros))
}

func TestIntegration_ServiceMacroLifecycle(t *testing.T) {
	client := newIntegrationClient(t)
	serverID := defaultMonitoringServerID(t, client)
	checkCommandID, serviceTemplateID := serviceCreatePrereq(t, client)

	hostName := fmt.Sprintf("go-it-svchost-%d", time.Now().UnixNano())
	hostID, err := client.Hosts.Create(t.Context(), &CreateHostRequest{
		MonitoringServerID: serverID,
		Name:               hostName,
		Address:            "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("Hosts.Create: %v", err)
	}
	t.Cleanup(func() {
		if err := client.Hosts.Delete(context.Background(), hostID); err != nil {
			t.Logf("cleanup: delete host %d: %v", hostID, err)
		}
	})

	req := &CreateServiceRequest{
		HostID: hostID,
		Name:   fmt.Sprintf("go-it-svc-%d", time.Now().UnixNano()),
		Macros: []Macro{
			{Name: "GOITSVCMACRO", Value: "roundtrip", IsPassword: false, Description: "integration test"},
			{Name: "GOITSVCSECRET", Value: "s3cret", IsPassword: true, Description: "secret"},
		},
	}
	if checkCommandID != 0 {
		req.CheckCommandID = checkCommandID
	} else {
		req.ServiceTemplateID = serviceTemplateID
	}
	svcID, err := client.Services.Create(t.Context(), req)
	if err != nil {
		t.Fatalf("Services.Create: %v", err)
	}
	// Registered after the host cleanup so LIFO deletes the service first.
	t.Cleanup(func() {
		if err := client.Services.Delete(context.Background(), svcID); err != nil {
			t.Logf("cleanup: delete service %d: %v", svcID, err)
		}
	})
	t.Logf("created service %d (%s) on host %d", svcID, req.Name, hostID)

	svc, err := client.Services.Get(t.Context(), svcID)
	if err != nil {
		t.Fatalf("Services.Get(%d): %v", svcID, err)
	}

	m := findServiceMacro(svc.Macros, "GOITSVCMACRO")
	if m == nil {
		t.Fatalf("macro GOITSVCMACRO not found in %d returned macro(s)", len(svc.Macros))
	}
	wantMacroValue(t, "GOITSVCMACRO", m.Value, "roundtrip")
	if m.IsPassword {
		t.Errorf("macro GOITSVCMACRO is_password = true, want false")
	}
	if m.Description != "integration test" {
		t.Errorf("macro GOITSVCMACRO description = %q, want %q", m.Description, "integration test")
	}

	secret := findServiceMacro(svc.Macros, "GOITSVCSECRET")
	if secret == nil {
		t.Fatalf("macro GOITSVCSECRET not found in %d returned macro(s)", len(svc.Macros))
	}
	wantMaskedMacro(t, "GOITSVCSECRET", secret.IsPassword, secret.Value)
	t.Logf("service macro round-trip OK: %d macro(s) returned", len(svc.Macros))
}
