package centreon

import (
	"context"
	"time"
)

// ResourceRef identifies a monitoring resource.
type ResourceRef struct {
	Type   string       `json:"type"` // "host" or "service"
	ID     int          `json:"id"`
	Parent *ResourceRef `json:"parent,omitempty"`
}

// AcknowledgeRequest is the request body for acknowledging resources.
type AcknowledgeRequest struct {
	Resources           []ResourceRef `json:"resources"`
	Comment             string        `json:"comment"`
	IsNotifyContacts    bool          `json:"is_notify_contacts"`
	IsPersistentComment bool          `json:"is_persistent_comment"`
	IsSticky            bool          `json:"is_sticky"`
}

// DowntimeRequest is the request body for scheduling downtime on resources.
type DowntimeRequest struct {
	Resources []ResourceRef `json:"resources"`
	Comment   string        `json:"comment"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Fixed     bool          `json:"is_fixed"`
	Duration  int           `json:"duration"`
}

// CheckRequest is the request body for forcing checks on resources.
type CheckRequest struct {
	Resources []ResourceRef `json:"resources"`
}

// SubmitResource is a single resource result to submit.
type SubmitResource struct {
	Type     string       `json:"type"`
	ID       int          `json:"id"`
	Parent   *ResourceRef `json:"parent,omitempty"`
	Status   int          `json:"status"`
	Output   string       `json:"output"`
	PerfData string       `json:"performance_data,omitzero"`
}

// SubmitResultRequest is the request body for submitting check results.
type SubmitResultRequest struct {
	Resources []SubmitResource `json:"resources"`
}

// CommentRequest is the request body for adding comments to resources.
type CommentRequest struct {
	Resources []ResourceRef `json:"resources"`
	Comment   string        `json:"comment"`
}

// OperationsService provides monitoring operations (acknowledge, downtime, check, submit, comment).
type OperationsService struct {
	client *Client
}

// Acknowledge acknowledges one or more resources.
func (s *OperationsService) Acknowledge(ctx context.Context, req *AcknowledgeRequest) error {
	return s.client.post(ctx, "/monitoring/resources/acknowledge", req, nil)
}

// Downtime schedules downtime for one or more resources.
func (s *OperationsService) Downtime(ctx context.Context, req *DowntimeRequest) error {
	return s.client.post(ctx, "/monitoring/resources/downtime", req, nil)
}

// Check forces an immediate check for one or more resources.
func (s *OperationsService) Check(ctx context.Context, req *CheckRequest) error {
	return s.client.post(ctx, "/monitoring/resources/check", req, nil)
}

// Submit submits passive check results for one or more resources.
func (s *OperationsService) Submit(ctx context.Context, req *SubmitResultRequest) error {
	return s.client.post(ctx, "/monitoring/resources/submit", req, nil)
}

// Comment adds a comment to one or more resources.
func (s *OperationsService) Comment(ctx context.Context, req *CommentRequest) error {
	return s.client.post(ctx, "/monitoring/resources/comments", req, nil)
}
