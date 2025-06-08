package routes

import (
	"bytes"
	"encoding/base64"
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

type modelParameters struct {
	MaxOutputTokens int     `json:"num_predict"`
	Temperature     float64 `json:"temperature"`
	TopP            float64 `json:"top_p"`
	TopK            int     `json:"top_k"`
	ThinkingBudget  int     `json:"thinking_budget"`
}

type generateInput struct {
	Model   string          `json:"model"`
	Prompt  string          `json:"prompt"`
	Images  []string        `json:"images"`
	Think   bool            `json:"think"`
	Stream  bool            `json:"stream"`
	Options modelParameters `json:"options"`
}

type inlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type geminiPart struct {
	Text       string      `json:"text,omitempty"`
	Thought    bool        `json:"thought,omitempty"`
	InlineData *inlineData `json:"inlineData,omitempty"`
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

type safetyRating struct {
	Category         string  `json:"category"`
	Probability      string  `json:"probability"`
	ProbabilityScore float64 `json:"probabilityScore"`
	Severity         string  `json:"severity"`
	SeverityScore    float64 `json:"severityScore"`
}

type geminiCompleteCandidate struct {
	geminiCandidate
	FinishReason  string         `json:"finishReason"`
	SafetyRatings []safetyRating `json:"safetyRatings"`
	AvgLogProbs   float64        `json:"avgLogprobs"`
}

type geminiPartialCandidate struct {
	geminiCandidate
	SafetyRatings []safetyRating `json:"safetyRatings"`
}

type geminiFinalCandidate struct {
	geminiCandidate
	FinishReason string `json:"finishReason"`
}

type geminiMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

type geminiOutput struct {
	ModelVersion string `json:"modelVersion"`
}

type geminiCompleteOutput struct {
	geminiOutput
	Candidates    []geminiCompleteCandidate `json:"candidates"`
	UsageMetadata geminiMetadata            `json:"usageMetadata"`
}

type geminiPartialOutput struct {
	geminiOutput
	Candidates []geminiPartialCandidate `json:"candidates"`
}

type geminiFinalOutput struct {
	geminiOutput
	Candidates    []geminiFinalCandidate `json:"candidates"`
	UsageMetadata geminiMetadata         `json:"usageMetadata"`
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
	if source.ThinkingBudget > 0 {
		target.ThinkingConfig.ThinkingBudget = source.ThinkingBudget
	}
}

func createGeminiParts(content string, images []string) ([]geminiPart, error) {
	parts := []geminiPart{
		{
			Text: content,
		},
	}
	for _, image := range images {
		bytes, err := base64.StdEncoding.DecodeString(image)
		if err != nil {
			return nil, fmt.Errorf("invalid image encoding: %s", err.Error())
		}
		mimeType := http.DetectContentType(bytes)
		if !strings.HasPrefix(mimeType, "image/") {
			return nil, fmt.Errorf("invalid image type: %s", mimeType)
		}
		part := geminiPart{
			InlineData: &inlineData{
				MimeType: mimeType,
				Data:     image,
			},
		}
		parts = append(parts, part)
	}
	return parts, nil
}

func createGenerateGeminiBody(input *generateInput) (interface{}, error) {
	generationConfig := cfg.Defaults.GeminiDefaults.GenerationConfig
	if input.Think {
		generationConfig.ThinkingConfig = cfg.ThinkingConfig{
			IncludeThoughts: true,
		}
	}
	mergeParameters(&generationConfig, &input.Options)
	parts, err := createGeminiParts(input.Prompt, input.Images)
	if err != nil {
		return nil, err
	}
	body := &geminiBody{
		Contents: []geminiContent{
			{
				Role:  "user",
				Parts: parts,
			},
		},
		GenerationConfig: generationConfig,
		SafetySettings:   cfg.Defaults.GeminiDefaults.SafetySettings,
	}
	return body, nil
}

func prepareGenerateBody(input *generateInput) (string, interface{}, interface{}, error) {
	urlPrefix := input.Model + ":generateContent"
	body, err := createGenerateGeminiBody(input)
	if err != nil {
		return "", nil, nil, err
	}
	return urlPrefix, body, &geminiCompleteOutput{}, nil
}

func prepareGenerateStream(input *generateInput) (string, interface{}, interface{}, interface{}, error) {
	urlPrefix := input.Model + ":streamGenerateContent?alt=sse"
	body, err := createGenerateGeminiBody(input)
	if err != nil {
		return "", nil, nil, nil, err
	}
	return urlPrefix, body, &geminiPartialOutput{}, &geminiFinalOutput{}, nil
}

func extractCompleteGeminiResponse(data interface{}) (string, string, string, int, int) {
	output, ok := data.(*geminiCompleteOutput)
	if !ok {
		log.Ftl("invalid gemini complete response type")
	}
	thoughts := ""
	answer := ""
	reason := ""
	if len(output.Candidates) > 0 {
		candidate := output.Candidates[0]
		parts := candidate.Content.Parts
		for _, part := range parts {
			if part.Thought {
				thoughts += part.Text
			} else {
				answer += part.Text
			}
		}
		reason = candidate.FinishReason
	}
	metadata := output.UsageMetadata
	return thoughts, answer, reason, metadata.PromptTokenCount, metadata.CandidatesTokenCount
}

func extractStreamGeminiResponse(resReader io.Reader, partialData interface{}, finalData interface{}) (string, string, string, []byte, int, int, error) {
	buf := make([]byte, 1024*1024)
	size, err := resReader.Read(buf)
	if err == io.EOF {
		if size == 0 {
			log.Dbg("response body stream ended unexpectedly")
			return "", "", "", nil, 0, 0, errors.New("response body stream ended unexpectedly")
		} else {
			if log.IsDbg {
				log.Dbg("< %d byte%s and EOF", size, log.GetPlural(size))
			}
		}
	} else if err != nil {
		log.Dbg("reading response body stream failed: %v", err)
		return "", "", "", nil, 0, 0, fmt.Errorf("reading response body stream failed: %v", err)
	} else {
		if log.IsDbg {
			log.Dbg("< %d byte%s", size, log.GetPlural(size))
		}
	}
	resBody := bytes.TrimSpace(buf[0:size])
	if prefixPos := bytes.Index(resBody, []byte("data: ")); prefixPos == 0 {
		resBody = resBody[6:]
	}
	var rest []byte
	if lineBreakPos := bytes.IndexByte(resBody, byte('\n')); lineBreakPos >= 0 {
		lineBreakPos++
		rest = bytes.TrimSpace(resBody[lineBreakPos:])
		resBody = resBody[0:lineBreakPos]
	}
	final := true
	if err = json.Unmarshal(resBody, finalData); err != nil {
		final = false
		if err = json.Unmarshal(resBody, partialData); err != nil {
			log.Dbg("receive response %s", resBody)
			log.Dbg("decoding response body failed: %v", err)
			return "", "", "", rest, 0, 0, errors.New("decoding response body failed")
		}
	}
	if final {
		output, ok := finalData.(*geminiFinalOutput)
		if !ok {
			log.Ftl("invalid gemini final response type")
		}
		thoughts := ""
		answer := ""
		reason := ""
		if len(output.Candidates) > 0 {
			candidate := output.Candidates[0]
			parts := candidate.Content.Parts
			for _, part := range parts {
				if part.Thought {
					thoughts += part.Text
				} else {
					answer += part.Text
				}
			}
			reason = candidate.FinishReason
		}
		if len(reason) > 0 {
			metadata := output.UsageMetadata
			return thoughts, answer, reason, rest, metadata.PromptTokenCount, metadata.CandidatesTokenCount, nil
		}
	}
	if err = json.Unmarshal(resBody, partialData); err != nil {
		log.Dbg("receive response %s", resBody)
		log.Dbg("decoding response body failed: %v", err)
		return "", "", "", rest, 0, 0, errors.New("decoding response body failed")
	}
	if log.IsDbg {
		var resLog bytes.Buffer
		if errLog := json.Indent(&resLog, resBody, "", "  "); errLog != nil {
			log.Net("receive response %s", resBody)
			// log.Printf("receive response %+v", output)
		} else {
			log.Net("receive response %s", resLog.Bytes())
		}
	}
	output, ok := partialData.(*geminiPartialOutput)
	if !ok {
		log.Ftl("invalid gemini partial response type")
	}
	thoughts := ""
	answer := ""
	if len(output.Candidates) > 0 {
		candidate := output.Candidates[0]
		parts := candidate.Content.Parts
		for _, part := range parts {
			if part.Thought {
				thoughts += part.Text
			} else {
				answer += part.Text
			}
		}
	}
	return thoughts, answer, "", rest, 0, 0, nil
}

type tokenCount struct {
	TotalTokens int `json:"totalTokens"`
}

type tokenMetadata struct {
	InputTokenCount  tokenCount `json:"inputTokenCount"`
	OutputTokenCount tokenCount `json:"outputTokenCount"`
}

type generateResponse struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Thinking  string `json:"thinking,omitempty"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
}

type generateCompleteResponse struct {
	generateResponse
	DoneReason         string `json:"done_reason"`
	Context            []int  `json:"context"`
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
		Think:  false,
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
	if len(input.Prompt) == 0 {
		return wrongInput(w, "prompt missing")
	}

	var forward bool
	if strings.HasPrefix(input.Model, "gemini") {
		forward = true
	} else if !canProxy {
		return wrongInput(w, fmt.Sprintf("unrecognised model %q", input.Model))
	}
	if log.IsDbg {
		log.Dbg("> generate from %d character%s using %s", len(input.Prompt),
			log.GetPlural(len(input.Prompt)), input.Model)
	}

	if !forward {
		if input.Stream {
			return proxyStream("generate", reqPayload, w, "result", input.Model)
		} else {
			return proxyRequest("generate", reqPayload, w, "result", input.Model)
		}
	}

	if input.Stream {
		urlSuffix, reqBody, partialOutput, finalOutput, err := prepareGenerateStream(&input)
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
			var thinking string
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
			thinking, content, reason, rest, promptTokens, contentTokens, err = extractStreamGeminiResponse(reader, partialOutput, finalOutput)
			var resBody any
			final := false
			if err != nil {
				break
			} else if len(reason) > 0 {
				duration := time.Since(start)
				promptDuration := int64(math.Round(float64(int64(duration) / 4)))
				resBody = &generateCompleteResponse{
					generateResponse: generateResponse{
						Model:     input.Model,
						CreatedAt: time.Now().UTC().Format(time.RFC3339),
						Thinking:  thinking,
						Response:  content,
						Done:      true,
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
				resBody = &generateResponse{
					Model:     input.Model,
					CreatedAt: time.Now().UTC().Format(time.RFC3339),
					Thinking:  thinking,
					Response:  content,
					Done:      false,
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
		urlSuffix, reqBody, output, err := prepareGenerateBody(&input)
		if err != nil {
			return wrongInput(w, err.Error())
		}
		status, duration, err := forwardRequest(urlSuffix, reqBody, output)
		if err != nil {
			return failRequest(w, status, err.Error())
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		thinking, content, reason, promptTokens, contentTokens := extractCompleteGeminiResponse(output)
		tokens := promptTokens + contentTokens
		if log.IsDbg {
			log.Dbg("< result by %s with %d character%s and %d token%s", input.Model,
				len(content), log.GetPlural(len(content)), tokens, log.GetPlural(tokens))
		}
		promptDuration := int64(math.Round(float64(int64(duration) / 4)))
		resBody := &generateCompleteResponse{
			generateResponse: generateResponse{
				Model:     input.Model,
				CreatedAt: time.Now().UTC().Format(time.RFC3339),
				Thinking:  thinking,
				Response:  content,
				Done:      true,
			},
			DoneReason:         strings.ToLower(reason),
			TotalDuration:      int64(duration),
			LoadDuration:       0,
			PromptEvalCount:    promptTokens,
			PromptEvalDuration: promptDuration,
			EvalCount:          contentTokens,
			EvalDuration:       int64(duration) - promptDuration,
		}
		if err = json.NewEncoder(w).Encode(resBody); err != nil {
			log.Dbg("! encoding response body failed: %v", err)
		}
	}
	return http.StatusOK
}
