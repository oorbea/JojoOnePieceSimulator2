package endpoints

import (
	"bytes"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/dto"
)

// defaultUserListLimit and maxUserListLimit bound GET /users' page size.
const (
	defaultUserListLimit = 50
	maxUserListLimit     = 200
)

// UserEndpoints wires the user profile/admin HTTP surface to the
// application service.
type UserEndpoints struct {
	svc *services.UserService
}

func NewUserEndpoints(svc *services.UserService) *UserEndpoints {
	return &UserEndpoints{svc: svc}
}

// Routes returns the /users sub-router. Every /me route resolves the
// authenticated caller's id EXCLUSIVELY from the request's claims (set by
// RequireAuth) - never from the path or the body - so a caller can only ever
// act on their own profile through these routes. GET /{id} is the one
// exception: it is reachable by any authenticated caller but only ever
// returns dto.PublicUserResponse, which has no email or role field to leak.
// The remaining admin-only routes sit behind RequireAdmin, mirroring
// stand_endpoints.go's Routes.
func (e *UserEndpoints) Routes(rateCfg RateLimitConfig) chi.Router {
	r := chi.NewRouter()
	read := readRateLimit(rateCfg)
	write := writeRateLimit(rateCfg)

	r.With(read).Get("/me", Wrap(e.getMe))
	r.With(write).Patch("/me", Wrap(e.updateMe))
	r.With(write).Patch("/me/picture", Wrap(e.patchMePicture))
	r.With(write).Delete("/me/picture", Wrap(e.deleteMePicture))
	r.With(write).Delete("/me", Wrap(e.deleteMe))
	r.With(read).Get("/{id}", Wrap(e.getByID))

	r.Group(func(r chi.Router) {
		r.Use(RequireAdmin)
		r.Use(write)
		r.Get("/", Wrap(e.list))
		r.Patch("/{id}", Wrap(e.adminUpdateUsername))
		r.Patch("/{id}/role", Wrap(e.updateRole))
		r.Delete("/{id}", Wrap(e.deleteByID))
	})
	return r
}

// callerID returns the authenticated caller's id from the request's claims.
// Every /me handler must use this, and only this, to decide whose profile to
// act on.
func callerID(r *http.Request) (user.UserID, error) {
	claims, ok := ClaimsFromRequest(r)
	if !ok {
		return user.UserID{}, ports.ErrUnauthenticated
	}
	return claims.UserID, nil
}

// getMe godoc
//
//	@Summary		Get the authenticated caller's profile
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.UserResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Router			/users/me [get]
func (e *UserEndpoints) getMe(w http.ResponseWriter, r *http.Request) error {
	id, err := callerID(r)
	if err != nil {
		return err
	}
	u, err := e.svc.GetByID(r.Context(), id)
	if err != nil {
		return err
	}
	resp, err := dto.NewUserResponse(r.Context(), u, e.svc.AvatarURL)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, resp)
	return nil
}

// updateMe godoc
//
//	@Summary		Update the authenticated caller's username
//	@Description	The only self-editable field. email, role, and completeName are never accepted here - sending them is a 400.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.UpdateProfileRequest	true	"New username"
//	@Success		200		{object}	dto.UserResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		409		{object}	dto.ErrorResponse
//	@Router			/users/me [patch]
func (e *UserEndpoints) updateMe(w http.ResponseWriter, r *http.Request) error {
	id, err := callerID(r)
	if err != nil {
		return err
	}
	var req dto.UpdateProfileRequest
	if err := decode(w, r, &req); err != nil {
		return err
	}
	if err := req.Validate(); err != nil {
		return err
	}
	u, err := e.svc.ChangeUsername(r.Context(), id, req.Username)
	if err != nil {
		return err
	}
	resp, err := dto.NewUserResponse(r.Context(), u, e.svc.AvatarURL)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, resp)
	return nil
}

// patchMePicture godoc
//
//	@Summary		Set the authenticated caller's avatar
//	@Description	Uploads the image to object storage; the response's avatar fields are the previous renditions until the background worker finishes.
//	@Tags			users
//	@Accept			mpfd
//	@Produce		json
//	@Security		BearerAuth
//	@Param			picture	formData	file	true	"Image file (WebP, AVIF, JPEG, PNG or GIF)"
//	@Success		202		{object}	dto.UserResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		413		{object}	dto.ErrorResponse
//	@Failure		429		{object}	dto.ErrorResponse
//	@Failure		503		{object}	dto.ErrorResponse
//	@Router			/users/me/picture [patch]
func (e *UserEndpoints) patchMePicture(w http.ResponseWriter, r *http.Request) error {
	id, err := callerID(r)
	if err != nil {
		return err
	}

	r.Body = http.MaxBytesReader(w, r.Body, e.svc.MaxPictureBytes()+maxMultipartMemory)
	if err := r.ParseMultipartForm(maxMultipartMemory); err != nil {
		return &dto.ValidationError{Errors: []string{err.Error()}}
	}
	defer func() {
		_ = r.MultipartForm.RemoveAll()
	}()

	file, _, err := r.FormFile("picture")
	if err != nil {
		return services.ErrPictureRequired
	}
	defer func() {
		_ = file.Close()
	}()

	maxBytes := e.svc.MaxPictureBytes()
	buf, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if err != nil {
		return err
	}
	if maxBytes > 0 && int64(len(buf)) > maxBytes {
		return services.ErrPictureTooLarge
	}

	head := buf
	if len(head) > sniffLen {
		head = head[:sniffLen]
	}
	contentType := sniffContentType(head)

	u, err := e.svc.SetAvatar(r.Context(), id, ports.Picture{
		Content:     bytes.NewReader(buf),
		ContentType: contentType,
		Size:        int64(len(buf)),
	})
	if err != nil {
		return err
	}

	resp, err := dto.NewUserResponse(r.Context(), u, e.svc.AvatarURL)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusAccepted, resp)
	return nil
}

// deleteMePicture godoc
//
//	@Summary		Delete the authenticated caller's avatar
//	@Description	Reverts display back to the Google-synced picture.
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.UserResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Router			/users/me/picture [delete]
func (e *UserEndpoints) deleteMePicture(w http.ResponseWriter, r *http.Request) error {
	id, err := callerID(r)
	if err != nil {
		return err
	}
	u, err := e.svc.DeleteAvatar(r.Context(), id)
	if err != nil {
		return err
	}
	resp, err := dto.NewUserResponse(r.Context(), u, e.svc.AvatarURL)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, resp)
	return nil
}

// deleteMe godoc
//
//	@Summary		Delete the authenticated caller's account
//	@Tags			users
//	@Security		BearerAuth
//	@Success		204
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		409	{object}	dto.ErrorResponse
//	@Router			/users/me [delete]
func (e *UserEndpoints) deleteMe(w http.ResponseWriter, r *http.Request) error {
	id, err := callerID(r)
	if err != nil {
		return err
	}
	if err := e.svc.Delete(r.Context(), id); err != nil {
		return err
	}
	writeJSON(w, http.StatusNoContent, nil)
	return nil
}

// getByID godoc
//
//	@Summary		Get another user's public profile
//	@Description	Never includes email or role.
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"User id (UUID)"
//	@Success		200	{object}	dto.PublicUserResponse
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Router			/users/{id} [get]
func (e *UserEndpoints) getByID(w http.ResponseWriter, r *http.Request) error {
	id, err := user.ParseUserID(chi.URLParam(r, "id"))
	if err != nil {
		return &dto.ValidationError{Errors: []string{"id: " + err.Error()}}
	}
	u, err := e.svc.GetByID(r.Context(), id)
	if err != nil {
		return err
	}
	resp, err := dto.NewPublicUserResponse(r.Context(), u, e.svc.AvatarURL)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, resp)
	return nil
}

// list godoc
//
//	@Summary		List users
//	@Description	Admin only.
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Param			limit	query		int	false	"Page size (default 50, max 200)"
//	@Param			offset	query		int	false	"Rows to skip"
//	@Success		200		{array}		dto.UserResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Router			/users [get]
func (e *UserEndpoints) list(w http.ResponseWriter, r *http.Request) error {
	limit := defaultUserListLimit
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			return &dto.ValidationError{Errors: []string{"limit: must be a positive integer"}}
		}
		limit = parsed
	}
	if limit > maxUserListLimit {
		limit = maxUserListLimit
	}

	offset := 0
	if raw := r.URL.Query().Get("offset"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			return &dto.ValidationError{Errors: []string{"offset: must be a non-negative integer"}}
		}
		offset = parsed
	}

	users, err := e.svc.List(r.Context(), int32(limit), int32(offset))
	if err != nil {
		return err
	}
	resp := make([]dto.UserResponse, 0, len(users))
	for _, u := range users {
		userResp, err := dto.NewUserResponse(r.Context(), u, e.svc.AvatarURL)
		if err != nil {
			return err
		}
		resp = append(resp, userResp)
	}
	writeJSON(w, http.StatusOK, resp)
	return nil
}

// adminUpdateUsername godoc
//
//	@Summary		Change another user's username
//	@Description	Admin only.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string							true	"User id (UUID)"
//	@Param			request	body		dto.AdminUpdateUserRequest	true	"New username"
//	@Success		200		{object}	dto.UserResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		409		{object}	dto.ErrorResponse
//	@Router			/users/{id} [patch]
func (e *UserEndpoints) adminUpdateUsername(w http.ResponseWriter, r *http.Request) error {
	id, err := user.ParseUserID(chi.URLParam(r, "id"))
	if err != nil {
		return &dto.ValidationError{Errors: []string{"id: " + err.Error()}}
	}
	var req dto.AdminUpdateUserRequest
	if err := decode(w, r, &req); err != nil {
		return err
	}
	if err := req.Validate(); err != nil {
		return err
	}
	u, err := e.svc.ChangeUsername(r.Context(), id, req.Username)
	if err != nil {
		return err
	}
	resp, err := dto.NewUserResponse(r.Context(), u, e.svc.AvatarURL)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, resp)
	return nil
}

// updateRole godoc
//
//	@Summary		Change another user's role
//	@Description	Admin only. An admin cannot change their own role through this route (use another admin account), and the last remaining admin cannot be demoted. A role change only takes effect on the affected user's next login, since the role is carried in the JWT. ADMIN_EMAILS re-promotes a listed email on every login, so demoting one of those emails is temporary.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"User id (UUID)"
//	@Param			request	body		dto.UpdateRoleRequest	true	"New role"
//	@Success		200		{object}	dto.UserResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		409		{object}	dto.ErrorResponse
//	@Router			/users/{id}/role [patch]
func (e *UserEndpoints) updateRole(w http.ResponseWriter, r *http.Request) error {
	id, err := user.ParseUserID(chi.URLParam(r, "id"))
	if err != nil {
		return &dto.ValidationError{Errors: []string{"id: " + err.Error()}}
	}
	me, err := callerID(r)
	if err != nil {
		return err
	}
	if me == id {
		return ports.ErrForbidden
	}
	var req dto.UpdateRoleRequest
	if err := decode(w, r, &req); err != nil {
		return err
	}
	role, err := req.Validate()
	if err != nil {
		return err
	}
	u, err := e.svc.ChangeRole(r.Context(), id, role)
	if err != nil {
		return err
	}
	resp, err := dto.NewUserResponse(r.Context(), u, e.svc.AvatarURL)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, resp)
	return nil
}

// deleteByID godoc
//
//	@Summary		Delete another user
//	@Description	Admin only. An admin cannot delete their own account through this route, and the last remaining admin cannot be deleted.
//	@Tags			users
//	@Security		BearerAuth
//	@Param			id	path	string	true	"User id (UUID)"
//	@Success		204
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		409	{object}	dto.ErrorResponse
//	@Router			/users/{id} [delete]
func (e *UserEndpoints) deleteByID(w http.ResponseWriter, r *http.Request) error {
	id, err := user.ParseUserID(chi.URLParam(r, "id"))
	if err != nil {
		return &dto.ValidationError{Errors: []string{"id: " + err.Error()}}
	}
	me, err := callerID(r)
	if err != nil {
		return err
	}
	if me == id {
		return ports.ErrForbidden
	}
	if err := e.svc.Delete(r.Context(), id); err != nil {
		return err
	}
	writeJSON(w, http.StatusNoContent, nil)
	return nil
}
