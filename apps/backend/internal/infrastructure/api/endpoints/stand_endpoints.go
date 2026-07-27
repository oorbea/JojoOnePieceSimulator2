package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/dto"
)

// maxRequestBodyBytes bounds decoded request bodies to guard against
// oversized payloads.
const maxRequestBodyBytes = 1 << 20 // 1 MiB

// StandEndpoints wires the Stand HTTP surface to the application service.
type StandEndpoints struct {
	svc *services.StandService
}

func NewStandEndpoints(svc *services.StandService) *StandEndpoints {
	return &StandEndpoints{svc: svc}
}

// Routes returns the /stands sub-router: GET/POST on the collection,
// GET/PUT/DELETE on a single stand by id.
func (e *StandEndpoints) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", e.list)
	r.Post("/", e.create)
	r.Get("/{id}", e.get)
	r.Put("/{id}", e.update)
	r.Delete("/{id}", e.delete)
	return r
}

func (e *StandEndpoints) list(w http.ResponseWriter, r *http.Request) {
	filters, hasFilters, err := dto.StandFiltersFromQuery(r.URL.Query())
	if err != nil {
		handleError(w, err)
		return
	}

	var stands []*powers.Stand
	if hasFilters {
		stands, err = e.svc.FilterStands(r.Context(), filters)
	} else {
		stands, err = e.svc.ListStands(r.Context())
	}
	if err != nil {
		handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewStandResponses(stands))
}

func (e *StandEndpoints) create(w http.ResponseWriter, r *http.Request) {
	var req dto.StandRequest
	if !decode(w, r, &req) {
		return
	}

	input, err := req.Validate()
	if err != nil {
		handleError(w, err)
		return
	}

	stand, err := e.svc.CreateStand(r.Context(), input)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/api/v1/stands/%s", stand.ID()))
	writeJSON(w, http.StatusCreated, dto.NewStandResponse(stand))
}

func (e *StandEndpoints) get(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePowerID(w, r)
	if !ok {
		return
	}

	stand, err := e.svc.GetStand(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewStandResponse(stand))
}

func (e *StandEndpoints) update(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePowerID(w, r)
	if !ok {
		return
	}

	var req dto.StandRequest
	if !decode(w, r, &req) {
		return
	}

	input, err := req.Validate()
	if err != nil {
		handleError(w, err)
		return
	}

	stand, err := e.svc.UpdateStand(r.Context(), id, input)
	if err != nil {
		handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.NewStandResponse(stand))
}

func (e *StandEndpoints) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePowerID(w, r)
	if !ok {
		return
	}

	if err := e.svc.DeleteStand(r.Context(), id); err != nil {
		handleError(w, err)
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}

func parsePowerID(w http.ResponseWriter, r *http.Request) (powers.PowerID, bool) {
	id, err := powers.ParsePowerID(chi.URLParam(r, "id"))
	if err != nil {
		handleError(w, err)
		return powers.NilPowerID, false
	}
	return id, true
}

// decode reads a JSON body into dst, rejecting unknown fields and oversized
// payloads. Returns false (having already written the error response) if
// decoding failed.
func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return false
	}
	return true
}
