package routes

import (
	"net/http"

	"github.com/prantlf/ovai/internal/log"
)

func HandleRoot(w http.ResponseWriter, r *http.Request) int {
	if r.Method == "HEAD" {
		w.WriteHeader(http.StatusOK)
		return http.StatusOK
	}
	log.Dbg(": root")
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ollama is running"))
	return http.StatusOK
}
