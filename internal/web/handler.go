package web

import (
	"net/http"
	"slices"
	"strings"

	"github.com/prantlf/ovai/internal/log"
)

func IsOK(status int) bool {
	return status >= 200 && status < 400
}

func DisallowMethod(w http.ResponseWriter, methods []string) int {
	w.Header().Set("Allow", strings.Join(methods, ","))
	w.WriteHeader(http.StatusMethodNotAllowed)
	return http.StatusMethodNotAllowed
}

func LogRequest(w http.ResponseWriter, r *http.Request) {
	log.Srv("request %s %s", r.Method, r.RequestURI)
}

func LogResponse(w http.ResponseWriter, r *http.Request, status int) {
	if IsOK(status) {
		log.Srv("respond %d: %s %s", status, r.Method, r.RequestURI)
	} else {
		log.Dbg("fail %d: %s %s", status, r.Method, r.RequestURI)
	}
}

type LoggedHandlerFunc func(http.ResponseWriter, *http.Request) int

func EnableCORS(w http.ResponseWriter, r *http.Request) {
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST")
	}
}

func WrapHandler(fn LoggedHandlerFunc, methods []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		LogRequest(w, r)
		EnableCORS(w, r)
		var status int
		if slices.Contains(methods, r.Method) {
			status = fn(w, r)
		} else {
			status = DisallowMethod(w, methods)
		}
		LogResponse(w, r, status)
	}
}
