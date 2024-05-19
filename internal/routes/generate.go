package routes

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/prantlf/ovai/internal/cfg"
	"github.com/prantlf/ovai/internal/log"
)

type modelParameters struct {
	MaxOutputTokens int     `json:"num_predict"`
	Temperature     float64 `json:"temperature"`
	TopP            float64 `json:"top_p"`
	TopK            int     `json:"top_k"`
}

type generateInput struct {
	Model   string          `json:"model"`
	Prompt  string          `json:"prompt"`
	Stream  bool            `json:"stream"`
	Options modelParameters `json:"options"`
}

type generateModelHandler interface {
	prepareBody(input *generateInput) (string, interface{}, interface{})
	extractResponse(data interface{}) (string, int, int)
}

type generateGeminiHandler struct{}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiBody struct {
	Contents         []geminiContent      `json:"contents"`
	GenerationConfig cfg.GenerationConfig `json:"generationConfig"`
	SafetySettings   []cfg.SafetySetting  `json:"safetySettings"`
}

type geminiCandidate struct {
	Content geminiContent `json:"content"`
}

type geminiMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
}

type geminiOutput struct {
	Candidates    []geminiCandidate `json:"candidates"`
	UsageMetadata geminiMetadata    `json:"usageMetadata"`
}

func mergeParameters(target *cfg.GenerationConfig, source *modelParameters) {
	if source.MaxOutputTokens > 0 {
		target.MaxOutputTokens = source.MaxOutputTokens
	}
	if source.Temperature >= 0 {
		target.Temperature = source.Temperature
	}
	if source.TopP >= 0 {
		target.TopP = source.TopP
	}
	if source.TopK > 0 {
		target.TopK = source.TopK
	}
}

func (h *generateGeminiHandler) prepareBody(input *generateInput) (string, interface{}, interface{}) {
	urlPrefix := input.Model + ":generateContent"
	generationConfig := cfg.Defaults.GeminiDefaults.GenerationConfig
	mergeParameters(&generationConfig, &input.Options)
	body := &geminiBody{
		Contents: []geminiContent{
			{
				Role: "user",
				Parts: []geminiPart{
					{
						Text: input.Prompt,
					},
				},
			},
		},
		GenerationConfig: generationConfig,
		SafetySettings:   cfg.Defaults.GeminiDefaults.SafetySettings,
	}
	return urlPrefix, body, &geminiOutput{}
}

func extractGeminiResponse(data interface{}) (string, int, int) {
	output, ok := data.(*geminiOutput)
	if !ok {
		log.Ftl("invalid gemini response type")
	}
	answer := ""
	if len(output.Candidates) > 0 {
		parts := output.Candidates[0].Content.Parts
		if len(parts) > 0 {
			answer = parts[0].Text
		}
	}
	metadata := output.UsageMetadata
	return answer, metadata.PromptTokenCount, metadata.CandidatesTokenCount
}

func (h *generateGeminiHandler) extractResponse(data interface{}) (string, int, int) {
	return extractGeminiResponse(data)
}

type generateBisonHandler struct{}

type generateBisonInstance struct {
	Prompt string `json:"prompt"`
}

type generateBisonBody struct {
	Instances  []generateBisonInstance `json:"instances"`
	Parameters cfg.GenerationConfig    `json:"parameters"`
}

type generatePrediction struct {
	Content string `json:"content"`
}

type tokenCount struct {
	TotalTokens int `json:"totalTokens"`
}

type tokenMetadata struct {
	InputTokenCount  tokenCount `json:"inputTokenCount"`
	OutputTokenCount tokenCount `json:"outputTokenCount"`
}

type bisonMetadata struct {
	TokenMetadata tokenMetadata `json:"tokenMetadata"`
}

type generateBisonOutput struct {
	Predictions []generatePrediction `json:"predictions"`
	Metadata    bisonMetadata        `json:"metadata"`
}

func (h *generateBisonHandler) prepareBody(input *generateInput) (string, interface{}, interface{}) {
	urlPrefix := input.Model + ":predict"
	parameters := cfg.Defaults.BisonDefaults.Parameters
	mergeParameters(&parameters, &input.Options)
	body := &generateBisonBody{
		Instances: []generateBisonInstance{
			{
				Prompt: input.Prompt,
			},
		},
		Parameters: parameters,
	}
	return urlPrefix, body, &generateBisonOutput{}
}

func (h *generateBisonHandler) extractResponse(data interface{}) (string, int, int) {
	output, ok := data.(*generateBisonOutput)
	if !ok {
		log.Ftl("invalid bison response type")
	}
	answer := ""
	if len(output.Predictions) > 0 {
		answer = output.Predictions[0].Content
	}
	metadata := output.Metadata.TokenMetadata
	return answer, metadata.InputTokenCount.TotalTokens, metadata.OutputTokenCount.TotalTokens
}

type generateResponse struct {
	Model              string `json:"model"`
	CreatedAt          string `json:"created_at"`
	Response           string `json:"response"`
	Done               bool   `json:"done"`
	TotalDuration      int64  `json:"total_duration"`
	LoadDuration       int64  `json:"load_duration"`
	PromptEvalCount    int    `json:"prompt_eval_count"`
	PromptEvalDuration int64  `json:"prompt_eval_duration"`
	EvalCount          int    `json:"eval_count"`
	EvalDuration       int64  `json:"eval_duration"`
}

func HandleGenerate(w http.ResponseWriter, r *http.Request) int {
	input := generateInput{
		Options: modelParameters{
			Temperature: -1,
			TopP:        -1,
		},
		Stream: true,
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return wrongInput(w, fmt.Sprintf("decoding request body failed: %v", err))
	}
	if len(input.Model) == 0 {
		return wrongInput(w, "model missing")
	}
	if len(input.Prompt) == 0 {
		return wrongInput(w, "prompt missing")
	}
	var handler generateModelHandler
	if strings.HasPrefix(input.Model, "gemini") {
		handler = &generateGeminiHandler{}
	} else if strings.HasPrefix(input.Model, "text-bison") {
		handler = &generateBisonHandler{}
	} else {
		return wrongInput(w, fmt.Sprintf("unrecognised model %q", input.Model))
	}
	if input.Stream {
		return wrongInput(w, "streaming not supported")
	}
	if log.IsDbg {
		log.Dbg("> generate from %d character%s using %s", len(input.Prompt),
			log.GetPlural(len(input.Prompt)), input.Model)
	}

	urlSuffix, reqBody, output := handler.prepareBody(&input)
	status, duration, err := forwardRequest(urlSuffix, reqBody, output)
	if err != nil {
		return failRequest(w, status, err.Error())
	}
	content, promptTokens, contentTokens := handler.extractResponse(output)
	tokens := promptTokens + contentTokens
	if log.IsDbg {
		log.Dbg("< result by %s with %d character%s and %d token%s", input.Model,
			len(content), log.GetPlural(len(content)), tokens, log.GetPlural(tokens))
	}
	promptDuration := int64(math.Round(float64(int64(duration) / 4)))
	resBody := &generateResponse{
		Model:              input.Model,
		CreatedAt:          time.Now().UTC().Format(time.RFC3339),
		Response:           content,
		Done:               true,
		TotalDuration:      int64(duration),
		LoadDuration:       0,
		PromptEvalCount:    promptTokens,
		PromptEvalDuration: promptDuration,
		EvalCount:          contentTokens,
		EvalDuration:       int64(duration) - promptDuration,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err = json.NewEncoder(w).Encode(resBody); err != nil {
		log.Dbg("! encoding response body failed: %v", err)
	}
	return http.StatusOK
}
