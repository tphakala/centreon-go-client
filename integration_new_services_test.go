//go:build integration

package centreon

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestIntegration_AuthenticationProviders reads all four providers and performs
// a safe idempotent PUT round-trip (echo the GET'd object back), proving the
// read struct is a valid full-replace write body. Secrets are left as returned;
// no real credential is written.
func TestIntegration_AuthenticationProviders(t *testing.T) {
	c := newIntegrationClient(t)
	ctx := t.Context()

	local, err := c.Authentication.GetLocal(ctx)
	if err != nil {
		t.Fatalf("GetLocal: %v", err)
	}
	if local.PasswordSecurityPolicy.PasswordMinLength <= 0 {
		t.Errorf("local password_min_length = %d, want > 0", local.PasswordSecurityPolicy.PasswordMinLength)
	}
	if err := c.Authentication.UpdateLocal(ctx, local); err != nil {
		t.Errorf("UpdateLocal (echo round-trip): %v", err)
	}

	openid, err := c.Authentication.GetOpenID(ctx)
	if err != nil {
		t.Fatalf("GetOpenID: %v", err)
	}
	if openid.AuthenticationType == "" {
		t.Error("openid authentication_type is empty, want a value")
	}
	// String must never leak a secret even if one is configured.
	if openid.ClientSecret != nil && strings.Contains(openid.String(), *openid.ClientSecret) {
		t.Error("OpenIDProvider.String leaked the client_secret")
	}
	if err := c.Authentication.UpdateOpenID(ctx, openid); err != nil {
		t.Errorf("UpdateOpenID (echo round-trip): %v", err)
	}

	saml, err := c.Authentication.GetSAML(ctx)
	if err != nil {
		t.Fatalf("GetSAML: %v", err)
	}
	if saml.Certificate != "" && strings.Contains(saml.String(), saml.Certificate) {
		t.Error("SAMLProvider.String leaked the certificate")
	}
	if err := c.Authentication.UpdateSAML(ctx, saml); err != nil {
		t.Errorf("UpdateSAML (echo round-trip): %v", err)
	}

	websso, err := c.Authentication.GetWebSSO(ctx)
	if err != nil {
		t.Fatalf("GetWebSSO: %v", err)
	}
	if err := c.Authentication.UpdateWebSSO(ctx, websso); err != nil {
		t.Errorf("UpdateWebSSO (echo round-trip): %v", err)
	}
}

// TestIntegration_Media creates a media via multipart upload, reads it back via
// Get and finds it in List, then deletes it.
func TestIntegration_Media(t *testing.T) {
	c := newIntegrationClient(t)
	ctx := t.Context()

	// 1x1 transparent PNG.
	png, err := base64.StdEncoding.DecodeString(
		"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==")
	if err != nil {
		t.Fatalf("decode png: %v", err)
	}

	dir := "gocli-itest"
	name := "gocli-itest.png"
	created, err := c.Medias.Create(ctx, &CreateMediaRequest{Filename: name, Directory: dir, Data: png})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID <= 0 {
		t.Fatalf("created media id = %d, want > 0", created.ID)
	}
	t.Cleanup(func() {
		// A fresh context is required: t.Context() is already cancelled by the time
		// t.Cleanup runs, so reusing it would make the cleanup request a no-op.
		//nolint:gocritic // intentional context.Background(): t.Context() is cancelled at cleanup time.
		if err := c.Medias.Delete(context.Background(), created.ID); err != nil {
			t.Logf("cleanup Delete(%d): %v", created.ID, err)
		}
	})

	got, err := c.Medias.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get(%d): %v", created.ID, err)
	}
	if got.Filename != name {
		t.Errorf("Get filename = %q, want %q", got.Filename, name)
	}
	if got.Directory != dir {
		t.Errorf("Get directory = %q, want %q", got.Directory, dir)
	}

	// Find it in the list; the list representation uses Name and carries a URL.
	found := findMediaByID(t, c, created.ID)
	if found == nil {
		t.Fatalf("created media id %d not found in List", created.ID)
	}
	if found.Name != name {
		t.Errorf("List name = %q, want %q", found.Name, name)
	}
	if found.URL == "" {
		t.Error("List url is empty, want the static image path")
	}

	// Folders list must contain the directory we just created.
	if !mediaFolderExists(t, c, dir) {
		t.Errorf("folder %q not found in ListFolders", dir)
	}
}

// findMediaByID scans the media list for the given id, returning a copy or nil.
func findMediaByID(t *testing.T, c *Client, id int) *Media {
	t.Helper()
	for m, err := range c.Medias.All(t.Context()) {
		if err != nil {
			t.Fatalf("All: %v", err)
		}
		if m.ID == id {
			mm := *m
			return &mm
		}
	}
	return nil
}

// mediaFolderExists reports whether a folder with the given name is listed.
func mediaFolderExists(t *testing.T, c *Client, name string) bool {
	t.Helper()
	for f, err := range c.Medias.AllFolders(t.Context()) {
		if err != nil {
			t.Fatalf("AllFolders: %v", err)
		}
		if f.Name == name {
			return true
		}
	}
	return false
}

// TestIntegration_CurrentUserUpdateParameters flips the theme via PATCH, confirms
// the change is reflected, then restores the original value.
func TestIntegration_CurrentUserUpdateParameters(t *testing.T) {
	c := newIntegrationClient(t)
	ctx := t.Context()

	before, err := c.CurrentUser.GetParameters(ctx)
	if err != nil {
		t.Fatalf("GetParameters: %v", err)
	}
	orig := before.Theme
	other := "dark"
	if orig == "dark" {
		other = "light"
	}
	t.Cleanup(func() {
		// A fresh context is required: t.Context() is already cancelled by the time
		// t.Cleanup runs, so reusing it would make the restore request a no-op.
		//nolint:gocritic // intentional context.Background(): t.Context() is cancelled at cleanup time.
		if err := c.CurrentUser.UpdateParameters(context.Background(), &UpdateCurrentUserParametersRequest{Theme: &orig}); err != nil {
			t.Logf("restore theme %q: %v", orig, err)
		}
	})

	if err := c.CurrentUser.UpdateParameters(ctx, &UpdateCurrentUserParametersRequest{Theme: &other}); err != nil {
		t.Fatalf("UpdateParameters(theme=%q): %v", other, err)
	}
	after, err := c.CurrentUser.GetParameters(ctx)
	if err != nil {
		t.Fatalf("GetParameters after update: %v", err)
	}
	if after.Theme != other {
		t.Errorf("theme after update = %q, want %q", after.Theme, other)
	}
}

// TestIntegration_ProxyRoundTrip exercises Proxy.Update against the live box and
// confirms the values land, then restores the original configuration. Update
// deliberately omits the "protocol" key (which returns HTTP 500 on 25.10.16), so
// a Get-modify-Update round-trip succeeds even though Get always returns
// protocol. This writes to the shared proxy singleton, so the original is saved
// and restored via t.Cleanup.
func TestIntegration_ProxyRoundTrip(t *testing.T) {
	c := newIntegrationClient(t)
	ctx := t.Context()

	original, err := c.Proxy.Get(ctx)
	if err != nil {
		t.Fatalf("Proxy.Get (original): %v", err)
	}
	t.Cleanup(func() {
		if err := c.Proxy.Update(context.Background(), original); err != nil {
			t.Logf("cleanup: restore proxy: %v", err)
		}
	})

	url, user, pw := "proxy.integration.example", "ituser", "it-secret-pw"
	port := 3130
	want := &ProxyConfiguration{URL: &url, Port: &port, User: &user, Password: &pw, Protocol: "http://"}
	if err := c.Proxy.Update(ctx, want); err != nil {
		t.Fatalf("Proxy.Update: %v (if this is an HTTP 500 mentioning protocol, the server defect is back; see #98)", err)
	}

	got, err := c.Proxy.Get(ctx)
	if err != nil {
		t.Fatalf("Proxy.Get (after update): %v", err)
	}
	if got.URL == nil || *got.URL != url {
		t.Errorf("URL after update = %v, want %q", got.URL, url)
	}
	if got.Port == nil || *got.Port != port {
		t.Errorf("Port after update = %v, want %d", got.Port, port)
	}
	if got.User == nil || *got.User != user {
		t.Errorf("User after update = %v, want %q", got.User, user)
	}
	if got.Password == nil || *got.Password != pw {
		t.Errorf("Password after update did not round-trip")
	}

	// A nil field is sent as JSON null and the server clears it: after setting a
	// populated url above, an Update with URL nil must read back as nil. This pins
	// the "nil clears" claim in the Update doc.
	cleared := &ProxyConfiguration{URL: nil, Port: got.Port, User: got.User, Password: got.Password}
	if err := c.Proxy.Update(ctx, cleared); err != nil {
		t.Fatalf("Proxy.Update (clear url): %v", err)
	}
	afterClear, err := c.Proxy.Get(ctx)
	if err != nil {
		t.Fatalf("Proxy.Get (after clear): %v", err)
	}
	if afterClear.URL != nil {
		t.Errorf("URL after clearing = %v, want nil (server should clear a null field)", *afterClear.URL)
	}
}

// TestIntegration_DashboardLifecycle exercises the full DashboardService surface
// against the live box: Create (with a panel and manual refresh), Get, Update,
// UpdateShares (add then clear), and Delete (via cleanup).
func TestIntegration_DashboardLifecycle(t *testing.T) {
	c := newIntegrationClient(t)
	ctx := t.Context()

	name := fmt.Sprintf("go-it-dash-%d", time.Now().UnixNano())
	create := &CreateDashboardRequest{
		Name:        name,
		Description: "GO IT dashboard",
		Panels: []DashboardPanelRequest{{
			Name:           "clock",
			Layout:         PanelLayout{X: 0, Y: 0, Width: 6, Height: 4, MinWidth: 2, MinHeight: 2},
			WidgetType:     "centreon-widget-clock",
			WidgetSettings: json.RawMessage(`{}`),
		}},
		Refresh: DashboardRefresh{Type: "manual", Interval: ptr(30)},
	}
	id, err := c.Dashboards.Create(ctx, create)
	if err != nil {
		t.Fatalf("Dashboards.Create: %v", err)
	}
	t.Cleanup(func() {
		if err := c.Dashboards.Delete(context.Background(), id); err != nil {
			t.Logf("cleanup: delete dashboard %d: %v", id, err)
		}
	})

	got, err := c.Dashboards.Get(ctx, id)
	if err != nil {
		t.Fatalf("Dashboards.Get: %v", err)
	}
	if got.ID != id {
		t.Errorf("Get ID = %d, want %d", got.ID, id)
	}
	if got.Name != name {
		t.Errorf("Get Name = %q, want %q", got.Name, name)
	}
	if len(got.Panels) != 1 {
		t.Fatalf("Get Panels = %d, want 1", len(got.Panels))
	}
	if got.Panels[0].WidgetType != "centreon-widget-clock" {
		t.Errorf("panel widget_type = %q, want centreon-widget-clock", got.Panels[0].WidgetType)
	}
	if got.Panels[0].Layout.Width != 6 {
		t.Errorf("panel layout width = %d, want 6", got.Panels[0].Layout.Width)
	}
	if got.Refresh == nil || got.Refresh.Type != "manual" {
		t.Errorf("refresh = %v, want type manual", got.Refresh)
	}

	// Update: change description (POST on the per-id path).
	update := &UpdateDashboardRequest{
		Name:        name,
		Description: "GO IT dashboard updated",
		Panels:      create.Panels,
		Refresh:     create.Refresh,
	}
	if err := c.Dashboards.Update(ctx, id, update); err != nil {
		t.Fatalf("Dashboards.Update: %v", err)
	}
	got2, err := c.Dashboards.Get(ctx, id)
	if err != nil {
		t.Fatalf("Dashboards.Get (after update): %v", err)
	}
	if got2.Description != "GO IT dashboard updated" {
		t.Errorf("description after update = %q, want updated", got2.Description)
	}

	// Shares: grant contact 1 (Admin) editor, then verify it is present.
	if err := c.Dashboards.UpdateShares(ctx, id, &UpdateDashboardSharesRequest{
		Contacts: []DashboardShareContactRequest{{ID: 1, Role: "editor"}},
	}); err != nil {
		t.Fatalf("Dashboards.UpdateShares: %v", err)
	}
	shared, err := c.Dashboards.Get(ctx, id)
	if err != nil {
		t.Fatalf("Dashboards.Get (after shares): %v", err)
	}
	if shared.Shares == nil || len(shared.Shares.Contacts) == 0 {
		t.Fatalf("Shares.Contacts empty after UpdateShares, want contact 1")
	}
	foundAdmin := false
	for _, sc := range shared.Shares.Contacts {
		if sc.ID == 1 {
			foundAdmin = true
		}
	}
	if !foundAdmin {
		t.Errorf("Shares.Contacts = %v, want to contain contact 1", shared.Shares.Contacts)
	}
}
