package centreon

import (
	"net/http"
	"testing"
)

func TestNotificationPolicyService_GetForHost(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/hosts/10/notification-policy", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"is_notification_enabled": true,
			"contacts": []map[string]any{
				{"id": 1, "name": "admin", "is_admin": true, "is_activated": true},
			},
			"contact_groups": []map[string]any{
				{"id": 2, "name": "admins", "is_activated": true},
			},
		})
	})

	np, err := c.NotificationPolicies.GetForHost(t.Context(), 10)
	if err != nil {
		t.Fatalf("GetForHost: %v", err)
	}
	if !np.IsNotificationEnabled {
		t.Error("IsNotificationEnabled = false, want true")
	}
	if len(np.Contacts) != 1 {
		t.Fatalf("len(Contacts) = %d, want 1", len(np.Contacts))
	}
	if np.Contacts[0].Name != "admin" {
		t.Errorf("Contacts[0].Name = %q, want %q", np.Contacts[0].Name, "admin")
	}
	if len(np.ContactGroups) != 1 {
		t.Fatalf("len(ContactGroups) = %d, want 1", len(np.ContactGroups))
	}
	if np.ContactGroups[0].Name != "admins" {
		t.Errorf("ContactGroups[0].Name = %q, want %q", np.ContactGroups[0].Name, "admins")
	}
}

func TestNotificationPolicyService_GetForService(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/hosts/10/services/20/notification-policy", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"is_notification_enabled": false,
			"contacts":                []map[string]any{},
			"contact_groups":          []map[string]any{},
		})
	})

	np, err := c.NotificationPolicies.GetForService(t.Context(), 10, 20)
	if err != nil {
		t.Fatalf("GetForService: %v", err)
	}
	if np.IsNotificationEnabled {
		t.Error("IsNotificationEnabled = true, want false")
	}
	if len(np.Contacts) != 0 {
		t.Errorf("len(Contacts) = %d, want 0", len(np.Contacts))
	}
	if len(np.ContactGroups) != 0 {
		t.Errorf("len(ContactGroups) = %d, want 0", len(np.ContactGroups))
	}
}
