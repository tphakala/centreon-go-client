package centreon

import (
	"context"
	"fmt"
)

// NotificationPolicy represents the notification policy for a host or service.
type NotificationPolicy struct {
	IsNotificationEnabled bool           `json:"is_notification_enabled"`
	Contacts              []User         `json:"contacts,omitzero"`
	ContactGroups         []ContactGroup `json:"contact_groups,omitzero"`
}

// NotificationPolicyService provides notification policy read operations.
//
// On Centreon Web 25.10.x the underlying endpoint is unreliable for some hosts.
// GET /configuration/hosts/{id}/notification-policy has been observed to return
// HTTP 500 ("ExtendedHost::setId: The value cannot be null") on some existing
// hosts and HTTP 404 ("Host not found") on freshly configured hosts. This is a
// server-side defect, not a client error; the client surfaces it as an *APIError
// carrying the corresponding HTTP status, so callers should tolerate 500 and 404
// from these reads on 25.10.x.
type NotificationPolicyService struct {
	client *Client
}

// GetForHost returns the notification policy for the given host ID.
//
// On Centreon Web 25.10.x this may return an *APIError with HTTP 500 or 404 for
// some hosts; see NotificationPolicyService for details.
func (s *NotificationPolicyService) GetForHost(ctx context.Context, hostID int) (*NotificationPolicy, error) {
	var np NotificationPolicy
	if err := s.client.get(ctx, fmt.Sprintf("/configuration/hosts/%d/notification-policy", hostID), &np); err != nil {
		return nil, err
	}
	return &np, nil
}

// GetForService returns the notification policy for the given host and service IDs.
//
// On Centreon Web 25.10.x this may return an *APIError with HTTP 500 or 404 for
// some hosts or services; see NotificationPolicyService for details.
func (s *NotificationPolicyService) GetForService(ctx context.Context, hostID, serviceID int) (*NotificationPolicy, error) {
	var np NotificationPolicy
	if err := s.client.get(ctx, fmt.Sprintf("/configuration/hosts/%d/services/%d/notification-policy", hostID, serviceID), &np); err != nil {
		return nil, err
	}
	return &np, nil
}
