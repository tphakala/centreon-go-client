package centreon

import (
	"errors"
	"net/http"
	"testing"
)

func TestAdministrationParametersService_Get(t *testing.T) {
	mux, c := newTestMux(t)
	// Distinct int values (3600/15/20) and a mix of true/false bools (notify is
	// the lone false among the acknowledgement flags, downtime_with_services the
	// lone false among the downtime flags) so a swapped, dropped, or retyped json
	// tag fails a specific assertion rather than passing by luck.
	mux.HandleFunc("GET /centreon/api/latest/administration/parameters", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"monitoring_default_downtime_duration":                   3600,
			"monitoring_default_refresh_interval":                    15,
			"statistics_default_refresh_interval":                    20,
			"monitoring_default_acknowledgement_persistent":          true,
			"monitoring_default_acknowledgement_sticky":              true,
			"monitoring_default_acknowledgement_notify":              false,
			"monitoring_default_acknowledgement_with_services":       true,
			"monitoring_default_acknowledgement_force_active_checks": true,
			"monitoring_default_downtime_fixed":                      true,
			"monitoring_default_downtime_with_services":              false,
			"is_resource_status_full_search_enabled":                 true,
		})
	})

	got, err := c.AdministrationParameters.Get(t.Context())
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	wantInt(t, "MonitoringDefaultDowntimeDuration", got.MonitoringDefaultDowntimeDuration, 3600)
	wantInt(t, "MonitoringDefaultRefreshInterval", got.MonitoringDefaultRefreshInterval, 15)
	wantInt(t, "StatisticsDefaultRefreshInterval", got.StatisticsDefaultRefreshInterval, 20)
	wantBool(t, "MonitoringDefaultAcknowledgementPersistent", got.MonitoringDefaultAcknowledgementPersistent, true)
	wantBool(t, "MonitoringDefaultAcknowledgementSticky", got.MonitoringDefaultAcknowledgementSticky, true)
	wantBool(t, "MonitoringDefaultAcknowledgementNotify", got.MonitoringDefaultAcknowledgementNotify, false)
	wantBool(t, "MonitoringDefaultAcknowledgementWithServices", got.MonitoringDefaultAcknowledgementWithServices, true)
	wantBool(t, "MonitoringDefaultAcknowledgementForceActiveChecks", got.MonitoringDefaultAcknowledgementForceActiveChecks, true)
	wantBool(t, "MonitoringDefaultDowntimeFixed", got.MonitoringDefaultDowntimeFixed, true)
	wantBool(t, "MonitoringDefaultDowntimeWithServices", got.MonitoringDefaultDowntimeWithServices, false)
	wantBool(t, "IsResourceStatusFullSearchEnabled", got.IsResourceStatusFullSearchEnabled, true)
}

func TestAdministrationParametersService_Get_Error(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("GET /centreon/api/latest/administration/parameters", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "boom"})
	})

	got, err := c.AdministrationParameters.Get(t.Context())
	if err == nil {
		t.Fatal("expected an error on HTTP 500, got nil")
	}
	if got != nil {
		t.Errorf("result = %v, want nil on error", got)
	}
	if _, ok := errors.AsType[*APIError](err); !ok {
		t.Errorf("expected *APIError, got %T", err)
	}
}
