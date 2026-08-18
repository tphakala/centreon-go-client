package centreon

import (
	"context"
	"iter"
)

// AccessGroup represents a Centreon ACL access group as returned by
// GET /configuration/access-groups (for example the built-in "ALL" group).
type AccessGroup struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Alias       string `json:"alias,omitzero"`
	HasChanged  bool   `json:"has_changed"`
	IsActivated bool   `json:"is_activated"`
}

// AccessGroupService provides read-only access group operations. The Centreon
// Web v2 API exposes access groups as a list-only collection: there is no
// per-id or write route (POST returns HTTP 405).
type AccessGroupService struct {
	client *Client
}

// List returns a paginated list of access groups.
func (s *AccessGroupService) List(ctx context.Context, opts ...ListOption) (*ListResponse[AccessGroup], error) {
	var resp ListResponse[AccessGroup]
	err := s.client.list(ctx, "/configuration/access-groups", opts, &resp)
	return &resp, err
}

// All returns an iterator over all access groups.
func (s *AccessGroupService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*AccessGroup, error] {
	return all(ctx, s.List, opts)
}
