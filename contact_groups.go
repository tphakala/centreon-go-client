package centreon

import (
	"context"
	"iter"
)

// ContactGroup represents a Centreon contact group.
type ContactGroup struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Alias       string `json:"alias,omitzero"`
	Type        string `json:"type,omitzero"`
	IsActivated bool   `json:"is_activated"`
}

// ContactGroupService provides contact group operations.
type ContactGroupService struct {
	client *Client
}

// List returns a paginated list of contact groups.
func (s *ContactGroupService) List(ctx context.Context, opts ...ListOption) (*ListResponse[ContactGroup], error) {
	var resp ListResponse[ContactGroup]
	err := s.client.list(ctx, "/configuration/users/contact-groups", opts, &resp)
	return &resp, err
}

// All returns an iterator over all contact groups.
func (s *ContactGroupService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*ContactGroup, error] {
	return all(ctx, s.List, opts)
}
