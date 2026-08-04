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

// LoginResult is returned by AuthService.LoginWithGoogle.
type LoginResult struct {
	User        *user.User
	AccessToken string
	ExpiresAt   time.Time
	Registered  bool // true the first time this Google account logs in
}

// AuthService coordinates the "log in or register via Google" use case.
type AuthService struct {
	users       ports.IUserRepository
	idGen       ports.IIdGenerator[user.UserID]
	verifier    ports.IGoogleTokenVerifier
	tokens      ports.ITokenIssuer
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
	return &AuthService{users: users, idGen: idGen, verifier: verifier, tokens: tokens, adminEmails: set, pictures: pictures}
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

	role := enums.Regular
	if _, ok := s.adminEmails[strings.ToLower(identity.Email)]; ok {
		role = enums.Admin
	}

	u, registered, err := s.findOrRegister(ctx, identity, role)
	if err != nil {
		return nil, err
	}

	token, expiresAt, err := s.tokens.Issue(u)
	if err != nil {
		return nil, err
	}

	return &LoginResult{User: u, AccessToken: token, ExpiresAt: expiresAt, Registered: registered}, nil
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
