package routes

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/prantlf/ovai/internal/log"
	"github.com/prantlf/ovai/internal/web"
)

var canProxy bool
var ollamaOrigin string

var _ = initProxy()

func initProxy() bool {
	ollamaOrigin = os.Getenv("OLLAMA_ORIGIN")
	canProxy = len(ollamaOrigin) > 0
	return true
}

func proxyRequest(name string, input []byte, w http.ResponseWriter, result string, model string) int {
	ollamaUrl := fmt.Sprintf("%s/api/%s", ollamaOrigin, name)
	req, err := web.CreateRawPostRequest(ollamaUrl, input)
	if err != nil {
		return failRequest(w, http.StatusInternalServerError, err.Error())
	}
	status, output, err := web.DispatchRawRequest(req)
	if err != nil {
		return failRequest(w, status, err.Error())
	}
	if log.IsDbg {
		if len(model) > 0 {
			log.Dbg("< %s by %s with %d byte%s", result, model,
				len(output), log.GetPlural(len(output)))
		} else {
			log.Dbg("< %s with %d byte%s", result,
				len(output), log.GetPlural(len(output)))
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(output); err != nil {
		log.Dbg("! writing response body failed: %v", err)
	}
	return status
}

func proxyStream(name string, input []byte, w http.ResponseWriter, r *http.Request, result string, model string) int {
	ollamaUrl := fmt.Sprintf("%s/api/%s", ollamaOrigin, name)
	req, err := web.CreateRawPostRequest(ollamaUrl, input)
	if err != nil {
		return failRequest(w, http.StatusInternalServerError, err.Error())
	}
	status, resReader, err := web.BeginRawRequest(req)
	if err != nil {
		return failRequest(w, status, err.Error())
	}
	defer func() {
		if err := resReader.Close(); err != nil {
			log.Dbg("closing response body stream failed: %v", err)
		}
	}()
	accept := r.Header.Get("Accept")
	sse := strings.HasPrefix(accept, "text/event-stream")
	if sse {
		w.Header().Set("Accept", "text/event-stream")
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	buf := make([]byte, 1024*1024)
	for {
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		size, err := resReader.Read(buf)
		if err == io.EOF {
			if size == 0 {
				break
			} else {
				if log.IsDbg {
					if len(model) > 0 {
						log.Dbg("< %s by %s with %d byte%s and EOF", result, model, size, log.GetPlural(size))
					} else {
						log.Dbg("< %s with %d byte%s and EOF", result, size, log.GetPlural(size))
					}
				}
			}
		} else if err != nil {
			log.Dbg("reading response body stream failed: %v", err)
		} else {
			if log.IsDbg {
				if len(model) > 0 {
					log.Dbg("< %s by %s with %d byte%s", result, model, size, log.GetPlural(size))
				} else {
					log.Dbg("< %s with %d byte%s", result, size, log.GetPlural(size))
				}
			}
		}
		if sse {
			web.WriteResponseString(w, "data: ")
		}
		if _, err := w.Write(buf[0:size]); err != nil {
			log.Dbg("! writing response body failed: %v", err)
		}
		if sse {
			web.WriteResponseString(w, "\n\n")
		}
	}
	return status
}
