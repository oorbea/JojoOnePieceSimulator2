package endpoints

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/characters"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/dto"
)

// JojoCharacterEndpoints wires the JojoCharacter HTTP surface to the
// application service - same shape as DevilFruitEndpoints, reusing the
// maxRequestBodyBytes/maxMultipartMemory/sniffLen constants and the
// sniffContentType/decode helpers declared in stand_endpoints.go.
type JojoCharacterEndpoints struct {
	svc   *services.JojoCharacterService
	media dto.MediaURLBuilder
}

func NewJojoCharacterEndpoints(svc *services.JojoCharacterService) *JojoCharacterEndpoints {
	return &JojoCharacterEndpoints{svc: svc}
}

// SetMediaURLBuilder - see StandEndpoints.SetMediaURLBuilder's doc.
func (e *JojoCharacterEndpoints) SetMediaURLBuilder(media dto.MediaURLBuilder) {
	e.media = media
}

// Routes returns the /jojo-characters sub-router - same shape as
// DevilFruitEndpoints.Routes.
func (e *JojoCharacterEndpoints) Routes(rateCfg RateLimitConfig, cacheCfg CacheConfig) chi.Router {
	r := chi.NewRouter()
	read := readRateLimit(rateCfg)
	cache := cacheHeaders(cacheCfg)
	r.With(read, cache).Get("/", Wrap(e.list))
	r.With(read, cache).Get("/{id}", Wrap(e.get))

	r.Group(func(r chi.Router) {
		r.Use(RequireAdmin)
		r.Use(writeRateLimit(rateCfg))
		r.Post("/", Wrap(e.create))
		r.Put("/{id}", Wrap(e.update))
		r.Patch("/{id}/picture", Wrap(e.patchPicture))
		r.Delete("/{id}", Wrap(e.delete))
		r.Get("/{id}/translations", Wrap(e.translations))
	})
	return r
}

// list godoc
//
//	@Summary		List or filter JoJo characters
//	@Description	Lists every JojoCharacter, or filters them if any query param is set.
//	@Tags			jojo-characters
//	@Produce		json
//	@Security		BearerAuth
//	@Param			rarity		query		string	false	"COMMON, RARE, EPIC, LEGENDARY, MYTHICAL"
//	@Param			hamon		query		string	false	"NONE, BASIC, ADVANCED, PERFECT"
//	@Param			spin		query		string	false	"NONE, BASIC, GOLDEN, INFINITE"
//	@Param			battleIq	query		int		false	"0-255"
//	@Param			q			query		string	false	"free-text search over name and description"
//	@Success		200			{array}		dto.JojoCharacterResponse
//	@Success		304
//	@Failure		400			{object}	dto.ErrorResponse
//	@Failure		401			{object}	dto.ErrorResponse
//	@Failure		429			{object}	dto.ErrorResponse
//	@Router			/jojo-characters [get]
func (e *JojoCharacterEndpoints) list(w http.ResponseWriter, r *http.Request) error {
	filters, hasFilters, err := dto.JojoCharacterFiltersFromQuery(r.URL.Query())
	if err != nil {
		return err
	}

	locale := LocaleFromRequest(r)

	pageParams, err := dto.PageParamsFromQuery(r.URL.Query())
	if err != nil {
		return err
	}
	if pageParams.Requested {
		return e.listPage(w, r, filters, locale, pageParams)
	}

	var list []*characters.JojoCharacter
	if hasFilters {
		list, err = e.svc.FilterJojoCharacters(r.Context(), filters, locale)
	} else {
		list, err = e.svc.ListJojoCharacters(r.Context(), locale)
	}
	if err != nil {
		return err
	}
	resp, err := dto.NewJojoCharacterResponses(r.Context(), list, e.svc.PictureURL, e.media)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, resp)
	return nil
}

// listPage serves the ?limit=/?cursor= paginated form of GET
// /jojo-characters - see StandEndpoints.listPage's doc.
func (e *JojoCharacterEndpoints) listPage(w http.ResponseWriter, r *http.Request, filters ports.JojoCharacterFilters, locale enums.Locale, params dto.PageParams) error {
	fingerprint := dto.JojoCharacterFiltersFingerprint(filters, locale)

	var afterName *string
	if params.HasCursor {
		cursor, err := dto.DecodeCursor[dto.JojoCharacterCursor](params.Cursor, fingerprint)
		if err != nil {
			return err
		}
		afterName = &cursor.Name
	}

	list, hasMore, err := e.svc.PageJojoCharacters(r.Context(), filters, locale, afterName, params.Limit)
	if err != nil {
		return err
	}

	var total *int
	if params.WithTotal && !params.HasCursor {
		count, err := e.svc.CountJojoCharacters(r.Context(), filters, locale)
		if err != nil {
			return err
		}
		total = &count
	}

	var nextCursor *string
	if hasMore && len(list) > 0 {
		encoded := dto.EncodeCursor(dto.JojoCharacterCursor{Name: list[len(list)-1].Name()}, fingerprint)
		nextCursor = &encoded
	}

	items, err := dto.NewJojoCharacterResponses(r.Context(), list, e.svc.PictureURL, e.media)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, dto.JojoCharacterPageResponse{
		PageInfo: dto.NewPageInfo(nextCursor, total),
		Items:    items,
	})
	return nil
}

// create godoc
//
//	@Summary		Create a JoJo character
//	@Description	Admin only.
//	@Tags			jojo-characters
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.JojoCharacterRequest	true	"Character to create"
//	@Success		201		{object}	dto.JojoCharacterResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		409		{object}	dto.ErrorResponse
//	@Failure		429		{object}	dto.ErrorResponse
//	@Router			/jojo-characters [post]
func (e *JojoCharacterEndpoints) create(w http.ResponseWriter, r *http.Request) error {
	var req dto.JojoCharacterRequest
	if err := decode(w, r, &req); err != nil {
		return err
	}

	input, err := req.Validate()
	if err != nil {
		return err
	}

	c, err := e.svc.CreateJojoCharacter(r.Context(), input)
	if err != nil {
		return err
	}

	resp, err := dto.NewJojoCharacterResponse(r.Context(), c, e.svc.PictureURL, e.media)
	if err != nil {
		return err
	}

	w.Header().Set("Location", fmt.Sprintf("/api/v1/jojo-characters/%s", c.ID()))
	writeJSON(w, http.StatusCreated, resp)
	return nil
}

// get godoc
//
//	@Summary		Get a JoJo character by id
//	@Tags			jojo-characters
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Character id (UUID)"
//	@Success		200	{object}	dto.JojoCharacterResponse
//	@Success		304
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		429	{object}	dto.ErrorResponse
//	@Router			/jojo-characters/{id} [get]
func (e *JojoCharacterEndpoints) get(w http.ResponseWriter, r *http.Request) error {
	id, err := parseCharacterID(r)
	if err != nil {
		return err
	}

	c, err := e.svc.GetJojoCharacter(r.Context(), id, LocaleFromRequest(r))
	if err != nil {
		return err
	}
	resp, err := dto.NewJojoCharacterResponse(r.Context(), c, e.svc.PictureURL, e.media)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, resp)
	return nil
}

// update godoc
//
//	@Summary		Replace a JoJo character
//	@Description	Admin only. Keeps the original id.
//	@Tags			jojo-characters
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"Character id (UUID)"
//	@Param			request	body		dto.JojoCharacterRequest	true	"Replacement character"
//	@Success		200		{object}	dto.JojoCharacterResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		409		{object}	dto.ErrorResponse
//	@Failure		429		{object}	dto.ErrorResponse
//	@Router			/jojo-characters/{id} [put]
func (e *JojoCharacterEndpoints) update(w http.ResponseWriter, r *http.Request) error {
	id, err := parseCharacterID(r)
	if err != nil {
		return err
	}

	var req dto.JojoCharacterRequest
	if err := decode(w, r, &req); err != nil {
		return err
	}

	input, err := req.Validate()
	if err != nil {
		return err
	}

	c, err := e.svc.UpdateJojoCharacter(r.Context(), id, input)
	if err != nil {
		return err
	}
	resp, err := dto.NewJojoCharacterResponse(r.Context(), c, e.svc.PictureURL, e.media)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, resp)
	return nil
}

// patchPicture godoc
//
//	@Summary		Set a JoJo character's picture
//	@Description	Admin only. Uploads the image to object storage and stores its key; the response's `picture` is a presigned URL.
//	@Tags			jojo-characters
//	@Accept			mpfd
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string	true	"Character id (UUID)"
//	@Param			picture	formData	file	true	"Image file (WebP, AVIF, JPEG, PNG or GIF)"
//	@Success		202		{object}	dto.JojoCharacterResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		413		{object}	dto.ErrorResponse
//	@Failure		429		{object}	dto.ErrorResponse
//	@Failure		503		{object}	dto.ErrorResponse
//	@Router			/jojo-characters/{id}/picture [patch]
func (e *JojoCharacterEndpoints) patchPicture(w http.ResponseWriter, r *http.Request) error {
	id, err := parseCharacterID(r)
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

	c, err := e.svc.SetJojoCharacterPicture(r.Context(), id, ports.Picture{
		Content:     bytes.NewReader(buf),
		ContentType: contentType,
		Size:        int64(len(buf)),
	})
	if err != nil {
		return err
	}

	resp, err := dto.NewJojoCharacterResponse(r.Context(), c, e.svc.PictureURL, e.media)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusAccepted, resp)
	return nil
}

// translations godoc
//
//	@Summary		Get every locale's translation for a JoJo character
//	@Description	Admin only. Always returns every locale's description at once, for an edit form.
//	@Tags			jojo-characters
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Character id (UUID)"
//	@Success		200	{object}	dto.CharacterTranslationsResponse
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		429	{object}	dto.ErrorResponse
//	@Router			/jojo-characters/{id}/translations [get]
func (e *JojoCharacterEndpoints) translations(w http.ResponseWriter, r *http.Request) error {
	id, err := parseCharacterID(r)
	if err != nil {
		return err
	}
	translations, err := e.svc.JojoCharacterTranslations(r.Context(), id)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, dto.NewCharacterTranslationsResponse(translations))
	return nil
}

// delete godoc
//
//	@Summary		Delete a JoJo character
//	@Description	Admin only.
//	@Tags			jojo-characters
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Character id (UUID)"
//	@Success		204
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		429	{object}	dto.ErrorResponse
//	@Router			/jojo-characters/{id} [delete]
func (e *JojoCharacterEndpoints) delete(w http.ResponseWriter, r *http.Request) error {
	id, err := parseCharacterID(r)
	if err != nil {
		return err
	}

	if err := e.svc.DeleteJojoCharacter(r.Context(), id); err != nil {
		return err
	}
	writeJSON(w, http.StatusNoContent, nil)
	return nil
}

func parseCharacterID(r *http.Request) (characters.CharacterID, error) {
	return characters.ParseCharacterID(chi.URLParam(r, "id"))
}
