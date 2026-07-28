package services_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// fakeUserRepository is an in-memory ports.IUserRepository.
type fakeUserRepository struct {
	mu    sync.Mutex
	users map[user.UserID]*user.User
}

func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{users: make(map[user.UserID]*user.User)}
}

func (f *fakeUserRepository) Save(_ context.Context, u *user.User) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for id, existing := range f.users {
		if id == u.ID() {
			continue
		}
		if existing.GoogleSub() == u.GoogleSub() || existing.Email() == u.Email() || existing.Username() == u.Username() {
			return ports.ErrUserAlreadyExists
		}
	}
	f.users[u.ID()] = u
	return nil
}

func (f *fakeUserRepository) FindByID(_ context.Context, id user.UserID) (*user.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.users[id]
	if !ok {
		return nil, ports.ErrUserNotFound
	}
	return u, nil
}

func (f *fakeUserRepository) FindByGoogleSub(_ context.Context, sub string) (*user.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, u := range f.users {
		if u.GoogleSub() == sub {
			return u, nil
		}
	}
	return nil, ports.ErrUserNotFound
}

func (f *fakeUserRepository) FindByEmail(_ context.Context, email string) (*user.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, u := range f.users {
		if u.Email() == email {
			return u, nil
		}
	}
	return nil, ports.ErrUserNotFound
}

func (f *fakeUserRepository) FindByUsername(_ context.Context, username string) (*user.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, u := range f.users {
		if u.Username() == username {
			return u, nil
		}
	}
	return nil, ports.ErrUserNotFound
}

var _ ports.IUserRepository = (*fakeUserRepository)(nil)

// fakeIDGenerator returns deterministic, incrementing ids.
type fakeIDGenerator struct {
	mu   sync.Mutex
	next byte
}

func (g *fakeIDGenerator) NewID() user.UserID {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.next++
	var id user.UserID
	id[15] = g.next
	return id
}

// fakeGoogleVerifier returns a fixed identity regardless of the raw token,
// unless configured to fail.
type fakeGoogleVerifier struct {
	identity ports.GoogleIdentity
	err      error
}

func (f fakeGoogleVerifier) Verify(_ context.Context, _ string) (ports.GoogleIdentity, error) {
	if f.err != nil {
		return ports.GoogleIdentity{}, f.err
	}
	return f.identity, nil
}

// fakeTokenIssuer returns a fixed token string carrying the user's role, so
// tests can assert on it without real JWT machinery.
type fakeTokenIssuer struct{}

func (fakeTokenIssuer) Issue(u *user.User) (string, time.Time, error) {
	return "token-for-" + u.Username(), time.Now().Add(time.Hour), nil
}

func (fakeTokenIssuer) Parse(string) (ports.Claims, error) {
	return ports.Claims{}, errors.New("not implemented")
}

func newAuthService(t *testing.T, verifier fakeGoogleVerifier, adminEmails []string) (*services.AuthService, *fakeUserRepository) {
	t.Helper()
	repo := newFakeUserRepository()
	svc := services.NewAuthService(repo, &fakeIDGenerator{}, verifier, fakeTokenIssuer{}, adminEmails)
	return svc, repo
}

func verifiedIdentity(sub, email, name string) ports.GoogleIdentity {
	return ports.GoogleIdentity{Subject: sub, Email: email, EmailVerified: true, Name: name, Picture: "pic.png"}
}

func TestLoginWithGoogle_NewUserRegisters(t *testing.T) {
	verifier := fakeGoogleVerifier{identity: verifiedIdentity("sub-1", "jotaro@example.com", "Jotaro Kujo")}
	svc, _ := newAuthService(t, verifier, nil)

	result, err := svc.LoginWithGoogle(context.Background(), "raw-token")
	if err != nil {
		t.Fatalf("LoginWithGoogle: %v", err)
	}
	if !result.Registered {
		t.Error("Registered = false, want true for a brand-new account")
	}
	if result.User.Username() != "jotaro" {
		t.Errorf("username = %q, want %q", result.User.Username(), "jotaro")
	}
	if result.User.Role() != enums.Regular {
		t.Errorf("role = %v, want Regular", result.User.Role())
	}
	if result.AccessToken == "" {
		t.Error("AccessToken is empty")
	}
}

func TestLoginWithGoogle_RepeatedLoginIsNotRegistration(t *testing.T) {
	verifier := fakeGoogleVerifier{identity: verifiedIdentity("sub-1", "jotaro@example.com", "Jotaro Kujo")}
	svc, _ := newAuthService(t, verifier, nil)

	first, err := svc.LoginWithGoogle(context.Background(), "raw-token")
	if err != nil {
		t.Fatalf("first login: %v", err)
	}

	second, err := svc.LoginWithGoogle(context.Background(), "raw-token")
	if err != nil {
		t.Fatalf("second login: %v", err)
	}
	if second.Registered {
		t.Error("Registered = true on second login, want false")
	}
	if second.User.ID() != first.User.ID() {
		t.Error("second login created a different user")
	}
}

func TestLoginWithGoogle_UsernameCollisionGetsSuffixed(t *testing.T) {
	verifier1 := fakeGoogleVerifier{identity: verifiedIdentity("sub-1", "jotaro@example.com", "Jotaro Kujo")}
	repo := newFakeUserRepository()
	svc := services.NewAuthService(repo, &fakeIDGenerator{}, verifier1, fakeTokenIssuer{}, nil)

	if _, err := svc.LoginWithGoogle(context.Background(), "raw-token"); err != nil {
		t.Fatalf("first registration: %v", err)
	}

	verifier2 := fakeGoogleVerifier{identity: verifiedIdentity("sub-2", "jotaro@other.com", "Jotaro Impostor")}
	svc2 := services.NewAuthService(repo, &fakeIDGenerator{}, verifier2, fakeTokenIssuer{}, nil)

	result, err := svc2.LoginWithGoogle(context.Background(), "raw-token")
	if err != nil {
		t.Fatalf("second registration: %v", err)
	}
	if result.User.Username() == "jotaro" {
		t.Error("username collision was not resolved with a suffix")
	}
}

func TestLoginWithGoogle_EmailNotVerified(t *testing.T) {
	identity := verifiedIdentity("sub-1", "jotaro@example.com", "Jotaro Kujo")
	identity.EmailVerified = false
	svc, _ := newAuthService(t, fakeGoogleVerifier{identity: identity}, nil)

	_, err := svc.LoginWithGoogle(context.Background(), "raw-token")
	if !errors.Is(err, ports.ErrEmailNotVerified) {
		t.Fatalf("err = %v, want ports.ErrEmailNotVerified", err)
	}
}

func TestLoginWithGoogle_AdminEmailPromotesRole(t *testing.T) {
	verifier := fakeGoogleVerifier{identity: verifiedIdentity("sub-1", "jotaro@example.com", "Jotaro Kujo")}
	svc, _ := newAuthService(t, verifier, []string{"JOTARO@example.com"})

	result, err := svc.LoginWithGoogle(context.Background(), "raw-token")
	if err != nil {
		t.Fatalf("LoginWithGoogle: %v", err)
	}
	if result.User.Role() != enums.Admin {
		t.Errorf("role = %v, want Admin", result.User.Role())
	}
}

func TestLoginWithGoogle_RemovingFromAdminEmailsDemotesOnNextLogin(t *testing.T) {
	verifier := fakeGoogleVerifier{identity: verifiedIdentity("sub-1", "jotaro@example.com", "Jotaro Kujo")}
	repo := newFakeUserRepository()

	adminSvc := services.NewAuthService(repo, &fakeIDGenerator{}, verifier, fakeTokenIssuer{}, []string{"jotaro@example.com"})
	if _, err := adminSvc.LoginWithGoogle(context.Background(), "raw-token"); err != nil {
		t.Fatalf("admin login: %v", err)
	}

	regularSvc := services.NewAuthService(repo, &fakeIDGenerator{}, verifier, fakeTokenIssuer{}, nil)
	result, err := regularSvc.LoginWithGoogle(context.Background(), "raw-token")
	if err != nil {
		t.Fatalf("second login: %v", err)
	}
	if result.User.Role() != enums.Regular {
		t.Errorf("role = %v, want Regular after being removed from ADMIN_EMAILS", result.User.Role())
	}
}

func TestLoginWithGoogle_LinksExistingAccountByEmail(t *testing.T) {
	repo := newFakeUserRepository()
	seedID := (&fakeIDGenerator{}).NewID()
	seed, err := user.NewUser(seedID, "seed-google-sub", "jotaro@example.com", "jotaro_seed", "", "", enums.Regular)
	if err != nil {
		t.Fatalf("building seed user: %v", err)
	}
	if err := repo.Save(context.Background(), seed); err != nil {
		t.Fatalf("saving seed user: %v", err)
	}

	verifier := fakeGoogleVerifier{identity: verifiedIdentity("real-google-sub", "jotaro@example.com", "Jotaro Kujo")}
	svc := services.NewAuthService(repo, &fakeIDGenerator{}, verifier, fakeTokenIssuer{}, nil)

	result, err := svc.LoginWithGoogle(context.Background(), "raw-token")
	if err != nil {
		t.Fatalf("LoginWithGoogle: %v", err)
	}
	if result.Registered {
		t.Error("Registered = true, want false when linking a pre-existing account")
	}
	if result.User.ID() != seedID {
		t.Error("linking created a new user instead of reusing the seeded one")
	}
	if result.User.GoogleSub() != "real-google-sub" {
		t.Errorf("google sub not updated on link: got %q", result.User.GoogleSub())
	}
	if result.User.Username() != "jotaro_seed" {
		t.Errorf("username changed on link: got %q, want the seeded username preserved", result.User.Username())
	}
}

func TestLoginWithGoogle_InvalidGoogleToken(t *testing.T) {
	svc, _ := newAuthService(t, fakeGoogleVerifier{err: ports.ErrInvalidGoogleToken}, nil)

	_, err := svc.LoginWithGoogle(context.Background(), "raw-token")
	if !errors.Is(err, ports.ErrInvalidGoogleToken) {
		t.Fatalf("err = %v, want ports.ErrInvalidGoogleToken", err)
	}
}
