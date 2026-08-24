package centreon

import "context"

// CurrentUserParameters is the profile and UI preferences of the authenticated
// user, returned by GET /configuration/users/current/parameters. DefaultPage is
// nullable (JSON null when the user has no default page), so it is a pointer.
type CurrentUserParameters struct {
	ID                       int                  `json:"id"`
	Name                     string               `json:"name"`
	Alias                    string               `json:"alias"`
	Email                    string               `json:"email"`
	Timezone                 string               `json:"timezone"`
	Locale                   string               `json:"locale"`
	IsAdmin                  bool                 `json:"is_admin"`
	UseDeprecatedPages       bool                 `json:"use_deprecated_pages"`
	UseDeprecatedCustomViews bool                 `json:"use_deprecated_custom_views"`
	IsExportButtonEnabled    bool                 `json:"is_export_button_enabled"`
	CanManageAPITokens       bool                 `json:"can_manage_api_tokens"`
	Theme                    string               `json:"theme"`
	UserInterfaceDensity     string               `json:"user_interface_density"`
	DefaultPage              *string              `json:"default_page"`
	Dashboard                CurrentUserDashboard `json:"dashboard"`
}

// CurrentUserDashboard is the authenticated user's dashboard role and rights,
// nested in CurrentUserParameters.
type CurrentUserDashboard struct {
	GlobalUserRole         string `json:"global_user_role"`
	ViewDashboards         bool   `json:"view_dashboards"`
	CreateDashboards       bool   `json:"create_dashboards"`
	AdministrateDashboards bool   `json:"administrate_dashboards"`
}

// ACLActions is the authenticated user's allowed real-time actions per object
// type, returned by GET /users/acl/actions. The three object types and the seven
// actions within each are a stable, bounded set, so they are fixed structs.
type ACLActions struct {
	Host        ACLActionSet `json:"host"`
	Service     ACLActionSet `json:"service"`
	Metaservice ACLActionSet `json:"metaservice"`
}

// ACLActionSet is the set of real-time actions allowed for one object type.
type ACLActionSet struct {
	Check              bool `json:"check"`
	ForcedCheck        bool `json:"forced_check"`
	Acknowledgement    bool `json:"acknowledgement"`
	Disacknowledgement bool `json:"disacknowledgement"`
	Downtime           bool `json:"downtime"`
	SubmitStatus       bool `json:"submit_status"`
	Comment            bool `json:"comment"`
}

// CurrentUserService provides read-only access to the authenticated user's own
// context: profile and UI preferences (/configuration/users/current/parameters),
// allowed actions (/users/acl/actions), and effective feature permissions
// (/users/acl/permissions). It is a distinct service from UserService because two
// of its routes live under /users, not /configuration/users, and it concerns the
// current user rather than user configuration.
//
// The parameters endpoint also accepts PATCH to update the current user's own
// preferences, but its writable field set is not yet modeled; a partial-update
// method is deferred to a follow-up (see issue #106).
type CurrentUserService struct {
	client *Client
}

// GetParameters returns the authenticated user's profile and UI preferences from
// GET /configuration/users/current/parameters.
func (s *CurrentUserService) GetParameters(ctx context.Context) (*CurrentUserParameters, error) {
	var result CurrentUserParameters
	if err := s.client.get(ctx, "/configuration/users/current/parameters", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetACLActions returns the authenticated user's allowed real-time actions per
// object type from GET /users/acl/actions.
func (s *CurrentUserService) GetACLActions(ctx context.Context) (*ACLActions, error) {
	var result ACLActions
	if err := s.client.get(ctx, "/users/acl/actions", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetACLPermissions returns the authenticated user's effective feature
// permissions from GET /users/acl/permissions. The response is a sparse,
// instance-dependent set of permission keys, so it is returned as a map; an
// absent key means the permission is not granted.
func (s *CurrentUserService) GetACLPermissions(ctx context.Context) (map[string]bool, error) {
	var result map[string]bool
	if err := s.client.get(ctx, "/users/acl/permissions", &result); err != nil {
		return nil, err
	}
	return result, nil
}
