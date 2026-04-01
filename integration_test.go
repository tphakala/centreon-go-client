//go:build integration

package centreon

import (
	"context"
	"os"
	"testing"
)

// Integration tests require a live Centreon instance.
// Run with: go test -tags integration -v ./...
//
// Required environment variables:
//   CENTREON_URL       - Base URL (e.g., https://centreon.example.com)
//   CENTREON_USERNAME  - Login username
//   CENTREON_PASSWORD  - Login password
//
// Optional:
//   CENTREON_VERSION   - API version (default: latest)
//   CENTREON_TOKEN     - Use API token instead of username/password

func newIntegrationClient(t *testing.T) *Client {
	t.Helper()

	baseURL := os.Getenv("CENTREON_URL")
	if baseURL == "" {
		t.Skip("CENTREON_URL not set, skipping integration test")
	}

	var opts []Option
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

func TestIntegration_ListHosts(t *testing.T) {
	client := newIntegrationClient(t)
	ctx := t.Context()

	resp, err := client.Hosts.List(ctx, WithLimit(5))
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
	ctx := t.Context()

	resp, err := client.Services.List(ctx, WithLimit(5))
	if err != nil {
		t.Fatalf("Services.List: %v", err)
	}
	t.Logf("Found %d services (total: %d)", len(resp.Result), resp.Meta.Total)
}

func TestIntegration_MonitoringStatus(t *testing.T) {
	client := newIntegrationClient(t)
	ctx := t.Context()

	hostCounts, err := client.MonitoringHosts.StatusCounts(ctx)
	if err != nil {
		t.Fatalf("MonitoringHosts.StatusCounts: %v", err)
	}
	t.Logf("Host status: UP=%d DOWN=%d Unreachable=%d Pending=%d",
		hostCounts.Up, hostCounts.Down, hostCounts.Unreachable, hostCounts.Pending)

	svcCounts, err := client.MonitoringServices.StatusCounts(ctx)
	if err != nil {
		t.Fatalf("MonitoringServices.StatusCounts: %v", err)
	}
	t.Logf("Service status: OK=%d Warning=%d Critical=%d Unknown=%d Pending=%d",
		svcCounts.OK, svcCounts.Warning, svcCounts.Critical, svcCounts.Unknown, svcCounts.Pending)
}

func TestIntegration_ListMonitoringServers(t *testing.T) {
	client := newIntegrationClient(t)
	ctx := t.Context()

	resp, err := client.MonitoringServers.List(ctx)
	if err != nil {
		t.Fatalf("MonitoringServers.List: %v", err)
	}
	t.Logf("Found %d monitoring servers", len(resp.Result))
	for _, s := range resp.Result {
		t.Logf("  %d: %s (default=%v)", s.ID, s.Name, s.IsDefault)
	}
}

func TestIntegration_ListTimePeriods(t *testing.T) {
	client := newIntegrationClient(t)
	ctx := t.Context()

	resp, err := client.TimePeriods.List(ctx)
	if err != nil {
		t.Fatalf("TimePeriods.List: %v", err)
	}
	t.Logf("Found %d time periods", len(resp.Result))
}

func TestIntegration_SearchFilter(t *testing.T) {
	client := newIntegrationClient(t)
	ctx := t.Context()

	// Search for hosts matching a pattern
	resp, err := client.Hosts.List(ctx,
		WithSearch(Lk("host.name", "%")),
		WithLimit(3),
	)
	if err != nil {
		t.Fatalf("Hosts.List with search: %v", err)
	}
	t.Logf("Search returned %d hosts (total: %d)", len(resp.Result), resp.Meta.Total)
}
