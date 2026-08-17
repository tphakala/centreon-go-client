//go:build integration

package centreon

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
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

// TestIntegration_MonitoringServerLastRestart verifies the last_restart field
// added for issue #48 against a live instance. The decode is pinned offline by
// TestMonitoringServer_LastRestartShapes; this test confirms the live RFC3339
// wire value parses into a real timestamp on a restarted poller. last_restart
// is null on a never-restarted poller, so the test skips if no poller reports a
// non-null value.
func TestIntegration_MonitoringServerLastRestart(t *testing.T) {
	client := newIntegrationClient(t)

	resp, err := client.MonitoringServers.List(t.Context())
	if err != nil {
		// Skip only when the token genuinely lacks permission (401/403); a
		// decode or transport error must fail, since catching a live decode
		// regression on last_restart is the point of this test.
		if apiErr, ok := errors.AsType[*APIError](err); ok &&
			(apiErr.HTTPStatus == http.StatusUnauthorized || apiErr.HTTPStatus == http.StatusForbidden) {
			t.Skipf("MonitoringServers.List: %v (token lacks permission)", err)
		}
		t.Fatalf("MonitoringServers.List: %v", err)
	}
	if len(resp.Result) == 0 {
		t.Skip("no monitoring servers returned")
	}

	var found bool
	for _, s := range resp.Result {
		if s.LastRestart == nil {
			t.Logf("poller %d (%s): last_restart null (engine not yet restarted)", s.ID, s.Name)
			continue
		}
		found = true
		if s.LastRestart.IsZero() {
			t.Errorf("poller %d (%s): last_restart present but parsed to the zero time", s.ID, s.Name)
		}
		t.Logf("poller %d (%s): last_restart = %s", s.ID, s.Name, s.LastRestart.Format(time.RFC3339))
	}
	if !found {
		t.Skip("no poller reported a non-null last_restart (no engine restarted); wire value unverified live this run")
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

// --- Operations (bulk monitoring actions) body-format validation ---

// syntheticResourceID is an assumed-nonexistent host ID used by the operations
// body-format tests. Centreon assigns small sequential host IDs, so this large
// sentinel is far above any real ID and will not collide on a live instance; on
// a fresh box (0 configured hosts) nothing exists at all. The operations
// therefore act on no real resource: acknowledge/check/comments no-op (204) and
// downtime/submit fail the resource lookup (404).
const syntheticResourceID = 999000001

// schemaValidationMarkers are lowercased substrings that the Centreon JSON
// schema validator emits when a request body has the wrong shape. They appear
// in the HTTP 500 response to a malformed operations body, e.g.
// "[acknowledgement] The property acknowledgement is required / The property
// comment is not defined and the definition does not allow additional
// properties". A resource-not-found body ("Host not found") contains none of
// them, so a 404 is not misclassified as a schema error.
var schemaValidationMarkers = []string{
	"is required",
	"additional properties",
	"not defined",
	"does not allow",
}

// isSchemaValidationError reports whether apiErr is a request-body JSON schema
// rejection: the signal that a bulk-operation body has the wrong shape. Only
// the negative control uses it, to confirm that a deliberately malformed body
// is rejected for the expected reason. The positive assertion below deliberately
// does NOT depend on message matching (see assertBodyFormatAccepted).
func isSchemaValidationError(apiErr *APIError) bool {
	switch apiErr.HTTPStatus {
	case http.StatusBadRequest, http.StatusUnprocessableEntity, http.StatusInternalServerError:
		msg := strings.ToLower(apiErr.Message)
		for _, marker := range schemaValidationMarkers {
			if strings.Contains(msg, marker) {
				return true
			}
		}
	}
	return false
}

// assertBodyFormatAccepted verifies that the live API accepted the request body
// STRUCTURE that an OperationsService method produced. A well-formed request is
// either applied (HTTP 204 -> nil error) or passes schema validation and then
// fails the resource-existence lookup for the absent synthetic target (HTTP
// 404). Any other status fails the test: 401/403 as an auth/permission error
// that blocks validation (fatal), and everything else (a 400/422/500 schema
// rejection) as a wrong nested-vs-flat wire format. Keying on 204/404 rather
// than on matching a schema-error message avoids a false pass when a broken
// body yields an error message this test does not recognize.
func assertBodyFormatAccepted(t *testing.T, op string, err error) {
	t.Helper()
	if err == nil {
		t.Logf("%s: body accepted (HTTP 204 no-op)", op)
		return
	}
	apiErr, ok := errors.AsType[*APIError](err)
	if !ok {
		t.Fatalf("%s: non-API error (transport/network?): %v", op, err)
	}
	switch apiErr.HTTPStatus {
	case http.StatusNotFound:
		// Body passed schema validation; only the resource lookup failed
		// (expected on an engine-less box where nothing is monitored).
		t.Logf("%s: body accepted by schema, resource lookup failed as expected (HTTP 404): %s",
			op, apiErr.Message)
	case http.StatusUnauthorized, http.StatusForbidden:
		t.Fatalf("%s: auth/permission error (HTTP %d): %s; cannot validate body format",
			op, apiErr.HTTPStatus, apiErr.Message)
	default:
		t.Errorf("%s: body rejected (HTTP %d): %s; want 204 or 404, so the nested-vs-flat wire format is wrong",
			op, apiErr.HTTPStatus, apiErr.Message)
	}
}

// TestIntegration_OperationsBodyFormat validates that the five OperationsService
// methods send the request body the Centreon API expects, answering the open
// nested-vs-flat question in issue #13. Three of the five nest their parameters
// in a sibling object next to the resource list (acknowledge -> "acknowledgement",
// check -> "check", downtime -> "downtime"); the other two embed their
// parameters into each element of the "resources" array (submit -> status/output,
// comments -> comment/date). In every case the parameters are NOT flat at the
// top level, which is the shape a flat (wrong) body would use.
//
// Verified live behavior (local podman Centreon Web 25.10.16, 2026-08-17), with
// the synthetic target absent (0 configured hosts on a fresh instance):
//   - acknowledge, check, comments: the correct body returns HTTP 204 (accepted,
//     applied as a no-op against the absent resource).
//   - downtime, submit: the correct body returns HTTP 404 "Host not found". The
//     server validates the request body schema before it resolves the target
//     resource, so a 404 (resource missing) means the body was schema-accepted.
//     That ordering is proven by TestIntegration_OperationsRejectsMalformedBody,
//     which shows a malformed body on these same endpoints is rejected with an
//     HTTP 500 schema error, not a 404.
//
// The test asserts only that the body STRUCTURE is accepted (204 or 404, both
// portable across instances), not that downtime and submit are applied
// end-to-end against a monitored resource (that needs a running monitoring
// engine, which a fresh box does not have).
func TestIntegration_OperationsBodyFormat(t *testing.T) {
	client := newIntegrationClient(t)
	uniq := fmt.Sprintf("go-it-bodyfmt-%d", time.Now().UnixNano())
	ref := ResourceRef{Type: "host", ID: syntheticResourceID}

	t.Run("acknowledge", func(t *testing.T) {
		err := client.Operations.Acknowledge(t.Context(), &AcknowledgeRequest{
			Resources: []ResourceRef{ref},
			Comment:   uniq,
			IsSticky:  true,
		})
		assertBodyFormatAccepted(t, "acknowledge", err)
	})

	t.Run("check", func(t *testing.T) {
		err := client.Operations.Check(t.Context(), &CheckRequest{
			Resources: []ResourceRef{ref},
		})
		assertBodyFormatAccepted(t, "check", err)
	})

	t.Run("comment", func(t *testing.T) {
		err := client.Operations.Comment(t.Context(), &CommentRequest{
			Resources: []ResourceRef{ref},
			Comment:   uniq,
		})
		assertBodyFormatAccepted(t, "comment", err)
	})

	t.Run("downtime", func(t *testing.T) {
		now := time.Now()
		err := client.Operations.Downtime(t.Context(), &DowntimeRequest{
			Resources: []ResourceRef{ref},
			Comment:   uniq,
			StartTime: now,
			EndTime:   now.Add(time.Hour),
			Fixed:     true,
			Duration:  3600,
		})
		assertBodyFormatAccepted(t, "downtime", err)
	})

	t.Run("submit", func(t *testing.T) {
		err := client.Operations.Submit(t.Context(), &SubmitResultRequest{
			Resources: []SubmitResource{{
				Type:   "host",
				ID:     syntheticResourceID,
				Status: 0,
				Output: uniq,
			}},
		})
		assertBodyFormatAccepted(t, "submit", err)
	})
}

// TestIntegration_OperationsRejectsMalformedBody is the negative control for
// TestIntegration_OperationsBodyFormat: it proves the 204/404 acceptance signal
// is not vacuous by sending a deliberately malformed body to each of the three
// wire shapes and asserting the live API rejects it as a JSON schema error. It
// deliberately covers the two operations whose accept signal is a 404 (downtime
// and submit), so a malformed body on those endpoints is shown to fail with a
// schema error rather than the same 404 the positive test treats as accepted.
// The three malformed shapes:
//   - acknowledge: the "acknowledgement" sibling wrapper removed, its fields
//     hoisted to the top level.
//   - downtime: the "downtime" sibling wrapper removed, its fields hoisted.
//   - submit: the per-resource status/output hoisted to the top level instead
//     of embedded in each "resources" element.
//
// Verified live (Centreon 25.10.16, 2026-08-17): all three return HTTP 500 with
// a schema message, which confirms the server validates the request body schema
// before it resolves the target resource.
func TestIntegration_OperationsRejectsMalformedBody(t *testing.T) {
	client := newIntegrationClient(t)

	ref := map[string]any{"type": "host", "id": syntheticResourceID, "parent": nil}
	now := time.Now().UTC()
	cases := []struct {
		name string
		path string
		body map[string]any
	}{
		{
			name: "acknowledge_flat",
			path: "/monitoring/resources/acknowledge",
			body: map[string]any{
				"resources":             []map[string]any{ref},
				"comment":               "go-it-malformed-negative-control",
				"is_notify_contacts":    false,
				"is_persistent_comment": false,
				"is_sticky":             true,
			},
		},
		{
			name: "downtime_flat",
			path: "/monitoring/resources/downtime",
			body: map[string]any{
				"resources":  []map[string]any{ref},
				"comment":    "go-it-malformed-negative-control",
				"start_time": now.Format(time.RFC3339),
				"end_time":   now.Add(time.Hour).Format(time.RFC3339),
				"is_fixed":   true,
				"duration":   0,
			},
		},
		{
			name: "submit_hoisted",
			path: "/monitoring/resources/submit",
			body: map[string]any{
				"resources": []map[string]any{ref},
				"status":    0,
				"output":    "go-it-malformed-negative-control",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// client.post runs the full request path (auth header, 401 re-auth,
			// body close) and returns the parsed *APIError.
			err := client.post(t.Context(), tc.path, tc.body, nil)
			if err == nil {
				t.Fatalf("%s: malformed body was ACCEPTED (nil error); the accept signal is vacuous", tc.name)
			}
			apiErr, ok := errors.AsType[*APIError](err)
			if !ok {
				t.Fatalf("%s: expected *APIError, got %v", tc.name, err)
			}
			if !isSchemaValidationError(apiErr) {
				t.Errorf("%s: malformed body rejected but not as a schema error (HTTP %d): %s; expected a JSON schema validation failure",
					tc.name, apiErr.HTTPStatus, apiErr.Message)
			}
			t.Logf("%s: correctly rejected as schema error (HTTP %d): %s", tc.name, apiErr.HTTPStatus, apiErr.Message)
		})
	}
}

// --- Configuration CRUD lifecycle for #57 (commands) and #58 (service
// categories / groups). Categories and groups have a delete route, so they
// round-trip Create -> Get -> Update -> Get and clean up. Commands have no
// per-id route on 25.10 (no delete), so that test is reuse-safe instead. ---

func TestIntegration_ServiceCategoryLifecycle(t *testing.T) {
	client := newIntegrationClient(t)

	name := fmt.Sprintf("go-it-svccat-%d", time.Now().UnixNano())
	id, err := client.ServiceCategories.Create(t.Context(), CreateServiceCategoryRequest{
		Name:  name,
		Alias: "GO IT category",
	})
	if err != nil {
		t.Fatalf("ServiceCategories.Create: %v", err)
	}
	t.Cleanup(func() {
		if err := client.ServiceCategories.Delete(context.Background(), id); err != nil {
			t.Logf("cleanup: delete service category %d: %v", id, err)
		}
	})

	cat, err := client.ServiceCategories.Get(t.Context(), id)
	if err != nil {
		t.Fatalf("ServiceCategories.Get: %v", err)
	}
	if cat.ID != id {
		t.Errorf("Get ID = %d, want %d", cat.ID, id)
	}
	if cat.Name != name {
		t.Errorf("Get Name = %q, want %q", cat.Name, name)
	}

	const newAlias = "GO IT category updated"
	if err := client.ServiceCategories.Update(t.Context(), id, UpdateServiceCategoryRequest{
		Name:  name,
		Alias: newAlias,
	}); err != nil {
		t.Fatalf("ServiceCategories.Update: %v", err)
	}

	cat2, err := client.ServiceCategories.Get(t.Context(), id)
	if err != nil {
		t.Fatalf("ServiceCategories.Get (after update): %v", err)
	}
	if cat2.Alias != newAlias {
		t.Errorf("Alias after update = %q, want %q", cat2.Alias, newAlias)
	}
}

func TestIntegration_ServiceGroupLifecycle(t *testing.T) {
	client := newIntegrationClient(t)

	name := fmt.Sprintf("go-it-svcgrp-%d", time.Now().UnixNano())
	id, err := client.ServiceGroups.Create(t.Context(), CreateServiceGroupRequest{
		Name:  name,
		Alias: "GO IT group",
	})
	if err != nil {
		t.Fatalf("ServiceGroups.Create: %v", err)
	}
	t.Cleanup(func() {
		if err := client.ServiceGroups.Delete(context.Background(), id); err != nil {
			t.Logf("cleanup: delete service group %d: %v", id, err)
		}
	})

	sg, err := client.ServiceGroups.Get(t.Context(), id)
	if err != nil {
		t.Fatalf("ServiceGroups.Get: %v", err)
	}
	if sg.ID != id {
		t.Errorf("Get ID = %d, want %d", sg.ID, id)
	}
	if sg.Name != name {
		t.Errorf("Get Name = %q, want %q", sg.Name, name)
	}

	const newAlias = "GO IT group updated"
	if err := client.ServiceGroups.Update(t.Context(), id, UpdateServiceGroupRequest{
		Name:  name,
		Alias: newAlias,
	}); err != nil {
		t.Fatalf("ServiceGroups.Update: %v", err)
	}

	sg2, err := client.ServiceGroups.Get(t.Context(), id)
	if err != nil {
		t.Fatalf("ServiceGroups.Get (after update): %v", err)
	}
	if sg2.Alias != newAlias {
		t.Errorf("Alias after update = %q, want %q", sg2.Alias, newAlias)
	}
}

func TestIntegration_CommandCreate(t *testing.T) {
	client := newIntegrationClient(t)

	// Commands have no delete route on 25.10, so this test reuses a stable name
	// rather than creating a fresh object every run.
	const name = "go-it-check-command"

	existing, err := client.Commands.List(t.Context(), WithSearch(Eq("name", name)))
	if err != nil {
		t.Fatalf("Commands.List: %v", err)
	}
	for i := range existing.Result {
		if existing.Result[i].Name == name {
			t.Logf("command %q already exists (id %d, type %d); reusing", name, existing.Result[i].ID, existing.Result[i].Type)
			return
		}
	}

	cmd, err := client.Commands.Create(t.Context(), CreateCommandRequest{
		Name:        name,
		Type:        2, // check command
		CommandLine: "/usr/lib/nagios/plugins/check_ping -H $HOSTADDRESS$ -w 100,20% -c 500,60% -p 5",
	})
	if err != nil {
		t.Fatalf("Commands.Create: %v", err)
	}
	if cmd.ID == 0 {
		t.Fatal("Create returned a command with zero ID")
	}
	if cmd.Name != name {
		t.Errorf("Name = %q, want %q", cmd.Name, name)
	}
	if cmd.Type != 2 {
		t.Errorf("Type = %d, want 2", cmd.Type)
	}

	back, err := client.Commands.List(t.Context(), WithSearch(Eq("name", name)))
	if err != nil {
		t.Fatalf("Commands.List (readback): %v", err)
	}
	found := false
	for i := range back.Result {
		if back.Result[i].Name == name && back.Result[i].Type == 2 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("created command %q (type 2) not found on readback", name)
	}
}
