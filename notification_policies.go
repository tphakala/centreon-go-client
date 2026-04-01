package centreon

import (
	"context"
	"fmt"
)

// NotificationPolicy represents the notification policy for a host or service.
type NotificationPolicy struct {
	IsNotificationEnabled bool           `json:"is_notification_enabled"`
	Users                 []User         `json:"users,omitzero"`
	ContactGroups         []ContactGroup `json:"contact_groups,omitzero"`
}

// NotificationPolicyService provides notification policy read operations.
type NotificationPolicyService struct {
	client *Client
}

// GetForHost returns the notification policy for the given host ID.
func (s *NotificationPolicyService) GetForHost(ctx context.Context, hostID int) (*NotificationPolicy, error) {
	var np NotificationPolicy
	if err := s.client.get(ctx, fmt.Sprintf("/configuration/hosts/%d/notification-policy", hostID), &np); err != nil {
		return nil, err
	}
	return &np, nil
}

// GetForService returns the notification policy for the given host and service IDs.
func (s *NotificationPolicyService) GetForService(ctx context.Context, hostID, serviceID int) (*NotificationPolicy, error) {
	var np NotificationPolicy
	if err := s.client.get(ctx, fmt.Sprintf("/configuration/hosts/%d/services/%d/notification-policy", hostID, serviceID), &np); err != nil {
		return nil, err
	}
	return &np, nil
}
