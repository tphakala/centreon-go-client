package centreon

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestCurrentUserService_GetParameters(t *testing.T) {
	mux, c := newTestMux(t)
	// default_page is null here (must decode to a nil *string). Asymmetric bools
	// (is_admin true, use_deprecated_pages false) and the nested dashboard object
	// catch a dropped or swapped tag.
	mux.HandleFunc("GET /centreon/api/latest/configuration/users/current/parameters", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"id": 1, "name": "Admin Centreon", "alias": "admin", "email": "admin@centreon.local",
			"timezone": "UTC", "locale": "en_US", "is_admin": true,
			// use_deprecated_pages and use_deprecated_custom_views are asymmetric
			// (false vs true) so a swap between these sibling tags is caught.
			"use_deprecated_pages": false, "use_deprecated_custom_views": true,
			"is_export_button_enabled": true, "can_manage_api_tokens": true,
			"theme": "light", "user_interface_density": "compact", "default_page": nil,
			"dashboard": map[string]any{
				"global_user_role": "administrator", "view_dashboards": true,
				"create_dashboards": true, "administrate_dashboards": false,
			},
		})
	})

	got, err := c.CurrentUser.GetParameters(t.Context())
	if err != nil {
		t.Fatalf("GetParameters: %v", err)
	}
	wantInt(t, "ID", got.ID, 1)
	wantStr(t, "Name", got.Name, "Admin Centreon")
	wantStr(t, "Alias", got.Alias, "admin")
	wantStr(t, "Email", got.Email, "admin@centreon.local")
	wantStr(t, "Timezone", got.Timezone, "UTC")
	wantStr(t, "Locale", got.Locale, "en_US")
	wantBool(t, "IsAdmin", got.IsAdmin, true)
	wantBool(t, "UseDeprecatedPages", got.UseDeprecatedPages, false)
	wantBool(t, "UseDeprecatedCustomViews", got.UseDeprecatedCustomViews, true)
	wantBool(t, "IsExportButtonEnabled", got.IsExportButtonEnabled, true)
	wantBool(t, "CanManageAPITokens", got.CanManageAPITokens, true)
	wantStr(t, "Theme", got.Theme, "light")
	wantStr(t, "UserInterfaceDensity", got.UserInterfaceDensity, "compact")
	wantNilStrPtr(t, "DefaultPage", got.DefaultPage)
	wantStr(t, "Dashboard.GlobalUserRole", got.Dashboard.GlobalUserRole, "administrator")
	wantBool(t, "Dashboard.ViewDashboards", got.Dashboard.ViewDashboards, true)
	wantBool(t, "Dashboard.CreateDashboards", got.Dashboard.CreateDashboards, true)
	wantBool(t, "Dashboard.AdministrateDashboards", got.Dashboard.AdministrateDashboards, false)
}

func TestCurrentUserService_GetParameters_DefaultPageSet(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("GET /centreon/api/latest/configuration/users/current/parameters", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"id": 2, "name": "u", "alias": "u", "email": "u@x", "timezone": "UTC",
			"locale": "en_US", "default_page": "/monitoring/resources",
			"dashboard": map[string]any{"global_user_role": "viewer"},
		})
	})

	got, err := c.CurrentUser.GetParameters(t.Context())
	if err != nil {
		t.Fatalf("GetParameters: %v", err)
	}
	wantStrPtr(t, "DefaultPage", got.DefaultPage, "/monitoring/resources")
}

func TestCurrentUserService_GetACLActions(t *testing.T) {
	mux, c := newTestMux(t)
	// Within each set the bools are asymmetric (one false) so a dropped tag on a
	// specific action fails; the three top-level object types differ too.
	mux.HandleFunc("GET /centreon/api/latest/users/acl/actions", func(w http.ResponseWriter, _ *http.Request) {
		// Within each set every adjacent pair is asymmetric (acknowledgement true
		// vs disacknowledgement false, downtime false vs submit_status true) so a
		// swap between any two sibling action tags is caught.
		set := func(comment bool) map[string]any {
			return map[string]any{
				"check": true, "forced_check": true, "acknowledgement": true,
				"disacknowledgement": false, "downtime": false, "submit_status": true,
				"comment": comment,
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"host":        set(true),
			"service":     set(false),
			"metaservice": set(true),
		})
	})

	got, err := c.CurrentUser.GetACLActions(t.Context())
	if err != nil {
		t.Fatalf("GetACLActions: %v", err)
	}
	wantBool(t, "Host.Check", got.Host.Check, true)
	wantBool(t, "Host.ForcedCheck", got.Host.ForcedCheck, true)
	wantBool(t, "Host.Acknowledgement", got.Host.Acknowledgement, true)
	wantBool(t, "Host.Disacknowledgement", got.Host.Disacknowledgement, false)
	wantBool(t, "Host.Downtime", got.Host.Downtime, false)
	wantBool(t, "Host.Comment", got.Host.Comment, true)
	wantBool(t, "Service.Comment", got.Service.Comment, false)
	wantBool(t, "Metaservice.SubmitStatus", got.Metaservice.SubmitStatus, true)
	wantBool(t, "Metaservice.Comment", got.Metaservice.Comment, true)
}

func TestCurrentUserService_GetACLPermissions(t *testing.T) {
	mux, c := newTestMux(t)
	// Sparse, instance-dependent key set -> map. Only granted permissions appear;
	// an absent key means not granted.
	mux.HandleFunc("GET /centreon/api/latest/users/acl/permissions", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"top_counter":                    true,
			"poller_statistics":              true,
			"configuration_host_group_write": false,
		})
	})

	got, err := c.CurrentUser.GetACLPermissions(t.Context())
	if err != nil {
		t.Fatalf("GetACLPermissions: %v", err)
	}
	wantInt(t, "len(permissions)", len(got), 3)
	wantBool(t, `permissions["top_counter"]`, got["top_counter"], true)
	wantBool(t, `permissions["configuration_host_group_write"]`, got["configuration_host_group_write"], false)
	wantBool(t, `permissions["absent"]`, got["absent"], false)
}

func TestCurrentUserService_UpdateParameters(t *testing.T) {
	mux, c := newTestMux(t)
	var gotBody string
	// The route is registered for PATCH only; calling any other verb 404s here,
	// which pins that UpdateParameters uses PATCH.
	mux.HandleFunc("PATCH /centreon/api/latest/configuration/users/current/parameters", func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusNoContent)
	})

	theme := "dark"
	if err := c.CurrentUser.UpdateParameters(t.Context(), &UpdateCurrentUserParametersRequest{Theme: &theme}); err != nil {
		t.Fatalf("UpdateParameters: %v", err)
	}
	// Only the set field is sent: theme present, user_interface_density omitted.
	// Removing ,omitempty from UserInterfaceDensity would serialize a null and
	// fail the absence assertion.
	if !strings.Contains(gotBody, `"theme":"dark"`) {
		t.Errorf("PATCH body = %s, want it to contain \"theme\":\"dark\"", gotBody)
	}
	if strings.Contains(gotBody, "user_interface_density") {
		t.Errorf("PATCH body = %s, want it to omit the unset user_interface_density", gotBody)
	}
}

func TestCurrentUserService_UpdateParameters_BothFields(t *testing.T) {
	mux, c := newTestMux(t)
	var gotBody string
	mux.HandleFunc("PATCH /centreon/api/latest/configuration/users/current/parameters", func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusNoContent)
	})

	theme, density := "light", "detailed"
	req := &UpdateCurrentUserParametersRequest{Theme: &theme, UserInterfaceDensity: &density}
	if err := c.CurrentUser.UpdateParameters(t.Context(), req); err != nil {
		t.Fatalf("UpdateParameters: %v", err)
	}
	if !strings.Contains(gotBody, `"theme":"light"`) {
		t.Errorf("PATCH body = %s, want \"theme\":\"light\"", gotBody)
	}
	if !strings.Contains(gotBody, `"user_interface_density":"detailed"`) {
		t.Errorf("PATCH body = %s, want \"user_interface_density\":\"detailed\"", gotBody)
	}
}

func TestCurrentUserService_UpdateParameters_Error(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("PATCH /centreon/api/latest/configuration/users/current/parameters", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "boom"})
	})

	theme := "dark"
	err := c.CurrentUser.UpdateParameters(t.Context(), &UpdateCurrentUserParametersRequest{Theme: &theme})
	if err == nil {
		t.Fatal("expected an error on HTTP 500, got nil")
	}
	if _, ok := errors.AsType[*APIError](err); !ok {
		t.Errorf("expected *APIError, got %T", err)
	}
}

func TestCurrentUserService_Errors(t *testing.T) {
	// Each call returns whether its own result is nil (checked inside the closure
	// where the concrete type is known: a typed-nil pointer or nil map boxed into
	// an `any` here would be a non-nil interface, so the nil check must happen on
	// the concrete value). On error every method must return a nil result.
	tests := []struct {
		name  string
		route string
		call  func(c *Client) (resultIsNil bool, err error)
	}{
		{"GetParameters", "GET /centreon/api/latest/configuration/users/current/parameters",
			func(c *Client) (bool, error) {
				r, err := c.CurrentUser.GetParameters(t.Context())
				return r == nil, err
			}},
		{"GetACLActions", "GET /centreon/api/latest/users/acl/actions",
			func(c *Client) (bool, error) {
				r, err := c.CurrentUser.GetACLActions(t.Context())
				return r == nil, err
			}},
		{"GetACLPermissions", "GET /centreon/api/latest/users/acl/permissions",
			func(c *Client) (bool, error) {
				r, err := c.CurrentUser.GetACLPermissions(t.Context())
				return r == nil, err
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux, c := newTestMux(t)
			mux.HandleFunc(tt.route, func(w http.ResponseWriter, _ *http.Request) {
				writeJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "boom"})
			})
			resultIsNil, err := tt.call(c)
			if err == nil {
				t.Fatal("expected an error on HTTP 500, got nil")
			}
			if _, ok := errors.AsType[*APIError](err); !ok {
				t.Errorf("expected *APIError, got %T", err)
			}
			if !resultIsNil {
				t.Error("expected a nil result on error")
			}
		})
	}
}
