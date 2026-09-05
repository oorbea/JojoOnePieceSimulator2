package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// usernameSanitizer replaces anything outside [a-z0-9_] with '_' when
// deriving a username from a Google account's email local part.
var usernameSanitizer = regexp.MustCompile(`[^a-z0-9_]`)

// maxUsernameAttempts bounds how many suffixed usernames LoginWithGoogle
// will try before giving up on a brand-new registration.
const maxUsernameAttempts = 5

// LoginResult is returned by AuthService.LoginWithGoogle and
// AuthService.Refresh.
type LoginResult struct {
	User        *user.User
	AccessToken string
	ExpiresAt   time.Time
	Registered  bool // true the first time this Google account logs in
	// RefreshToken/RefreshExpiresAt are the freshly minted refresh token
	// paired with AccessToken - see AuthEndpoints.loginWithGoogle/refresh for
	// how it's returned to the caller (cookie for web, body for native).
	RefreshToken     string
	RefreshExpiresAt time.Time
}

// AuthService coordinates the "log in or register via Google" use case, plus
// refresh-token rotation and logout.
type AuthService struct {
	users       ports.IUserRepository
	idGen       ports.IIdGenerator[user.UserID]
	verifier    ports.IGoogleTokenVerifier
	tokens      ports.ITokenIssuer
	refresh     ports.IRefreshTokenStore
	adminEmails map[string]struct{}
	pictures    ports.IPictureStorage
}

// NewAuthService builds an AuthService. adminEmails is matched
// case-insensitively against the verified Google account email to decide
// whether a user should hold the ADMIN role. pictures is used only to
// presign a caller's self-uploaded avatar key in the login response - see
// PictureURL.
func NewAuthService(
	users ports.IUserRepository,
	idGen ports.IIdGenerator[user.UserID],
	verifier ports.IGoogleTokenVerifier,
	tokens ports.ITokenIssuer,
	refresh ports.IRefreshTokenStore,
	adminEmails []string,
	pictures ports.IPictureStorage,
) *AuthService {
	set := make(map[string]struct{}, len(adminEmails))
	for _, email := range adminEmails {
		email = strings.ToLower(strings.TrimSpace(email))
		if email == "" {
			continue
		}
		set[email] = struct{}{}
	}
	return &AuthService{users: users, idGen: idGen, verifier: verifier, tokens: tokens, refresh: refresh, adminEmails: set, pictures: pictures}
}

// resolveRole reports whether email (matched case-insensitively) is listed
// in ADMIN_EMAILS. Shared by LoginWithGoogle (against the freshly verified
// Google identity) and Refresh (against the stored user's own email), so a
// caller removed from ADMIN_EMAILS is demoted on either path rather than
// only at next Google login.
func (s *AuthService) resolveRole(email string) enums.UserRole {
	if _, ok := s.adminEmails[strings.ToLower(email)]; ok {
		return enums.Admin
	}
	return enums.Regular
}

// PictureURL resolves a stored object-storage key into a URL a client can
// GET, or "" if key is empty. Mirrors StandService.PictureURL - used to
// presign a User's self-uploaded avatar key in dto.NewUserResponse.
func (s *AuthService) PictureURL(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", nil
	}
	return s.pictures.PresignGetURL(ctx, key)
}

// LoginWithGoogle verifies rawIDToken, then logs the caller in - creating a
// new User the first time a given Google account is seen.
func (s *AuthService) LoginWithGoogle(ctx context.Context, rawIDToken string) (*LoginResult, error) {
	identity, err := s.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, err
	}
	if !identity.EmailVerified {
		return nil, ports.ErrEmailNotVerified
	}

	role := s.resolveRole(identity.Email)

	u, registered, err := s.findOrRegister(ctx, identity, role)
	if err != nil {
		return nil, err
	}

	token, expiresAt, err := s.tokens.Issue(u)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshExpiresAt, err := s.refresh.Issue(ctx, ports.RefreshToken{UserID: u.ID(), Role: u.Role(), FamilyID: ""})
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		User:             u,
		AccessToken:      token,
		ExpiresAt:        expiresAt,
		Registered:       registered,
		RefreshToken:     refreshToken,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

// Refresh redeems rawToken (rotating it: the returned token is a new,
// single-use replacement in the same family) and mints a fresh access token
// for the user it names. A replayed token surfaces as ports.ErrRefreshReuse
// (and kills the whole family server-side); an unknown/expired/revoked one
// as ports.ErrRefreshInvalid - both returned as-is for the endpoint to map
// onto a 401.
func (s *AuthService) Refresh(ctx context.Context, rawToken string) (*LoginResult, error) {
	rt, err := s.refresh.Redeem(ctx, rawToken)
	if err != nil {
		return nil, err
	}

	u, err := s.users.FindByID(ctx, rt.UserID)
	if err != nil {
		if errors.Is(err, ports.ErrUserNotFound) {
			return nil, ports.ErrUnauthenticated
		}
		return nil, err
	}

	// Recompute (and, if it has drifted since the user's last login/refresh,
	// persist) the ADMIN_EMAILS-derived role - a caller removed from
	// ADMIN_EMAILS must be demoted on their very next refresh, not only the
	// next time they re-authenticate with Google.
	role := s.resolveRole(u.Email())
	if role != u.Role() {
		if err := u.ChangeRole(role); err != nil {
			return nil, err
		}
		if err := s.users.UpdateRole(ctx, u.ID(), role); err != nil {
			return nil, err
		}
	}

	token, expiresAt, err := s.tokens.Issue(u)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshExpiresAt, err := s.refresh.Issue(ctx, ports.RefreshToken{UserID: u.ID(), Role: role, FamilyID: rt.FamilyID})
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		User:             u,
		AccessToken:      token,
		ExpiresAt:        expiresAt,
		Registered:       false,
		RefreshToken:     refreshToken,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

// Logout redeems rawToken and revokes its whole refresh-token family.
// Redeeming an already-invalid/reused token is not treated as an error here
// - logging out with a stale token must still succeed - so only a genuine
// infra error from RevokeFamily (after a successful redeem) is returned.
func (s *AuthService) Logout(ctx context.Context, rawToken string) error {
	rt, err := s.refresh.Redeem(ctx, rawToken)
	if err != nil {
		// Already invalid/reused: nothing left to revoke, but that's not a
		// failure from the caller's point of view.
		return nil
	}
	return s.refresh.RevokeFamily(ctx, rt.FamilyID)
}

func (s *AuthService) findOrRegister(ctx context.Context, identity ports.GoogleIdentity, role enums.UserRole) (*user.User, bool, error) {
	u, err := s.users.FindByGoogleSub(ctx, identity.Subject)
	switch {
	case err == nil:
		return s.syncExisting(ctx, u, identity, role)
	case errors.Is(err, ports.ErrUserNotFound):
		// fall through to email-linking / registration below
	default:
		return nil, false, err
	}

	// No user is linked to this Google sub yet. If a user already exists
	// with the same email (e.g. seeded by hand), link it instead of
	// creating a duplicate account.
	u, err = s.users.FindByEmail(ctx, identity.Email)
	switch {
	case err == nil:
		linked, buildErr := user.NewUser(u.ID(), identity.Subject, identity.Email, u.Username(), identity.Name, identity.Picture, role)
		if buildErr != nil {
			return nil, false, buildErr
		}
		linked.SetAvatarRenditions(u.AvatarKey(), u.AvatarThumbKey(), u.AvatarStatus())
		if saveErr := s.users.Save(ctx, linked); saveErr != nil {
			return nil, false, saveErr
		}
		return linked, false, nil
	case errors.Is(err, ports.ErrUserNotFound):
		return s.register(ctx, identity, role)
	default:
		return nil, false, err
	}
}

func (s *AuthService) syncExisting(ctx context.Context, u *user.User, identity ports.GoogleIdentity, role enums.UserRole) (*user.User, bool, error) {
	if u.Email() == identity.Email && u.CompleteName() == identity.Name && u.GooglePicture() == identity.Picture && u.Role() == role {
		return u, false, nil
	}

	synced, err := user.NewUser(u.ID(), u.GoogleSub(), identity.Email, u.Username(), identity.Name, identity.Picture, role)
	if err != nil {
		return nil, false, err
	}
	// NewUser always starts with no avatar; Save never touches the avatar
	// columns anyway, but the returned User must still report the caller's
	// existing self-uploaded avatar (if any) rather than silently dropping
	// it from the response.
	synced.SetAvatarRenditions(u.AvatarKey(), u.AvatarThumbKey(), u.AvatarStatus())
	if err := s.users.Save(ctx, synced); err != nil {
		return nil, false, err
	}
	return synced, false, nil
}

func (s *AuthService) register(ctx context.Context, identity ports.GoogleIdentity, role enums.UserRole) (*user.User, bool, error) {
	id := s.idGen.NewID()

	username, err := s.uniqueUsername(ctx, identity.Email, id)
	if err != nil {
		return nil, false, err
	}

	u, err := user.NewUser(id, identity.Subject, identity.Email, username, identity.Name, identity.Picture, role)
	if err != nil {
		return nil, false, err
	}
	if err := s.users.Save(ctx, u); err != nil {
		return nil, false, err
	}
	return u, true, nil
}

// uniqueUsername derives a username from the local part of email, falling
// back to suffixed variants (using hex digits of id) if it is already taken.
func (s *AuthService) uniqueUsername(ctx context.Context, email string, id user.UserID) (string, error) {
	base := usernameSanitizer.ReplaceAllString(strings.ToLower(localPart(email)), "_")
	if base == "" {
		base = "player"
	}

	idHex := id.String()
	candidate := base
	for attempt := 0; attempt < maxUsernameAttempts; attempt++ {
		_, err := s.users.FindByUsername(ctx, candidate)
		switch {
		case errors.Is(err, ports.ErrUserNotFound):
			return candidate, nil
		case err != nil:
			return "", err
		}
		end := 4 + attempt
		if end > len(idHex) {
			end = len(idHex)
		}
		candidate = fmt.Sprintf("%s_%s", base, idHex[:end])
	}
	return "", fmt.Errorf("could not derive a unique username for %q", email)
}

func localPart(email string) string {
	if at := strings.IndexByte(email, '@'); at >= 0 {
		return email[:at]
	}
	return email
}
