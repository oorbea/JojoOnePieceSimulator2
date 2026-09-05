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
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/refreshtoken"
)

// newRefreshStore builds a real (cheap, in-memory) refresh token store for
// tests, rather than a hand-rolled fake - it's the same store production
// uses when REDIS_URL is unset, so tests exercise real reuse/family
// semantics instead of a stub that might drift from them.
func newRefreshStore() *refreshtoken.MemoryStore {
	return refreshtoken.NewMemoryStore(refreshtoken.Config{TTL: time.Hour})
}

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

func (f *fakeUserRepository) UpdateUsername(_ context.Context, id user.UserID, username string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.users[id]
	if !ok {
		return ports.ErrUserNotFound
	}
	return u.ChangeUsername(username)
}

func (f *fakeUserRepository) UpdateLanguage(_ context.Context, id user.UserID, language enums.Locale) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.users[id]
	if !ok {
		return ports.ErrUserNotFound
	}
	return u.ChangeLanguage(language)
}

func (f *fakeUserRepository) UpdateAvatar(_ context.Context, id user.UserID, main, thumb *string, status enums.PictureStatus) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.users[id]
	if !ok {
		return ports.ErrUserNotFound
	}
	newMain, newThumb := u.AvatarKey(), u.AvatarThumbKey()
	if main != nil {
		newMain = *main
	}
	if thumb != nil {
		newThumb = *thumb
	}
	u.SetAvatarRenditions(newMain, newThumb, status)
	return nil
}

func (f *fakeUserRepository) AvatarKeys(_ context.Context, id user.UserID) (string, string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.users[id]
	if !ok {
		return "", "", ports.ErrUserNotFound
	}
	return u.AvatarKey(), u.AvatarThumbKey(), nil
}

func (f *fakeUserRepository) UpdateRole(_ context.Context, id user.UserID, role enums.UserRole) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.users[id]
	if !ok {
		return ports.ErrUserNotFound
	}
	return u.ChangeRole(role)
}

func (f *fakeUserRepository) Delete(_ context.Context, id user.UserID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.users[id]; !ok {
		return ports.ErrUserNotFound
	}
	delete(f.users, id)
	return nil
}

func (f *fakeUserRepository) List(_ context.Context, limit, offset int32) ([]*user.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	all := make([]*user.User, 0, len(f.users))
	for _, u := range f.users {
		all = append(all, u)
	}
	start := int(offset)
	if start > len(all) {
		start = len(all)
	}
	end := start + int(limit)
	if end > len(all) {
		end = len(all)
	}
	return all[start:end], nil
}

func (f *fakeUserRepository) CountAdmins(_ context.Context) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var count int64
	for _, u := range f.users {
		if u.IsAdmin() {
			count++
		}
	}
	return count, nil
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
	svc := services.NewAuthService(repo, &fakeIDGenerator{}, verifier, fakeTokenIssuer{}, newRefreshStore(), adminEmails, nil)
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
	svc := services.NewAuthService(repo, &fakeIDGenerator{}, verifier1, fakeTokenIssuer{}, newRefreshStore(), nil, nil)

	if _, err := svc.LoginWithGoogle(context.Background(), "raw-token"); err != nil {
		t.Fatalf("first registration: %v", err)
	}

	verifier2 := fakeGoogleVerifier{identity: verifiedIdentity("sub-2", "jotaro@other.com", "Jotaro Impostor")}
	svc2 := services.NewAuthService(repo, &fakeIDGenerator{}, verifier2, fakeTokenIssuer{}, newRefreshStore(), nil, nil)

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

	adminSvc := services.NewAuthService(repo, &fakeIDGenerator{}, verifier, fakeTokenIssuer{}, newRefreshStore(), []string{"jotaro@example.com"}, nil)
	if _, err := adminSvc.LoginWithGoogle(context.Background(), "raw-token"); err != nil {
		t.Fatalf("admin login: %v", err)
	}

	regularSvc := services.NewAuthService(repo, &fakeIDGenerator{}, verifier, fakeTokenIssuer{}, newRefreshStore(), nil, nil)
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
	svc := services.NewAuthService(repo, &fakeIDGenerator{}, verifier, fakeTokenIssuer{}, newRefreshStore(), nil, nil)

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

func TestRefresh_HappyPath_RotatesTokenAndRereadsUser(t *testing.T) {
	verifier := fakeGoogleVerifier{identity: verifiedIdentity("sub-1", "jotaro@example.com", "Jotaro Kujo")}
	svc, _ := newAuthService(t, verifier, nil)

	login, err := svc.LoginWithGoogle(context.Background(), "raw-token")
	if err != nil {
		t.Fatalf("LoginWithGoogle: %v", err)
	}
	if login.RefreshToken == "" {
		t.Fatal("LoginWithGoogle did not mint a refresh token")
	}

	result, err := svc.Refresh(context.Background(), login.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if result.User.ID() != login.User.ID() {
		t.Error("Refresh returned a different user")
	}
	if result.AccessToken == "" {
		t.Error("Refresh did not mint a new access token")
	}
	if result.RefreshToken == "" || result.RefreshToken == login.RefreshToken {
		t.Error("Refresh did not rotate the refresh token")
	}

	// The original (now burned) token must no longer work.
	if _, err := svc.Refresh(context.Background(), login.RefreshToken); err == nil {
		t.Error("replaying the original refresh token after rotation should fail")
	}
}

func TestRefresh_DemotedAdminLosesRoleOnNextRefresh(t *testing.T) {
	verifier := fakeGoogleVerifier{identity: verifiedIdentity("sub-1", "jotaro@example.com", "Jotaro Kujo")}
	repo := newFakeUserRepository()
	store := newRefreshStore()

	adminSvc := services.NewAuthService(repo, &fakeIDGenerator{}, verifier, fakeTokenIssuer{}, store, []string{"jotaro@example.com"}, nil)
	login, err := adminSvc.LoginWithGoogle(context.Background(), "raw-token")
	if err != nil {
		t.Fatalf("admin login: %v", err)
	}
	if login.User.Role() != enums.Admin {
		t.Fatalf("role = %v, want Admin", login.User.Role())
	}

	// Same repo/store, but this instance no longer lists the user as admin -
	// mirrors ADMIN_EMAILS being edited between refreshes.
	regularSvc := services.NewAuthService(repo, &fakeIDGenerator{}, verifier, fakeTokenIssuer{}, store, nil, nil)
	result, err := regularSvc.Refresh(context.Background(), login.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if result.User.Role() != enums.Regular {
		t.Errorf("role = %v, want Regular after being removed from ADMIN_EMAILS", result.User.Role())
	}
}

func TestRefresh_ReplayedToken_ReturnsErrorAndKillsFamily(t *testing.T) {
	verifier := fakeGoogleVerifier{identity: verifiedIdentity("sub-1", "jotaro@example.com", "Jotaro Kujo")}
	svc, _ := newAuthService(t, verifier, nil)

	login, err := svc.LoginWithGoogle(context.Background(), "raw-token")
	if err != nil {
		t.Fatalf("LoginWithGoogle: %v", err)
	}

	rotated, err := svc.Refresh(context.Background(), login.RefreshToken)
	if err != nil {
		t.Fatalf("first refresh: %v", err)
	}

	// Replaying the already-burned original token must fail and kill the
	// whole family.
	if _, err := svc.Refresh(context.Background(), login.RefreshToken); err == nil {
		t.Error("replayed refresh token should fail")
	}

	// The legitimately rotated token from the same family must now also
	// fail, since the family was just revoked.
	if _, err := svc.Refresh(context.Background(), rotated.RefreshToken); err == nil {
		t.Error("a legit token from a revoked family should now fail too")
	}
}

func TestLogout_RevokesFamily(t *testing.T) {
	verifier := fakeGoogleVerifier{identity: verifiedIdentity("sub-1", "jotaro@example.com", "Jotaro Kujo")}
	svc, _ := newAuthService(t, verifier, nil)

	login, err := svc.LoginWithGoogle(context.Background(), "raw-token")
	if err != nil {
		t.Fatalf("LoginWithGoogle: %v", err)
	}

	if err := svc.Logout(context.Background(), login.RefreshToken); err != nil {
		t.Fatalf("Logout: %v", err)
	}

	if _, err := svc.Refresh(context.Background(), login.RefreshToken); err == nil {
		t.Error("refreshing with a logged-out token should fail")
	}
}

func TestLogout_SecondCallWithStaleTokenDoesNotError(t *testing.T) {
	verifier := fakeGoogleVerifier{identity: verifiedIdentity("sub-1", "jotaro@example.com", "Jotaro Kujo")}
	svc, _ := newAuthService(t, verifier, nil)

	login, err := svc.LoginWithGoogle(context.Background(), "raw-token")
	if err != nil {
		t.Fatalf("LoginWithGoogle: %v", err)
	}

	if err := svc.Logout(context.Background(), login.RefreshToken); err != nil {
		t.Fatalf("first Logout: %v", err)
	}
	if err := svc.Logout(context.Background(), login.RefreshToken); err != nil {
		t.Fatalf("second Logout with the same stale token must not error, got: %v", err)
	}
}
