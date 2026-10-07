package endpoints

import (
	"net/http"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/dto"
)

// SetPushSubscriptions enables the Web Push routes under /users/me/push.
// Without it (no VAPID keys configured) GET /users/me/push reports
// enabled=false and the write routes answer 404, so the frontend simply
// hides the feature. Set before the router starts serving, like
// SetMediaURLBuilder.
func (e *UserEndpoints) SetPushSubscriptions(repo ports.IPushSubscriptionRepository, vapidPublicKey string) {
	e.push = repo
	e.pushPublicKey = vapidPublicKey
}

func (e *UserEndpoints) pushEnabled() bool {
	return e.push != nil && e.pushPublicKey != ""
}

func writePushDisabled(w http.ResponseWriter) {
	writeJSON(w, http.StatusNotFound, dto.ErrorResponse{Error: "push notifications are not enabled on this server"})
}

// getPushConfig godoc
//
//	@Summary		Web Push availability and the VAPID public key
//	@Description	The key the browser needs to subscribe. enabled=false when the server has no VAPID keys configured.
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.PushConfigResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Router			/users/me/push [get]
func (e *UserEndpoints) getPushConfig(w http.ResponseWriter, r *http.Request) error {
	if _, err := callerID(r); err != nil {
		return err
	}
	if !e.pushEnabled() {
		writeJSON(w, http.StatusOK, dto.PushConfigResponse{})
		return nil
	}
	writeJSON(w, http.StatusOK, dto.PushConfigResponse{Enabled: true, PublicKey: e.pushPublicKey})
	return nil
}

// subscribePush godoc
//
//	@Summary		Register this device for push notifications
//	@Description	Idempotent: re-registering an endpoint refreshes its keys, and an endpoint last registered by another account moves to the caller (a shared device). Only real browser push services are accepted as endpoints.
//	@Tags			users
//	@Accept			json
//	@Security		BearerAuth
//	@Param			request	body	dto.PushSubscribeRequest	true	"The browser's PushSubscription"
//	@Success		204
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Router			/users/me/push/subscriptions [post]
func (e *UserEndpoints) subscribePush(w http.ResponseWriter, r *http.Request) error {
	id, err := callerID(r)
	if err != nil {
		return err
	}
	if !e.pushEnabled() {
		writePushDisabled(w)
		return nil
	}
	var req dto.PushSubscribeRequest
	if err := decode(w, r, &req); err != nil {
		return err
	}
	if err := req.Validate(); err != nil {
		return err
	}
	if err := e.push.Upsert(r.Context(), ports.PushSubscription{
		UserID:   id,
		Endpoint: req.Endpoint,
		P256dh:   req.Keys.P256dh,
		Auth:     req.Keys.Auth,
	}); err != nil {
		return err
	}
	writeJSON(w, http.StatusNoContent, nil)
	return nil
}

// unsubscribePush godoc
//
//	@Summary		Unregister this device from push notifications
//	@Description	Only ever removes the caller's own subscription; an unknown endpoint is a no-op.
//	@Tags			users
//	@Accept			json
//	@Security		BearerAuth
//	@Param			request	body	dto.PushUnsubscribeRequest	true	"The subscription's endpoint"
//	@Success		204
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Router			/users/me/push/subscriptions [delete]
func (e *UserEndpoints) unsubscribePush(w http.ResponseWriter, r *http.Request) error {
	id, err := callerID(r)
	if err != nil {
		return err
	}
	if !e.pushEnabled() {
		writePushDisabled(w)
		return nil
	}
	var req dto.PushUnsubscribeRequest
	if err := decode(w, r, &req); err != nil {
		return err
	}
	if err := req.Validate(); err != nil {
		return err
	}
	if err := e.push.DeleteForUser(r.Context(), id, req.Endpoint); err != nil {
		return err
	}
	writeJSON(w, http.StatusNoContent, nil)
	return nil
}
