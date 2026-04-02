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
