package cfg

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"os"

	"github.com/prantlf/ovai/internal/log"
)

type ThinkingConfig struct {
	IncludeThoughts bool `json:"includeThoughts"`
	ThinkingBudget  *int `json:"thinkingBudget,omitempty"` // flash: 0-2456, pro: 128-32768
}

type GenerationConfig struct {
	MaxOutputTokens *int           `json:"maxOutputTokens,omitempty"`
	Temperature     *float64       `json:"temperature,omitempty"`
	TopP            *float64       `json:"topP,omitempty"`
	TopK            *int           `json:"topK,omitempty"`
	Scope           string         `json:"scope,omitempty"`
	ThinkingConfig  ThinkingConfig `json:"thinkingConfig,omitempty"`
}

type SafetySetting struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

type geminiDefaults struct {
	GenerationConfig GenerationConfig `json:"generationConfig"`
	SafetySettings   []SafetySetting  `json:"safetySettings"`
}

type defaults struct {
	ApiLocation    string         `json:"apiLocation"`
	ApiEndpoint    string         `json:"apiEndpoint"`
	GeminiDefaults geminiDefaults `json:"geminiDefaults"`
}

//go:embed model-defaults.json
var builtins []byte

var Defaults = readDefaults()

func mergeParameters(target *GenerationConfig, source *GenerationConfig) {
	if source.MaxOutputTokens != nil {
		target.MaxOutputTokens = source.MaxOutputTokens
	}
	if source.Temperature != nil {
		target.Temperature = source.Temperature
	}
	if source.TopP != nil {
		target.TopP = source.TopP
	}
	if source.TopK != nil {
		target.TopK = source.TopK
	}
	if len(source.Scope) > 0 {
		target.Scope = source.Scope
	}
	if source.ThinkingConfig.ThinkingBudget != nil {
		target.ThinkingConfig.ThinkingBudget = source.ThinkingConfig.ThinkingBudget
	}
}

func mergeDefaults(target *defaults, source *defaults) {
	if len(source.ApiLocation) > 0 {
		target.ApiLocation = source.ApiLocation
	}
	if len(source.ApiEndpoint) > 0 {
		target.ApiEndpoint = source.ApiEndpoint
	}
	mergeParameters(&target.GeminiDefaults.GenerationConfig, &source.GeminiDefaults.GenerationConfig)
	if len(source.GeminiDefaults.SafetySettings) > 0 {
		target.GeminiDefaults.SafetySettings = source.GeminiDefaults.SafetySettings
	}
}

func readDefaults() *defaults {
	defaultsFile := os.Getenv("OVAI_DEFAULTS")
	if len(defaultsFile) == 0 {
		defaultsFile = "model-defaults.json"
	}
	var deflts defaults
	if err := json.Unmarshal(builtins, &deflts); err != nil {
		log.Ftl("decoding built-in defaults failed: %v", err)
	}
	log.Dbg("read %s", defaultsFile)
	defaultsJson, err := os.ReadFile(defaultsFile)
	if err != nil {
		log.Dbg("reading %s failed: %v", defaultsFile, err)
	} else {
		over := defaults{
			GeminiDefaults: geminiDefaults{
				GenerationConfig: GenerationConfig{},
			},
		}
		if err := json.Unmarshal(defaultsJson, &over); err != nil {
			log.Ftl("decoding %s failed: %v", defaultsFile, err)
		}
		mergeDefaults(&deflts, &over)
		if log.IsDbg {
			var overJson bytes.Buffer
			if errLog := json.Indent(&overJson, defaultsJson, "", "  "); errLog != nil {
				log.Dbg("override defaults: %s", defaultsJson)
			} else {
				log.Dbg("override defaults: %s", overJson.Bytes())
			}
			defltsJson, errLog := json.MarshalIndent(deflts, "", "  ")
			if errLog != nil {
				log.Dbg("customised defaults: %+v", deflts)
			} else {
				log.Dbg("customised defaults: %s", defltsJson)
			}
		}
	}
	return &deflts
}
