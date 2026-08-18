package centreon

import (
	"context"
	"iter"
)

// Icon represents a Centreon media image as returned by
// GET /configuration/icons. It resolves the icon_id fields carried by hosts,
// host templates, and severities. URL is the path the Centreon web UI serves
// the image from.
type Icon struct {
	ID        int    `json:"id"`
	Directory string `json:"directory"`
	Name      string `json:"name"`
	URL       string `json:"url"`
}

// IconService provides read-only icon operations. The Centreon Web v2 API
// exposes icons as a list-only collection: there is no per-id or write route
// (POST returns HTTP 405).
type IconService struct {
	client *Client
}

// List returns a paginated list of icons.
func (s *IconService) List(ctx context.Context, opts ...ListOption) (*ListResponse[Icon], error) {
	var resp ListResponse[Icon]
	err := s.client.list(ctx, "/configuration/icons", opts, &resp)
	return &resp, err
}

// All returns an iterator over all icons.
func (s *IconService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*Icon, error] {
	return all(ctx, s.List, opts)
}
