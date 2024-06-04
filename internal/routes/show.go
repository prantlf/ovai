package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/prantlf/ovai/internal/log"
)

type showInput struct {
	Name string `json:"name"`
}

type showOutput struct {
	License    string       `json:"license"`
	ModelFile  string       `json:"modelfile"`
	Parameters string       `json:"parameters"`
	Template   string       `json:"template"`
	Details    modelDetails `json:"details"`
}

func HandleShow(w http.ResponseWriter, r *http.Request) int {
	var input showInput
	reqPayload, err := io.ReadAll(r.Body)
	if err != nil {
		return wrongInput(w, fmt.Sprintf("reading request body failed: %v", err))
	}
	if err := json.Unmarshal(reqPayload, &input); err != nil {
		return wrongInput(w, fmt.Sprintf("decoding request body failed: %v", err))
	}
	// if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
	// 	return wrongInput(w, fmt.Sprintf("decoding request body failed: %v", err))
	// }
	if len(input.Name) == 0 {
		return wrongInput(w, "model name missing")
	}
	log.Dbg("> look for %s", input.Name)
	if strings.HasPrefix(input.Name, "gemini") ||
		strings.HasPrefix(input.Name, "text-bison") ||
		strings.HasPrefix(input.Name, "chat-bison") ||
		strings.HasPrefix(input.Name, "textembedding-gecko") {
		var details *modelDetails
		for _, model := range googleModels {
			if model.Name == input.Name {
				details = &model.Details
			}
		}
		if details != nil {
			if log.IsDbg {
				log.Dbg("< found %s", input.Name)
			}
			output := &showOutput{
				Details: *details,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(output); err != nil {
				log.Dbg("! encoding response body failed: %v", err)
			}
			return http.StatusOK
		}
	} else if canProxy {
		return proxyRequest("show", reqPayload, w, "model", "")
	}
	return wrongInput(w, fmt.Sprintf("unrecognised model %q", input.Name))
}
