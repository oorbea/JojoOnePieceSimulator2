package endpoints

import (
	"encoding/json"
	"net/http"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/dto"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code string, msg string, details ...string) {
	writeJSON(w, status, dto.ErrorResponse{Error: msg, Code: code, Details: details})
}
