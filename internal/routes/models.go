package routes

import (
	"encoding/json"
	"net/http"

	"github.com/prantlf/ovai/internal/log"
)

type openaiModel struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

type modelsOutput struct {
	Object string        `json:"object"`
	Data   []openaiModel `json:"data"`
}

func convertModelsToOpenAI(models []modelInfo) []openaiModel {
	var data []openaiModel
	for _, m := range models {
		data = append(data, openaiModel{
			ID:      m.Name,
			Object:  "model",
			Created: 0,
			OwnedBy: "ovai",
		})
	}
	return data
}

func HandleModels(w http.ResponseWriter, r *http.Request) int {
	if r.Method == "HEAD" {
		w.WriteHeader(http.StatusOK)
		return http.StatusOK
	}
	models, status := getAllModels(w)
	if status != http.StatusOK {
		return status
	}
	data := convertModelsToOpenAI(models)
	resp := modelsOutput{
		Object: "list",
		Data:   data,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Dbg("! encoding response body failed: %v", err)
	}
	return http.StatusOK
}
