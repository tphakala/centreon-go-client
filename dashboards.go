package centreon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"time"
)

// Dashboard is a Centreon dashboard as returned by the /configuration/dashboards
// endpoints on Centreon Web 25.10.x. It is a superset read model: the three read
// shapes return different subsets of these fields, and a single struct decodes
// all of them because an absent or null JSON key leaves the corresponding Go
// field at its zero value (nil for the pointer/slice fields).
//
//   - List (GET /configuration/dashboards) omits Panels and Refresh, but carries
//     Shares, Thumbnail, and IsFavorite.
//   - Get (GET /configuration/dashboards/{id}) carries every field.
//   - The Create response (POST) carries Panels and Refresh but omits Shares,
//     Thumbnail, and IsFavorite.
//
// Panels, Refresh, and Shares are therefore modeled as slice/pointer types so a
// caller can tell "absent" (nil) from "present but empty".
type Dashboard struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedBy   NamedRef  `json:"created_by"`
	UpdatedBy   NamedRef  `json:"updated_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	OwnRole     string    `json:"own_role"`
	// Panels, Refresh, Shares, and Thumbnail are absent from some of the three read
	// shapes, so they carry omitzero: re-marshaling a Dashboard then reproduces the
	// server's omit-when-absent shape instead of emitting explicit nulls.
	Panels  []DashboardPanel  `json:"panels,omitzero"`
	Refresh *DashboardRefresh `json:"refresh,omitzero"`
	Shares  *DashboardShares  `json:"shares,omitzero"`
	// Thumbnail is null on every reachable dashboard on 25.10.16; its populated
	// shape was never observed, so it is kept as raw JSON (nil for null/absent) to
	// avoid a wrong Go type breaking the decode of the whole list.
	Thumbnail  *json.RawMessage `json:"thumbnail,omitzero"`
	IsFavorite bool             `json:"is_favorite"`
}

// DashboardPanel is a single panel (widget) on a dashboard. WidgetSettings is
// kept as raw JSON because its shape depends on WidgetType (each widget kind has
// its own settings object); callers decode it per widget kind.
type DashboardPanel struct {
	ID             int             `json:"id"`
	Name           string          `json:"name"`
	Layout         PanelLayout     `json:"layout"`
	WidgetType     string          `json:"widget_type"`
	WidgetSettings json.RawMessage `json:"widget_settings"`
}

// PanelLayout is a panel's position and size on the dashboard grid. The API
// requires all six properties on write.
type PanelLayout struct {
	X         int `json:"x"`
	Y         int `json:"y"`
	Width     int `json:"width"`
	Height    int `json:"height"`
	MinWidth  int `json:"min_width"`
	MinHeight int `json:"min_height"`
}

// DashboardRefresh is a dashboard's auto-refresh setting. Type is "global"
// (follow the platform refresh interval) or "manual". Interval is the manual
// refresh interval in seconds, null for the global type.
type DashboardRefresh struct {
	Type     string `json:"type"`
	Interval *int   `json:"interval"`
}

// DashboardShares lists who a dashboard is shared with. ContactGroups is empty on
// a box without dashboard-ACL-enabled contact groups; its populated element shape
// was not observed live, so DashboardContactGroupShare is modeled by symmetry with
// the contact share (id/name/role, without an email).
type DashboardShares struct {
	Contacts      []DashboardContactShare      `json:"contacts"`
	ContactGroups []DashboardContactGroupShare `json:"contact_groups"`
}

// DashboardContactShare is a single contact a dashboard is shared with, together
// with the role granted to that contact ("editor" or "viewer").
type DashboardContactShare struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// DashboardContactGroupShare is a single contact group a dashboard is shared with.
type DashboardContactGroupShare struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// CreateDashboardRequest is the request body for creating a dashboard
// (POST /configuration/dashboards). Centreon Web 25.10.x requires all four
// fields. Refresh must be a JSON object with a non-empty Type (a null or scalar
// refresh returns HTTP 400). Panels may be empty but not null; a nil Panels is
// normalized to an empty array by Create. See UpdateDashboardRequest for the
// replace path.
type CreateDashboardRequest struct {
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Panels      []DashboardPanelRequest `json:"panels"`
	Refresh     DashboardRefresh        `json:"refresh"`
}

// UpdateDashboardRequest is the request body for replacing a dashboard
// (POST /configuration/dashboards/{id}). It carries the same four required fields
// as create; it is a distinct type to match the per-resource Create/Update
// request convention used elsewhere in this package.
//
// Each panel's WidgetSettings is given as a JSON object here, exactly as on
// create, even though the update endpoint wants it as a JSON-encoded string on
// the wire (see dashboardUpdatePanel); Update performs that re-encoding, so the
// caller-facing shape stays uniform across Create and Update.
type UpdateDashboardRequest struct {
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Panels      []DashboardPanelRequest `json:"panels"`
	Refresh     DashboardRefresh        `json:"refresh"`
}

// dashboardUpdatePanel is the update-side panel wire shape. Unlike create, the
// dashboard update route (POST /configuration/dashboards/{id}) expects
// widget_settings as a JSON-encoded STRING, not an object: sending an object
// returns HTTP 500 ("json_decode(): Argument #1 ($json) must be of type string,
// array given"), live-verified on 25.10.16. Update re-encodes each panel's
// WidgetSettings object into this string form.
type dashboardUpdatePanel struct {
	Name           string      `json:"name"`
	Layout         PanelLayout `json:"layout"`
	WidgetType     string      `json:"widget_type"`
	WidgetSettings string      `json:"widget_settings"`
}

// dashboardUpdateBody is the POST /configuration/dashboards/{id} request body,
// identical to a create body except that panel widget_settings are strings (see
// dashboardUpdatePanel).
type dashboardUpdateBody struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Panels      []dashboardUpdatePanel `json:"panels"`
	Refresh     DashboardRefresh       `json:"refresh"`
}

// DashboardPanelRequest is a panel on the write side. It carries no id (the
// server assigns one). WidgetSettings is a JSON object; a nil or empty value is
// defaulted to an empty object ("{}") by both Create and Update, because the API
// rejects a null or empty widget_settings.
type DashboardPanelRequest struct {
	Name           string          `json:"name"`
	Layout         PanelLayout     `json:"layout"`
	WidgetType     string          `json:"widget_type"`
	WidgetSettings json.RawMessage `json:"widget_settings"`
}

// UpdateDashboardSharesRequest is the request body for replacing a dashboard's share list
// (PUT /configuration/dashboards/{id}/shares). Each entry grants a role to a
// contact or contact group by id; the name and email in the read model are not
// sent. Nil slices are normalized to empty arrays. An empty request clears all
// shares.
type UpdateDashboardSharesRequest struct {
	Contacts      []DashboardShareContactRequest      `json:"contacts"`
	ContactGroups []DashboardShareContactGroupRequest `json:"contact_groups"`
}

// DashboardShareContactRequest grants a role to one contact when updating dashboard shares.
type DashboardShareContactRequest struct {
	ID   int    `json:"id"`
	Role string `json:"role"`
}

// DashboardShareContactGroupRequest grants a role to one contact group when updating
// dashboard shares.
type DashboardShareContactGroupRequest struct {
	ID   int    `json:"id"`
	Role string `json:"role"`
}

// DashboardService provides dashboard configuration operations (issue #69). It
// wraps /configuration/dashboards on Centreon Web 25.10+.
//
// Two routes break the conventions used elsewhere in this package: the update
// verb is POST on the per-id path (not PUT/PATCH), and share management is a
// separate PUT /configuration/dashboards/{id}/shares (see UpdateShares).
type DashboardService struct {
	client *Client
}

// List returns a paginated list of dashboards. The list representation omits each
// dashboard's Panels and Refresh (use Get for those).
func (s *DashboardService) List(ctx context.Context, opts ...ListOption) (*ListResponse[Dashboard], error) {
	var resp ListResponse[Dashboard]
	err := s.client.list(ctx, "/configuration/dashboards", opts, &resp)
	return &resp, err
}

// All returns an iterator over all dashboards.
func (s *DashboardService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*Dashboard, error] {
	return all(ctx, s.List, opts)
}

// Get returns the full dashboard for the given ID, including its panels, refresh
// setting, and shares. Returns an *APIError with HTTP status 404 when no
// dashboard has that id.
func (s *DashboardService) Get(ctx context.Context, id int) (*Dashboard, error) {
	var d Dashboard
	if err := s.client.get(ctx, fmt.Sprintf("/configuration/dashboards/%d", id), &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// normalizeCreatePanels returns a copy of panels safe to send to the create
// route: it is a non-nil slice (so a nil Panels marshals to [], which the API
// requires instead of null), and each panel's empty WidgetSettings is defaulted
// to an empty JSON object (the API rejects a null/empty widget_settings). The
// copy leaves the caller's slice and its panels unmodified.
func normalizeCreatePanels(panels []DashboardPanelRequest) []DashboardPanelRequest {
	out := make([]DashboardPanelRequest, len(panels))
	for i, p := range panels {
		if len(p.WidgetSettings) == 0 {
			p.WidgetSettings = json.RawMessage("{}")
		}
		out[i] = p
	}
	return out
}

// Create creates a new dashboard and returns its ID. The POST route echoes the
// full created dashboard; only the id is decoded here. Panels is normalized (nil
// slice to [], empty widget_settings to "{}") on a copy so the caller's request
// struct is left unmodified.
func (s *DashboardService) Create(ctx context.Context, req *CreateDashboardRequest) (int, error) {
	if req == nil {
		return 0, errors.New("centreon: nil CreateDashboardRequest")
	}
	body := *req
	body.Panels = normalizeCreatePanels(req.Panels)
	var result struct {
		ID int `json:"id"`
	}
	if err := s.client.post(ctx, "/configuration/dashboards", &body, &result); err != nil {
		return 0, err
	}
	return result.ID, nil
}

// Update replaces an existing dashboard. The update verb is POST on the per-id
// path (NOT PUT: PUT /configuration/dashboards/{id} returns HTTP 405). Each
// panel's WidgetSettings (a JSON object, as on create) is re-encoded as a
// JSON-encoded string, which this endpoint requires; a nil/empty WidgetSettings
// becomes "{}". Panels is emitted as an array (empty, not null, when there are
// none). The caller's request struct is not mutated.
func (s *DashboardService) Update(ctx context.Context, id int, req *UpdateDashboardRequest) error {
	if req == nil {
		return errors.New("centreon: nil UpdateDashboardRequest")
	}
	body := dashboardUpdateBody{
		Name:        req.Name,
		Description: req.Description,
		Refresh:     req.Refresh,
		Panels:      make([]dashboardUpdatePanel, len(req.Panels)),
	}
	for i, p := range req.Panels {
		settings := string(p.WidgetSettings)
		if settings == "" {
			settings = "{}"
		}
		body.Panels[i] = dashboardUpdatePanel{
			Name:           p.Name,
			Layout:         p.Layout,
			WidgetType:     p.WidgetType,
			WidgetSettings: settings,
		}
	}
	return s.client.post(ctx, fmt.Sprintf("/configuration/dashboards/%d", id), &body, nil)
}

// Delete deletes a dashboard by ID.
func (s *DashboardService) Delete(ctx context.Context, id int) error {
	return s.client.delete(ctx, fmt.Sprintf("/configuration/dashboards/%d", id))
}

// UpdateShares replaces the share list of a dashboard
// (PUT /configuration/dashboards/{id}/shares). Nil contact and contact-group
// slices are normalized to empty arrays on a copy (so the caller's request struct
// is left unmodified); passing an empty request clears all shares.
func (s *DashboardService) UpdateShares(ctx context.Context, id int, req *UpdateDashboardSharesRequest) error {
	if req == nil {
		return errors.New("centreon: nil UpdateDashboardSharesRequest")
	}
	body := *req
	body.Contacts = nilToEmpty(body.Contacts)
	body.ContactGroups = nilToEmpty(body.ContactGroups)
	return s.client.put(ctx, fmt.Sprintf("/configuration/dashboards/%d/shares", id), body)
}
