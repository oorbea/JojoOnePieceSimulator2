package endpoints

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/dto"
)

// maxRequestBodyBytes bounds decoded request bodies to guard against
// oversized payloads.
const maxRequestBodyBytes = 1 << 20 // 1 MiB

// maxMultipartMemory bounds how much of a PATCH .../picture upload is
// buffered in memory before net/http spills the rest to a temp file.
const maxMultipartMemory = 1 << 20 // 1 MiB

// sniffLen is how many leading bytes of an uploaded picture are read to
// detect its real content type via http.DetectContentType.
const sniffLen = 512

// StandEndpoints wires the Stand HTTP surface to the application service.
type StandEndpoints struct {
	svc *services.StandService
}

func NewStandEndpoints(svc *services.StandService) *StandEndpoints {
	return &StandEndpoints{svc: svc}
}

// Routes returns the /stands sub-router: GET/POST on the collection,
// GET/PUT/PATCH/DELETE on a single stand by id. Reads and writes are
// rate-limited as separate tiers (see rateCfg), each keyed by the
// authenticated user id; both GETs share a single readRateLimit instance so
// they draw from one budget instead of doubling it.
func (e *StandEndpoints) Routes(rateCfg RateLimitConfig) chi.Router {
	r := chi.NewRouter()
	read := readRateLimit(rateCfg)
	r.With(read).Get("/", Wrap(e.list))
	r.With(read).Get("/{id}", Wrap(e.get))

	r.Group(func(r chi.Router) {
		r.Use(RequireAdmin)
		r.Use(writeRateLimit(rateCfg))
		r.Post("/", Wrap(e.create))
		r.Put("/{id}", Wrap(e.update))
		r.Patch("/{id}/picture", Wrap(e.patchPicture))
		r.Delete("/{id}", Wrap(e.delete))
	})
	return r
}

// list godoc
//
//	@Summary		List or filter stands
//	@Description	Lists every Stand, or filters them if any query param is set.
//	@Tags			stands
//	@Produce		json
//	@Security		BearerAuth
//	@Param			rarity		query		string	false	"COMMON, RARE, EPIC, LEGENDARY"
//	@Param			attackPower	query		string	false	"E, D, C, B, A, INFINITE, NULL"
//	@Param			speed		query		string	false	"E, D, C, B, A, INFINITE, NULL"
//	@Param			attackRange	query		string	false	"E, D, C, B, A, INFINITE, NULL"
//	@Param			endurance	query		string	false	"E, D, C, B, A, INFINITE, NULL"
//	@Param			precision	query		string	false	"E, D, C, B, A, INFINITE, NULL"
//	@Param			potential	query		string	false	"E, D, C, B, A, INFINITE, NULL"
//	@Param			evolvesFrom	query		string	false	"name of the Stand this one evolves from"
//	@Success		200			{array}		dto.StandResponse
//	@Failure		400			{object}	dto.ErrorResponse
//	@Failure		401			{object}	dto.ErrorResponse
//	@Failure		429			{object}	dto.ErrorResponse
//	@Router			/stands [get]
func (e *StandEndpoints) list(w http.ResponseWriter, r *http.Request) error {
	filters, hasFilters, err := dto.StandFiltersFromQuery(r.URL.Query())
	if err != nil {
		return err
	}

	var stands []*powers.Stand
	if hasFilters {
		stands, err = e.svc.FilterStands(r.Context(), filters)
	} else {
		stands, err = e.svc.ListStands(r.Context())
	}
	if err != nil {
		return err
	}
	resp, err := dto.NewStandResponses(r.Context(), stands, e.svc.PictureURL)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, resp)
	return nil
}

// create godoc
//
//	@Summary		Create a stand
//	@Description	Admin only.
//	@Tags			stands
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.StandRequest	true	"Stand to create"
//	@Success		201		{object}	dto.StandResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		409		{object}	dto.ErrorResponse
//	@Failure		429		{object}	dto.ErrorResponse
//	@Router			/stands [post]
func (e *StandEndpoints) create(w http.ResponseWriter, r *http.Request) error {
	var req dto.StandRequest
	if err := decode(w, r, &req); err != nil {
		return err
	}

	input, err := req.Validate()
	if err != nil {
		return err
	}

	stand, err := e.svc.CreateStand(r.Context(), input)
	if err != nil {
		return err
	}

	resp, err := dto.NewStandResponse(r.Context(), stand, e.svc.PictureURL)
	if err != nil {
		return err
	}

	w.Header().Set("Location", fmt.Sprintf("/api/v1/stands/%s", stand.ID()))
	writeJSON(w, http.StatusCreated, resp)
	return nil
}

// get godoc
//
//	@Summary		Get a stand by id
//	@Tags			stands
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Stand id (UUID)"
//	@Success		200	{object}	dto.StandResponse
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		429	{object}	dto.ErrorResponse
//	@Router			/stands/{id} [get]
func (e *StandEndpoints) get(w http.ResponseWriter, r *http.Request) error {
	id, err := parsePowerID(r)
	if err != nil {
		return err
	}

	stand, err := e.svc.GetStand(r.Context(), id)
	if err != nil {
		return err
	}
	resp, err := dto.NewStandResponse(r.Context(), stand, e.svc.PictureURL)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, resp)
	return nil
}

// update godoc
//
//	@Summary		Replace a stand
//	@Description	Admin only. Keeps the original id.
//	@Tags			stands
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string				true	"Stand id (UUID)"
//	@Param			request	body		dto.StandRequest	true	"Replacement Stand"
//	@Success		200		{object}	dto.StandResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		409		{object}	dto.ErrorResponse
//	@Failure		429		{object}	dto.ErrorResponse
//	@Router			/stands/{id} [put]
func (e *StandEndpoints) update(w http.ResponseWriter, r *http.Request) error {
	id, err := parsePowerID(r)
	if err != nil {
		return err
	}

	var req dto.StandRequest
	if err := decode(w, r, &req); err != nil {
		return err
	}

	input, err := req.Validate()
	if err != nil {
		return err
	}

	stand, err := e.svc.UpdateStand(r.Context(), id, input)
	if err != nil {
		return err
	}
	resp, err := dto.NewStandResponse(r.Context(), stand, e.svc.PictureURL)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, resp)
	return nil
}

// patchPicture godoc
//
//	@Summary		Set a stand's picture
//	@Description	Admin only. Uploads the image to object storage and stores its key; the response's `picture` is a presigned URL.
//	@Tags			stands
//	@Accept			mpfd
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string	true	"Stand id (UUID)"
//	@Param			picture	formData	file	true	"Image file (JPEG, PNG or WebP)"
//	@Success		200		{object}	dto.StandResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		413		{object}	dto.ErrorResponse
//	@Failure		429		{object}	dto.ErrorResponse
//	@Router			/stands/{id}/picture [patch]
func (e *StandEndpoints) patchPicture(w http.ResponseWriter, r *http.Request) error {
	id, err := parsePowerID(r)
	if err != nil {
		return err
	}

	// The maxRequestBodyBytes cap above is JSON-only; multipart bodies need
	// their own guard, sized to the configured picture limit plus room for
	// the multipart form overhead itself.
	r.Body = http.MaxBytesReader(w, r.Body, e.svc.MaxPictureBytes()+maxMultipartMemory)
	if err := r.ParseMultipartForm(maxMultipartMemory); err != nil {
		return &dto.ValidationError{Errors: []string{err.Error()}}
	}
	defer func() {
		_ = r.MultipartForm.RemoveAll()
	}()

	file, header, err := r.FormFile("picture")
	if err != nil {
		return services.ErrPictureRequired
	}
	defer func() {
		_ = file.Close()
	}()

	// Sniff the real content type; a client-supplied Content-Type header is
	// not trustworthy.
	head := make([]byte, sniffLen)
	n, err := io.ReadFull(file, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return err
	}
	contentType := http.DetectContentType(head[:n])

	stand, err := e.svc.SetStandPicture(r.Context(), id, ports.Picture{
		Content:     io.MultiReader(bytes.NewReader(head[:n]), file),
		ContentType: contentType,
		Size:        header.Size,
	})
	if err != nil {
		return err
	}

	resp, err := dto.NewStandResponse(r.Context(), stand, e.svc.PictureURL)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, resp)
	return nil
}

// delete godoc
//
//	@Summary		Delete a stand
//	@Description	Admin only.
//	@Tags			stands
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Stand id (UUID)"
//	@Success		204
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		429	{object}	dto.ErrorResponse
//	@Router			/stands/{id} [delete]
func (e *StandEndpoints) delete(w http.ResponseWriter, r *http.Request) error {
	id, err := parsePowerID(r)
	if err != nil {
		return err
	}

	if err := e.svc.DeleteStand(r.Context(), id); err != nil {
		return err
	}
	writeJSON(w, http.StatusNoContent, nil)
	return nil
}

func parsePowerID(r *http.Request) (powers.PowerID, error) {
	return powers.ParsePowerID(chi.URLParam(r, "id"))
}

// decode reads a JSON body into dst, rejecting unknown fields and oversized
// payloads.
func decode(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return &dto.ValidationError{Errors: []string{err.Error()}}
	}
	return nil
}
