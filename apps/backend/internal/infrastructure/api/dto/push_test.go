package dto_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/dto"
)

func validPushRequest(endpoint string) dto.PushSubscribeRequest {
	return dto.PushSubscribeRequest{
		Endpoint: endpoint,
		Keys: dto.PushKeys{
			P256dh: "BNcRdreALRFXTkOOUHK1EtK2wtaz5Ry4YfYCA_0QTpQtUbVlUls0VJXg7A8u-Ts1XbjhazAkj7I99e8QcYP7DkM",
			Auth:   "tBHItJI5svbpez7KI4CCXg",
		},
	}
}

func TestPushSubscribeRequest_AcceptsRealPushServices(t *testing.T) {
	endpoints := []string{
		"https://fcm.googleapis.com/fcm/send/abc:def",
		"https://updates.push.services.mozilla.com/wpush/v2/gAAAA",
		"https://updates-autopush.push.services.mozilla.com/wpush/v2/gAAAA",
		"https://web.push.apple.com/QGxyz",
		"https://par02p.notify.windows.com/w/?token=abc",
		"https://FCM.GoogleAPIs.com:443/fcm/send/x",
	}
	for _, endpoint := range endpoints {
		if err := validPushRequest(endpoint).Validate(); err != nil {
			t.Errorf("Validate(%q) = %v, want accepted", endpoint, err)
		}
	}
}

func TestPushSubscribeRequest_RejectsEndpointsThatCouldReachInternalHosts(t *testing.T) {
	endpoints := map[string]string{
		"empty":               "",
		"plain http":          "http://fcm.googleapis.com/fcm/send/x",
		"loopback":            "https://127.0.0.1/push",
		"internal service":    "https://backend:8080/api/v1/users/me",
		"cloud metadata":      "https://169.254.169.254/latest/meta-data",
		"lookalike suffix":    "https://evilfcm.googleapis.com.attacker.example/x",
		"lookalike prefix":    "https://notfcm.googleapis.com.evil/x",
		"userinfo trick":      "https://fcm.googleapis.com@attacker.example/x",
		"other port":          "https://fcm.googleapis.com:8443/x",
		"not a url":           "::::",
		"no host":             "https:///path",
		"unrelated google":    "https://www.googleapis.com/x",
		"oversized endpoint":  "https://fcm.googleapis.com/" + strings.Repeat("a", 3000),
		"javascript scheme":   "javascript:alert(1)",
		"uppercase bad host":  "https://ATTACKER.example/x",
		"subdomain of nobody": "https://fcm.googleapis.com.evil.example/x",
	}
	for name, endpoint := range endpoints {
		err := validPushRequest(endpoint).Validate()
		var verr *dto.ValidationError
		if !errors.As(err, &verr) {
			t.Errorf("%s: Validate(%q) = %v, want a ValidationError", name, endpoint, err)
			continue
		}
		if verr.Errors[0].Field != "endpoint" {
			t.Errorf("%s: first error field = %q, want endpoint", name, verr.Errors[0].Field)
		}
	}
}

func TestPushSubscribeRequest_RejectsBadKeys(t *testing.T) {
	cases := map[string]dto.PushKeys{
		"missing p256dh": {Auth: "tBHItJI5svbpez7KI4CCXg"},
		"missing auth":   {P256dh: "BNcRdreALRFXTkOOUHK1EtK2wtaz5Ry4"},
		"bad characters": {P256dh: "not valid!", Auth: "tBHItJI5svbpez7KI4CCXg"},
		"oversized":      {P256dh: strings.Repeat("A", 300), Auth: "tBHItJI5svbpez7KI4CCXg"},
	}
	for name, keys := range cases {
		req := validPushRequest("https://fcm.googleapis.com/fcm/send/x")
		req.Keys = keys
		var verr *dto.ValidationError
		if err := req.Validate(); !errors.As(err, &verr) {
			t.Errorf("%s: Validate = %v, want a ValidationError", name, err)
		}
	}
}

func TestPushSubscribeRequest_CollectsEveryProblem(t *testing.T) {
	err := dto.PushSubscribeRequest{Endpoint: "http://evil.example"}.Validate()
	var verr *dto.ValidationError
	if !errors.As(err, &verr) {
		t.Fatalf("Validate = %v, want a ValidationError", err)
	}
	if len(verr.Errors) != 3 {
		t.Errorf("got %d field errors, want 3 (endpoint, p256dh, auth): %+v", len(verr.Errors), verr.Errors)
	}
}

func TestPushUnsubscribeRequest_RequiresAnEndpoint(t *testing.T) {
	if err := (dto.PushUnsubscribeRequest{}).Validate(); err == nil {
		t.Error("empty endpoint accepted")
	}
	if err := (dto.PushUnsubscribeRequest{Endpoint: "https://fcm.googleapis.com/x"}).Validate(); err != nil {
		t.Errorf("Validate = %v, want accepted", err)
	}
}
