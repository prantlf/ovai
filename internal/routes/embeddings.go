package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/prantlf/ovai/internal/log"
)

type embeddingsInput struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type instance struct {
	Content string `json:"content"`
}

type embeddingsBody struct {
	Instances []instance `json:"instances"`
}

type statistics struct {
	TokenCount int `json:"token_count"`
}

type embeddings struct {
	Values     []float64  `json:"values"`
	Statistics statistics `json:"statistics"`
}

type embeddingPrediction struct {
	Embeddings embeddings `json:"embeddings"`
}

type embeddingsResponse struct {
	Predictions []embeddingPrediction `json:"predictions"`
}

type embeddingsOutput struct {
	Embedding []float64 `json:"embedding"`
}

func extractEmbeddingsResponse(res *embeddingsResponse) ([]float64, int) {
	if len(res.Predictions) > 0 {
		embeddings := res.Predictions[0].Embeddings
		return embeddings.Values, embeddings.Statistics.TokenCount
	}
	return []float64{}, 0
}

func HandleEmbeddings(w http.ResponseWriter, r *http.Request) int {
	var input embeddingsInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return wrongInput(w, fmt.Sprintf("decoding request body failed: %v", err))
	}
	if len(input.Model) == 0 {
		return wrongInput(w, "model missing")
	}
	if len(input.Prompt) == 0 {
		return wrongInput(w, "prompt missing")
	}
	if log.IsDbg {
		log.Dbg("> vectorise %d character%s using %s", len(input.Prompt),
			log.GetPlural(len(input.Prompt)), input.Model)
	}

	reqBody := &embeddingsBody{
		Instances: []instance{
			{
				Content: input.Prompt,
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
	output := &embeddingsOutput{
		Embedding: values,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err = json.NewEncoder(w).Encode(output); err != nil {
		log.Dbg("! encoding response body failed: %v", err)
	}
	return http.StatusOK
}
