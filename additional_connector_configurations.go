package centreon

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"time"
)

// AdditionalConnectorConfiguration is the list representation of a Centreon
// additional connector configuration as returned by
// GET /configuration/additional-connector-configurations on Centreon Web
// 25.10.x. The list view omits the type-dependent parameters object and the
// pollers list (both are available only from Get) and carries audit metadata.
type AdditionalConnectorConfiguration struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Description string    `json:"description,omitzero"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedBy   NamedRef  `json:"created_by"`
	UpdatedAt   time.Time `json:"updated_at"`
	UpdatedBy   NamedRef  `json:"updated_by"`
}

// AdditionalConnectorConfigurationDetail is the full additional connector
// configuration returned by AdditionalConnectorConfigurationService.Get
// (GET /configuration/additional-connector-configurations/{id}) on Centreon Web
// 25.10.x. It carries every list field plus the type-dependent parameters object
// and the resolved pollers.
//
// Parameters is kept as a json.RawMessage because its shape depends on Type. The
// only type registered on 25.10.16 is "vmware_v6", whose parameters carry a port
// and a non-empty vcenters array; each vcenter has name, url, username, and
// password. On read the password is masked to null and each vcenter gains a
// server-assigned id. Callers decode the raw bytes per connector kind.
type AdditionalConnectorConfigurationDetail struct {
	ID          int             `json:"id"`
	Name        string          `json:"name"`
	Type        string          `json:"type"`
	Description string          `json:"description,omitzero"`
	Parameters  json.RawMessage `json:"parameters"`
	Pollers     []NamedRef      `json:"pollers,omitzero"`
	CreatedAt   time.Time       `json:"created_at"`
	CreatedBy   NamedRef        `json:"created_by"`
	UpdatedAt   time.Time       `json:"updated_at"`
	UpdatedBy   NamedRef        `json:"updated_by"`
}

// CreateAdditionalConnectorConfigurationRequest is the request body for creating
// an additional connector configuration. All five fields are required by
// Centreon Web 25.10.x. Type must be "vmware_v6" (the only kind registered on
// 25.10.16). Pollers is a list of poller IDs. Parameters is the type-dependent
// object described on AdditionalConnectorConfigurationDetail and is sent
// verbatim; for vmware_v6 each vcenters[] entry needs name, url, username, and
// password.
type CreateAdditionalConnectorConfigurationRequest struct {
	Type        string          `json:"type"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Pollers     []int           `json:"pollers"`
	Parameters  json.RawMessage `json:"parameters"`
}

// UpdateAdditionalConnectorConfigurationRequest is the request body for replacing
// an additional connector configuration (PUT). Centreon Web 25.10.x requires the
// same five fields as create. The update parameters schema differs from create:
// for vmware_v6 each parameters.vcenters[] entry must include an id (int for an
// existing vcenter, null for a new one), otherwise the API returns HTTP 500
// ("Invalid argument for 'parameters.vcenters[].id'"). Build Parameters
// accordingly (read the current parameters via Get to recover each vcenter id).
type UpdateAdditionalConnectorConfigurationRequest struct {
	Type        string          `json:"type"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Pollers     []int           `json:"pollers"`
	Parameters  json.RawMessage `json:"parameters"`
}

// AdditionalConnectorConfigurationService provides additional connector
// configuration operations (issue #67). It wraps
// /configuration/additional-connector-configurations on Centreon Web 25.10+.
type AdditionalConnectorConfigurationService struct {
	client *Client
}

// List returns a paginated list of additional connector configurations.
func (s *AdditionalConnectorConfigurationService) List(ctx context.Context, opts ...ListOption) (*ListResponse[AdditionalConnectorConfiguration], error) {
	var resp ListResponse[AdditionalConnectorConfiguration]
	err := s.client.list(ctx, "/configuration/additional-connector-configurations", opts, &resp)
	return &resp, err
}

// All returns an iterator over all additional connector configurations.
func (s *AdditionalConnectorConfigurationService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*AdditionalConnectorConfiguration, error] {
	return all(ctx, s.List, opts)
}

// Get returns the full additional connector configuration for the given ID,
// including the type-dependent parameters object and the resolved pollers.
func (s *AdditionalConnectorConfigurationService) Get(ctx context.Context, id int) (*AdditionalConnectorConfigurationDetail, error) {
	var acc AdditionalConnectorConfigurationDetail
	if err := s.client.get(ctx, fmt.Sprintf("/configuration/additional-connector-configurations/%d", id), &acc); err != nil {
		return nil, err
	}
	return &acc, nil
}

// Create creates a new additional connector configuration and returns its ID.
// The POST route echoes the full object; only the id is decoded here. A nil
// Pollers is normalized to an empty array because Centreon Web 25.10.x rejects a
// null pollers with HTTP 400 ("NULL value found, but an array is required").
// Normalization is applied to a copy so the caller's request struct is left
// unmodified.
func (s *AdditionalConnectorConfigurationService) Create(ctx context.Context, req *CreateAdditionalConnectorConfigurationRequest) (int, error) {
	body := *req
	if body.Pollers == nil {
		body.Pollers = []int{}
	}
	var result struct {
		ID int `json:"id"`
	}
	if err := s.client.post(ctx, "/configuration/additional-connector-configurations", &body, &result); err != nil {
		return 0, err
	}
	return result.ID, nil
}

// Update replaces an existing additional connector configuration using PUT. A
// nil Pollers is normalized to an empty array (the API rejects null);
// normalization is applied to a copy so the caller's request struct is left
// unmodified.
func (s *AdditionalConnectorConfigurationService) Update(ctx context.Context, id int, req *UpdateAdditionalConnectorConfigurationRequest) error {
	body := *req
	if body.Pollers == nil {
		body.Pollers = []int{}
	}
	return s.client.put(ctx, fmt.Sprintf("/configuration/additional-connector-configurations/%d", id), &body)
}

// Delete deletes an additional connector configuration by ID.
func (s *AdditionalConnectorConfigurationService) Delete(ctx context.Context, id int) error {
	return s.client.delete(ctx, fmt.Sprintf("/configuration/additional-connector-configurations/%d", id))
}
