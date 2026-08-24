//go:build integration

package centreon

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"
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
