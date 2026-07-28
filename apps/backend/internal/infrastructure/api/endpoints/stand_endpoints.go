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
	r.Get("/", Wrap(e.list))
	r.Post("/", Wrap(e.create))
	r.Get("/{id}", Wrap(e.get))
	r.Put("/{id}", Wrap(e.update))
	r.Delete("/{id}", Wrap(e.delete))
	return r
}

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
	writeJSON(w, http.StatusOK, dto.NewStandResponses(stands))
	return nil
}

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

	w.Header().Set("Location", fmt.Sprintf("/api/v1/stands/%s", stand.ID()))
	writeJSON(w, http.StatusCreated, dto.NewStandResponse(stand))
	return nil
}

func (e *StandEndpoints) get(w http.ResponseWriter, r *http.Request) error {
	id, err := parsePowerID(r)
	if err != nil {
		return err
	}

	stand, err := e.svc.GetStand(r.Context(), id)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, dto.NewStandResponse(stand))
	return nil
}

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
	writeJSON(w, http.StatusOK, dto.NewStandResponse(stand))
	return nil
}

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
