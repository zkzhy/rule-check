package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type PathsConfig struct {
	StateDir   string `json:"state_dir"`
	OutputFile string `json:"output_file"`
}

type YuhengConfig struct {
	BaseURL        string            `json:"base_url"`
	VerifySSL      bool              `json:"verify_ssl"`
	TimeoutS       float64           `json:"timeout_s"`
	Username       string            `json:"username"`
	Password       string            `json:"password"`
	ListEndpoint   string            `json:"list_endpoint"`
	ListMethod     string            `json:"list_method"`
	ListPageSize   int               `json:"list_page_size"`
	ListFilters    map[string]any    `json:"list_filters"`
	ListTimeFields map[string]string `json:"list_time_fields"`
	ListSendStyle  string            `json:"list_send_style"`
}

type AIConfig struct {
	Provider     string          `json:"provider"`
	Model        string          `json:"model"`
	TimeoutS     float64         `json:"timeout_s"`
	BaseURL      string          `json:"base_url"`
	PromptPath   string          `json:"prompt_path"`
	Concurrency  int             `json:"concurrency"`
	RateLimitQPS int             `json:"rate_limit_qps"`
	Context      AIContextConfig `json:"context"`
	ATTCK        AIAttckConfig   `json:"attck"`
	APIKey       string          `json:"-"`
}

type AIContextConfig struct {
	TotalMaxRunes       int `json:"total_max_runes"`
	NameMaxRunes        int `json:"name_max_runes"`
	DescriptionMaxRunes int `json:"description_max_runes"`
	POCMaxRunes         int `json:"poc_max_runes"`
	ReqMaxRunes         int `json:"req_max_runes"`
	RespMaxRunes        int `json:"resp_max_runes"`
}

type AIAttckConfig struct {
	CSVPath            string   `json:"csv_path"`
	TacticAllowlist    []string `json:"tactic_allowlist"`
	TechniqueTopK      int      `json:"technique_top_k"`
	CandidateMaxRunes  int      `json:"candidate_max_runes"`
	SubMaxPerTechnique int      `json:"sub_max_per_technique"`
}

type RootConfig struct {
	Paths  PathsConfig  `json:"paths"`
	Yuheng YuhengConfig `json:"yuheng"`
	AI     AIConfig     `json:"ai"`
}

func (c *RootConfig) StateDir() string {
	if c == nil {
		return "data"
	}
	s := strings.TrimSpace(c.Paths.StateDir)
	if s == "" {
		return "data"
	}
	return s
}

func (c *RootConfig) PendingAuditsPath() string {
	return filepath.Join(c.StateDir(), "pending_audits.jsonl")
}

func (c *RootConfig) PendingAuditsResultsPath() string {
	return filepath.Join(c.StateDir(), "pending_audits_results.jsonl")
}

func (c *RootConfig) SubmittedIDsPath() string {
	return filepath.Join(c.StateDir(), "submitted_ids.jsonl")
}

func Load() (*RootConfig, error) {
	appPath := os.Getenv("YH_CONFIG")
	if appPath == "" {
		appPath = "config/app.json"
	}
	secretsPath := os.Getenv("YH_SECRETS")
	if secretsPath == "" {
		secretsPath = "config/secrets.local.json"
	}

	base, err := readJSON[RootConfig](appPath)
	if err != nil {
		return nil, err
	}
	secrets, err := readRawJSON(secretsPath)
	if err != nil {
		return nil, err
	}

	var aiSecrets map[string]any
	if y, ok := secrets["yuheng"].(map[string]any); ok {
		if v, ok := y["username"].(string); ok {
			base.Yuheng.Username = v
		}
		if v, ok := y["password"].(string); ok {
			base.Yuheng.Password = v
		}
	}
	if a, ok := secrets["ai"].(map[string]any); ok {
		aiSecrets = a
	}

	if base.Paths.StateDir == "" {
		base.Paths.StateDir = "data" // Simplified state dir for Go workflow
	}
	if base.Paths.OutputFile == "" {
		base.Paths.OutputFile = "data/pending_audits.jsonl"
	}

	if base.Yuheng.ListEndpoint == "" {
		base.Yuheng.ListEndpoint = "/api/lines/operation"
	}
	if base.Yuheng.ListMethod == "" {
		base.Yuheng.ListMethod = "GET"
	}
	if base.Yuheng.ListSendStyle == "" {
		base.Yuheng.ListSendStyle = "query"
	}
	if base.Yuheng.ListPageSize <= 0 {
		base.Yuheng.ListPageSize = 1000
	}

	if base.AI.Provider == "" {
		base.AI.Provider = "doubao-ai"
	}
	if base.AI.Concurrency <= 0 {
		base.AI.Concurrency = 1
	}

	if base.AI.Context.TotalMaxRunes <= 0 {
		base.AI.Context.TotalMaxRunes = 2600
	}
	if base.AI.Context.NameMaxRunes <= 0 {
		base.AI.Context.NameMaxRunes = 200
	}
	if base.AI.Context.DescriptionMaxRunes <= 0 {
		base.AI.Context.DescriptionMaxRunes = 1200
	}
	if base.AI.Context.POCMaxRunes <= 0 {
		base.AI.Context.POCMaxRunes = 1000
	}
	if base.AI.Context.ReqMaxRunes <= 0 {
		base.AI.Context.ReqMaxRunes = 800
	}
	if base.AI.Context.RespMaxRunes <= 0 {
		base.AI.Context.RespMaxRunes = 800
	}

	if base.AI.ATTCK.TechniqueTopK <= 0 {
		base.AI.ATTCK.TechniqueTopK = 30
	}
	if base.AI.ATTCK.CandidateMaxRunes <= 0 {
		base.AI.ATTCK.CandidateMaxRunes = 1800
	}
	if base.AI.ATTCK.SubMaxPerTechnique <= 0 {
		base.AI.ATTCK.SubMaxPerTechnique = 8
	}

	if p := os.Getenv("AI_PROVIDER"); p != "" {
		base.AI.Provider = p
	}
	if p := os.Getenv("AI_MODEL"); p != "" {
		base.AI.Model = p
	}
	if p := os.Getenv("AI_BASE_URL"); p != "" {
		base.AI.BaseURL = p
	}
	if p := os.Getenv("AI_PROMPT_PATH"); p != "" {
		base.AI.PromptPath = p
	}
	if p := os.Getenv("AI_CONCURRENCY"); p != "" {
		if v, err := strconv.Atoi(strings.TrimSpace(p)); err == nil && v > 0 {
			base.AI.Concurrency = v
		}
	}
	if p := os.Getenv("AI_RATE_LIMIT_QPS"); p != "" {
		if v, err := strconv.Atoi(strings.TrimSpace(p)); err == nil && v >= 0 {
			base.AI.RateLimitQPS = v
		}
	}

	if p := os.Getenv("AI_API_KEY"); p != "" {
		base.AI.APIKey = p
	} else if v := resolveAPIKeyFromSecrets(aiSecrets, base.AI.Provider); v != "" {
		base.AI.APIKey = v
	}

	return &base, nil
}

func resolveAPIKeyFromSecrets(aiSecrets map[string]any, provider string) string {
	if aiSecrets == nil {
		return ""
	}

	normalizedProvider := normalizeAIProvider(provider)

	if keysAny, ok := aiSecrets["api_keys"]; ok {
		if keys, ok := keysAny.(map[string]any); ok {
			if v, ok := keys[normalizedProvider].(string); ok && strings.TrimSpace(v) != "" {
				return v
			}
			if v, ok := keys[provider].(string); ok && strings.TrimSpace(v) != "" {
				return v
			}
		}
	}

	suffix := strings.NewReplacer("-", "_", " ", "_").Replace(normalizedProvider)
	if v, ok := aiSecrets["api_key_"+suffix].(string); ok && strings.TrimSpace(v) != "" {
		return v
	}

	if v, ok := aiSecrets["api_key"].(string); ok && strings.TrimSpace(v) != "" {
		return v
	}
	return ""
}

func normalizeAIProvider(p string) string {
	p = strings.ToLower(strings.TrimSpace(p))
	switch p {
	case "":
		return ""
	case "ark", "doubao", "doubao-ai":
		return "doubao-ai"
	case "openai", "openai_compat", "openai-compatible":
		return "openai-compatible"
	case "chaitin":
		return "chaitin"
	default:
		return p
	}
}

func readJSON[T any](path string) (T, error) {
	var zero T
	data, err := os.ReadFile(path)
	if err != nil {
		return zero, err
	}
	if len(data) == 0 {
		return zero, errors.New("empty config file")
	}
	if err := json.Unmarshal(data, &zero); err != nil {
		return zero, err
	}
	return zero, nil
}

func readRawJSON(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return map[string]any{}, nil
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if m == nil {
		m = map[string]any{}
	}
	return m, nil
}
