package centreon

import (
	"net/http"
	"testing"
)

func TestNotificationPolicyService_GetForHost(t *testing.T) {
	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/configuration/hosts/10/notification-policy", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, NotificationPolicy{
			IsNotificationEnabled: true,
			Users: []User{
				{ID: 1, Name: "admin", IsAdmin: true, IsActivated: true},
			},
			ContactGroups: []ContactGroup{
				{ID: 2, Name: "admins", IsActivated: true},
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
	if len(np.Users) != 1 {
		t.Fatalf("len(Users) = %d, want 1", len(np.Users))
	}
	if np.Users[0].Name != "admin" {
		t.Errorf("Users[0].Name = %q, want %q", np.Users[0].Name, "admin")
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
		writeJSON(w, http.StatusOK, NotificationPolicy{
			IsNotificationEnabled: false,
			Users:                 []User{},
			ContactGroups:         []ContactGroup{},
		})
	})

	np, err := c.NotificationPolicies.GetForService(t.Context(), 10, 20)
	if err != nil {
		t.Fatalf("GetForService: %v", err)
	}
	if np.IsNotificationEnabled {
		t.Error("IsNotificationEnabled = true, want false")
	}
	if len(np.Users) != 0 {
		t.Errorf("len(Users) = %d, want 0", len(np.Users))
	}
	if len(np.ContactGroups) != 0 {
		t.Errorf("len(ContactGroups) = %d, want 0", len(np.ContactGroups))
	}
}
