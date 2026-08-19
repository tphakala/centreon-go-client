package centreon

import (
	"context"
	"fmt"
	"iter"
)

// TimePeriod represents a Centreon time period configuration resource.
//
// Exceptions is []TimePeriodException, a breaking change from the []any used
// through v2.0.0. Each read exception carries a server-assigned, read-only id.
// To set exceptions on an update, use []TimePeriodExceptionRequest (which has no
// id field), mirroring how Templates is read as []NamedRef but written as []int.
type TimePeriod struct {
	ID         int                   `json:"id"`
	Name       string                `json:"name"`
	Alias      string                `json:"alias"`
	Days       []TimePeriodDay       `json:"days,omitzero"`
	Templates  []NamedRef            `json:"templates,omitzero"`
	Exceptions []TimePeriodException `json:"exceptions,omitzero"`
	InPeriod   bool                  `json:"in_period"`
}

// TimePeriodDay represents a day definition within a time period.
// Day is an integer where 1=Monday through 7=Sunday.
// TimeRange is a string like "00:00-24:00".
type TimePeriodDay struct {
	Day       int    `json:"day"`
	TimeRange string `json:"time_range"`
}

// TimePeriodException is a date-range exception within a time period, as
// returned by reads (List and Get). On Centreon 25.10.x an exception round-trips
// as {"id":2,"day_range":"january 1","time_range":"00:00-00:00"} (live-verified
// on 25.10.16). DayRange is a human date expression (for example "january 1")
// and TimeRange is a "HH:MM-HH:MM" window, paralleling TimePeriodDay.
//
// ID is server-assigned and read-only. To set exceptions on an update, build
// []TimePeriodExceptionRequest instead; it has no id field, so the read-only id
// cannot be sent by mistake.
type TimePeriodException struct {
	ID        int    `json:"id"`
	DayRange  string `json:"day_range"`
	TimeRange string `json:"time_range"`
}

// TimePeriodExceptionRequest is a single exception to send on a time-period
// update. It deliberately omits the id field: the id is server-assigned and
// read-only, and the live-verified write shape is {day_range, time_range}
// (25.10.16). This read/write split mirrors a time period's templates, which are
// read as []NamedRef but written as []int.
type TimePeriodExceptionRequest struct {
	DayRange  string `json:"day_range"`
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
//
// Exceptions is []TimePeriodExceptionRequest, a breaking change from the []any
// used through v2.0.0. Each entry carries only day_range and time_range; the
// read-only, server-assigned id has no field here and so is never sent.
type UpdateTimePeriodRequest struct {
	Name       string                       `json:"name"`
	Alias      string                       `json:"alias"`
	Days       []TimePeriodDay              `json:"days"`
	Templates  []int                        `json:"templates"`
	Exceptions []TimePeriodExceptionRequest `json:"exceptions"`
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
	body.Days = nilToEmpty(body.Days)
	body.Templates = nilToEmpty(body.Templates)
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
	body.Days = nilToEmpty(body.Days)
	body.Templates = nilToEmpty(body.Templates)
	body.Exceptions = nilToEmpty(body.Exceptions)
	return s.client.put(ctx, fmt.Sprintf("/configuration/timeperiods/%d", id), &body)
}

// Delete deletes a time period by ID.
func (s *TimePeriodService) Delete(ctx context.Context, id int) error {
	return s.client.delete(ctx, fmt.Sprintf("/configuration/timeperiods/%d", id))
}
