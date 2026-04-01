package centreon

import (
	"context"
	"iter"
)

// ContactTemplate represents a Centreon contact template.
type ContactTemplate struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Alias       string `json:"alias,omitzero"`
	IsActivated bool   `json:"is_activated"`
}

// ContactTemplateService provides contact template operations.
type ContactTemplateService struct {
	client *Client
}

// List returns a paginated list of contact templates.
func (s *ContactTemplateService) List(ctx context.Context, opts ...ListOption) (*ListResponse[ContactTemplate], error) {
	var resp ListResponse[ContactTemplate]
	err := s.client.list(ctx, "/users/contact-templates", opts, &resp)
	return &resp, err
}

// All returns an iterator over all contact templates.
func (s *ContactTemplateService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*ContactTemplate, error] {
	return all(ctx, s.List, opts)
}
