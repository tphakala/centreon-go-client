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
	Name  string          `json:"name"`
	Alias string          `json:"alias"`
	Days  []TimePeriodDay `json:"days,omitzero"`
}

// UpdateTimePeriodRequest is the request body for replacing a time period (PUT).
type UpdateTimePeriodRequest struct {
	Name  string          `json:"name"`
	Alias string          `json:"alias"`
	Days  []TimePeriodDay `json:"days,omitzero"`
}

// TimePeriodService provides time period configuration operations.
type TimePeriodService struct {
	client *Client
}

// List returns a paginated list of time periods.
func (s *TimePeriodService) List(ctx context.Context, opts ...ListOption) (*ListResponse[TimePeriod], error) {
	var resp ListResponse[TimePeriod]
	err := s.client.list(ctx, "/configuration/time-periods", opts, &resp)
	return &resp, err
}

// All returns an iterator over all time periods.
func (s *TimePeriodService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*TimePeriod, error] {
	return all(ctx, s.List, opts)
}

// Get returns the time period with the given ID.
func (s *TimePeriodService) Get(ctx context.Context, id int) (*TimePeriod, error) {
	var tp TimePeriod
	if err := s.client.get(ctx, fmt.Sprintf("/configuration/time-periods/%d", id), &tp); err != nil {
		return nil, err
	}
	return &tp, nil
}

// Create creates a new time period and returns its ID.
func (s *TimePeriodService) Create(ctx context.Context, req CreateTimePeriodRequest) (int, error) {
	var result struct {
		ID int `json:"id"`
	}
	if err := s.client.post(ctx, "/configuration/time-periods", req, &result); err != nil {
		return 0, err
	}
	return result.ID, nil
}

// Update replaces an existing time period using PUT.
func (s *TimePeriodService) Update(ctx context.Context, id int, req UpdateTimePeriodRequest) error {
	return s.client.put(ctx, fmt.Sprintf("/configuration/time-periods/%d", id), req)
}

// Delete deletes a time period by ID.
func (s *TimePeriodService) Delete(ctx context.Context, id int) error {
	return s.client.delete(ctx, fmt.Sprintf("/configuration/time-periods/%d", id))
}
