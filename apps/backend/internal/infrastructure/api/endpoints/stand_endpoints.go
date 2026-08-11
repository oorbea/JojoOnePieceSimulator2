package endpoints

import (
	"bytes"
	"encoding/json"
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
// detect its real content type.
const sniffLen = 512

// avifFtypBrands are the ISOBMFF brand codes that mark a file as AVIF -
// checked manually because http.DetectContentType has no AVIF signature.
var avifFtypBrands = [][]byte{[]byte("avif"), []byte("avis")}

// sniffContentType detects head's real content type, recognizing AVIF on
// top of everything http.DetectContentType already covers.
func sniffContentType(head []byte) string {
	if isAVIF(head) {
		return "image/avif"
	}
	return http.DetectContentType(head)
}

// isAVIF reports whether head starts with an ISOBMFF "ftyp" box whose major
// or a compatible brand is "avif" (still image) or "avis" (image sequence).
func isAVIF(head []byte) bool {
	if len(head) < 16 || !bytes.Equal(head[4:8], []byte("ftyp")) {
		return false
	}
	for _, brand := range avifFtypBrands {
		if bytes.Equal(head[8:12], brand) {
			return true
		}
	}
	for i := 16; i+4 <= len(head); i += 4 {
		for _, brand := range avifFtypBrands {
			if bytes.Equal(head[i:i+4], brand) {
				return true
			}
		}
	}
	return false
}

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
// they draw from one budget instead of doubling it. Both GETs also get the
// ETag/Cache-Control layer (see cacheCfg, cache_headers.go); writes never do,
// since their responses aren't meant to be cached by the client.
func (e *StandEndpoints) Routes(rateCfg RateLimitConfig, cacheCfg CacheConfig) chi.Router {
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
//	@Param			q			query		string	false	"free-text search over name and description"
//	@Description	Responses carry an ETag; a request with a matching If-None-Match
//	@Description	gets 304 Not Modified with no body instead of the full list.
//	@Success		200			{array}		dto.StandResponse
//	@Success		304
//	@Failure		400			{object}	dto.ErrorResponse
//	@Failure		401			{object}	dto.ErrorResponse
//	@Failure		429			{object}	dto.ErrorResponse
//	@Router			/stands [get]
func (e *StandEndpoints) list(w http.ResponseWriter, r *http.Request) error {
	filters, hasFilters, err := dto.StandFiltersFromQuery(r.URL.Query())
	if err != nil {
		return err
	}

	locale := LocaleFromRequest(r)
	var stands []*powers.Stand
	if hasFilters {
		stands, err = e.svc.FilterStands(r.Context(), filters, locale)
	} else {
		stands, err = e.svc.ListStands(r.Context(), locale)
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
//	@Description	The response carries an ETag; a request with a matching If-None-Match
//	@Description	gets 304 Not Modified with no body instead of the full Stand.
//	@Success		200	{object}	dto.StandResponse
//	@Success		304
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

	stand, err := e.svc.GetStand(r.Context(), id, LocaleFromRequest(r))
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
//	@Param			picture	formData	file	true	"Image file (WebP, AVIF, JPEG, PNG or GIF)"
//	@Description	The image is re-encoded to WebP by a background worker: the response is
//	@Description	202 Accepted with `pictureStatus` set to PENDING and the previous
//	@Description	`picture`/`pictureThumb` URLs (or "" on a first upload); poll GET
//	@Description	/stands/{id} until `pictureStatus` becomes READY or FAILED.
//	@Success		202		{object}	dto.StandResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		413		{object}	dto.ErrorResponse
//	@Failure		429		{object}	dto.ErrorResponse
//	@Failure		503		{object}	dto.ErrorResponse
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

	file, _, err := r.FormFile("picture")
	if err != nil {
		return services.ErrPictureRequired
	}
	defer func() {
		_ = file.Close()
	}()

	// The pipeline needs the whole image in memory: it is both probed and
	// handed to the background worker as a single buffer. Read one byte past
	// the configured limit so an oversized file is detected reliably -
	// header.Size is client-supplied and not trustworthy.
	maxBytes := e.svc.MaxPictureBytes()
	buf, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if err != nil {
		return err
	}
	if maxBytes > 0 && int64(len(buf)) > maxBytes {
		return services.ErrPictureTooLarge
	}

	// Sniff the real content type; a client-supplied Content-Type header is
	// not trustworthy.
	head := buf
	if len(head) > sniffLen {
		head = head[:sniffLen]
	}
	contentType := sniffContentType(head)

	stand, err := e.svc.SetStandPicture(r.Context(), id, ports.Picture{
		Content:     bytes.NewReader(buf),
		ContentType: contentType,
		Size:        int64(len(buf)),
	})
	if err != nil {
		return err
	}

	resp, err := dto.NewStandResponse(r.Context(), stand, e.svc.PictureURL)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusAccepted, resp)
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

// translations godoc
//
//	@Summary		Get every locale's translation for a stand
//	@Description	Admin only. Unlike GET /stands/{id}, always returns every
//	@Description	locale's description/skills at once, for an edit form.
//	@Tags			stands
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Stand id (UUID)"
//	@Success		200	{object}	dto.PowerTranslationsResponse
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Failure		429	{object}	dto.ErrorResponse
//	@Router			/stands/{id}/translations [get]
func (e *StandEndpoints) translations(w http.ResponseWriter, r *http.Request) error {
	id, err := parsePowerID(r)
	if err != nil {
		return err
	}
	translations, err := e.svc.StandTranslations(r.Context(), id)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, dto.NewPowerTranslationsResponse(translations))
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
