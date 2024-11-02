package routes

import (
	"encoding/json"
	"fmt"
	"io"
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

	forwardRequest := func() (int, time.Duration, error) {
		start := time.Now()
		req, err := web.CreatePostRequest(url, input)
		if err != nil {
			return http.StatusInternalServerError, 0, err
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		status, err := web.DispatchRequest(req, output)
		duration := time.Since(start)
		return status, duration, err
	}

	status, duration, err := forwardRequest()
	if err != nil && status == 401 {
		auth.RefreshAccessToken()
		status, duration, err = forwardRequest()
	}
	return status, duration, err
}

func forwardStream(urlSuffix string, input interface{}) (int, time.Time, io.ReadCloser, error) {
	url := fmt.Sprintf("https://%s/v1/projects/%s/locations/%s/publishers/google/models/%s",
		cfg.Defaults.ApiEndpoint, auth.Account.ProjectId, cfg.Defaults.ApiLocation, urlSuffix)
	accessToken, err := auth.UseAccessToken()
	if err != nil {
		return http.StatusInternalServerError, time.Time{}, nil, err
	}

	forwardStream := func() (int, time.Time, io.ReadCloser, error) {
		start := time.Now()
		req, err := web.CreatePostRequest(url, input)
		if err != nil {
			return http.StatusInternalServerError, time.Time{}, nil, err
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		var status int
		status, resReader, err := web.BeginRawRequest(req)
		if err != nil {
			return status, time.Time{}, nil, err
		}
		return status, start, resReader, err
	}

	status, start, resReader, err := forwardStream()
	if err != nil && status == 401 {
		auth.RefreshAccessToken()
		status, start, resReader, err = forwardStream()
	}
	return status, start, resReader, err
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
	if _, err := w.Write(resBody); err != nil {
		log.Dbg("! writing response body failed: %v", err)
	}
	return status
}

func wrongInput(w http.ResponseWriter, msg string) int {
	log.Dbg("! %s", msg)
	return failRequest(w, http.StatusBadRequest, msg)
}
