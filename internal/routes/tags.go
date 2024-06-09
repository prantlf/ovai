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
	Size       int64        `json:"message"`
	Digest     string       `json:"digest"`
	Details    modelDetails `json:"details"`
	ExpiresAt  string       `json:"expires_at"`
}

type tagsOutput struct {
	Models []modelInfo `json:"models"`
}

var googleModels = []modelInfo{
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
		Name:       "chat-bison-32k@002",
		Model:      "chat-bison-32k@002",
		ModifiedAt: "2023-12-04T00:00:00.000Z",
		Details: modelDetails{
			Family: "chat-bison-32k",
			Families: []string{
				"chat-bison-32k",
			},
			Purposes: []string{"chat"},
		},
		ExpiresAt: "2024-10-09T00:00:00.000Z",
	},
	{
		Name:       "chat-bison@002",
		Model:      "chat-bison@002",
		ModifiedAt: "2023-12-06T00:00:00.000Z",
		Details: modelDetails{
			Family: "chat-bison",
			Families: []string{
				"chat-bison",
			},
			Purposes: []string{"chat"},
		},
		ExpiresAt: "2024-10-09T00:00:00.000Z",
	},
	{
		Name:       "chat-bison@001",
		Model:      "chat-bison@001",
		ModifiedAt: "2023-07-10T00:00:00.000Z",
		Details: modelDetails{
			Family: "chat-bison",
			Families: []string{
				"chat-bison",
			},
			Purposes: []string{"chat"},
		},
		ExpiresAt: "2024-07-06T00:00:00.000Z",
	},
	{
		Name:       "text-bison-32k@002",
		Model:      "text-bison-32k@002",
		ModifiedAt: "2023-12-04T00:00:00.000Z",
		Details: modelDetails{
			Family: "text-bison-32k",
			Families: []string{
				"text-bison-32k",
			},
			Purposes: []string{"generate"},
		},
		ExpiresAt: "2024-10-09T00:00:00.000Z",
	},
	{
		Name:       "text-bison@002",
		Model:      "text-bison@002",
		ModifiedAt: "2023-12-06T00:00:00.000Z",
		Details: modelDetails{
			Family: "text-bison",
			Families: []string{
				"text-bison",
			},
			Purposes: []string{"generate"},
		},
		ExpiresAt: "2024-10-09T00:00:00.000Z",
	},
	{
		Name:       "text-bison@001",
		Model:      "text-bison@001",
		ModifiedAt: "2023-06-07T00:00:00.000Z",
		Details: modelDetails{
			Family: "text-bison",
			Families: []string{
				"text-bison",
			},
			Purposes: []string{"generate"},
		},
		ExpiresAt: "2024-07-06T00:00:00.000Z",
	},
	{
		Name:       "text-unicorn@001",
		Model:      "text-unicorn@001",
		ModifiedAt: "2023-11-30T00:00:00.000Z",
		Details: modelDetails{
			Family: "text-unicorn",
			Families: []string{
				"text-unicorn",
			},
			Purposes: []string{"generate"},
		},
		ExpiresAt: "2024-11-30T00:00:00.000Z",
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
		ExpiresAt: "2025-02-12T00:00:00.000Z",
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
		ExpiresAt: "0001-01-01T00:00:00Z",
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
		ExpiresAt: "2024-07-06T00:00:00.000Z",
	},
}

func HandleTags(w http.ResponseWriter, r *http.Request) int {
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
