//go:build integration

package centreon

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

// Poll windows for the engine-backed end-to-end operations tests. The durable
// list signals (per-host downtime and acknowledgement) need time for the
// external command to travel from the web layer through gorgone to the engine,
// then back through the broker into the realtime database, so the window is
// generous; measured propagation on a local engine-backed box was well under
// this. The submit window is shorter because that check is best-effort.
const (
	opsE2EPollTimeout   = 60 * time.Second
	opsE2EPollInterval  = 2 * time.Second
	opsE2ESubmitTimeout = 30 * time.Second
)

// newestFirst sorts a monitoring list by descending id so a just-created entry
// lands on the first page regardless of how many (including cancelled) entries
// have accumulated on a long-lived instance.
func newestFirst() ListOption {
	return WithSort(map[string]string{"id": "DESC"})
}

// pollUntil calls cond every interval until it returns true or timeout elapses,
// returning true on the first success. A cond error is transient: it is
// remembered and returned as the second value but does not stop the poll, so a
// momentary read failure does not abort verification. It stops early if the
// test context is cancelled.
func pollUntil(t *testing.T, timeout, interval time.Duration, cond func() (bool, error)) (bool, error) {
	t.Helper()
	ctx := t.Context()
	deadline := time.Now().Add(timeout)
	var lastErr error
	for {
		ok, err := cond()
		switch {
		case ok:
			return true, nil
		case err != nil:
			lastErr = err
		}
		if !time.Now().Before(deadline) {
			return false, lastErr
		}
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(interval):
		}
	}
}

// failNotVisible reports that an expected signal never appeared within the
// durable-signal window. The last read error is appended only when one occurred,
// so a clean timeout does not print a misleading "<nil>".
func failNotVisible(t *testing.T, what string, lastErr error) {
	t.Helper()
	msg := fmt.Sprintf("%s not visible within %s; engine, broker, or gorgone command delivery "+
		"may be misconfigured or slow", what, opsE2EPollTimeout)
	if lastErr != nil {
		msg += fmt.Sprintf(" (last read error: %v)", lastErr)
	}
	t.Error(msg)
}

// requireVisible polls cond until it succeeds and reports a test error if the
// signal never appears within the durable-signal window.
func requireVisible(t *testing.T, what string, cond func() (bool, error)) {
	t.Helper()
	if found, perr := pollUntil(t, opsE2EPollTimeout, opsE2EPollInterval, cond); !found {
		failNotVisible(t, what, perr)
	}
}

// findMonitoringHost returns the first live host resource matching pred (a nil
// pred matches any host), or skips the test when the realtime listing exposes
// none. On an engine-less or CI box /monitoring/resources returns no host
// resources, so the end-to-end tests skip rather than fail.
func findMonitoringHost(t *testing.T, client *Client, pred func(*MonitoringResource) bool) *MonitoringResource {
	t.Helper()
	resp, err := client.Monitoring.List(t.Context(), WithResourceTypes("host"), WithLimit(100))
	if err != nil {
		t.Fatalf("Monitoring.List(host): %v", err)
	}
	for i := range resp.Result {
		if pred == nil || pred(&resp.Result[i]) {
			return &resp.Result[i]
		}
	}
	t.Skip("no live host resource found; monitoring engine may be absent")
	return nil
}

// isHostProblem reports whether a host resource is in a problem state (DOWN or
// UNREACHABLE), the precondition for a meaningful acknowledge.
func isHostProblem(status ResourceStatus) bool {
	return status.Code == 1 || status.Code == 2
}

// TestIntegration_OperationsDowntimeE2E complements TestIntegration_OperationsBodyFormat.
// The body-format test proves the wire shape is accepted independent of resource
// existence (a synthetic id on this endpoint returns HTTP 404); this test proves a downtime
// actually applies against a real monitored host (a real id returns HTTP 204)
// and appears in the durable per-host downtime list. It needs a running
// monitoring engine and skips cleanly when none is present.
func TestIntegration_OperationsDowntimeE2E(t *testing.T) {
	client := newIntegrationClient(t)
	host := findMonitoringHost(t, client, nil)
	t.Logf("using live host resource id=%d name=%q status=%s (code %d)",
		host.ID, host.Name, host.Status.Name, host.Status.Code)
	ctx := t.Context()
	comment := fmt.Sprintf("go-it-e2e-downtime-%d", time.Now().UnixNano())

	// Register teardown before scheduling so a failed assertion still cancels.
	// A fresh context is used because the test context is cancelled by the time
	// cleanup runs. CancelForHost cancels every non-cancelled downtime on the
	// host, which also clears leftovers from any prior aborted run.
	t.Cleanup(func() {
		//nolint:gocritic // t.Context() is cancelled once cleanup runs; a live context is required to cancel.
		if err := client.Downtimes.CancelForHost(context.Background(), host.ID); err != nil {
			t.Logf("cleanup: cancel downtimes for host %d: %v", host.ID, err)
		}
	})

	now := time.Now()
	err := client.Operations.Downtime(ctx, &DowntimeRequest{
		Resources: []ResourceRef{{Type: "host", ID: host.ID}}, // Parent nil for a host
		Comment:   comment,
		StartTime: now,
		EndTime:   now.Add(time.Hour),
		Fixed:     true,
		Duration:  3600,
	})
	if err != nil {
		t.Fatalf("Operations.Downtime on live host %d: want HTTP 204 (nil error), got %v", host.ID, err)
	}

	requireVisible(t, fmt.Sprintf("scheduled downtime (comment %q) in Downtimes.ListForHost(%d)", comment, host.ID),
		func() (bool, error) {
			resp, err := client.Downtimes.ListForHost(ctx, host.ID, WithLimit(100), newestFirst())
			if err != nil {
				return false, err
			}
			for i := range resp.Result {
				if !resp.Result[i].IsCancelled && strings.Contains(resp.Result[i].Comment, comment) {
					return true, nil
				}
			}
			return false, nil
		})

	if res, err := client.Monitoring.GetHost(ctx, host.ID); err == nil {
		t.Logf("realtime is_in_downtime=%v (informational)", res.IsInDowntime)
	}
}

// TestIntegration_OperationsAcknowledgeE2E schedules an acknowledgement on a live
// host in a problem state and verifies it appears in the durable per-host
// acknowledgement list. It skips when no host or no problem-state host exists,
// and also skips if the host recovers mid-test (the engine then drops the
// acknowledgement, which is an environment change rather than a client fault).
func TestIntegration_OperationsAcknowledgeE2E(t *testing.T) {
	client := newIntegrationClient(t)
	host := findMonitoringHost(t, client, func(r *MonitoringResource) bool { return isHostProblem(r.Status) })
	t.Logf("using live host resource id=%d name=%q status=%s (code %d)",
		host.ID, host.Name, host.Status.Name, host.Status.Code)
	ctx := t.Context()
	comment := fmt.Sprintf("go-it-e2e-ack-%d", time.Now().UnixNano())

	t.Cleanup(func() {
		//nolint:gocritic // t.Context() is cancelled once cleanup runs; a live context is required to cancel.
		if err := client.Acknowledgements.CancelForHost(context.Background(), host.ID); err != nil {
			t.Logf("cleanup: cancel acknowledgement for host %d: %v", host.ID, err)
		}
	})

	err := client.Operations.Acknowledge(ctx, &AcknowledgeRequest{
		Resources: []ResourceRef{{Type: "host", ID: host.ID}},
		Comment:   comment,
		IsSticky:  true,
	})
	if err != nil {
		t.Fatalf("Operations.Acknowledge on live host %d: want HTTP 204 (nil error), got %v", host.ID, err)
	}

	found, perr := pollUntil(t, opsE2EPollTimeout, opsE2EPollInterval, func() (bool, error) {
		resp, err := client.Acknowledgements.ListForHost(ctx, host.ID, WithLimit(100), newestFirst())
		if err != nil {
			return false, err
		}
		for i := range resp.Result {
			if resp.Result[i].DeletionTime == nil && strings.Contains(resp.Result[i].Comment, comment) {
				return true, nil
			}
		}
		return false, nil
	})
	if !found {
		if res, gerr := client.Monitoring.GetHost(ctx, host.ID); gerr == nil && !isHostProblem(res.Status) {
			t.Skipf("host %d recovered to %s (code %d) during the test; acknowledgement not applicable",
				host.ID, res.Status.Name, res.Status.Code)
		}
		failNotVisible(t, fmt.Sprintf("acknowledgement (comment %q) in Acknowledgements.ListForHost(%d)", comment, host.ID), perr)
	}

	// #61: GetHost must decode the detail endpoint's top-level "acknowledged"
	// flag (not the list-shaped "is_acknowledged") and derive HostID from the
	// call argument. The realtime detail flag can lag the durable list by a few
	// seconds, so poll for it.
	var lastRes *MonitoringResource
	ackReflected, perr2 := pollUntil(t, opsE2EPollTimeout, opsE2EPollInterval, func() (bool, error) {
		res, err := client.Monitoring.GetHost(ctx, host.ID)
		if err != nil {
			return false, err
		}
		lastRes = res
		return res.IsAcknowledged, nil
	})
	if lastRes != nil && lastRes.HostID != host.ID {
		t.Errorf("GetHost.HostID = %d, want %d (must be derived from the request argument)", lastRes.HostID, host.ID)
	}
	if !ackReflected {
		if lastRes != nil && !isHostProblem(lastRes.Status) {
			t.Skipf("host %d recovered during the test; realtime acknowledgement not applicable", host.ID)
		}
		failNotVisible(t, fmt.Sprintf("GetHost(%d).IsAcknowledged", host.ID), perr2)
	}
}

// TestIntegration_OperationsSubmitE2E submits a passive check result for a live
// host and hard-asserts only the HTTP 204 acceptance (the end-to-end proof that
// a real resource is accepted, unlike the HTTP 404 the synthetic-id body-format
// test gets). Output propagation is best-effort because the engine processes the
// passive result asynchronously and an active check can overwrite the output
// before it is read.
func TestIntegration_OperationsSubmitE2E(t *testing.T) {
	client := newIntegrationClient(t)
	host := findMonitoringHost(t, client, nil)
	t.Logf("using live host resource id=%d name=%q status=%s (code %d)",
		host.ID, host.Name, host.Status.Name, host.Status.Code)
	ctx := t.Context()
	output := fmt.Sprintf("go-it-e2e-submit-%d", time.Now().UnixNano())

	// Submit the host's current status so the passive result does not itself flip
	// its state; the goal is only to prove the passive-result path applies. A host
	// passive result accepts UP (0), DOWN (1), or UNREACHABLE (2), so clamp any
	// other value (for example a transient PENDING) to UP to keep the submission
	// valid.
	status := host.Status.Code
	if status < 0 || status > 2 {
		status = 0
	}

	err := client.Operations.Submit(ctx, &SubmitResultRequest{
		Resources: []SubmitResource{{
			Type:   "host",
			ID:     host.ID,
			Status: status,
			Output: output,
		}},
	})
	if err != nil {
		t.Fatalf("Operations.Submit on live host %d: want HTTP 204 (nil error), got %v", host.ID, err)
	}

	applied, _ := pollUntil(t, opsE2ESubmitTimeout, opsE2EPollInterval, func() (bool, error) {
		res, err := client.Monitoring.GetHost(ctx, host.ID)
		if err != nil {
			return false, err // transient; pollUntil records it and keeps polling
		}
		return strings.Contains(res.Information, output), nil
	})
	if applied {
		t.Logf("passive result applied: host information now contains %q", output)
	} else {
		t.Logf("passive result accepted (HTTP 204) but output not observed within %s "+
			"(active-check overwrite race; best-effort, not a failure)", opsE2ESubmitTimeout)
	}
}
