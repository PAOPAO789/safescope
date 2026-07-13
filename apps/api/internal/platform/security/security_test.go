package security

import (
	"testing"
	"time"

	"github.com/safescope/safescope/apps/api/internal/domain"
)

func TestJWTManagerRoundTrip(t *testing.T) {
	manager := JWTManager{Secret: []byte("a-secret-that-is-long-enough-for-tests"), TTL: time.Hour}
	token, _, err := manager.Issue(&domain.User{ID: "user-1", Email: "a@example.com", Name: "A", Role: domain.RoleAdmin})
	if err != nil {
		t.Fatal(err)
	}
	user, err := manager.Parse(token)
	if err != nil {
		t.Fatal(err)
	}
	if user.ID != "user-1" || user.Role != domain.RoleAdmin {
		t.Fatalf("unexpected user: %#v", user)
	}
}
