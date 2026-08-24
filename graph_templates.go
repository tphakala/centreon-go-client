package centreon

import (
	"context"
	"iter"
)

// GraphTemplate represents a Centreon performance-graph template as returned by
// GET /configuration/graphs/templates. Hosts and services reference a template
// by its id (see the GraphTemplateID fields on services and service templates).
type GraphTemplate struct {
	ID                        int               `json:"id"`
	Name                      string            `json:"name"`
	VerticalAxisLabel         string            `json:"vertical_axis_label"`
	Width                     int               `json:"width"`
	Height                    int               `json:"height"`
	Grid                      GraphTemplateGrid `json:"grid"`
	Base                      int               `json:"base"`
	IsGraphScaled             bool              `json:"is_graph_scaled"`
	IsDefaultCentreonTemplate bool              `json:"is_default_centreon_template"`
	IsDefault                 bool              `json:"is_default"`
}

// GraphTemplateGrid holds a graph template's value-axis bounds. UpperLimit is a
// pointer because the API sends JSON null for "no upper bound" (live-verified on
// 25.10.16), which must stay distinct from a real 0; LowerLimit is never null on
// the wire.
type GraphTemplateGrid struct {
	LowerLimit             float64  `json:"lower_limit"`
	UpperLimit             *float64 `json:"upper_limit"`
	IsUpperLimitSizedToMax bool     `json:"is_upper_limit_sized_to_max"`
}

// GraphTemplateService provides read-only access to performance-graph templates.
// The Centreon Web v2 API exposes them as a list-only collection: the collection
// route serves GET only, so POST/PUT/PATCH/DELETE return HTTP 405 and there is no
// per-id route (full CRUD is legacy-only). Verified against 25.10.16.
type GraphTemplateService struct {
	client *Client
}

// List returns a paginated list of graph templates.
func (s *GraphTemplateService) List(ctx context.Context, opts ...ListOption) (*ListResponse[GraphTemplate], error) {
	var resp ListResponse[GraphTemplate]
	err := s.client.list(ctx, "/configuration/graphs/templates", opts, &resp)
	return &resp, err
}

// All returns an iterator over all graph templates.
func (s *GraphTemplateService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*GraphTemplate, error] {
	return all(ctx, s.List, opts)
}
