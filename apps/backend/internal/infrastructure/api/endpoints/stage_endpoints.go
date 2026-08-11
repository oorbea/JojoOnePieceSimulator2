package endpoints

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/dto"
)

// StageEndpoints wires the Stage HTTP surface (the round-content catalogue
// game.IGameMode plays through) to StageService. Reads are open to any
// authenticated user, same as Stands/DevilFruits; writes are admin-only.
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
		r.Delete("/{id}", Wrap(e.delete))
	})
	return r
}

// list godoc
//
//	@Summary		List stages
//	@Description	Lists every Stage, or only manga's if the query param is set.
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
	if raw := r.URL.Query().Get("manga"); raw != "" {
		manga, err := enums.ParseManga(raw)
		if err != nil {
			return &dto.ValidationError{Errors: []string{fmt.Sprintf("manga: %v", err)}}
		}
		stages, err := e.svc.StagesByManga(r.Context(), manga)
		if err != nil {
			return err
		}
		writeJSON(w, http.StatusOK, dto.NewStageResponses(stages))
		return nil
	}

	stages, err := e.svc.ListStages(r.Context())
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, dto.NewStageResponses(stages))
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
	w.Header().Set("Location", fmt.Sprintf("/api/v1/stages/%s", st.ID()))
	writeJSON(w, http.StatusCreated, dto.NewStageResponse(st))
	return nil
}

// get godoc
//
//	@Summary		Get a stage by id
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
	st, err := e.svc.GetStage(r.Context(), id)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, dto.NewStageResponse(st))
	return nil
}

// update godoc
//
//	@Summary		Replace a stage
//	@Description	Admin only. Keeps the original id.
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
	writeJSON(w, http.StatusOK, dto.NewStageResponse(st))
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
