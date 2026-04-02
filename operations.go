package centreon

import (
	"context"
	"time"
)

// ParentRef identifies a parent host for a service resource. Only ID is sent
// on the wire because the monitoring endpoints reject additional fields in the
// parent object (e.g. "type" or nested "parent").
type ParentRef struct {
	ID int `json:"id"`
}

// ResourceRef identifies a monitoring resource.
type ResourceRef struct {
	Type   string     `json:"type"` // "host" or "service"
	ID     int        `json:"id"`
	Parent *ParentRef `json:"parent"`
}

// AcknowledgeRequest is the request body for acknowledging resources.
type AcknowledgeRequest struct {
	Resources           []ResourceRef
	Comment             string
	IsNotifyContacts    bool
	IsPersistentComment bool
	IsSticky            bool
}

// DowntimeRequest is the request body for scheduling downtime on resources.
type DowntimeRequest struct {
	Resources []ResourceRef
	Comment   string
	StartTime time.Time
	EndTime   time.Time
	Fixed     bool
	Duration  int
}

// CheckRequest is the request body for forcing checks on resources.
type CheckRequest struct {
	Resources []ResourceRef
}

// SubmitResource is a single resource result to submit.
type SubmitResource struct {
	Type     string     `json:"type"`
	ID       int        `json:"id"`
	Parent   *ParentRef `json:"parent"`
	Status   int        `json:"status"`
	Output   string     `json:"output"`
	PerfData string     `json:"performance_data,omitempty"`
}

// SubmitResultRequest is the request body for submitting check results.
type SubmitResultRequest struct {
	Resources []SubmitResource `json:"resources"`
}

// CommentRequest is the request body for adding comments to resources.
type CommentRequest struct {
	Resources []ResourceRef
	Comment   string
}

// Wire-format types for JSON serialization.

type acknowledgeBody struct {
	Resources       []ResourceRef    `json:"resources"`
	Acknowledgement acknowledgeParam `json:"acknowledgement"`
}
type acknowledgeParam struct {
	Comment             string `json:"comment"`
	IsNotifyContacts    bool   `json:"is_notify_contacts"`
	IsPersistentComment bool   `json:"is_persistent_comment"`
	IsSticky            bool   `json:"is_sticky"`
}

type downtimeBody struct {
	Resources []ResourceRef `json:"resources"`
	Downtime  downtimeParam `json:"downtime"`
}
type downtimeParam struct {
	Comment   string    `json:"comment"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	IsFixed   bool      `json:"is_fixed"`
	Duration  int       `json:"duration"`
}

type checkBody struct {
	Resources []ResourceRef `json:"resources"`
	Check     checkParam    `json:"check"`
}
type checkParam struct {
	IsForced bool `json:"is_forced"`
}

type commentResource struct {
	Type    string     `json:"type"`
	ID      int        `json:"id"`
	Parent  *ParentRef `json:"parent"`
	Comment string     `json:"comment"`
	Date    time.Time  `json:"date"`
}
type commentBody struct {
	Resources []commentResource `json:"resources"`
}

// OperationsService provides monitoring operations (acknowledge, downtime, check, submit, comment).
type OperationsService struct {
	client *Client
}

// Acknowledge acknowledges one or more resources.
func (s *OperationsService) Acknowledge(ctx context.Context, req *AcknowledgeRequest) error {
	body := acknowledgeBody{
		Resources: req.Resources,
		Acknowledgement: acknowledgeParam{
			Comment:             req.Comment,
			IsNotifyContacts:    req.IsNotifyContacts,
			IsPersistentComment: req.IsPersistentComment,
			IsSticky:            req.IsSticky,
		},
	}
	return s.client.post(ctx, "/monitoring/resources/acknowledge", body, nil)
}

// Downtime schedules downtime for one or more resources.
func (s *OperationsService) Downtime(ctx context.Context, req *DowntimeRequest) error {
	body := downtimeBody{
		Resources: req.Resources,
		Downtime: downtimeParam{
			Comment:   req.Comment,
			StartTime: req.StartTime,
			EndTime:   req.EndTime,
			IsFixed:   req.Fixed,
			Duration:  req.Duration,
		},
	}
	return s.client.post(ctx, "/monitoring/resources/downtime", body, nil)
}

// Check forces an immediate check for one or more resources.
func (s *OperationsService) Check(ctx context.Context, req *CheckRequest) error {
	body := checkBody{
		Resources: req.Resources,
		Check:     checkParam{IsForced: true},
	}
	return s.client.post(ctx, "/monitoring/resources/check", body, nil)
}

// Submit submits passive check results for one or more resources.
func (s *OperationsService) Submit(ctx context.Context, req *SubmitResultRequest) error {
	return s.client.post(ctx, "/monitoring/resources/submit", req, nil)
}

// Comment adds a comment to one or more resources.
func (s *OperationsService) Comment(ctx context.Context, req *CommentRequest) error {
	now := time.Now()
	resources := make([]commentResource, len(req.Resources))
	for i, r := range req.Resources {
		resources[i] = commentResource{
			Type:    r.Type,
			ID:      r.ID,
			Parent:  r.Parent,
			Comment: req.Comment,
			Date:    now,
		}
	}
	return s.client.post(ctx, "/monitoring/resources/comments", commentBody{Resources: resources}, nil)
}
