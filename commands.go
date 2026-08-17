package centreon

import (
	"context"
	"iter"
)

// Command represents a Centreon check command.
type Command struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Type        int    `json:"type"`
	CommandLine string `json:"command_line"`
	IsShell     bool   `json:"is_shell"`
	IsLocked    bool   `json:"is_locked"`
	IsActivated bool   `json:"is_activated"`
}

// CreateCommandRequest is the request body for creating a command.
//
// Type selects the command type: 1=notification, 2=check, 3=misc, 4=discovery.
// Name, Type, and CommandLine are required by the API.
type CreateCommandRequest struct {
	Name        string `json:"name"`
	Type        int    `json:"type"`
	CommandLine string `json:"command_line"`
	IsShell     bool   `json:"is_shell"`
}

type CommandService struct {
	client *Client
}

func (s *CommandService) List(ctx context.Context, opts ...ListOption) (*ListResponse[Command], error) {
	var resp ListResponse[Command]
	err := s.client.list(ctx, "/configuration/commands", opts, &resp)
	return &resp, err
}

func (s *CommandService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*Command, error] {
	return all(ctx, s.List, opts)
}

// Create creates a new command and returns it.
//
// Unlike the other Create methods in this package, which return only the new
// ID, Create returns the full Command decoded from the API response. Centreon
// Web 25.10 exposes no per-id route for commands (GET, PUT, PATCH, and DELETE
// on /configuration/commands/{id} all return 404), so returning the full
// object here spares callers a List-with-filter to recover what they just
// created. There is deliberately no Get, Update, or Delete, and a created
// command cannot be removed via this client.
func (s *CommandService) Create(ctx context.Context, req CreateCommandRequest) (*Command, error) {
	var cmd Command
	if err := s.client.post(ctx, "/configuration/commands", req, &cmd); err != nil {
		return nil, err
	}
	return &cmd, nil
}
