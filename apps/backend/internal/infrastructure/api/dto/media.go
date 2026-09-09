package dto

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// MediaURLBuilder turns a content-addressed group id + variant into a URL
// under the immutable media proxy (GET /api/v1/media/...) - pure string
// concatenation and HMAC signing, no ctx, no I/O, no error return. This is
// the entire point of T2: NewStandResponse (and its DevilFruit/Stage/User
// siblings) stop making any network call to build a picture URL once an
// entity has a PictureMediaID.
//
// Public URLs never expire and carry no signature - see
// ObsidianVault/media-proxy-content-addressed.md for why that's safe for
// Stand/DevilFruit/Stage (identical for every viewer, already readable by
// any logged-in user). Private URLs (User avatars) are quantized to a fixed
// window so repeated calls within that window return the byte-identical
// URL - the property the browser HTTP cache and the response ETag over a
// list containing many avatar URLs depend on. The frontend service worker
// deliberately does NOT cache this scope (see
// ObsidianVault/sw-cache-media-scope-privado-2026-09-09.md): a rotating URL
// under Cache Storage's no-TTL, insertion-order eviction would churn out
// the truly-forever-valid public entries instead.
type MediaURLBuilder struct {
	BaseURL    string
	Secret     []byte
	PrivateTTL time.Duration
}

// NewMediaURLBuilder builds a MediaURLBuilder. secret should be
// config.MediaURLSecret; privateTTL should be config.MediaPrivateURLTTL.
func NewMediaURLBuilder(baseURL, secret string, privateTTL time.Duration) MediaURLBuilder {
	return MediaURLBuilder{BaseURL: strings.TrimSuffix(baseURL, "/"), Secret: []byte(secret), PrivateTTL: privateTTL}
}

// Public builds an unsigned, never-expiring URL for a public-scope group
// (Stand/DevilFruit/Stage). Returns "" if groupID is empty (not backfilled
// yet - the caller falls back to the presign path in that case).
func (b MediaURLBuilder) Public(groupID, variant string) string {
	if groupID == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s/%s.webp", b.BaseURL, groupID, variant)
}

// privateWindow is half of PrivateTTL - see (MediaURLBuilder).Private's doc
// for why halving (not using the full TTL as the quantization step) still
// guarantees every URL stays valid for at least PrivateTTL/2 from the
// moment it's handed out.
func (b MediaURLBuilder) privateWindow() time.Duration {
	if b.PrivateTTL <= 0 {
		return time.Hour
	}
	return b.PrivateTTL / 2
}

// Private builds a signed URL for a private-scope group (User avatars),
// quantizing exp to the current privateWindow boundary: every call within
// the same window returns the byte-identical URL, which is what lets the
// browser HTTP cache and the response ETag over a list containing many
// avatar URLs actually hit. Returns "" if groupID is empty.
func (b MediaURLBuilder) Private(groupID, variant string, now time.Time) string {
	if groupID == "" {
		return ""
	}
	window := b.privateWindow()
	exp := quantizeExp(now, window)
	sig := b.sign(exp, groupID, variant)
	return fmt.Sprintf("%s/p/%d/%s/%s/%s.webp", b.BaseURL, exp, sig, groupID, variant)
}

// quantizeExp rounds now up to the next window boundary (unix seconds), so
// two calls within the same window agree on exp exactly.
func quantizeExp(now time.Time, window time.Duration) int64 {
	secs := int64(window / time.Second)
	if secs <= 0 {
		secs = 1
	}
	nowSecs := now.Unix()
	return ((nowSecs / secs) + 1) * secs
}

// sign computes the signature covering exp|groupID|variant, base64url
// (no padding), truncated to 22 chars (132 bits - HMAC-SHA256 truncated
// this far is still infeasible to forge).
func (b MediaURLBuilder) sign(exp int64, groupID, variant string) string {
	mac := hmac.New(sha256.New, b.Secret)
	mac.Write([]byte(strconv.FormatInt(exp, 10)))
	mac.Write([]byte("|"))
	mac.Write([]byte(groupID))
	mac.Write([]byte("|"))
	mac.Write([]byte(variant))
	sum := mac.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(sum)[:22]
}

// VerifyPrivate reports whether sig is a valid, unexpired signature for
// (groupID, variant, exp) as of now - the media handler's entire private-URL
// auth check. Uses subtle.ConstantTimeCompare so signature verification
// cannot be timed to leak information about the expected value.
func (b MediaURLBuilder) VerifyPrivate(exp int64, groupID, variant, sig string, now time.Time) bool {
	if now.Unix() > exp {
		return false
	}
	want := b.sign(exp, groupID, variant)
	return subtle.ConstantTimeCompare([]byte(want), []byte(sig)) == 1
}
