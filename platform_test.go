package centreon

import (
	"errors"
	"net/http"
	"testing"
)

func TestPlatformService_Versions(t *testing.T) {
	mux, c := newTestMux(t)
	// A module entry carries "minor":"03": the numeric parts are quoted strings
	// on the wire. Retyping any of them to int fails the build here (wantStr
	// takes a string) and would also abort the decode of the quoted value, so a
	// reviewer cannot silently "simplify" the fields to int.
	mux.HandleFunc("GET /centreon/api/latest/platform/versions", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"web": map[string]any{"version": "25.10.16", "major": "25", "minor": "10", "fix": "16"},
			"modules": map[string]any{
				"centreon-map4-web-client": map[string]any{"version": "25.03.0", "major": "25", "minor": "03", "fix": "0"},
			},
			"widgets": map[string]any{
				"centreon-widget-graph": map[string]any{"version": "25.10.16", "major": "25", "minor": "10", "fix": "16"},
			},
		})
	})

	got, err := c.Platform.Versions(t.Context())
	if err != nil {
		t.Fatalf("Versions: %v", err)
	}
	wantStr(t, "Web.Version", got.Web.Version, "25.10.16")
	wantStr(t, "Web.Major", got.Web.Major, "25")
	wantStr(t, "Web.Minor", got.Web.Minor, "10")
	wantStr(t, "Web.Fix", got.Web.Fix, "16")

	wantInt(t, "len(Modules)", len(got.Modules), 1)
	mod, ok := got.Modules["centreon-map4-web-client"]
	if !ok {
		t.Fatalf("Modules missing centreon-map4-web-client key: %v", got.Modules)
	}
	wantStr(t, "module.Version", mod.Version, "25.03.0")
	wantStr(t, "module.Minor", mod.Minor, "03")

	wantInt(t, "len(Widgets)", len(got.Widgets), 1)
	wid, ok := got.Widgets["centreon-widget-graph"]
	if !ok {
		t.Fatalf("Widgets missing centreon-widget-graph key: %v", got.Widgets)
	}
	wantStr(t, "widget.Version", wid.Version, "25.10.16")
}

func TestPlatformService_Versions_EmptyModules(t *testing.T) {
	mux, c := newTestMux(t)
	// A module-less install returns empty JSON objects {} (live-verified on both
	// 24.10 and 25.10), which decode to zero-length maps.
	mux.HandleFunc("GET /centreon/api/latest/platform/versions", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"web":     map[string]any{"version": "25.10.16", "major": "25", "minor": "10", "fix": "16"},
			"modules": map[string]any{},
			"widgets": map[string]any{},
		})
	})

	got, err := c.Platform.Versions(t.Context())
	if err != nil {
		t.Fatalf("Versions: %v", err)
	}
	wantStr(t, "Web.Version", got.Web.Version, "25.10.16")
	wantInt(t, "len(Modules)", len(got.Modules), 0)
	wantInt(t, "len(Widgets)", len(got.Widgets), 0)
}

func TestPlatformService_InstallationStatus(t *testing.T) {
	mux, c := newTestMux(t)
	// Asymmetric true/false so a swapped or dropped json tag is caught.
	mux.HandleFunc("GET /centreon/api/latest/platform/installation/status", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"is_installed":          true,
			"has_upgrade_available": false,
		})
	})

	got, err := c.Platform.InstallationStatus(t.Context())
	if err != nil {
		t.Fatalf("InstallationStatus: %v", err)
	}
	wantBool(t, "IsInstalled", got.IsInstalled, true)
	wantBool(t, "HasUpgradeAvailable", got.HasUpgradeAvailable, false)
}

func TestPlatformService_Versions_Error(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("GET /centreon/api/latest/platform/versions", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "boom"})
	})

	got, err := c.Platform.Versions(t.Context())
	if err == nil {
		t.Fatal("expected an error on HTTP 500, got nil")
	}
	if got != nil {
		t.Errorf("result = %v, want nil on error", got)
	}
	if _, ok := errors.AsType[*APIError](err); !ok {
		t.Errorf("expected *APIError, got %T", err)
	}
}

func TestPlatformService_InstallationStatus_Error(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("GET /centreon/api/latest/platform/installation/status", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "boom"})
	})

	got, err := c.Platform.InstallationStatus(t.Context())
	if err == nil {
		t.Fatal("expected an error on HTTP 500, got nil")
	}
	if got != nil {
		t.Errorf("result = %v, want nil on error", got)
	}
	if _, ok := errors.AsType[*APIError](err); !ok {
		t.Errorf("expected *APIError, got %T", err)
	}
}

func TestPlatformService_Features(t *testing.T) {
	mux, c := newTestMux(t)
	// Mixed true/false flags so a swapped/dropped json tag or a struct-vs-map
	// mistake is caught; is_cloud_platform:false is asymmetric to the true flags.
	mux.HandleFunc("GET /centreon/api/latest/platform/features", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"is_cloud_platform": false,
			"feature_flags": map[string]any{
				"notification":               true,
				"vault":                      false,
				"resource_access_management": true,
			},
		})
	})

	got, err := c.Platform.Features(t.Context())
	if err != nil {
		t.Fatalf("Features: %v", err)
	}
	wantBool(t, "IsCloudPlatform", got.IsCloudPlatform, false)
	wantInt(t, "len(FeatureFlags)", len(got.FeatureFlags), 3)
	wantBool(t, `IsEnabled("notification")`, got.IsEnabled("notification"), true)
	wantBool(t, `IsEnabled("vault")`, got.IsEnabled("vault"), false)
	wantBool(t, `IsEnabled("resource_access_management")`, got.IsEnabled("resource_access_management"), true)
	wantBool(t, `IsEnabled("absent")`, got.IsEnabled("absent"), false)
}

func TestPlatformFeatures_IsEnabled(t *testing.T) {
	f := &PlatformFeatures{FeatureFlags: map[string]bool{"on": true, "off": false}}
	tests := []struct {
		name string
		pf   *PlatformFeatures
		flag string
		want bool
	}{
		{"present true", f, "on", true},
		{"present false", f, "off", false},
		{"absent key", f, "missing", false},
		{"nil receiver", nil, "on", false},
		{"nil feature map", &PlatformFeatures{}, "on", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pf.IsEnabled(tt.flag); got != tt.want {
				t.Errorf("IsEnabled(%q) = %v, want %v", tt.flag, got, tt.want)
			}
		})
	}
}

func TestPlatformService_Features_Error(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("GET /centreon/api/latest/platform/features", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "boom"})
	})

	got, err := c.Platform.Features(t.Context())
	if err == nil {
		t.Fatal("expected an error on HTTP 500, got nil")
	}
	if got != nil {
		t.Errorf("result = %v, want nil on error", got)
	}
	if _, ok := errors.AsType[*APIError](err); !ok {
		t.Errorf("expected *APIError, got %T", err)
	}
}

func TestPlatformVersions_AtLeast(t *testing.T) {
	v := func(major, minor string) *PlatformVersions {
		return &PlatformVersions{Web: ComponentVersion{Major: major, Minor: minor}}
	}
	tests := []struct {
		name         string
		pv           *PlatformVersions
		major, minor int
		want         bool
	}{
		{"exact", v("25", "10"), 25, 10, true},
		{"same major higher minor wanted", v("25", "10"), 25, 11, false},
		{"same major lower minor wanted", v("25", "10"), 25, 9, true},
		{"higher major", v("26", "0"), 25, 10, true},
		{"lower major", v("24", "10"), 25, 10, false},
		{"leading zero minor parses", v("25", "03"), 25, 3, true},
		{"leading zero minor too old", v("25", "03"), 25, 4, false},
		{"unparseable major is too old", v("", "10"), 25, 10, false},
		{"unparseable minor is too old", v("25", "x"), 25, 10, false},
		{"higher major wins even with unparseable minor", v("26", "x"), 25, 10, true},
		{"lower major loses even with unparseable minor", v("24", "x"), 25, 10, false},
		{"nil receiver is too old", nil, 25, 10, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pv.AtLeast(tt.major, tt.minor); got != tt.want {
				t.Errorf("AtLeast(%d, %d) = %v, want %v", tt.major, tt.minor, got, tt.want)
			}
		})
	}
}
