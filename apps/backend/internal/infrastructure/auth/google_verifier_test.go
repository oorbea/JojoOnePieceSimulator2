package auth_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/auth"
)

// What this file can and cannot cover, and why.
//
// GoogleVerifier.Verify is a thin shell around google.golang.org/api/idtoken's
// package-level idtoken.Validate. That function reads a package-global
// defaultValidator, so there is no seam to inject a fake HTTP client (the
// library's own NewValidator does take option.WithHTTPClient, but Verify
// can't reach it without changing GoogleVerifier's construction - a
// behaviour change this test pass deliberately does not make).
//
// What that costs us: everything after the signature check needs Google's
// real signing keys, so the happy path - and with it the payload.Claims ->
// ports.GoogleIdentity mapping and the Subject copy - is not reachable
// offline. Faking it would mean either standing up a fake OIDC/JWKS server
// or monkey-patching the library, both of which test the fake rather than
// our code.
//
// What we get for free: idtoken's validate() checks the audience and the
// expiry *before* it ever fetches a cert (see validate.go in
// google.golang.org/api@v0.293.0 - the aud/exp comparisons sit above the
// RS256/ES256 switch that calls getCert). So the audience rejection this
// backend's security actually rests on IS covered here, offline and with no
// network call, along with malformed input, expiry, and unsupported
// algorithms. Every test below asserts the error is wrapped as
// ports.ErrInvalidGoogleToken, which is what endpoints/errors.go maps to 401.
//
// Note also that this version of idtoken does not check the `iss` claim at
// all; the issuer is pinned implicitly by the signature having to verify
// against Google's own JWKS. Nothing to fix on our side, but it means an
// "issuer rejected" test here would assert a behaviour the library doesn't
// have.

const testClientID = "test-client-id.apps.googleusercontent.com"

// segment base64url-encodes v as one JWT segment.
func segment(t *testing.T, v map[string]any) string {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshalling segment: %v", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}

// unsignedToken builds a structurally valid three-segment JWT with a
// syntactically valid (but meaningless) signature segment. Good enough to
// reach every pre-signature check in idtoken.validate.
func unsignedToken(t *testing.T, header, payload map[string]any) string {
	t.Helper()
	return segment(t, header) + "." + segment(t, payload) + ".AAAA"
}

func googleHeader() map[string]any {
	return map[string]any{"alg": "RS256", "typ": "JWT", "kid": "test-key-id"}
}

func googlePayload(audience string, expires time.Time) map[string]any {
	return map[string]any{
		"iss":            "https://accounts.google.com",
		"aud":            audience,
		"sub":            "1234567890",
		"exp":            expires.Unix(),
		"iat":            time.Now().Add(-time.Minute).Unix(),
		"email":          "jotaro@example.com",
		"email_verified": true,
		"name":           "Jotaro Kujo",
		"picture":        "https://google.example/pic.png",
	}
}

func TestGoogleVerifier_ImplementsPort(t *testing.T) {
	var v ports.IGoogleTokenVerifier = auth.NewGoogleVerifier(testClientID)
	if v == nil {
		t.Fatal("NewGoogleVerifier returned nil")
	}
}

// The audience check is the one that matters: without it, an ID token minted
// for any *other* Google OAuth client would be accepted here, letting a
// third-party app's token log its users into this backend.
func TestGoogleVerifier_Verify_RejectsWrongAudience(t *testing.T) {
	v := auth.NewGoogleVerifier(testClientID)
	token := unsignedToken(t, googleHeader(), googlePayload("someone-elses-client-id.apps.googleusercontent.com", time.Now().Add(time.Hour)))

	identity, err := v.Verify(context.Background(), token)
	if err == nil {
		t.Fatal("Verify accepted a token minted for another audience")
	}
	if !errors.Is(err, ports.ErrInvalidGoogleToken) {
		t.Errorf("err = %v, want wrapping ports.ErrInvalidGoogleToken", err)
	}
	if identity != (ports.GoogleIdentity{}) {
		t.Errorf("identity = %+v, want the zero value on failure", identity)
	}
}

func TestGoogleVerifier_Verify_RejectsExpiredToken(t *testing.T) {
	v := auth.NewGoogleVerifier(testClientID)
	token := unsignedToken(t, googleHeader(), googlePayload(testClientID, time.Now().Add(-time.Hour)))

	identity, err := v.Verify(context.Background(), token)
	if err == nil {
		t.Fatal("Verify accepted an expired token")
	}
	if !errors.Is(err, ports.ErrInvalidGoogleToken) {
		t.Errorf("err = %v, want wrapping ports.ErrInvalidGoogleToken", err)
	}
	if identity != (ports.GoogleIdentity{}) {
		t.Errorf("identity = %+v, want the zero value on failure", identity)
	}
}

// alg is checked before any cert fetch too, so this stays offline. It covers
// the JWT equivalent of the alg=none downgrade that jwt_issuer_test.go pins
// for our own tokens.
func TestGoogleVerifier_Verify_RejectsUnsupportedAlgorithms(t *testing.T) {
	v := auth.NewGoogleVerifier(testClientID)
	for _, alg := range []string{"none", "HS256", "RS512", ""} {
		header := googleHeader()
		header["alg"] = alg
		token := unsignedToken(t, header, googlePayload(testClientID, time.Now().Add(time.Hour)))

		if _, err := v.Verify(context.Background(), token); err == nil {
			t.Errorf("alg %q: Verify accepted a token idtoken cannot verify", alg)
		} else if !errors.Is(err, ports.ErrInvalidGoogleToken) {
			t.Errorf("alg %q: err = %v, want wrapping ports.ErrInvalidGoogleToken", alg, err)
		}
	}
}

func TestGoogleVerifier_Verify_RejectsMalformedTokens(t *testing.T) {
	v := auth.NewGoogleVerifier(testClientID)
	cases := map[string]string{
		"empty":            "",
		"not a jwt":        "not-a-jwt",
		"two segments":     "aGVhZGVy.cGF5bG9hZA",
		"four segments":    "aGVhZGVy.cGF5bG9hZA.c2ln.ZXh0cmE",
		"non base64":       "!!!.!!!.!!!",
		"non json payload": base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256"}`)) + "." + base64.RawURLEncoding.EncodeToString([]byte("not json")) + ".AAAA",
		"only dots":        "..",
	}
	for name, token := range cases {
		identity, err := v.Verify(context.Background(), token)
		if err == nil {
			t.Errorf("%s: Verify accepted a malformed token", name)
			continue
		}
		if !errors.Is(err, ports.ErrInvalidGoogleToken) {
			t.Errorf("%s: err = %v, want wrapping ports.ErrInvalidGoogleToken", name, err)
		}
		if identity != (ports.GoogleIdentity{}) {
			t.Errorf("%s: identity = %+v, want the zero value on failure", name, identity)
		}
	}
}

// stubGoogleVerifier is a ports.IGoogleTokenVerifier the rest of the suite
// can build on. It exists here, next to the real implementation, so the port
// contract the happy path must satisfy is written down even though
// GoogleVerifier's own happy path can't be exercised offline (see the file
// header): a successful Verify returns the caller's identity with
// EmailVerified carried through untouched, and a failure returns the zero
// identity alongside ports.ErrInvalidGoogleToken.
type stubGoogleVerifier struct {
	identity ports.GoogleIdentity
	err      error
}

func (s stubGoogleVerifier) Verify(_ context.Context, _ string) (ports.GoogleIdentity, error) {
	if s.err != nil {
		return ports.GoogleIdentity{}, s.err
	}
	return s.identity, nil
}

var _ ports.IGoogleTokenVerifier = stubGoogleVerifier{}

func TestGoogleTokenVerifierPort_ContractShape(t *testing.T) {
	want := ports.GoogleIdentity{
		Subject:       "1234567890",
		Email:         "jotaro@example.com",
		EmailVerified: false,
		Name:          "Jotaro Kujo",
		Picture:       "https://google.example/pic.png",
	}

	got, err := stubGoogleVerifier{identity: want}.Verify(context.Background(), "any")
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	// EmailVerified is deliberately false here: the verifier's job is to
	// report the claim, not to enforce it - AuthService.LoginWithGoogle is
	// what turns EmailVerified == false into ports.ErrEmailNotVerified. A
	// verifier that silently normalised it to true would break that check.
	if got != want {
		t.Errorf("identity = %+v, want %+v", got, want)
	}

	if _, err := (stubGoogleVerifier{err: ports.ErrInvalidGoogleToken}).Verify(context.Background(), "any"); !errors.Is(err, ports.ErrInvalidGoogleToken) {
		t.Errorf("err = %v, want ports.ErrInvalidGoogleToken", err)
	}
}
