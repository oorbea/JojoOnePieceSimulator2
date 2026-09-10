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

// OnePieceCharacterEndpoints wires the OnePieceCharacter HTTP surface to
// the application service - same shape as JojoCharacterEndpoints.
type OnePieceCharacterEndpoints struct {
	svc   *services.OnePieceCharacterService
	media dto.MediaURLBuilder
}

func NewOnePieceCharacterEndpoints(svc *services.OnePieceCharacterService) *OnePieceCharacterEndpoints {
	return &OnePieceCharacterEndpoints{svc: svc}
}

func (e *OnePieceCharacterEndpoints) SetMediaURLBuilder(media dto.MediaURLBuilder) {
	e.media = media
}

// Routes returns the /one-piece-characters sub-router - same shape as
// JojoCharacterEndpoints.Routes.
func (e *OnePieceCharacterEndpoints) Routes(rateCfg RateLimitConfig, cacheCfg CacheConfig) chi.Router {
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
//	@Summary		List or filter One Piece characters
//	@Description	Lists every OnePieceCharacter, or filters them if any query param is set.
//	@Tags			one-piece-characters
//	@Produce		json
//	@Security		BearerAuth
//	@Param			rarity			query		string	false	"COMMON, RARE, EPIC, LEGENDARY, MYTHICAL"
//	@Param			physicalForm	query		string	false	"PRIVATE, STRONG_FISHMAN, MARINE_CAPTAIN, VICE_ADMIRAL, YONKO_COMMANDER, YONKO_PLUS"
//	@Param			armamentHaki	query		string	false	"NONE, PRIVATE, VICE_ADMIRAL, YONKO_COMMANDER, YONKO_PLUS"
//	@Param			observationHaki	query		string	false	"NONE, PRIVATE, VICE_ADMIRAL, YONKO_COMMANDER, YONKO_PLUS"
//	@Param			conquerorHaki	query		string	false	"NONE, PRIVATE, VICE_ADMIRAL, YONKO_COMMANDER, YONKO_PLUS"
//	@Param			fruitMastery	query		string	false	"NONE, REGULAR, ADVANCED, AWAKENED"
//	@Param			q				query		string	false	"free-text search over name and description"
//	@Success		200				{array}		dto.OnePieceCharacterResponse
//	@Success		304
//	@Failure		400				{object}	dto.ErrorResponse
//	@Failure		401				{object}	dto.ErrorResponse
//	@Failure		429				{object}	dto.ErrorResponse
//	@Router			/one-piece-characters [get]
func (e *OnePieceCharacterEndpoints) list(w http.ResponseWriter, r *http.Request) error {
	filters, hasFilters, err := dto.OnePieceCharacterFiltersFromQuery(r.URL.Query())
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

	var list []*characters.OnePieceCharacter
	if hasFilters {
		list, err = e.svc.FilterOnePieceCharacters(r.Context(), filters, locale)
	} else {
		list, err = e.svc.ListOnePieceCharacters(r.Context(), locale)
	}
	if err != nil {
		return err
	}
	resp, err := dto.NewOnePieceCharacterResponses(r.Context(), list, e.svc.PictureURL, e.media)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, resp)
	return nil
}

// listPage serves the ?limit=/?cursor= paginated form of GET
// /one-piece-characters - see JojoCharacterEndpoints.listPage's doc.
func (e *OnePieceCharacterEndpoints) listPage(w http.ResponseWriter, r *http.Request, filters ports.OnePieceCharacterFilters, locale enums.Locale, params dto.PageParams) error {
	fingerprint := dto.OnePieceCharacterFiltersFingerprint(filters, locale)

	var afterName *string
	if params.HasCursor {
		cursor, err := dto.DecodeCursor[dto.OnePieceCharacterCursor](params.Cursor, fingerprint)
		if err != nil {
			return err
		}
		afterName = &cursor.Name
	}

	list, hasMore, err := e.svc.PageOnePieceCharacters(r.Context(), filters, locale, afterName, params.Limit)
	if err != nil {
		return err
	}

	var total *int
	if params.WithTotal && !params.HasCursor {
		count, err := e.svc.CountOnePieceCharacters(r.Context(), filters, locale)
		if err != nil {
			return err
		}
		total = &count
	}

	var nextCursor *string
	if hasMore && len(list) > 0 {
		encoded := dto.EncodeCursor(dto.OnePieceCharacterCursor{Name: list[len(list)-1].Name()}, fingerprint)
		nextCursor = &encoded
	}

	items, err := dto.NewOnePieceCharacterResponses(r.Context(), list, e.svc.PictureURL, e.media)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, dto.OnePieceCharacterPageResponse{
		PageInfo: dto.NewPageInfo(nextCursor, total),
		Items:    items,
	})
	return nil
}

// create godoc
//
//	@Summary		Create a One Piece character
//	@Description	Admin only.
//	@Tags			one-piece-characters
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.OnePieceCharacterRequest	true	"Character to create"
//	@Success		201		{object}	dto.OnePieceCharacterResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		409		{object}	dto.ErrorResponse
//	@Failure		429		{object}	dto.ErrorResponse
//	@Router			/one-piece-characters [post]
func (e *OnePieceCharacterEndpoints) create(w http.ResponseWriter, r *http.Request) error {
	var req dto.OnePieceCharacterRequest
	if err := decode(w, r, &req); err != nil {
		return err
	}

	input, err := req.Validate()
	if err != nil {
		return err
	}

	c, err := e.svc.CreateOnePieceCharacter(r.Context(), input)
	if err != nil {
		return err
	}

	resp, err := dto.NewOnePieceCharacterResponse(r.Context(), c, e.svc.PictureURL, e.media)
	if err != nil {
		return err
	}

	w.Header().Set("Location", fmt.Sprintf("/api/v1/one-piece-characters/%s", c.ID()))
	writeJSON(w, http.StatusCreated, resp)
	return nil
}

// get godoc
//
//	@Summary		Get a One Piece character by id
//	@Tags			one-piece-characters
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Character id (UUID)"
//	@Success		200	{object}	dto.OnePieceCharacterResponse
//	@Success		304
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		429	{object}	dto.ErrorResponse
//	@Router			/one-piece-characters/{id} [get]
func (e *OnePieceCharacterEndpoints) get(w http.ResponseWriter, r *http.Request) error {
	id, err := parseCharacterID(r)
	if err != nil {
		return err
	}

	c, err := e.svc.GetOnePieceCharacter(r.Context(), id, LocaleFromRequest(r))
	if err != nil {
		return err
	}
	resp, err := dto.NewOnePieceCharacterResponse(r.Context(), c, e.svc.PictureURL, e.media)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, resp)
	return nil
}

// update godoc
//
//	@Summary		Replace a One Piece character
//	@Description	Admin only. Keeps the original id.
//	@Tags			one-piece-characters
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string							true	"Character id (UUID)"
//	@Param			request	body		dto.OnePieceCharacterRequest	true	"Replacement character"
//	@Success		200		{object}	dto.OnePieceCharacterResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		409		{object}	dto.ErrorResponse
//	@Failure		429		{object}	dto.ErrorResponse
//	@Router			/one-piece-characters/{id} [put]
func (e *OnePieceCharacterEndpoints) update(w http.ResponseWriter, r *http.Request) error {
	id, err := parseCharacterID(r)
	if err != nil {
		return err
	}

	var req dto.OnePieceCharacterRequest
	if err := decode(w, r, &req); err != nil {
		return err
	}

	input, err := req.Validate()
	if err != nil {
		return err
	}

	c, err := e.svc.UpdateOnePieceCharacter(r.Context(), id, input)
	if err != nil {
		return err
	}
	resp, err := dto.NewOnePieceCharacterResponse(r.Context(), c, e.svc.PictureURL, e.media)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, resp)
	return nil
}

// patchPicture godoc
//
//	@Summary		Set a One Piece character's picture
//	@Description	Admin only. Uploads the image to object storage and stores its key; the response's `picture` is a presigned URL.
//	@Tags			one-piece-characters
//	@Accept			mpfd
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string	true	"Character id (UUID)"
//	@Param			picture	formData	file	true	"Image file (WebP, AVIF, JPEG, PNG or GIF)"
//	@Success		202		{object}	dto.OnePieceCharacterResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		413		{object}	dto.ErrorResponse
//	@Failure		429		{object}	dto.ErrorResponse
//	@Failure		503		{object}	dto.ErrorResponse
//	@Router			/one-piece-characters/{id}/picture [patch]
func (e *OnePieceCharacterEndpoints) patchPicture(w http.ResponseWriter, r *http.Request) error {
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

	c, err := e.svc.SetOnePieceCharacterPicture(r.Context(), id, ports.Picture{
		Content:     bytes.NewReader(buf),
		ContentType: contentType,
		Size:        int64(len(buf)),
	})
	if err != nil {
		return err
	}

	resp, err := dto.NewOnePieceCharacterResponse(r.Context(), c, e.svc.PictureURL, e.media)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusAccepted, resp)
	return nil
}

// translations godoc
//
//	@Summary		Get every locale's translation for a One Piece character
//	@Description	Admin only. Always returns every locale's description at once, for an edit form.
//	@Tags			one-piece-characters
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Character id (UUID)"
//	@Success		200	{object}	dto.CharacterTranslationsResponse
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		429	{object}	dto.ErrorResponse
//	@Router			/one-piece-characters/{id}/translations [get]
func (e *OnePieceCharacterEndpoints) translations(w http.ResponseWriter, r *http.Request) error {
	id, err := parseCharacterID(r)
	if err != nil {
		return err
	}
	translations, err := e.svc.OnePieceCharacterTranslations(r.Context(), id)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, dto.NewCharacterTranslationsResponse(translations))
	return nil
}

// delete godoc
//
//	@Summary		Delete a One Piece character
//	@Description	Admin only.
//	@Tags			one-piece-characters
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Character id (UUID)"
//	@Success		204
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		429	{object}	dto.ErrorResponse
//	@Router			/one-piece-characters/{id} [delete]
func (e *OnePieceCharacterEndpoints) delete(w http.ResponseWriter, r *http.Request) error {
	id, err := parseCharacterID(r)
	if err != nil {
		return err
	}

	if err := e.svc.DeleteOnePieceCharacter(r.Context(), id); err != nil {
		return err
	}
	writeJSON(w, http.StatusNoContent, nil)
	return nil
}
