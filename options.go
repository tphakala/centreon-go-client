package centreon

import (
	"context"
	"encoding/json"
	"iter"
	"net/url"
	"strconv"
)

// Meta holds pagination metadata from the API response.
type Meta struct {
	Page   int `json:"page"`
	Limit  int `json:"limit"`
	Total  int `json:"total"`
	Search any `json:"search,omitzero"`
	SortBy any `json:"sort_by,omitzero"`
}

// ListResponse is a generic API list response with pagination metadata.
type ListResponse[T any] struct {
	Result []T  `json:"result"`
	Meta   Meta `json:"meta"`
}

// ListOptions configures a list request.
type ListOptions struct {
	Page   int
	Limit  int
	Search Filter
	SortBy map[string]string
}

// ListOption is a functional option for list requests.
type ListOption func(*ListOptions)

// WithPage sets the page number.
func WithPage(page int) ListOption {
	return func(o *ListOptions) { o.Page = page }
}

// WithLimit sets the number of results per page.
func WithLimit(limit int) ListOption {
	return func(o *ListOptions) { o.Limit = limit }
}

// WithSearch sets the search filter.
func WithSearch(f Filter) ListOption {
	return func(o *ListOptions) { o.Search = f }
}

// WithSort sets the sort order.
func WithSort(sortBy map[string]string) ListOption {
	return func(o *ListOptions) { o.SortBy = sortBy }
}

// applyOptions applies functional options to a ListOptions struct.
func applyOptions(opts []ListOption) *ListOptions {
	o := &ListOptions{}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// queryParams converts ListOptions into URL query parameters.
func (o *ListOptions) queryParams() url.Values {
	q := url.Values{}
	if o.Page > 0 {
		q.Set("page", strconv.Itoa(o.Page))
	}
	if o.Limit > 0 {
		q.Set("limit", strconv.Itoa(o.Limit))
	}
	if o.Search != nil {
		data, err := json.Marshal(o.Search.Build())
		if err == nil {
			q.Set("search", string(data))
		}
	}
	if len(o.SortBy) > 0 {
		data, err := json.Marshal(o.SortBy)
		if err == nil {
			q.Set("sort_by", string(data))
		}
	}
	return q
}

// list performs a paginated GET request and decodes into result.
func (c *Client) list(ctx context.Context, path string, opts []ListOption, result any) error {
	o := applyOptions(opts)
	q := o.queryParams()
	fullPath := path
	if encoded := q.Encode(); encoded != "" {
		fullPath = path + "?" + encoded
	}
	return c.get(ctx, fullPath, result)
}

// all returns an iterator that fetches all pages of a list endpoint.
func all[T any](
	ctx context.Context,
	list func(context.Context, ...ListOption) (*ListResponse[T], error),
	opts []ListOption,
) iter.Seq2[*T, error] {
	return func(yield func(*T, error) bool) {
		page := 1
		// Extract the user's limit from opts, default to 0 (let server decide)
		o := applyOptions(opts)
		limit := o.Limit

		for {
			pageOpts := make([]ListOption, len(opts))
			copy(pageOpts, opts)
			pageOpts = append(pageOpts, WithPage(page))

			resp, err := list(ctx, pageOpts...)
			if err != nil {
				yield(nil, err)
				return
			}

			for i := range resp.Result {
				if !yield(&resp.Result[i], nil) {
					return
				}
			}

			// Determine if we've fetched all items
			pageSize := limit
			if pageSize <= 0 {
				pageSize = resp.Meta.Limit
			}
			if pageSize <= 0 || page*pageSize >= resp.Meta.Total {
				return
			}
			page++
		}
	}
}

// getByID finds a single resource by ID using a filtered list lookup.
// Returns *NotFoundError if no matching resource is found.
func getByID[T any](
	ctx context.Context,
	list func(context.Context, ...ListOption) (*ListResponse[T], error),
	resource string,
	id int,
) (*T, error) {
	resp, err := list(ctx, WithSearch(Eq("id", id)))
	if err != nil {
		return nil, err
	}
	if len(resp.Result) == 0 {
		return nil, &NotFoundError{Resource: resource, ID: id}
	}
	return &resp.Result[0], nil
}
