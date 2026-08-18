package centreon

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
)

// AgentConfiguration is the list representation of a Centreon agent
// configuration as returned by GET /configuration/agent-configurations on
// Centreon Web 25.10.x. The list view is reduced: it omits connection_mode and
// the type-dependent configuration object (both are available only from Get),
// and it adds is_agent_initiated plus an is_central flag on each poller that the
// detail view does not report.
type AgentConfiguration struct {
	ID               int           `json:"id"`
	Name             string        `json:"name"`
	Type             string        `json:"type"`
	IsAgentInitiated bool          `json:"is_agent_initiated"`
	Pollers          []AgentPoller `json:"pollers,omitzero"`
}

// AgentPoller is a poller reference in the AgentConfiguration list view. Unlike
// the plain {id, name} references in the detail view, the list view also reports
// whether the poller is the central server.
type AgentPoller struct {
	ID        int    `json:"id"`
	Name      string `json:"name,omitzero"`
	IsCentral bool   `json:"is_central"`
}

// AgentConfigurationDetail is the full agent configuration returned by
// AgentConfigurationService.Get (GET /configuration/agent-configurations/{id})
// on Centreon Web 25.10.x. Unlike the list view it carries connection_mode and
// the type-dependent configuration object, and its pollers are plain {id, name}
// references without is_central.
//
// Configuration is kept as a json.RawMessage because its shape depends on Type.
// On Centreon Web 25.10.16 the required keys differ by type:
// a "telegraf" agent takes otel_public_certificate, otel_ca_certificate,
// otel_private_key, conf_server_port, conf_certificate, and conf_private_key
// (verified by a create round-trip), while a "centreon-agent" takes
// agent_initiated, poller_initiated, otel_public_certificate,
// otel_ca_certificate, otel_private_key, port, hosts, create_host_auto, and
// tokens (observed from the create validation error, not a full round-trip).
// Callers decode the raw bytes per agent kind.
type AgentConfigurationDetail struct {
	ID             int             `json:"id"`
	Name           string          `json:"name"`
	Type           string          `json:"type"`
	ConnectionMode string          `json:"connection_mode"`
	Configuration  json.RawMessage `json:"configuration"`
	Pollers        []NamedRef      `json:"pollers,omitzero"`
}

// CreateAgentConfigurationRequest is the request body for creating an agent
// configuration. All five fields are required by Centreon Web 25.10.x. Type is
// "telegraf" or "centreon-agent"; ConnectionMode is "secure", "no-tls", or
// "insecure". The request key is poller_ids (a list of poller IDs), though the
// response echoes them under pollers. Configuration is the type-dependent object
// described on AgentConfigurationDetail and is sent verbatim.
type CreateAgentConfigurationRequest struct {
	Type           string          `json:"type"`
	Name           string          `json:"name"`
	ConnectionMode string          `json:"connection_mode"`
	PollerIDs      []int           `json:"poller_ids"`
	Configuration  json.RawMessage `json:"configuration"`
}

// UpdateAgentConfigurationRequest is the request body for replacing an agent
// configuration (PUT). On Centreon Web 25.10.x, PUT requires name, poller_ids,
// and configuration; type and connection_mode are optional on update, so they
// are omitted when empty (sending an empty string would fail enum validation).
// Set ConnectionMode to change the connection mode as part of a replacement.
type UpdateAgentConfigurationRequest struct {
	Type           string          `json:"type,omitzero"`
	Name           string          `json:"name"`
	ConnectionMode string          `json:"connection_mode,omitzero"`
	PollerIDs      []int           `json:"poller_ids"`
	Configuration  json.RawMessage `json:"configuration"`
}

// AgentConfigurationService provides agent configuration operations (issue #66).
// It wraps /configuration/agent-configurations on Centreon Web 25.10+.
type AgentConfigurationService struct {
	client *Client
}

// List returns a paginated list of agent configurations.
func (s *AgentConfigurationService) List(ctx context.Context, opts ...ListOption) (*ListResponse[AgentConfiguration], error) {
	var resp ListResponse[AgentConfiguration]
	err := s.client.list(ctx, "/configuration/agent-configurations", opts, &resp)
	return &resp, err
}

// All returns an iterator over all agent configurations.
func (s *AgentConfigurationService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*AgentConfiguration, error] {
	return all(ctx, s.List, opts)
}

// Get returns the full agent configuration for the given ID, including the
// connection mode and the type-dependent configuration object.
func (s *AgentConfigurationService) Get(ctx context.Context, id int) (*AgentConfigurationDetail, error) {
	var ac AgentConfigurationDetail
	if err := s.client.get(ctx, fmt.Sprintf("/configuration/agent-configurations/%d", id), &ac); err != nil {
		return nil, err
	}
	return &ac, nil
}

// Create creates a new agent configuration and returns its ID. The POST route
// echoes the full object; only the id is decoded here, matching the other
// configuration Create methods. A nil PollerIDs is normalized to an empty array
// because Centreon Web 25.10.x rejects a null poller_ids with HTTP 400
// ("NULL value found, but an array is required"). Normalization is applied to a
// copy so the caller's request struct is left unmodified.
func (s *AgentConfigurationService) Create(ctx context.Context, req *CreateAgentConfigurationRequest) (int, error) {
	body := *req
	body.PollerIDs = nilToEmpty(body.PollerIDs)
	var result struct {
		ID int `json:"id"`
	}
	if err := s.client.post(ctx, "/configuration/agent-configurations", &body, &result); err != nil {
		return 0, err
	}
	return result.ID, nil
}

// Update replaces an existing agent configuration using PUT. A nil PollerIDs is
// normalized to an empty array (the API rejects null); normalization is applied
// to a copy so the caller's request struct is left unmodified.
func (s *AgentConfigurationService) Update(ctx context.Context, id int, req *UpdateAgentConfigurationRequest) error {
	body := *req
	body.PollerIDs = nilToEmpty(body.PollerIDs)
	return s.client.put(ctx, fmt.Sprintf("/configuration/agent-configurations/%d", id), &body)
}

// Delete deletes an agent configuration by ID.
func (s *AgentConfigurationService) Delete(ctx context.Context, id int) error {
	return s.client.delete(ctx, fmt.Sprintf("/configuration/agent-configurations/%d", id))
}
