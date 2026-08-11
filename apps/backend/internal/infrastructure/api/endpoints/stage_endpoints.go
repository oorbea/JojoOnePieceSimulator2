package endpoints

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/dto"
)

// StageEndpoints wires the Stage HTTP surface (the round-content catalogue
// game.IGameMode plays through) to StageService. Reads are open to any
// authenticated user, same as Stands/DevilFruits; writes are admin-only.
// Reuses stand_endpoints.go's maxMultipartMemory/sniffLen/sniffContentType
// constants/helpers - the picture pipeline is identical.
type StageEndpoints struct {
	svc *services.StageService
}

func NewStageEndpoints(svc *services.StageService) *StageEndpoints {
	return &StageEndpoints{svc: svc}
}

// Routes returns the /stages sub-router.
func (e *StageEndpoints) Routes(rateCfg RateLimitConfig, cacheCfg CacheConfig) chi.Router {
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
//	@Summary		List stages
//	@Description	Lists every Stage, or only manga's if the query param is set. Description resolved for the request's locale.
//	@Tags			stages
//	@Produce		json
//	@Security		BearerAuth
//	@Param			manga	query		string	false	"JOJO, ONE_PIECE"
//	@Success		200		{array}		dto.StageResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		429		{object}	dto.ErrorResponse
//	@Router			/stages [get]
func (e *StageEndpoints) list(w http.ResponseWriter, r *http.Request) error {
	locale := LocaleFromRequest(r)

	if raw := r.URL.Query().Get("manga"); raw != "" {
		manga, err := enums.ParseManga(raw)
		if err != nil {
			return &dto.ValidationError{Errors: []string{fmt.Sprintf("manga: %v", err)}}
		}
		stages, err := e.svc.StagesByManga(r.Context(), manga, locale)
		if err != nil {
			return err
		}
		resp, err := dto.NewStageResponses(r.Context(), stages, e.svc.PictureURL)
		if err != nil {
			return err
		}
		writeJSON(w, http.StatusOK, resp)
		return nil
	}

	stages, err := e.svc.ListStages(r.Context(), locale)
	if err != nil {
		return err
	}
	resp, err := dto.NewStageResponses(r.Context(), stages, e.svc.PictureURL)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, resp)
	return nil
}

// create godoc
//
//	@Summary		Create a stage
//	@Description	Admin only.
//	@Tags			stages
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.StageRequest	true	"Stage to create"
//	@Success		201		{object}	dto.StageResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		409		{object}	dto.ErrorResponse
//	@Failure		429		{object}	dto.ErrorResponse
//	@Router			/stages [post]
func (e *StageEndpoints) create(w http.ResponseWriter, r *http.Request) error {
	var req dto.StageRequest
	if err := decode(w, r, &req); err != nil {
		return err
	}
	input, err := req.Validate()
	if err != nil {
		return err
	}
	st, err := e.svc.CreateStage(r.Context(), input)
	if err != nil {
		return err
	}
	resp, err := dto.NewStageResponse(r.Context(), st, e.svc.PictureURL)
	if err != nil {
		return err
	}
	w.Header().Set("Location", fmt.Sprintf("/api/v1/stages/%s", st.ID()))
	writeJSON(w, http.StatusCreated, resp)
	return nil
}

// get godoc
//
//	@Summary		Get a stage by id
//	@Description	Description resolved for the request's locale.
//	@Tags			stages
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Stage id (UUID)"
//	@Success		200	{object}	dto.StageResponse
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		429	{object}	dto.ErrorResponse
//	@Router			/stages/{id} [get]
func (e *StageEndpoints) get(w http.ResponseWriter, r *http.Request) error {
	id, err := parseStageID(r)
	if err != nil {
		return err
	}
	st, err := e.svc.GetStage(r.Context(), id, LocaleFromRequest(r))
	if err != nil {
		return err
	}
	resp, err := dto.NewStageResponse(r.Context(), st, e.svc.PictureURL)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, resp)
	return nil
}

// update godoc
//
//	@Summary		Replace a stage
//	@Description	Admin only. Keeps the original id and picture.
//	@Tags			stages
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string				true	"Stage id (UUID)"
//	@Param			request	body		dto.StageRequest	true	"Replacement stage"
//	@Success		200		{object}	dto.StageResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		409		{object}	dto.ErrorResponse
//	@Failure		429		{object}	dto.ErrorResponse
//	@Router			/stages/{id} [put]
func (e *StageEndpoints) update(w http.ResponseWriter, r *http.Request) error {
	id, err := parseStageID(r)
	if err != nil {
		return err
	}
	var req dto.StageRequest
	if err := decode(w, r, &req); err != nil {
		return err
	}
	input, err := req.Validate()
	if err != nil {
		return err
	}
	st, err := e.svc.UpdateStage(r.Context(), id, input)
	if err != nil {
		return err
	}
	resp, err := dto.NewStageResponse(r.Context(), st, e.svc.PictureURL)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, resp)
	return nil
}

// patchPicture godoc
//
//	@Summary		Set a stage's picture
//	@Description	Admin only. Uploads the image to object storage and stores its key; the response's `picture` is a presigned URL.
//	@Tags			stages
//	@Accept			mpfd
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string	true	"Stage id (UUID)"
//	@Param			picture	formData	file	true	"Image file (WebP, AVIF, JPEG, PNG or GIF)"
//	@Description	The image is re-encoded to WebP by a background worker: the response is
//	@Description	202 Accepted with `pictureStatus` set to PENDING and the previous
//	@Description	`picture`/`pictureThumb` URLs (or "" on a first upload); poll GET
//	@Description	/stages/{id} until `pictureStatus` becomes READY or FAILED.
//	@Success		202		{object}	dto.StageResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		413		{object}	dto.ErrorResponse
//	@Failure		429		{object}	dto.ErrorResponse
//	@Failure		503		{object}	dto.ErrorResponse
//	@Router			/stages/{id}/picture [patch]
func (e *StageEndpoints) patchPicture(w http.ResponseWriter, r *http.Request) error {
	id, err := parseStageID(r)
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

	st, err := e.svc.SetStagePicture(r.Context(), id, ports.Picture{
		Content:     bytes.NewReader(buf),
		ContentType: contentType,
		Size:        int64(len(buf)),
	})
	if err != nil {
		return err
	}

	resp, err := dto.NewStageResponse(r.Context(), st, e.svc.PictureURL)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusAccepted, resp)
	return nil
}

// translations godoc
//
//	@Summary		Get every locale's translation for a stage
//	@Description	Admin only. Unlike GET /stages/{id}, always returns every
//	@Description	locale's description at once, for an edit form.
//	@Tags			stages
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Stage id (UUID)"
//	@Success		200	{object}	dto.StageTranslationsResponse
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		429	{object}	dto.ErrorResponse
//	@Router			/stages/{id}/translations [get]
func (e *StageEndpoints) translations(w http.ResponseWriter, r *http.Request) error {
	id, err := parseStageID(r)
	if err != nil {
		return err
	}
	translations, err := e.svc.StageTranslations(r.Context(), id)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, dto.NewStageTranslationsResponse(translations))
	return nil
}

// delete godoc
//
//	@Summary		Delete a stage
//	@Description	Admin only.
//	@Tags			stages
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Stage id (UUID)"
//	@Success		204
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		429	{object}	dto.ErrorResponse
//	@Router			/stages/{id} [delete]
func (e *StageEndpoints) delete(w http.ResponseWriter, r *http.Request) error {
	id, err := parseStageID(r)
	if err != nil {
		return err
	}
	if err := e.svc.DeleteStage(r.Context(), id); err != nil {
		return err
	}
	writeJSON(w, http.StatusNoContent, nil)
	return nil
}

func parseStageID(r *http.Request) (game.StageID, error) {
	return game.ParseStageID(chi.URLParam(r, "id"))
}
