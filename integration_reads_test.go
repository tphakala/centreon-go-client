//go:build integration

package centreon

import (
	"strings"
	"testing"
)

// Integration tests for the read-only configuration/platform/admin endpoints
// added for issues #87, #98, #99, #105, #106. Each exercises the live decode
// path against a real Centreon instance.

func TestIntegration_PlatformFeatures(t *testing.T) {
	client := newIntegrationClient(t)
	f, err := client.Platform.Features(t.Context())
	if err != nil {
		t.Fatalf("Platform.Features: %v", err)
	}
	// The feature-flag map is always present (possibly empty). IsEnabled must be
	// nil-safe and consistent with the map contents.
	t.Logf("is_cloud_platform=%v feature_flags=%v", f.IsCloudPlatform, f.FeatureFlags)
	for name, enabled := range f.FeatureFlags {
		if got := f.IsEnabled(name); got != enabled {
			t.Errorf("IsEnabled(%q) = %v, want %v (map value)", name, got, enabled)
		}
	}
	if f.IsEnabled("this-flag-does-not-exist") {
		t.Error("IsEnabled returned true for an absent flag")
	}
}

func TestIntegration_AdministrationParameters(t *testing.T) {
	client := newIntegrationClient(t)
	p, err := client.AdministrationParameters.Get(t.Context())
	if err != nil {
		t.Fatalf("AdministrationParameters.Get: %v", err)
	}
	// Refresh intervals are positive on any real instance; a zero here would
	// signal a wire-key mismatch (all fields decoding to their zero value).
	if p.MonitoringDefaultRefreshInterval <= 0 {
		t.Errorf("MonitoringDefaultRefreshInterval = %d, want > 0 (possible tag mismatch)", p.MonitoringDefaultRefreshInterval)
	}
	t.Logf("downtime_duration=%d refresh=%d", p.MonitoringDefaultDowntimeDuration, p.MonitoringDefaultRefreshInterval)
}

func TestIntegration_GraphTemplates(t *testing.T) {
	client := newIntegrationClient(t)
	resp, err := client.GraphTemplates.List(t.Context())
	if err != nil {
		t.Fatalf("GraphTemplates.List: %v", err)
	}
	if len(resp.Result) == 0 {
		t.Skip("no graph templates on this instance")
	}
	// A stock instance ships named templates; assert more than just Name so a
	// renamed/retyped grid/base/id key surfaces instead of silently zeroing.
	for i := range resp.Result {
		g := resp.Result[i]
		if g.ID == 0 {
			t.Errorf("graph template %q has zero ID (possible tag mismatch)", g.Name)
		}
		if g.Name == "" {
			t.Errorf("graph template %d has empty Name (possible tag mismatch)", g.ID)
		}
		if g.Base <= 0 {
			t.Errorf("graph template %q has Base=%d, want > 0 (possible grid/base tag mismatch)", g.Name, g.Base)
		}
	}
	t.Logf("graph templates: %d", len(resp.Result))
}

func TestIntegration_Proxy(t *testing.T) {
	client := newIntegrationClient(t)
	p, err := client.Proxy.Get(t.Context())
	if err != nil {
		t.Fatalf("Proxy.Get: %v", err)
	}
	// Protocol always decodes to a non-empty default (e.g. "http://").
	if p.Protocol == "" {
		t.Error("Proxy.Protocol is empty, want a default such as http://")
	}
	// Formatting the config must never leak a configured password.
	if s := p.String(); p.Password != nil && strings.Contains(s, *p.Password) {
		t.Error("Proxy.String() leaked the configured password")
	}
	t.Logf("proxy: %s", p)
}

func TestIntegration_CurrentUser(t *testing.T) {
	client := newIntegrationClient(t)

	params, err := client.CurrentUser.GetParameters(t.Context())
	if err != nil {
		t.Fatalf("CurrentUser.GetParameters: %v", err)
	}
	if params.Name == "" {
		t.Error("CurrentUser.GetParameters returned empty Name (possible tag mismatch)")
	}
	t.Logf("current user: id=%d name=%q admin=%v", params.ID, params.Name, params.IsAdmin)

	actions, err := client.CurrentUser.GetACLActions(t.Context())
	if err != nil {
		t.Fatalf("CurrentUser.GetACLActions: %v", err)
	}
	t.Logf("acl actions host.check=%v service.comment=%v", actions.Host.Check, actions.Service.Comment)

	perms, err := client.CurrentUser.GetACLPermissions(t.Context())
	if err != nil {
		t.Fatalf("CurrentUser.GetACLPermissions: %v", err)
	}
	t.Logf("acl permissions: %d keys", len(perms))
}
