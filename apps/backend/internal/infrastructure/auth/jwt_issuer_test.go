package auth_test

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/auth"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/idgen"
)

func newTestUser(t *testing.T, role enums.UserRole) *user.User {
	t.Helper()
	id := idgen.UUIDGenerator[user.UserID]{}.NewID()
	u, err := user.NewUser(id, "google-sub", "jotaro@example.com", "jotaro", "Jotaro Kujo", "", role)
	if err != nil {
		t.Fatalf("building test user: %v", err)
	}
	return u
}

func TestJWTIssuer_IssueAndParse_RoundTrip(t *testing.T) {
	issuer := auth.NewJWTIssuer([]byte("test-secret-at-least-32-bytes!!"), "test-issuer", time.Hour)
	u := newTestUser(t, enums.Admin)

	token, expiresAt, err := issuer.Issue(u)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if token == "" {
		t.Fatal("Issue returned empty token")
	}
	if !expiresAt.After(time.Now()) {
		t.Errorf("expiresAt = %v, want in the future", expiresAt)
	}

	claims, err := issuer.Parse(token)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if claims.UserID != u.ID() {
		t.Errorf("claims.UserID = %v, want %v", claims.UserID, u.ID())
	}
	if claims.Role != enums.Admin {
		t.Errorf("claims.Role = %v, want %v", claims.Role, enums.Admin)
	}
}

func TestJWTIssuer_Parse_RejectsTamperedSignature(t *testing.T) {
	issuer := auth.NewJWTIssuer([]byte("test-secret-at-least-32-bytes!!"), "test-issuer", time.Hour)
	u := newTestUser(t, enums.Regular)

	token, _, err := issuer.Issue(u)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	tampered := token[:len(token)-1] + "x"
	if _, err := issuer.Parse(tampered); err == nil {
		t.Fatal("Parse accepted a tampered signature")
	}
}

func TestJWTIssuer_Parse_RejectsAlgNone(t *testing.T) {
	issuer := auth.NewJWTIssuer([]byte("test-secret-at-least-32-bytes!!"), "test-issuer", time.Hour)

	claims := jwt.MapClaims{
		"sub":  "whatever",
		"iss":  "test-issuer",
		"role": "ADMIN",
		"exp":  time.Now().Add(time.Hour).Unix(),
	}
	unsigned, err := jwt.NewWithClaims(jwt.SigningMethodNone, claims).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("building alg=none token: %v", err)
	}

	if _, err := issuer.Parse(unsigned); err == nil {
		t.Fatal("Parse accepted an alg=none token")
	}
}

func TestJWTIssuer_Parse_RejectsExpiredToken(t *testing.T) {
	issuer := auth.NewJWTIssuer([]byte("test-secret-at-least-32-bytes!!"), "test-issuer", -time.Second)
	u := newTestUser(t, enums.Regular)

	token, _, err := issuer.Issue(u)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if _, err := issuer.Parse(token); err == nil {
		t.Fatal("Parse accepted an expired token")
	}
}

func TestJWTIssuer_Parse_RejectsDifferentIssuer(t *testing.T) {
	issuerA := auth.NewJWTIssuer([]byte("test-secret-at-least-32-bytes!!"), "issuer-a", time.Hour)
	issuerB := auth.NewJWTIssuer([]byte("test-secret-at-least-32-bytes!!"), "issuer-b", time.Hour)
	u := newTestUser(t, enums.Regular)

	token, _, err := issuerA.Issue(u)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if _, err := issuerB.Parse(token); err == nil {
		t.Fatal("Parse accepted a token from a different issuer")
	}
}

func TestJWTIssuer_Parse_RejectsGarbage(t *testing.T) {
	issuer := auth.NewJWTIssuer([]byte("test-secret-at-least-32-bytes!!"), "test-issuer", time.Hour)
	_, err := issuer.Parse("not.a.jwt")
	if err == nil {
		t.Fatal("Parse accepted garbage input")
	}
	if !errors.Is(err, ports.ErrUnauthenticated) {
		t.Errorf("err = %v, want wrapping ports.ErrUnauthenticated", err)
	}
}
