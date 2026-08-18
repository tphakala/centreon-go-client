package centreon

import (
	"context"
	"iter"
)

// Connector represents a Centreon connector (for example the Perl or SSH
// connector) as returned by GET /configuration/connectors. It resolves the
// connector reference carried by commands.
type Connector struct {
	ID          int                `json:"id"`
	Name        string             `json:"name"`
	CommandLine string             `json:"command_line"`
	Description *string            `json:"description"`
	Commands    []ConnectorCommand `json:"commands,omitzero"`
	IsActivated bool               `json:"is_activated"`
}

// ConnectorCommand is a lightweight reference to a command that uses a
// connector, as returned in Connector.Commands. Type is the Centreon command
// type (for example 2 for a check command).
type ConnectorCommand struct {
	ID   int    `json:"id"`
	Type int    `json:"type"`
	Name string `json:"name"`
}

// ConnectorService provides read-only connector operations. The Centreon Web
// v2 API exposes connectors as a list-only collection: there is no per-id or
// write route (POST returns HTTP 405).
type ConnectorService struct {
	client *Client
}

// List returns a paginated list of connectors.
func (s *ConnectorService) List(ctx context.Context, opts ...ListOption) (*ListResponse[Connector], error) {
	var resp ListResponse[Connector]
	err := s.client.list(ctx, "/configuration/connectors", opts, &resp)
	return &resp, err
}

// All returns an iterator over all connectors.
func (s *ConnectorService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*Connector, error] {
	return all(ctx, s.List, opts)
}
