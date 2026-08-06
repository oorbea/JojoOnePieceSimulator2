package endpoints

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/dto"
)

// DevilFruitEndpoints wires the DevilFruit HTTP surface to the application
// service. It reuses the maxRequestBodyBytes/maxMultipartMemory/sniffLen
// constants and the sniffContentType/parsePowerID/decode helpers declared in
// stand_endpoints.go - both catalogues share the same JSON/multipart rules.
type DevilFruitEndpoints struct {
	svc *services.DevilFruitService
}

func NewDevilFruitEndpoints(svc *services.DevilFruitService) *DevilFruitEndpoints {
	return &DevilFruitEndpoints{svc: svc}
}

// Routes returns the /devil-fruits sub-router: GET/POST on the collection,
// GET/PUT/PATCH/DELETE on a single devil fruit by id. Same rate-limit and
// cache-header layering as StandEndpoints.Routes.
func (e *DevilFruitEndpoints) Routes(rateCfg RateLimitConfig, cacheCfg CacheConfig) chi.Router {
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
//	@Summary		List or filter devil fruits
//	@Description	Lists every DevilFruit, or filters them if any query param is set.
//	@Tags			devil-fruits
//	@Produce		json
//	@Security		BearerAuth
//	@Param			rarity		query		string	false	"COMMON, RARE, EPIC, LEGENDARY"
//	@Param			fruitType	query		string	false	"PARAMECIA, ZOAN, LOGIA, SPECIAL_PARAMECIA, ANCIENT_ZOAN, MYTHICAL_ZOAN"
//	@Description	Responses carry an ETag; a request with a matching If-None-Match
//	@Description	gets 304 Not Modified with no body instead of the full list.
//	@Success		200			{array}		dto.DevilFruitResponse
//	@Success		304
//	@Failure		400			{object}	dto.ErrorResponse
//	@Failure		401			{object}	dto.ErrorResponse
//	@Failure		429			{object}	dto.ErrorResponse
//	@Router			/devil-fruits [get]
func (e *DevilFruitEndpoints) list(w http.ResponseWriter, r *http.Request) error {
	filters, hasFilters, err := dto.DevilFruitFiltersFromQuery(r.URL.Query())
	if err != nil {
		return err
	}

	locale := LocaleFromRequest(r)
	var fruits []*powers.DevilFruit
	if hasFilters {
		fruits, err = e.svc.FilterDevilFruits(r.Context(), filters, locale)
	} else {
		fruits, err = e.svc.ListDevilFruits(r.Context(), locale)
	}
	if err != nil {
		return err
	}
	resp, err := dto.NewDevilFruitResponses(r.Context(), fruits, e.svc.PictureURL)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, resp)
	return nil
}

// create godoc
//
//	@Summary		Create a devil fruit
//	@Description	Admin only.
//	@Tags			devil-fruits
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.DevilFruitRequest	true	"Devil fruit to create"
//	@Success		201		{object}	dto.DevilFruitResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		409		{object}	dto.ErrorResponse
//	@Failure		429		{object}	dto.ErrorResponse
//	@Router			/devil-fruits [post]
func (e *DevilFruitEndpoints) create(w http.ResponseWriter, r *http.Request) error {
	var req dto.DevilFruitRequest
	if err := decode(w, r, &req); err != nil {
		return err
	}

	input, err := req.Validate()
	if err != nil {
		return err
	}

	fruit, err := e.svc.CreateDevilFruit(r.Context(), input)
	if err != nil {
		return err
	}

	resp, err := dto.NewDevilFruitResponse(r.Context(), fruit, e.svc.PictureURL)
	if err != nil {
		return err
	}

	w.Header().Set("Location", fmt.Sprintf("/api/v1/devil-fruits/%s", fruit.ID()))
	writeJSON(w, http.StatusCreated, resp)
	return nil
}

// get godoc
//
//	@Summary		Get a devil fruit by id
//	@Tags			devil-fruits
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Devil fruit id (UUID)"
//	@Description	The response carries an ETag; a request with a matching If-None-Match
//	@Description	gets 304 Not Modified with no body instead of the full DevilFruit.
//	@Success		200	{object}	dto.DevilFruitResponse
//	@Success		304
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		429	{object}	dto.ErrorResponse
//	@Router			/devil-fruits/{id} [get]
func (e *DevilFruitEndpoints) get(w http.ResponseWriter, r *http.Request) error {
	id, err := parsePowerID(r)
	if err != nil {
		return err
	}

	fruit, err := e.svc.GetDevilFruit(r.Context(), id, LocaleFromRequest(r))
	if err != nil {
		return err
	}
	resp, err := dto.NewDevilFruitResponse(r.Context(), fruit, e.svc.PictureURL)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, resp)
	return nil
}

// update godoc
//
//	@Summary		Replace a devil fruit
//	@Description	Admin only. Keeps the original id.
//	@Tags			devil-fruits
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Devil fruit id (UUID)"
//	@Param			request	body		dto.DevilFruitRequest	true	"Replacement devil fruit"
//	@Success		200		{object}	dto.DevilFruitResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		409		{object}	dto.ErrorResponse
//	@Failure		429		{object}	dto.ErrorResponse
//	@Router			/devil-fruits/{id} [put]
func (e *DevilFruitEndpoints) update(w http.ResponseWriter, r *http.Request) error {
	id, err := parsePowerID(r)
	if err != nil {
		return err
	}

	var req dto.DevilFruitRequest
	if err := decode(w, r, &req); err != nil {
		return err
	}

	input, err := req.Validate()
	if err != nil {
		return err
	}

	fruit, err := e.svc.UpdateDevilFruit(r.Context(), id, input)
	if err != nil {
		return err
	}
	resp, err := dto.NewDevilFruitResponse(r.Context(), fruit, e.svc.PictureURL)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, resp)
	return nil
}

// patchPicture godoc
//
//	@Summary		Set a devil fruit's picture
//	@Description	Admin only. Uploads the image to object storage and stores its key; the response's `picture` is a presigned URL.
//	@Tags			devil-fruits
//	@Accept			mpfd
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string	true	"Devil fruit id (UUID)"
//	@Param			picture	formData	file	true	"Image file (WebP, AVIF, JPEG, PNG or GIF)"
//	@Description	The image is re-encoded to WebP by a background worker: the response is
//	@Description	202 Accepted with `pictureStatus` set to PENDING and the previous
//	@Description	`picture`/`pictureThumb` URLs (or "" on a first upload); poll GET
//	@Description	/devil-fruits/{id} until `pictureStatus` becomes READY or FAILED.
//	@Success		202		{object}	dto.DevilFruitResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		413		{object}	dto.ErrorResponse
//	@Failure		429		{object}	dto.ErrorResponse
//	@Failure		503		{object}	dto.ErrorResponse
//	@Router			/devil-fruits/{id}/picture [patch]
func (e *DevilFruitEndpoints) patchPicture(w http.ResponseWriter, r *http.Request) error {
	id, err := parsePowerID(r)
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

	fruit, err := e.svc.SetDevilFruitPicture(r.Context(), id, ports.Picture{
		Content:     bytes.NewReader(buf),
		ContentType: contentType,
		Size:        int64(len(buf)),
	})
	if err != nil {
		return err
	}

	resp, err := dto.NewDevilFruitResponse(r.Context(), fruit, e.svc.PictureURL)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusAccepted, resp)
	return nil
}

// translations godoc
//
//	@Summary		Get every locale's translation for a devil fruit
//	@Description	Admin only. Unlike GET /devil-fruits/{id}, always returns
//	@Description	every locale's description/skills at once, for an edit form.
//	@Tags			devil-fruits
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Devil fruit id (UUID)"
//	@Success		200	{object}	dto.PowerTranslationsResponse
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		429	{object}	dto.ErrorResponse
//	@Router			/devil-fruits/{id}/translations [get]
func (e *DevilFruitEndpoints) translations(w http.ResponseWriter, r *http.Request) error {
	id, err := parsePowerID(r)
	if err != nil {
		return err
	}
	translations, err := e.svc.DevilFruitTranslations(r.Context(), id)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, dto.NewPowerTranslationsResponse(translations))
	return nil
}

// delete godoc
//
//	@Summary		Delete a devil fruit
//	@Description	Admin only.
//	@Tags			devil-fruits
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Devil fruit id (UUID)"
//	@Success		204
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		429	{object}	dto.ErrorResponse
//	@Router			/devil-fruits/{id} [delete]
func (e *DevilFruitEndpoints) delete(w http.ResponseWriter, r *http.Request) error {
	id, err := parsePowerID(r)
	if err != nil {
		return err
	}

	if err := e.svc.DeleteDevilFruit(r.Context(), id); err != nil {
		return err
	}
	writeJSON(w, http.StatusNoContent, nil)
	return nil
}
