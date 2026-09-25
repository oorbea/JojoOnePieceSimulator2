package services_test

import (
	"context"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

func TestLoginDev_CreatesUserOnDevDomain(t *testing.T) {
	svc, _ := newAuthService(t, fakeGoogleVerifier{}, nil)

	result, err := svc.LoginDev(context.Background(), "alice", false)
	if err != nil {
		t.Fatalf("LoginDev: %v", err)
	}
	if !result.Registered {
		t.Error("Registered = false, want true for a brand-new dev account")
	}
	if result.User.Email() != "alice"+services.DevEmailDomain {
		t.Errorf("email = %q, want alice%s", result.User.Email(), services.DevEmailDomain)
	}
	if result.User.Username() != "alice" {
		t.Errorf("username = %q, want %q", result.User.Username(), "alice")
	}
	if result.User.Role() != enums.Regular {
		t.Errorf("role = %v, want Regular", result.User.Role())
	}
	if result.RefreshToken == "" {
		t.Error("LoginDev did not mint a refresh token")
	}
}

func TestLoginDev_AdminChecked_CreatesAdmin(t *testing.T) {
	svc, _ := newAuthService(t, fakeGoogleVerifier{}, nil)

	result, err := svc.LoginDev(context.Background(), "bob", true)
	if err != nil {
		t.Fatalf("LoginDev: %v", err)
	}
	if result.User.Role() != enums.Admin {
		t.Errorf("role = %v, want Admin", result.User.Role())
	}
}

func TestLoginDev_RepeatedCallReusesSameUser(t *testing.T) {
	svc, _ := newAuthService(t, fakeGoogleVerifier{}, nil)

	first, err := svc.LoginDev(context.Background(), "alice", false)
	if err != nil {
		t.Fatalf("first LoginDev: %v", err)
	}

	second, err := svc.LoginDev(context.Background(), "alice", false)
	if err != nil {
		t.Fatalf("second LoginDev: %v", err)
	}
	if second.Registered {
		t.Error("Registered = true on second dev login, want false")
	}
	if second.User.ID() != first.User.ID() {
		t.Error("second dev login created a different user")
	}
}

func TestLoginDev_RepeatedCallWithDifferentAdminFlagChangesRole(t *testing.T) {
	svc, _ := newAuthService(t, fakeGoogleVerifier{}, nil)

	first, err := svc.LoginDev(context.Background(), "alice", false)
	if err != nil {
		t.Fatalf("first LoginDev: %v", err)
	}
	if first.User.Role() != enums.Regular {
		t.Fatalf("role = %v, want Regular", first.User.Role())
	}

	second, err := svc.LoginDev(context.Background(), "alice", true)
	if err != nil {
		t.Fatalf("second LoginDev: %v", err)
	}
	if second.User.ID() != first.User.ID() {
		t.Fatal("second dev login created a different user")
	}
	if second.User.Role() != enums.Admin {
		t.Errorf("role = %v, want Admin after re-login with admin=true", second.User.Role())
	}
}

func TestLoginDev_TwoDifferentNamesAreTwoUsers(t *testing.T) {
	svc, _ := newAuthService(t, fakeGoogleVerifier{}, nil)

	alice, err := svc.LoginDev(context.Background(), "alice", false)
	if err != nil {
		t.Fatalf("alice LoginDev: %v", err)
	}
	bob, err := svc.LoginDev(context.Background(), "bob", false)
	if err != nil {
		t.Fatalf("bob LoginDev: %v", err)
	}
	if alice.User.ID() == bob.User.ID() {
		t.Error("two different dev names resolved to the same user")
	}
}

func TestRefresh_PreservesDevAccountRoleRegardlessOfAdminEmails(t *testing.T) {
	// A dev account's email is never in ADMIN_EMAILS, so a naive resolveRole
	// recompute on every Refresh would silently demote it back to Regular -
	// the role LoginDev set (via its explicit admin flag) must stick.
	svc, _ := newAuthService(t, fakeGoogleVerifier{}, nil)

	login, err := svc.LoginDev(context.Background(), "alice", true)
	if err != nil {
		t.Fatalf("LoginDev: %v", err)
	}
	if login.User.Role() != enums.Admin {
		t.Fatalf("role = %v, want Admin", login.User.Role())
	}

	result, err := svc.Refresh(context.Background(), login.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if result.User.Role() != enums.Admin {
		t.Errorf("role = %v, want Admin preserved across Refresh", result.User.Role())
	}
}

func TestRefresh_StillAppliesAdminEmailsForRealAccounts(t *testing.T) {
	// Regression guard: the dev-account carve-out in Refresh must not affect
	// the existing ADMIN_EMAILS resync path for real Google accounts.
	verifier := fakeGoogleVerifier{identity: verifiedIdentity("sub-1", "jotaro@example.com", "Jotaro Kujo")}
	repo := newFakeUserRepository()
	store := newRefreshStore()

	adminSvc := services.NewAuthService(repo, &fakeIDGenerator{}, verifier, fakeTokenIssuer{}, store, []string{"jotaro@example.com"}, nil)
	login, err := adminSvc.LoginWithGoogle(context.Background(), "raw-token")
	if err != nil {
		t.Fatalf("admin login: %v", err)
	}

	regularSvc := services.NewAuthService(repo, &fakeIDGenerator{}, verifier, fakeTokenIssuer{}, store, nil, nil)
	result, err := regularSvc.Refresh(context.Background(), login.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if result.User.Role() != enums.Regular {
		t.Errorf("role = %v, want Regular after being removed from ADMIN_EMAILS", result.User.Role())
	}
}
