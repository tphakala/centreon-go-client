package centreon

import (
	"context"
	"fmt"
	"iter"
)

// TimePeriod represents a Centreon time period configuration resource.
type TimePeriod struct {
	ID         int             `json:"id"`
	Name       string          `json:"name"`
	Alias      string          `json:"alias"`
	Days       []TimePeriodDay `json:"days,omitzero"`
	Templates  []NamedRef      `json:"templates,omitzero"`
	Exceptions []any           `json:"exceptions,omitzero"`
	InPeriod   bool            `json:"in_period"`
}

// TimePeriodDay represents a day definition within a time period.
// Day is an integer where 1=Monday through 7=Sunday.
// TimeRange is a string like "00:00-24:00".
type TimePeriodDay struct {
	Day       int    `json:"day"`
	TimeRange string `json:"time_range"`
}

// CreateTimePeriodRequest is the request body for creating a time period.
type CreateTimePeriodRequest struct {
	Name      string          `json:"name"`
	Alias     string          `json:"alias"`
	Days      []TimePeriodDay `json:"days"`
	Templates []int           `json:"templates"`
}

// UpdateTimePeriodRequest is the request body for replacing a time period (PUT).
//
// On Centreon 25.10.x, PUT /configuration/timeperiods/{id} requires the
// exceptions field, so Update always sends it (normalizing a nil slice to an
// empty array); Create does not require it. See TimePeriodService.Update.
type UpdateTimePeriodRequest struct {
	Name       string          `json:"name"`
	Alias      string          `json:"alias"`
	Days       []TimePeriodDay `json:"days"`
	Templates  []int           `json:"templates"`
	Exceptions []any           `json:"exceptions"`
}

// TimePeriodService provides time period configuration operations.
type TimePeriodService struct {
	client *Client
}

// List returns a paginated list of time periods.
func (s *TimePeriodService) List(ctx context.Context, opts ...ListOption) (*ListResponse[TimePeriod], error) {
	var resp ListResponse[TimePeriod]
	err := s.client.list(ctx, "/configuration/timeperiods", opts, &resp)
	return &resp, err
}

// All returns an iterator over all time periods.
func (s *TimePeriodService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*TimePeriod, error] {
	return all(ctx, s.List, opts)
}

// Get returns the time period with the given ID.
func (s *TimePeriodService) Get(ctx context.Context, id int) (*TimePeriod, error) {
	var tp TimePeriod
	if err := s.client.get(ctx, fmt.Sprintf("/configuration/timeperiods/%d", id), &tp); err != nil {
		return nil, err
	}
	return &tp, nil
}

// Create creates a new time period and returns its ID.
func (s *TimePeriodService) Create(ctx context.Context, req *CreateTimePeriodRequest) (int, error) {
	// Normalize nil slices to empty arrays on a copy so the caller's request
	// struct is left unmodified (the API rejects null for these fields).
	body := *req
	if body.Days == nil {
		body.Days = []TimePeriodDay{}
	}
	if body.Templates == nil {
		body.Templates = []int{}
	}
	var result struct {
		ID int `json:"id"`
	}
	if err := s.client.post(ctx, "/configuration/timeperiods", &body, &result); err != nil {
		return 0, err
	}
	return result.ID, nil
}

// Update replaces an existing time period using PUT.
//
// On Centreon 25.10.x the exceptions field is required on update (unlike
// create), so a nil Exceptions slice is normalized to an empty array to avoid
// an HTTP 400 "[exceptions] required" response. Normalization is applied to a
// copy so the caller's request struct is left unmodified.
func (s *TimePeriodService) Update(ctx context.Context, id int, req *UpdateTimePeriodRequest) error {
	body := *req
	if body.Days == nil {
		body.Days = []TimePeriodDay{}
	}
	if body.Templates == nil {
		body.Templates = []int{}
	}
	if body.Exceptions == nil {
		body.Exceptions = []any{}
	}
	return s.client.put(ctx, fmt.Sprintf("/configuration/timeperiods/%d", id), &body)
}

// Delete deletes a time period by ID.
func (s *TimePeriodService) Delete(ctx context.Context, id int) error {
	return s.client.delete(ctx, fmt.Sprintf("/configuration/timeperiods/%d", id))
}
