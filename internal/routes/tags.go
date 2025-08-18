package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/prantlf/ovai/internal/log"
	"github.com/prantlf/ovai/internal/web"
)

type modelDetails struct {
	ParentModel       string   `json:"parent_model"`
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
	Purposes          []string `json:"purposes"`
}

type modelInfo struct {
	Name       string       `json:"name"`
	Model      string       `json:"model"`
	ModifiedAt string       `json:"modified_at"`
	Size       int64        `json:"size"`
	Digest     string       `json:"digest"`
	Details    modelDetails `json:"details"`
	ExpiresAt  string       `json:"expires_at"`
}

type tagsOutput struct {
	Models []modelInfo `json:"models"`
}

var googleModels = []modelInfo{
	{
		Name:       "gemini-2.5-flash-lite",
		Model:      "gemini-2.5-flash-lite",
		ModifiedAt: "2025-07-22T00:00:00.000Z",
		Details: modelDetails{
			Family: "gemini-2.5-flash-lite",
			Families: []string{
				"gemini-2.5-flash-lite",
			},
			Purposes: []string{"chat", "generate"},
		},
		ExpiresAt: "2026-07-22T00:00:00.000Z",
	},
	{
		Name:       "gemini-2.5-flash-lite-preview-06-17",
		Model:      "gemini-2.5-flash-lite-preview-06-17",
		ModifiedAt: "2025-06-17T00:00:00.000Z",
		Details: modelDetails{
			Family: "gemini-2.5-flash-lite",
			Families: []string{
				"gemini-2.5-flash-lite",
			},
			Purposes: []string{"chat", "generate"},
		},
		ExpiresAt: "0001-01-01T00:00:00Z",
	},
	{
		Name:       "gemini-2.5-pro",
		Model:      "gemini-2.5-pro",
		ModifiedAt: "2025-06-17T00:00:00.000Z",
		Details: modelDetails{
			Family: "gemini-2.5-pro",
			Families: []string{
				"gemini-2.5-pro",
			},
			Purposes: []string{"chat", "generate"},
		},
		ExpiresAt: "2025-06-17T00:00:00.000Z",
	},
	{
		Name:       "gemini-2.5-flash",
		Model:      "gemini-2.5-flash",
		ModifiedAt: "2025-06-17T00:00:00.000Z",
		Details: modelDetails{
			Family: "gemini-2.5-flash",
			Families: []string{
				"gemini-2.5-flash",
			},
			Purposes: []string{"chat", "generate"},
		},
		ExpiresAt: "2025-06-17T00:00:00.000Z",
	},
	{
		Name:       "gemini-2.5-pro-preview-06-05",
		Model:      "gemini-2.5-pro-preview-06-05",
		ModifiedAt: "2025-01-01T00:00:00.000Z",
		Details: modelDetails{
			Family: "gemini-2.5-pro",
			Families: []string{
				"gemini-2.5-pro",
			},
			Purposes: []string{"chat", "generate"},
		},
		ExpiresAt: "0001-01-01T00:00:00Z",
	},
	{
		Name:       "gemini-2.5-flash-preview-05-20",
		Model:      "gemini-2.5-flash-preview-05-20",
		ModifiedAt: "2025-01-01T00:00:00.000Z",
		Details: modelDetails{
			Family: "gemini-2.5-flash",
			Families: []string{
				"gemini-2.5-flash",
			},
			Purposes: []string{"chat", "generate"},
		},
		ExpiresAt: "0001-01-01T00:00:00Z",
	},
	{
		Name:       "gemini-2.0-flash-thinking-exp-01-21",
		Model:      "gemini-2.0-flash-thinking-exp-01-21",
		ModifiedAt: "2025-01-25T00:00:00.000Z",
		Details: modelDetails{
			Family: "gemini-2.0-flash-thinking",
			Families: []string{
				"gemini-2.0-flash-thinking",
			},
			Purposes: []string{"chat", "generate"},
		},
		ExpiresAt: "0001-01-01T00:00:00Z",
	},
	{
		Name:       "gemini-2.5-pro-exp-03-25",
		Model:      "gemini-2.5-pro-exp-03-25",
		ModifiedAt: "2025-01-01T00:00:00.000Z",
		Details: modelDetails{
			Family: "gemini-2.5-pro",
			Families: []string{
				"gemini-2.5-pro",
			},
			Purposes: []string{"chat", "generate"},
		},
		ExpiresAt: "0001-01-01T00:00:00Z",
	},
	{
		Name:       "gemini-2.0-flash-lite-001",
		Model:      "gemini-2.0-flash-lite-001",
		ModifiedAt: "2025-02-25T00:00:00.000Z",
		Details: modelDetails{
			Family: "gemini-2.0-flash-lite",
			Families: []string{
				"gemini-2.0-flash-lite",
			},
			Purposes: []string{"chat", "generate"},
		},
		ExpiresAt: "2026-02-25T00:00:00.000Z",
	},
	{
		Name:       "gemini-2.0-flash-001",
		Model:      "gemini-2.0-flash-001",
		ModifiedAt: "2025-02-05T00:00:00.000Z",
		Details: modelDetails{
			Family: "gemini-2.0-flash",
			Families: []string{
				"gemini-2.0-flash",
			},
			Purposes: []string{"chat", "generate"},
		},
		ExpiresAt: "2026-02-05T00:00:00.000Z",
	},
	{
		Name:       "gemini-2.0-flash-exp",
		Model:      "gemini-2.0-flash-exp",
		ModifiedAt: "2024-12-01T00:00:00.000Z",
		Details: modelDetails{
			Family: "gemini-2.0-flash",
			Families: []string{
				"gemini-2.0-flash",
			},
			Purposes: []string{"chat", "generate"},
		},
		ExpiresAt: "0001-01-01T00:00:00Z",
	},
	{
		Name:       "gemini-1.5-flash-002",
		Model:      "gemini-1.5-flash-002",
		ModifiedAt: "2024-09-24T00:00:00.000Z",
		Details: modelDetails{
			Family: "gemini-1.5-flash",
			Families: []string{
				"gemini-1.5-flash",
			},
			Purposes: []string{"chat", "generate"},
		},
		ExpiresAt: "2025-09-24T00:00:00.000Z",
	},
	{
		Name:       "gemini-1.5-flash-8b-001",
		Model:      "gemini-1.5-flash-8b-001",
		ModifiedAt: "2024-05-24T00:00:00.000Z",
		Details: modelDetails{
			Family: "gemini-1.5-flash",
			Families: []string{
				"gemini-1.5-flash",
			},
			Purposes: []string{"chat", "generate"},
		},
		ExpiresAt: "2025-05-24T00:00:00.000Z",
	},
	{
		Name:       "gemini-1.5-flash-001",
		Model:      "gemini-1.5-flash-001",
		ModifiedAt: "2024-05-24T00:00:00.000Z",
		Details: modelDetails{
			Family: "gemini-1.5-flash",
			Families: []string{
				"gemini-1.5-flash",
			},
			Purposes: []string{"chat", "generate"},
		},
		ExpiresAt: "2025-05-24T00:00:00.000Z",
	},
	{
		Name:       "gemini-1.5-pro-002",
		Model:      "gemini-1.5-pro-002",
		ModifiedAt: "2024-09-24T00:00:00.000Z",
		Details: modelDetails{
			Family: "gemini-1.5-pro",
			Families: []string{
				"gemini-1.5-pro",
			},
			Purposes: []string{"chat", "generate"},
		},
		ExpiresAt: "2025-09-24T00:00:00.000Z",
	},
	{
		Name:       "gemini-1.5-pro-001",
		Model:      "gemini-1.5-pro-001",
		ModifiedAt: "2024-05-24T00:00:00.000Z",
		Details: modelDetails{
			Family: "gemini-1.5-pro",
			Families: []string{
				"gemini-1.5-pro",
			},
			Purposes: []string{"chat", "generate"},
		},
		ExpiresAt: "2025-05-24T00:00:00.000Z",
	},
	{
		Name:       "gemini-1.0-pro-vision-001",
		Model:      "gemini-1.0-pro-vision-001",
		ModifiedAt: "2024-02-15T00:00:00.000Z",
		Details: modelDetails{
			Family: "gemini-1.0-pro-vision",
			Families: []string{
				"gemini-1.0-pro-vision",
			},
			Purposes: []string{"chat", "generate"},
		},
		ExpiresAt: "2025-02-15T00:00:00.000Z",
	},
	{
		Name:       "gemini-1.0-pro-002",
		Model:      "gemini-1.0-pro-002",
		ModifiedAt: "2024-04-09T00:00:00.000Z",
		Details: modelDetails{
			Family: "gemini-1.0-pro",
			Families: []string{
				"gemini-1.0-pro",
			},
			Purposes: []string{"chat", "generate"},
		},
		ExpiresAt: "2025-04-09T00:00:00.000Z",
	},
	{
		Name:       "gemini-1.0-pro-001",
		Model:      "gemini-1.0-pro-001",
		ModifiedAt: "2024-02-15T00:00:00.000Z",
		Details: modelDetails{
			Family: "gemini-1.0-pro",
			Families: []string{
				"gemini-1.0-pro",
			},
			Purposes: []string{"chat", "generate"},
		},
		ExpiresAt: "2025-02-15T00:00:00.000Z",
	},
	{
		Name:       "gemini-embedding-001",
		Model:      "gemini-embedding-001",
		ModifiedAt: "2025-05-20T00:00:00.000Z",
		Details: modelDetails{
			Family: "gemini-embedding",
			Families: []string{
				"gemini-embedding",
			},
			Purposes: []string{"embeddings"},
		},
		ExpiresAt: "0001-01-01T00:00:00Z",
	},
	{
		Name:       "multimodalembedding@001",
		Model:      "multimodalembedding@001",
		ModifiedAt: "2024-12-02T00:00:00.000Z",
		Details: modelDetails{
			Family: "multimodalembedding",
			Families: []string{
				"multimodalembedding",
			},
			Purposes: []string{"embeddings"},
		},
		ExpiresAt: "0001-01-01T00:00:00Z",
	},
	{
		Name:       "textembedding-gecko@002",
		Model:      "textembedding-gecko@002",
		ModifiedAt: "2023-11-02T00:00:00.000Z",
		Details: modelDetails{
			Family: "textembedding-gecko",
			Families: []string{
				"textembedding-gecko",
			},
			Purposes: []string{"embeddings"},
		},
		ExpiresAt: "2024-10-09T00:00:00.000Z",
	},
	{
		Name:       "text-multilingual-embedding-002",
		Model:      "text-multilingual-embedding-002",
		ModifiedAt: "2024-05-14T00:00:00.000Z",
		Details: modelDetails{
			Family: "text-multilingual-embedding",
			Families: []string{
				"text-multilingual-embedding",
			},
			Purposes: []string{"embeddings"},
		},
		ExpiresAt: "0001-01-01T00:00:00Z",
	},
	{
		Name:       "text-embedding-005",
		Model:      "text-embedding-005",
		ModifiedAt: "2024-11-18T00:00:00.000Z",
		Details: modelDetails{
			Family: "text-embedding",
			Families: []string{
				"textembedding-gecko-multilingual",
			},
			Purposes: []string{"embeddings"},
		},
		ExpiresAt: "0001-01-01T00:00:00Z",
	},
	{
		Name:       "text-embedding-004",
		Model:      "text-embedding-004",
		ModifiedAt: "2024-05-14T00:00:00.000Z",
		Details: modelDetails{
			Family: "text-embedding",
			Families: []string{
				"textembedding-gecko-multilingual",
			},
			Purposes: []string{"embeddings"},
		},
		ExpiresAt: "2025-11-18T00:00:00.000Z",
	},
	{
		Name:       "textembedding-gecko-multilingual@001",
		Model:      "textembedding-gecko-multilingual@001",
		ModifiedAt: "2023-11-02T00:00:00.000Z",
		Details: modelDetails{
			Family: "textembedding-gecko-multilingual",
			Families: []string{
				"textembedding-gecko-multilingual",
			},
			Purposes: []string{"embeddings"},
		},
		ExpiresAt: "2025-05-14T00:00:00.000Z",
	},
	{
		Name:       "textembedding-gecko@003",
		Model:      "textembedding-gecko@003",
		ModifiedAt: "2023-12-12T00:00:00.000Z",
		Details: modelDetails{
			Family: "textembedding-gecko",
			Families: []string{
				"textembedding-gecko",
			},
			Purposes: []string{"embeddings"},
		},
		ExpiresAt: "2025-05-14T00:00:00.000Z",
	},
	{
		Name:       "textembedding-gecko@002",
		Model:      "textembedding-gecko@002",
		ModifiedAt: "2023-11-02T00:00:00.000Z",
		Details: modelDetails{
			Family: "textembedding-gecko",
			Families: []string{
				"textembedding-gecko",
			},
			Purposes: []string{"embeddings"},
		},
		ExpiresAt: "2025-04-09T00:00:00.000Z",
	},
	{
		Name:       "textembedding-gecko@001",
		Model:      "textembedding-gecko@001",
		ModifiedAt: "2023-06-07T00:00:00.000Z",
		Details: modelDetails{
			Family: "textembedding-gecko",
			Families: []string{
				"textembedding-gecko",
			},
			Purposes: []string{"embeddings"},
		},
		ExpiresAt: "2025-04-09T00:00:00.000Z",
	},
}

func HandleTags(w http.ResponseWriter, r *http.Request) int {
	if r.Method == "HEAD" {
		w.WriteHeader(http.StatusOK)
		return http.StatusOK
	}
	log.Dbg("> list models")
	output := &tagsOutput{}
	if canProxy {
		ollamaUrl := fmt.Sprintf("%s/api/tags", ollamaOrigin)
		req, err := web.CreateGetRequest(ollamaUrl)
		if err != nil {
			return failRequest(w, http.StatusInternalServerError, err.Error())
		}
		status, err := web.DispatchRequest(req, output)
		if err != nil {
			return failRequest(w, status, err.Error())
		}
	}
	models := append(googleModels, output.Models...)
	if log.IsDbg {
		var names []string
		for _, model := range models {
			names = append(names, model.Name)
		}
		log.Dbg("< %d model%s: %v", len(models), log.GetPlural(len(models)), names)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	output.Models = models
	if err := json.NewEncoder(w).Encode(output); err != nil {
		log.Dbg("! encoding response body failed: %v", err)
	}
	return http.StatusOK
}
