package centreon

import (
	"context"
	"encoding/json"
	"fmt"
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

	// ArrayFilters holds dedicated name-based array query parameters for the
	// /monitoring/resources listing (statuses, types, host/service group and
	// category names, monitoring-server names, states). Centreon parses these
	// as a scalar param whose value is a JSON-encoded array, for example
	// statuses=["OK","WARNING"], which the search DSL cannot express. Each key
	// maps to the values sent for it. queryParams JSON-encodes each entry.
	// These are only meaningful for /monitoring/resources; setting them on
	// another list endpoint sends parameters that endpoint does not define.
	// Param names and wire format verified against centreon-web
	// FindResourcesRequestValidator.php; not yet confirmed against a live 25.10
	// instance.
	ArrayFilters map[string][]string
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

// WithArrayFilter adds a dedicated name-based array filter for the
// /monitoring/resources listing. Values accumulate across calls with the
// same key. Centreon sends these as statuses=["OK","WARNING"] (a scalar
// param holding a JSON array), so they cannot be expressed via WithSearch.
// Prefer the typed helpers (WithStatuses, WithResourceTypes, ...) for known
// filters; use this for filters the typed helpers do not cover. Applies to
// /monitoring/resources only. Do not pass a reserved key (page, limit, search,
// sort_by); those are owned by WithPage/WithLimit/WithSearch/WithSort, and a
// colliding array filter would override the value they set.
func WithArrayFilter(key string, values ...string) ListOption {
	return func(o *ListOptions) {
		if len(values) == 0 {
			return
		}
		if o.ArrayFilters == nil {
			o.ArrayFilters = make(map[string][]string)
		}
		o.ArrayFilters[key] = append(o.ArrayFilters[key], values...)
	}
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
func (o *ListOptions) queryParams() (url.Values, error) {
	q := url.Values{}
	if o.Page > 0 {
		q.Set("page", strconv.Itoa(o.Page))
	}
	if o.Limit > 0 {
		q.Set("limit", strconv.Itoa(o.Limit))
	}
	if o.Search != nil {
		data, err := json.Marshal(o.Search.Build())
		if err != nil {
			return nil, fmt.Errorf("centreon: marshal search filter: %w", err)
		}
		q.Set("search", string(data))
	}
	if len(o.SortBy) > 0 {
		data, err := json.Marshal(o.SortBy)
		if err != nil {
			return nil, fmt.Errorf("centreon: marshal sort_by: %w", err)
		}
		q.Set("sort_by", string(data))
	}
	// url.Values.Encode sorts keys when the query string is built, so the
	// emitted order is already deterministic; no need to sort here.
	for k, vals := range o.ArrayFilters {
		if len(vals) == 0 {
			continue
		}
		data, err := json.Marshal(vals)
		if err != nil {
			return nil, fmt.Errorf("centreon: marshal array filter %q: %w", k, err)
		}
		q.Set(k, string(data))
	}
	return q, nil
}

// list performs a paginated GET request and decodes into result.
func (c *Client) list(ctx context.Context, path string, opts []ListOption, result any) error {
	o := applyOptions(opts)
	q, err := o.queryParams()
	if err != nil {
		return err
	}
	fullPath := path
	if encoded := q.Encode(); encoded != "" {
		fullPath = path + "?" + encoded
	}
	return c.get(ctx, fullPath, result)
}

// isLastPage determines whether the current page is the final one.
func isLastPage(page, userLimit int, meta Meta, resultCount int) bool {
	pageSize := userLimit
	if pageSize <= 0 {
		pageSize = meta.Limit
	}
	if pageSize <= 0 {
		pageSize = resultCount
	}
	if pageSize <= 0 {
		return true
	}
	if meta.Total > 0 {
		return page*pageSize >= meta.Total
	}
	return resultCount < pageSize
}

// all returns an iterator that fetches all pages of a list endpoint.
func all[T any](
	ctx context.Context,
	list func(context.Context, ...ListOption) (*ListResponse[T], error),
	opts []ListOption,
) iter.Seq2[*T, error] {
	return func(yield func(*T, error) bool) {
		page := 1
		userLimit := applyOptions(opts).Limit

		for {
			pageOpts := make([]ListOption, 0, len(opts)+1)
			pageOpts = append(pageOpts, opts...)
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

			if isLastPage(page, userLimit, resp.Meta, len(resp.Result)) {
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
