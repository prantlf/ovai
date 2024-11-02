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

func prepareChatBody(input *chatInput) (string, interface{}, interface{}, error) {
	urlPrefix := input.Model + ":generateContent"
	body, err := createChatGeminiBody(input)
	if err != nil {
		return "", nil, nil, err
	}
	return urlPrefix, body, &geminiCompleteOutput{}, nil
}

func prepareChatStream(input *chatInput) (string, interface{}, interface{}, interface{}, error) {
	urlPrefix := input.Model + ":streamGenerateContent?alt=sse"
	body, err := createChatGeminiBody(input)
	if err != nil {
		return "", nil, nil, nil, err
	}
	return urlPrefix, body, &geminiPartialOutput{}, &geminiFinalOutput{}, nil
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

	var forward bool
	if strings.HasPrefix(input.Model, "gemini") {
		forward = true
	} else if !canProxy {
		return wrongInput(w, fmt.Sprintf("unrecognised model %q", input.Model))
	}
	if log.IsDbg {
		log.Dbg("> ask with %d message%s using %s", len(input.Messages),
			log.GetPlural(len(input.Messages)), input.Model)
	}

	if !forward {
		if input.Stream {
			return proxyStream("chat", reqPayload, w, "answer", input.Model)
		} else {
			return proxyRequest("chat", reqPayload, w, "answer", input.Model)
		}
	}

	if input.Stream {
		urlSuffix, reqBody, partialOutput, finalOutput, err := prepareChatStream(&input)
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
			var reader io.Reader
			if len(rest) > 0 {
				reader = bytes.NewReader(rest)
			} else {
				reader = resReader
			}
			content, reason, rest, promptTokens, contentTokens, err = extractStreamGeminiResponse(reader, partialOutput, finalOutput)
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
		urlSuffix, reqBody, output, err := prepareChatBody(&input)
		if err != nil {
			return wrongInput(w, err.Error())
		}
		status, duration, err := forwardRequest(urlSuffix, reqBody, output)
		if err != nil {
			return failRequest(w, status, err.Error())
		}
		content, reason, promptTokens, contentTokens := extractCompleteGeminiResponse(output)
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
