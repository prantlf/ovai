package routes

import (
	"net/http"

	"github.com/prantlf/ovai/internal/log"
)

func HandlePing(w http.ResponseWriter, r *http.Request) int {
	log.Dbg(": ping")
	w.WriteHeader(http.StatusNoContent)
	return http.StatusNoContent
}
