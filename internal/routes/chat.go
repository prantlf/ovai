package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/prantlf/ovai/internal/cfg"
	"github.com/prantlf/ovai/internal/log"
)

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatInput struct {
	Model    string          `json:"model"`
	Messages []message       `json:"messages"`
	Stream   bool            `json:"stream"`
	Options  modelParameters `json:"options"`
}

type chatModelHandler interface {
	prepareBody(input *chatInput) (string, interface{}, interface{}, error)
	extractResponse(data interface{}) (string, int, int)
}

type chatGeminiHandler struct{}

func convertGeminiMessages(messages []message) ([]geminiContent, error) {
	systemMessages := make([]geminiPart, 0, 1)
	chatMessages := make([]geminiContent, 0, len(messages))
	for _, msg := range messages {
		if msg.Role == "system" {
			systemMessages = append(systemMessages, geminiPart{
				Text: msg.Content,
			})
		} else {
			var role string
			if msg.Role == "user" {
				role = "user"
			} else if msg.Role == "assistant" {
				role = "model"
			} else {
				return []geminiContent{}, fmt.Errorf("invalid chat message role: %q", msg.Role)
			}
			chatMessages = append(chatMessages, geminiContent{
				Role: role,
				Parts: []geminiPart{
					{
						Text: msg.Content,
					},
				},
			})
		}
	}
	if len(chatMessages) == 0 {
		return []geminiContent{}, errors.New("no user message found")
	}
	if len(systemMessages) > 0 {
		chatMessages[0].Parts = append(systemMessages, chatMessages[0].Parts...)
	}
	return chatMessages, nil
}

func (h *chatGeminiHandler) prepareBody(input *chatInput) (string, interface{}, interface{}, error) {
	urlPrefix := input.Model + ":generateContent"
	chatMessages, err := convertGeminiMessages(input.Messages)
	if err != nil {
		return "", nil, nil, err
	}
	generationConfig := cfg.Defaults.GeminiDefaults.GenerationConfig
	mergeParameters(&generationConfig, &input.Options)
	body := &geminiBody{
		Contents:         chatMessages,
		GenerationConfig: generationConfig,
		SafetySettings:   cfg.Defaults.GeminiDefaults.SafetySettings,
	}
	return urlPrefix, body, &geminiOutput{}, nil
}

func (h *chatGeminiHandler) extractResponse(data interface{}) (string, int, int) {
	return extractGeminiResponse(data)
}

type chatBisonHandler struct{}

type bisonMessage struct {
	Author  string `json:"author"`
	Content string `json:"content"`
}

type chatBisonInstance struct {
	Context  string         `json:"context"`
	Examples []string       `json:"examples"`
	Messages []bisonMessage `json:"messages"`
}

type bisonBody struct {
	Instances  []chatBisonInstance  `json:"instances"`
	Parameters cfg.GenerationConfig `json:"parameters"`
}

type bisonCandidate struct {
	Content string `json:"content"`
}

type chatPrediction struct {
	Candidates []bisonCandidate `json:"candidates"`
}

type bisonOutput struct {
	Predictions []chatPrediction `json:"predictions"`
	Metadata    bisonMetadata    `json:"metadata"`
}

func convertBisonMessages(messages []message) (string, []bisonMessage, error) {
	systemMessages := make([]string, 0, 1)
	chatMessages := make([]bisonMessage, 0, len(messages))
	for _, msg := range messages {
		if msg.Role == "system" {
			systemMessages = append(systemMessages, msg.Content)
		} else {
			var role string
			if msg.Role == "user" {
				role = "user"
			} else if msg.Role == "assistant" {
				role = "bot"
			} else {
				return "", []bisonMessage{}, fmt.Errorf("invalid chat message role: %q", msg.Role)
			}
			chatMessages = append(chatMessages, bisonMessage{
				Author:  role,
				Content: msg.Content,
			})
		}
	}
	if len(chatMessages) == 0 {
		return "", []bisonMessage{}, errors.New("no user message found")
	}
	var context string
	if len(systemMessages) > 0 {
		context = strings.Join(systemMessages, "\n")
	} else {
		context = ""
	}
	return context, chatMessages, nil
}

func (h *chatBisonHandler) prepareBody(input *chatInput) (string, interface{}, interface{}, error) {
	urlPrefix := input.Model + ":predict"
	context, chatMessages, err := convertBisonMessages(input.Messages)
	if err != nil {
		return "", nil, nil, err
	}
	parameters := cfg.Defaults.BisonDefaults.Parameters
	mergeParameters(&parameters, &input.Options)
	body := &bisonBody{
		Instances: []chatBisonInstance{
			{
				Context:  context,
				Examples: []string{},
				Messages: chatMessages,
			},
		},
		Parameters: parameters,
	}
	return urlPrefix, body, &bisonOutput{}, nil
}

func (h *chatBisonHandler) extractResponse(data interface{}) (string, int, int) {
	output, ok := data.(*bisonOutput)
	if !ok {
		log.Ftl("invalid bison response type")
	}
	answer := ""
	if len(output.Predictions) > 0 {
		prediction := output.Predictions[0]
		if len(prediction.Candidates) > 0 {
			answer = prediction.Candidates[0].Content
		}
	}
	metadata := output.Metadata.TokenMetadata
	return answer, metadata.InputTokenCount.TotalTokens, metadata.OutputTokenCount.TotalTokens
}

type chatResponse struct {
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

func HandleChat(w http.ResponseWriter, r *http.Request) int {
	input := chatInput{
		Options: modelParameters{
			Temperature: -1,
			TopP:        -1,
		},
		Stream: true,
	}
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		return wrongInput(w, fmt.Sprintf("decoding request body failed: %v", err))
	}
	if len(input.Model) == 0 {
		return wrongInput(w, "model missing")
	}
	if len(input.Messages) == 0 {
		return wrongInput(w, "messages missing")
	}
	var handler chatModelHandler
	if strings.HasPrefix(input.Model, "gemini") {
		handler = &chatGeminiHandler{}
	} else if strings.HasPrefix(input.Model, "chat-bison") {
		handler = &chatBisonHandler{}
	} else {
		return wrongInput(w, fmt.Sprintf("unrecognised model %q", input.Model))
	}
	if input.Stream {
		return wrongInput(w, "streaming not supported")
	}
	if log.IsDbg {
		log.Dbg("> ask with %d message%s using %s", len(input.Messages),
			log.GetPlural(len(input.Messages)), input.Model)
	}

	urlSuffix, reqBody, output, err := handler.prepareBody(&input)
	if err != nil {
		return wrongInput(w, err.Error())
	}
	status, duration, err := forwardRequest(urlSuffix, reqBody, output)
	if err != nil {
		return failRequest(w, status, err.Error())
	}
	content, promptTokens, contentTokens := handler.extractResponse(output)
	tokens := promptTokens + contentTokens
	if log.IsDbg {
		log.Dbg("< answer by %s with %d character%s and %d token%s", input.Model,
			len(content), log.GetPlural(len(content)), tokens, log.GetPlural(tokens))
	}
	promptDuration := int64(math.Round(float64(int64(duration) / 4)))
	resBody := &chatResponse{
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
	err = json.NewEncoder(w).Encode(resBody)
	if err != nil {
		log.Dbg("! encoding response body failed: %v", err)
	}
	return http.StatusOK
}
