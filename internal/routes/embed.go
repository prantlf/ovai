package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/prantlf/ovai/internal/log"
)

type inputPrompt []string

type embedInput struct {
	Model string      `json:"model"`
	Input inputPrompt `json:"input"`
}

type embedOutput struct {
	Embeddings [][]float64 `json:"embeddings"`
}

func (t *inputPrompt) UnmarshalJSON(data []byte) error {
	if len(data) > 1 && data[0] == '[' {
		var array []string
		if err := json.Unmarshal(data, &array); err != nil {
			return err
		}
		*t = array
	} else {
		var scalar string
		if err := json.Unmarshal(data, &scalar); err != nil {
			return err
		}
		array := make([]string, 1)
		array[0] = scalar
		*t = array
	}
	return nil
}

func HandleEmbed(w http.ResponseWriter, r *http.Request) int {
	var input embedInput
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
	if len(input.Model) == 0 {
		return wrongInput(w, "model missing")
	}
	if len(input.Input) == 0 {
		return wrongInput(w, "input missing")
	}
	totalChars := 0
	for i, text := range input.Input {
		if len(text) == 0 {
			return wrongInput(w, "input "+strconv.Itoa(i)+" empty")
		}
		totalChars += len(text)
	}

	var forward bool
	if strings.HasPrefix(input.Model, "textembedding-gecko") ||
		strings.HasPrefix(input.Model, "textembedding-gecko-multilingual") ||
		strings.HasPrefix(input.Model, "text-embedding") ||
		strings.HasPrefix(input.Model, "multimodalembedding") ||
		strings.HasPrefix(input.Model, "text-multilingual-embedding") {
		forward = true
	} else if !canProxy {
		return wrongInput(w, fmt.Sprintf("unrecognised model %q", input.Model))
	}
	if log.IsDbg {
		log.Dbg("> vectorise %d text%s with %d character%s using %s", len(input.Input),
			log.GetPlural(len(input.Input)), totalChars, log.GetPlural(totalChars), input.Model)
	}

	if !forward {
		return proxyRequest("embed", reqPayload, w, "embeddings", input.Model)
	}
	embeddings := make([][]float64, len(input.Input))
	for i, text := range input.Input {
		reqBody := &embeddingsBody{
			Instances: []instance{
				{
					Content: text,
				},
			},
		}
		var resBody embeddingsResponse
		status, _, err := forwardRequest(input.Model+":predict", reqBody, &resBody)
		if err != nil {
			return failRequest(w, status, err.Error())
		}
		values, tokens := extractEmbeddingsResponse(&resBody)
		if log.IsDbg {
			log.Dbg("< embedding by %s with %d float%s from %d token%s", input.Model,
				len(values), log.GetPlural(len(values)), tokens, log.GetPlural(tokens))
		}
		embeddings[i] = values
	}
	output := &embedOutput{
		Embeddings: embeddings,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(output); err != nil {
		log.Dbg("! encoding response body failed: %v", err)
	}
	return http.StatusOK
}
