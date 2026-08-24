package centreon

import "context"

// AdministrationParameters is the response of GET /administration/parameters on
// Centreon Web 25.10.x: the global monitoring default parameters shown on the
// administration parameters page. Every field is a plain int or bool; none is
// nullable on the wire (live-verified against 25.10.16), so no pointers are
// needed. Durations and intervals are in seconds.
type AdministrationParameters struct {
	MonitoringDefaultDowntimeDuration                 int  `json:"monitoring_default_downtime_duration"`
	MonitoringDefaultRefreshInterval                  int  `json:"monitoring_default_refresh_interval"`
	StatisticsDefaultRefreshInterval                  int  `json:"statistics_default_refresh_interval"`
	MonitoringDefaultAcknowledgementPersistent        bool `json:"monitoring_default_acknowledgement_persistent"`
	MonitoringDefaultAcknowledgementSticky            bool `json:"monitoring_default_acknowledgement_sticky"`
	MonitoringDefaultAcknowledgementNotify            bool `json:"monitoring_default_acknowledgement_notify"`
	MonitoringDefaultAcknowledgementWithServices      bool `json:"monitoring_default_acknowledgement_with_services"`
	MonitoringDefaultAcknowledgementForceActiveChecks bool `json:"monitoring_default_acknowledgement_force_active_checks"`
	MonitoringDefaultDowntimeFixed                    bool `json:"monitoring_default_downtime_fixed"`
	MonitoringDefaultDowntimeWithServices             bool `json:"monitoring_default_downtime_with_services"`
	IsResourceStatusFullSearchEnabled                 bool `json:"is_resource_status_full_search_enabled"`
}

// AdministrationParametersService provides read-only access to the global
// administration parameters (GET /administration/parameters). The endpoint is
// read-only on Centreon Web 25.10.x: it is a GET-only route, so POST/PUT/PATCH/
// DELETE all return HTTP 405 (live-verified against 25.10.16). There is no update
// method.
type AdministrationParametersService struct {
	client *Client
}

// Get returns the global administration parameters.
func (s *AdministrationParametersService) Get(ctx context.Context) (*AdministrationParameters, error) {
	var result AdministrationParameters
	if err := s.client.get(ctx, "/administration/parameters", &result); err != nil {
		return nil, err
	}
	return &result, nil
}
