package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/prantlf/ovai/internal/cfg"
	"github.com/prantlf/ovai/internal/log"
)

type message struct {
	Role    string   `json:"role"`
	Content string   `json:"content"`
	Images  []string `json:"images"`
}

type chatInput struct {
	Model    string          `json:"model"`
	Messages []message       `json:"messages"`
	Stream   bool            `json:"stream"`
	Options  modelParameters `json:"options"`
}

type chatModelHandler interface {
	prepareBody(input *chatInput) (string, interface{}, interface{}, error)
	prepareStream(input *chatInput) (string, interface{}, interface{}, interface{}, error)
	extractCompleteResponse(data interface{}) (string, string, int, int)
	extractStreamResponse(resReader io.Reader, partialData interface{}, finalData interface{}) (string, string, []byte, int, int, error)
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
				return nil, fmt.Errorf("invalid chat message role: %q", msg.Role)
			}
			parts, err := createGeminiParts(msg.Content, msg.Images)
			if err != nil {
				return nil, err
			}
			parts[0].Text = msg.Content
			chatMessages = append(chatMessages, geminiContent{
				Role:  role,
				Parts: parts,
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

func createChatGeminiBody(input *chatInput) (interface{}, error) {
	chatMessages, err := convertGeminiMessages(input.Messages)
	if err != nil {
		return nil, err
	}
	generationConfig := cfg.Defaults.GeminiDefaults.GenerationConfig
	mergeParameters(&generationConfig, &input.Options)
	body := &geminiBody{
		Contents:         chatMessages,
		GenerationConfig: generationConfig,
		SafetySettings:   cfg.Defaults.GeminiDefaults.SafetySettings,
	}
	return body, nil
}

func (h *chatGeminiHandler) prepareBody(input *chatInput) (string, interface{}, interface{}, error) {
	urlPrefix := input.Model + ":generateContent"
	body, err := createChatGeminiBody(input)
	if err != nil {
		return "", nil, nil, err
	}
	return urlPrefix, body, &geminiCompleteOutput{}, nil
}

func (h *chatGeminiHandler) prepareStream(input *chatInput) (string, interface{}, interface{}, interface{}, error) {
	urlPrefix := input.Model + ":streamGenerateContent?alt=sse"
	body, err := createChatGeminiBody(input)
	if err != nil {
		return "", nil, nil, nil, err
	}
	return urlPrefix, body, &geminiPartialOutput{}, &geminiFinalOutput{}, nil
}

func (h *chatGeminiHandler) extractCompleteResponse(data interface{}) (string, string, int, int) {
	return extractCompleteGeminiResponse(data)
}

func (h *chatGeminiHandler) extractStreamResponse(resReader io.Reader, partialData interface{}, finalData interface{}) (string, string, []byte, int, int, error) {
	return extractStreamGeminiResponse(resReader, partialData, finalData)
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
				return "", nil, fmt.Errorf("invalid chat message role: %q", msg.Role)
			}
			chatMessages = append(chatMessages, bisonMessage{
				Author:  role,
				Content: msg.Content,
			})
		}
	}
	if len(chatMessages) == 0 {
		return "", nil, errors.New("no user message found")
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

func (h *chatBisonHandler) prepareStream(input *chatInput) (string, interface{}, interface{}, interface{}, error) {
	log.Dbg("streaming for bison models not implemented")
	return "", nil, nil, nil, errors.New("streaming for bison models not implemented")
}

func (h *chatBisonHandler) extractCompleteResponse(data interface{}) (string, string, int, int) {
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
	return answer, "STOP", metadata.InputTokenCount.TotalTokens, metadata.OutputTokenCount.TotalTokens
}

func (h *chatBisonHandler) extractStreamResponse(resReader io.Reader, partialData interface{}, finalData interface{}) (string, string, []byte, int, int, error) {
	log.Dbg("streaming for bison models not implemented")
	return "", "", nil, 0, 0, errors.New("streaming for bison models not implemented")
}

type chatResponse struct {
	Model     string  `json:"model"`
	CreatedAt string  `json:"created_at"`
	Message   message `json:"message"`
	Done      bool    `json:"done"`
}

type chatCompleteResponse struct {
	chatResponse
	DoneReason         string `json:"done_reason"`
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
	if len(input.Messages) == 0 {
		return wrongInput(w, "messages missing")
	}
	var handler chatModelHandler
	if strings.HasPrefix(input.Model, "gemini") {
		handler = &chatGeminiHandler{}
	} else if strings.HasPrefix(input.Model, "chat-bison") {
		handler = &chatBisonHandler{}
	} else if !canProxy {
		return wrongInput(w, fmt.Sprintf("unrecognised model %q", input.Model))
	}
	if log.IsDbg {
		log.Dbg("> ask with %d message%s using %s", len(input.Messages),
			log.GetPlural(len(input.Messages)), input.Model)
	}

	if handler == nil {
		if input.Stream {
			return proxyStream("chat", reqPayload, w, "answer", input.Model)
		} else {
			return proxyRequest("chat", reqPayload, w, "answer", input.Model)
		}
	}

	if input.Stream {
		urlSuffix, reqBody, partialOutput, finalOutput, err := handler.prepareStream(&input)
		if err != nil {
			return wrongInput(w, err.Error())
		}
		status, start, resReader, err := forwardStream(urlSuffix, reqBody)
		if err != nil {
			return failRequest(w, status, err.Error())
		}
		defer func() {
			if err := resReader.Close(); err != nil {
				log.Dbg("closing response body stream failed: %v", err)
			}
		}()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		var rest []byte
		for {
			var content string
			var reason string
			var promptTokens int
			var contentTokens int
			if rest != nil && len(rest) > 0 {
				content, reason, rest, promptTokens, contentTokens, err = handler.extractStreamResponse(bytes.NewReader(rest), partialOutput, finalOutput)
			} else {
				content, reason, rest, promptTokens, contentTokens, err = handler.extractStreamResponse(resReader, partialOutput, finalOutput)
			}
			var resBody any
			final := false
			if err != nil {
				break
			} else if len(reason) > 0 {
				duration := time.Since(start)
				promptDuration := int64(math.Round(float64(int64(duration) / 4)))
				resBody = &chatCompleteResponse{
					chatResponse: chatResponse{
						Model:     input.Model,
						CreatedAt: time.Now().UTC().Format(time.RFC3339),
						Message: message{
							Role:    "assistant",
							Content: content,
						},
						Done: true,
					},
					DoneReason:         strings.ToLower(reason),
					TotalDuration:      int64(duration),
					LoadDuration:       0,
					PromptEvalCount:    promptTokens,
					PromptEvalDuration: promptDuration,
					EvalCount:          contentTokens,
					EvalDuration:       int64(duration) - promptDuration,
				}
				final = true
			} else {
				resBody = &chatResponse{
					Model:     input.Model,
					CreatedAt: time.Now().UTC().Format(time.RFC3339),
					Message: message{
						Role:    "assistant",
						Content: content,
					},
					Done: false,
				}
			}
			if err = json.NewEncoder(w).Encode(resBody); err != nil {
				log.Dbg("! encoding response body failed: %v", err)
			}
			if final {
				break
			}
		}
	} else {
		urlSuffix, reqBody, output, err := handler.prepareBody(&input)
		if err != nil {
			return wrongInput(w, err.Error())
		}
		status, duration, err := forwardRequest(urlSuffix, reqBody, output)
		if err != nil {
			return failRequest(w, status, err.Error())
		}
		content, reason, promptTokens, contentTokens := handler.extractCompleteResponse(output)
		tokens := promptTokens + contentTokens
		if log.IsDbg {
			log.Dbg("< answer by %s with %d character%s and %d token%s", input.Model,
				len(content), log.GetPlural(len(content)), tokens, log.GetPlural(tokens))
		}
		promptDuration := int64(math.Round(float64(int64(duration) / 4)))
		resBody := &chatCompleteResponse{
			chatResponse: chatResponse{
				Model:     input.Model,
				CreatedAt: time.Now().UTC().Format(time.RFC3339),
				Message: message{
					Role:    "assistant",
					Content: content,
				},
				Done: true,
			},
			DoneReason:         strings.ToLower(reason),
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
	}
	return http.StatusOK
}
