package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/prantlf/ovai/internal/cfg"
	"github.com/prantlf/ovai/internal/log"
)

type streamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type completionsInput struct {
	Model               string               `json:"model"`
	Messages            []completionsMessage `json:"messages"`
	Tools               []FunctionTool       `json:"tools"`
	Stream              bool                 `json:"stream"`
	StreamOptions       streamOptions        `json:"stream_options"`
	ReasoningEffort     string               `json:"reasoning_effort"`
	MaxCompletionTokens *int                 `json:"max_completion_tokens"`
	MaxTokens           *int                 `json:"max_tokens"`
	Temperature         *float64             `json:"temperature"`
	TopP                *float64             `json:"top_p"`
	ThinkingBudget      *int                 `json:"thinking_budget,omitempty"`
}

type imageUrl struct {
	URL string `json:"url"`
}

type completionsContent struct {
	Type     string   `json:"type"`
	Text     string   `json:"text"`
	ImageUrl imageUrl `json:"image_url"`
}

type completionsContentArray []completionsContent

type completionsMessage struct {
	Content    completionsContentArray `json:"content"`
	Role       string                  `json:"role"`
	Name       string                  `json:"name,omitempty"`
	ToolCallId string                  `json:"tool_call_id,omitempty"`
}

func (t *completionsContentArray) UnmarshalJSON(data []byte) error {
	if len(data) > 1 && data[0] == '[' {
		var array []completionsContent
		if err := json.Unmarshal(data, &array); err != nil {
			return err
		}
		*t = array
	} else {
		var scalar string
		if err := json.Unmarshal(data, &scalar); err != nil {
			return err
		}
		array := make([]completionsContent, 1)
		array[0] = completionsContent{
			Type: "text",
			Text: scalar,
		}
		*t = array
	}
	return nil
}

func convertCompletionsContentToGeminiParts(contents []completionsContent) ([]geminiPart, error) {
	parts := []geminiPart{}
	for _, content := range contents {
		var part geminiPart
		switch content.Type {
		case "text":
			part = geminiPart{
				Text: content.Text,
			}
		case "image_url":
			url := content.ImageUrl.URL
			if !strings.HasPrefix(content.ImageUrl.URL, "data:") {
				return nil, fmt.Errorf("invalid data URI prefix: %s", url[:5])
			}
			semicolon := strings.IndexByte(url, ';')
			if semicolon > 0 {
				return nil, fmt.Errorf("missing comma in data URI: %s", url[:5])
			}
			mimeType := url[5:semicolon]
			if !strings.HasPrefix(mimeType, "image/") {
				return nil, fmt.Errorf("invalid image type: %s", mimeType)
			}
			comma := strings.IndexByte(url, ',')
			if comma > 0 {
				return nil, fmt.Errorf("missing comma in data URI: %s", url[:5])
			}
			encoding := url[semicolon:comma]
			if encoding != "base64" {
				return nil, fmt.Errorf("invalid image encoding: %s", encoding)
			}
			image := url[comma:]
			part = geminiPart{
				InlineData: &inlineData{
					MimeType: mimeType,
					Data:     image,
				},
			}
		default:
			return nil, fmt.Errorf("invalid content type: %q", content.Type)
		}
		parts = append(parts, part)
	}
	return parts, nil
}

func convertCompletionsToolContentToGeminiParts(contents []completionsContent, toolName string) ([]geminiPart, error) {
	parts := []geminiPart{}
	var builder strings.Builder
	for i, content := range contents {
		if content.Type != "text" {
			return nil, fmt.Errorf("invalid content type of tool result: %s", content.Type)
		}
		if i > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(content.Text)
	}
	part := geminiPart{
		FunctionResponse: &functionResponse{
			Name: toolName,
			Response: map[string]string{
				"result": builder.String(),
			},
		},
	}
	parts = append(parts, part)
	return parts, nil
}

func convertCompletionsMessagesToGemini(messages []completionsMessage) ([]geminiContent, error) {
	systemMessages := make([]geminiPart, 0, 1)
	chatMessages := make([]geminiContent, 0, len(messages))
	for _, msg := range messages {
		if msg.Role == "system" || msg.Role == "developer" {
			if parts, err := convertCompletionsContentToGeminiParts(msg.Content); err == nil {
				systemMessages = append(systemMessages, parts...)
			} else {
				return nil, err
			}
		} else {
			var role string
			switch msg.Role {
			case "user":
				role = "user"
			case "assistant":
				role = "model"
			case "tool":
				role = "user"
			default:
				return nil, fmt.Errorf("invalid chat message role: %q", msg.Role)
			}
			var parts []geminiPart
			var err error
			if len(msg.ToolCallId) > 0 {
				parts, err = convertCompletionsToolContentToGeminiParts(msg.Content, msg.ToolCallId)
			} else {
				parts, err = convertCompletionsContentToGeminiParts(msg.Content)
			}
			if err != nil {
				return nil, err
			}
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

func mergeCompletionsParameters(target *cfg.GenerationConfig, source *completionsInput) error {
	if source.MaxTokens != nil {
		target.MaxOutputTokens = source.MaxTokens
	}
	if source.MaxCompletionTokens != nil {
		target.MaxOutputTokens = source.MaxCompletionTokens
	}
	if source.Temperature != nil {
		target.Temperature = source.Temperature
	}
	if source.TopP != nil {
		target.TopP = source.TopP
	}
	if len(source.ReasoningEffort) > 0 {
		var think thinkLevel
		if source.ReasoningEffort == "minimal" {
			think = "none"
		} else {
			think = thinkLevel(source.ReasoningEffort)
		}
		var thoughts bool
		if think != "none" {
			thoughts = true
		} else {
			if strings.HasPrefix(source.Model, "gemini-2.5-pro") {
				thoughts = true
			} else {
				thoughts = false
			}
		}
		target.ThinkingConfig = cfg.ThinkingConfig{
			IncludeThoughts: thoughts,
		}
		var thinkingBudget int
		if source.ThinkingBudget != nil {
			thinkingBudget = *source.ThinkingBudget
		} else {
			var err error
			thinkingBudget, err = GetThinkingBudget(source.Model, think)
			if err != nil {
				return err
			}
		}
		target.ThinkingConfig.ThinkingBudget = &thinkingBudget
	}
	return nil
}

func convertCompletionsBodyToGemini(input *completionsInput) (interface{}, error) {
	chatMessages, err := convertCompletionsMessagesToGemini(input.Messages)
	if err != nil {
		return nil, err
	}
	generationConfig := cfg.Defaults.GeminiDefaults.GenerationConfig
	if err := mergeCompletionsParameters(&generationConfig, input); err != nil {
		return nil, err
	}
	tools := convertToolsToGemini(input.Tools)
	body := &geminiBody{
		Contents:         chatMessages,
		GenerationConfig: generationConfig,
		SafetySettings:   cfg.Defaults.GeminiDefaults.SafetySettings,
		Tools:            tools,
	}
	return body, nil
}

func prepareCompletionsBody(input *completionsInput) (string, interface{}, interface{}, error) {
	urlPrefix := input.Model + ":generateContent"
	body, err := convertCompletionsBodyToGemini(input)
	if err != nil {
		return "", nil, nil, err
	}
	return urlPrefix, body, &geminiCompleteOutput{}, nil
}

func prepareCompletionsStream(input *completionsInput) (string, interface{}, interface{}, interface{}, error) {
	urlPrefix := input.Model + ":streamGenerateContent?alt=sse"
	body, err := convertCompletionsBodyToGemini(input)
	if err != nil {
		return "", nil, nil, nil, err
	}
	return urlPrefix, body, &geminiPartialOutput{}, &geminiFinalOutput{}, nil
}

type outputMessage struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []toolCall `json:"tool_calls,omitempty"`
}

type deltaChoice struct {
	Index        int           `json:"index"`
	Delta        outputMessage `json:"delta"`
	FinishReason *string       `json:"finish_reason"`
}

type completeChoice struct {
	Index        int           `json:"index"`
	Message      outputMessage `json:"message"`
	FinishReason *string       `json:"finish_reason"`
}

type completionsResponse struct {
	Model             string `json:"model"`
	Created           int64  `json:"created"`
	ID                string `json:"id"`
	Object            string `json:"object"`
	SystemFingerprint string `json:"system_fingerprint"`
}

type completionsUsage struct {
	CompletionTokens int `json:"completion_tokens"`
	PromptTokens     int `json:"prompt_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type completionsCompleteResponse struct {
	completionsResponse
	Choices []completeChoice `json:"choices"`
	Usage   completionsUsage `json:"usage"`
}

type completionsDeltaResponse struct {
	completionsResponse
	Choices []deltaChoice `json:"choices"`
}

func createCompletionsResponse(model string, chunked bool) completionsResponse {
	now := time.Now().UTC()
	var object string
	if chunked {
		object = "chat.completion.chunk"
	} else {
		object = "chat.completion"
	}
	return completionsResponse{
		Model:             model,
		Created:           now.Unix(),
		ID:                now.Format(time.RFC3339),
		Object:            object,
		SystemFingerprint: "fp_gemini",
	}
}

func HandleCompletions(w http.ResponseWriter, r *http.Request) int {
	input := completionsInput{
		ReasoningEffort: "medium",
		Stream:          false,
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
			return proxyStream("chat/completions", reqPayload, w, "answer", input.Model)
		}
		return proxyRequest("chat/completions", reqPayload, w, "answer", input.Model)
	}

	if input.Stream {
		urlSuffix, reqBody, partialOutput, finalOutput, err := prepareCompletionsStream(&input)
		if err != nil {
			return wrongInput(w, err.Error())
		}
		status, _, resReader, err := forwardStream(urlSuffix, reqBody)
		if err != nil {
			return failRequest(w, status, err.Error())
		}
		defer func() {
			if err := resReader.Close(); err != nil {
				log.Dbg("closing response body stream failed: %v", err)
			}
		}()
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(status)
		var rest []byte
		for {
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			var content string
			var functionCalls []functionCall
			var reason string
			var promptTokens int
			var contentTokens int
			var reader io.Reader
			if len(rest) > 0 {
				reader = bytes.NewReader(rest)
			} else {
				reader = resReader
			}
			_, content, functionCalls, reason, rest, promptTokens, contentTokens, err = extractStreamGeminiResponse(reader, partialOutput, finalOutput)
			var resBody any
			final := false
			if err != nil {
				break
			}
			toolCalls := convertFunctionCallsToToolCalls(functionCalls)
			final = len(reason) > 0
			var outputReason *string
			if final {
				stringReason := strings.ToLower(reason)
				outputReason = &stringReason
			}
			resBody = &completionsDeltaResponse{
				completionsResponse: createCompletionsResponse(input.Model, true),
				Choices: []deltaChoice{
					{
						Delta: outputMessage{
							Role:      "assistant",
							Content:   content,
							ToolCalls: toolCalls,
						},
						FinishReason: outputReason,
					},
				},
			}
			w.Write([]byte("data: "))
			if err = json.NewEncoder(w).Encode(resBody); err != nil {
				log.Dbg("! encoding response body failed: %v", err)
			}
			w.Write([]byte("\n\n"))
			if final {
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
				if input.StreamOptions.IncludeUsage {
					resBody = &completionsCompleteResponse{
						completionsResponse: createCompletionsResponse(input.Model, true),
						Choices:             []completeChoice{},
						Usage: completionsUsage{
							CompletionTokens: contentTokens,
							PromptTokens:     promptTokens,
							TotalTokens:      promptTokens + contentTokens,
						},
					}
					w.Write([]byte("data: "))
					if err = json.NewEncoder(w).Encode(resBody); err != nil {
						log.Dbg("! encoding response body failed: %v", err)
					}
					w.Write([]byte("\n\n"))
					if f, ok := w.(http.Flusher); ok {
						f.Flush()
					}
				}
				w.Write([]byte("data: [DONE]\n\n"))
				break
			}
		}
	} else {
		urlSuffix, reqBody, output, err := prepareCompletionsBody(&input)
		if err != nil {
			return wrongInput(w, err.Error())
		}
		status, _, err := forwardRequest(urlSuffix, reqBody, output)
		if err != nil {
			return failRequest(w, status, err.Error())
		}
		_, content, functionCalls, reason, promptTokens, contentTokens := extractCompleteGeminiResponse(output)
		tokens := promptTokens + contentTokens
		if log.IsDbg {
			log.Dbg("< answer by %s with %d character%s and %d token%s", input.Model,
				len(content), log.GetPlural(len(content)), tokens, log.GetPlural(tokens))
		}
		toolCalls := convertFunctionCallsToToolCalls(functionCalls)
		outputReason := strings.ToLower(reason)
		resBody := &completionsCompleteResponse{
			completionsResponse: createCompletionsResponse(input.Model, false),
			Choices: []completeChoice{
				{
					Message: outputMessage{
						Role:      "assistant",
						Content:   content,
						ToolCalls: toolCalls,
					},
					FinishReason: &outputReason,
				},
			},
			Usage: completionsUsage{
				CompletionTokens: contentTokens,
				PromptTokens:     promptTokens,
				TotalTokens:      promptTokens + contentTokens,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if err = json.NewEncoder(w).Encode(resBody); err != nil {
			log.Dbg("! encoding response body failed: %v", err)
		}
	}
	return http.StatusOK
}
