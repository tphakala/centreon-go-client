//go:build integration

package centreon

import (
	"context"
	"os"
	"testing"
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
