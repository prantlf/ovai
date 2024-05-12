package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/prantlf/ovai/internal/auth"
	"github.com/prantlf/ovai/internal/cfg"
	"github.com/prantlf/ovai/internal/log"
	"github.com/prantlf/ovai/internal/web"
)

func forwardRequest(urlSuffix string, input interface{}, output interface{}) (int, time.Duration, error) {
	url := fmt.Sprintf("https://%s/v1/projects/%s/locations/%s/publishers/google/models/%s",
		cfg.Defaults.ApiEndpoint, auth.Account.ProjectId, cfg.Defaults.ApiLocation, urlSuffix)
	accessToken, err := auth.UseAccessToken()
	if err != nil {
		return http.StatusInternalServerError, 0, err
	}

	var duration time.Duration
	forwardRequest := func() (int, error) {
		start := time.Now()
		req, err := web.CreatePostRequest(url, input)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		status, err := web.DispatchRequest(req, output)
		duration = time.Since(start)
		return status, err
	}

	status, err := forwardRequest()
	if err != nil && status == 401 {
		auth.RefreshAccessToken()
		status, err = forwardRequest()
	}
	return status, duration, err
}

type failResponse struct {
	Error string `json:"error"`
}

func failRequest(w http.ResponseWriter, status int, msg string) int {
	resObj := &failResponse{
		Error: msg,
	}
	resBody, err := json.Marshal(resObj)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		resBody = []byte(msg)
	} else {
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(status)
	w.Write(resBody)
	return status
}

func wrongInput(w http.ResponseWriter, msg string) int {
	log.Dbg("! %s", msg)
	return failRequest(w, http.StatusBadRequest, msg)
}
