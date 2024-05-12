package routes

import (
	"net/http"

	"github.com/prantlf/ovai/internal/log"
	"github.com/prantlf/ovai/internal/web"
)

func HandleShutdown(w http.ResponseWriter, r *http.Request) bool {
	web.LogRequest(w, r)
	web.EnableCORS(w, r)
	if r.Method == "POST" {
		log.Dbg(": shut down")
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	status := web.DisallowMethod(w, []string{"POST"})
	web.LogResponse(w, r, status)
	return false
}
