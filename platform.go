package centreon

import (
	"context"
	"strconv"
)

// ComponentVersion is the version of one Centreon component (the web core, or a
// module or widget) as returned by GET /platform/versions. Version, Major,
// Minor, and Fix are strings, not ints: the API sends every part as quoted JSON,
// which a numeric Go field cannot decode at all, and some parts carry a leading
// zero such as "03", which further confirms they are textual.
type ComponentVersion struct {
	Version string `json:"version"`
	Major   string `json:"major"`
	Minor   string `json:"minor"`
	Fix     string `json:"fix"`
}

// PlatformVersions is the response of GET /platform/versions. Web is the
// Centreon Web core version; Modules and Widgets are keyed by component name. On
// an install with none of them the field is an empty JSON object {} (not [] and
// not absent), so map[string]ComponentVersion decodes it to a zero-length map;
// this was live-verified against both Centreon 24.10 and 25.10. The payload
// carries no engine version.
type PlatformVersions struct {
	Web     ComponentVersion            `json:"web"`
	Modules map[string]ComponentVersion `json:"modules"`
	Widgets map[string]ComponentVersion `json:"widgets"`
}

// AtLeast reports whether the Web core version is at least major.minor. Web.Major
// and Web.Minor are strings on the wire (they may carry a leading zero, e.g.
// "03"), so they are parsed with strconv.Atoi. The comparison is major first,
// then minor only when the majors are equal; the fix (patch) level is not
// compared. It returns false if the receiver is nil, if Web.Major does not parse,
// or if the majors are equal and Web.Minor does not parse, so a caller gating a
// newer-only endpoint treats an unparseable version as too old. A strictly
// higher major is enough on its own: it returns true without needing Web.Minor.
func (v *PlatformVersions) AtLeast(major, minor int) bool {
	if v == nil {
		return false
	}
	gotMajor, err := strconv.Atoi(v.Web.Major)
	if err != nil {
		return false
	}
	if gotMajor != major {
		return gotMajor > major
	}
	gotMinor, err := strconv.Atoi(v.Web.Minor)
	if err != nil {
		return false
	}
	return gotMinor >= minor
}

// InstallationStatus is the response of GET /platform/installation/status.
type InstallationStatus struct {
	IsInstalled         bool `json:"is_installed"`
	HasUpgradeAvailable bool `json:"has_upgrade_available"`
}

// PlatformFeatures is the response of GET /platform/features. IsCloudPlatform
// reports whether this is a Centreon Cloud instance. FeatureFlags is keyed by
// feature name (for example "notification", "vault", "resource_access_management")
// with a bool for whether it is enabled. It is a map, not a fixed struct, because
// the set of feature keys is version dependent and grows across Centreon releases;
// a map decodes any key set without dropping unknown flags. Some v2 REST route
// families register only when their flag is enabled, so a consumer can call
// Features once and gate those calls with IsEnabled instead of probing for a 404.
type PlatformFeatures struct {
	IsCloudPlatform bool            `json:"is_cloud_platform"`
	FeatureFlags    map[string]bool `json:"feature_flags"`
}

// IsEnabled reports whether the named feature flag is present and true. An absent
// key returns false (the zero value of the map lookup), and a nil receiver returns
// false, so a caller gating a feature treats "unknown" as disabled. This mirrors
// PlatformVersions.AtLeast treating an unparseable version as too old.
func (f *PlatformFeatures) IsEnabled(flag string) bool {
	if f == nil {
		return false
	}
	return f.FeatureFlags[flag]
}

// PlatformService provides read-only access to Centreon platform metadata
// (GET /platform/versions and /platform/installation/status). A consumer can
// call Versions once at connect to detect the Centreon version and gate
// endpoints that exist only on newer releases, instead of inferring support
// from a 404. These endpoints live under /platform, not /configuration.
type PlatformService struct {
	client *Client
}

// Versions returns the Centreon Web core version plus the module and widget
// versions from GET /platform/versions.
func (s *PlatformService) Versions(ctx context.Context) (*PlatformVersions, error) {
	var result PlatformVersions
	if err := s.client.get(ctx, "/platform/versions", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// InstallationStatus returns the platform installation and upgrade status from
// GET /platform/installation/status.
func (s *PlatformService) InstallationStatus(ctx context.Context) (*InstallationStatus, error) {
	var result InstallationStatus
	if err := s.client.get(ctx, "/platform/installation/status", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Features returns the platform feature flags from GET /platform/features:
// whether this is a Cloud instance and which optional features are enabled. Use
// the result's IsEnabled to gate calls to feature-flagged endpoints (for example
// /configuration/notifications, which registers only when "notification" is on).
func (s *PlatformService) Features(ctx context.Context) (*PlatformFeatures, error) {
	var result PlatformFeatures
	if err := s.client.get(ctx, "/platform/features", &result); err != nil {
		return nil, err
	}
	return &result, nil
}
