package prompt

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"audit-workflow/internal/config"

	einoprompt "github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

type promptConfig struct {
	Template string   `json:"template"`
	Sections []string `json:"sections"`
}

type ChatTemplate interface {
	Format(ctx context.Context, vars map[string]any, opts ...einoprompt.Option) ([]*schema.Message, error)
}

func BuildRiskTemplate(cfg *config.RootConfig) (ChatTemplate, error) {
	templateStr, err := loadPromptTemplate(cfg)
	if err != nil {
		return nil, err
	}
	return buildTemplateFromString(templateStr), nil
}

func BuildATTCKTacticTemplate() ChatTemplate {
	return buildTemplateFromString(
		"你将收到一条 HTTP 漏洞记录的精简上下文，以及候选战术列表。\n" +
			"你必须从候选列表中选择一个最匹配的战术名称，并输出严格 JSON。\n\n" +
			"漏洞上下文：\n{context}\n\n" +
			"候选战术列表（只能从中选择，不要输出数值 ID）：\n{tactic_candidates}\n\n" +
			"输出格式（严格 JSON，不要 Markdown）：\n" +
			"{\n" +
			"  \"tactic_name\": \"<string>\"\n" +
			"}\n",
	)
}

func buildTemplateFromString(templateStr string) ChatTemplate {
	templateStr = strings.ReplaceAll(templateStr, "{{#context#}}", "{context}")
	templateStr = strings.ReplaceAll(templateStr, "{#context#}", "{context}")

	vars := []string{
		"context",
		"description",
		"name",
		"tactic_candidates",
		"tactic_name_selected",
		"technique_candidates",
	}
	for _, v := range vars {
		templateStr = strings.ReplaceAll(templateStr, "{"+v+"}", "__VAR_"+v+"__")
	}

	templateStr = strings.ReplaceAll(templateStr, "{", "{{")
	templateStr = strings.ReplaceAll(templateStr, "}", "}}")

	for _, v := range vars {
		templateStr = strings.ReplaceAll(templateStr, "__VAR_"+v+"__", "{"+v+"}")
	}

	return einoprompt.FromMessages(schema.FString, schema.UserMessage(templateStr))
}

func loadPromptTemplate(cfg *config.RootConfig) (string, error) {
	pathsToCheck := []string{
		os.Getenv("AI_PROMPT_PATH"),
		cfg.AI.PromptPath,
		"internal/components/prompts/risk.json",
	}

	for _, p := range pathsToCheck {
		if p == "" {
			continue
		}
		content, err := os.ReadFile(p)
		if err != nil || len(content) == 0 {
			continue
		}

		var pc promptConfig
		if jsonErr := json.Unmarshal(content, &pc); jsonErr == nil {
			if pc.Template != "" {
				return pc.Template, nil
			}
			if len(pc.Sections) > 0 {
				return strings.Join(pc.Sections, "\n"), nil
			}
		}
		return string(content), nil
	}
	return "", fmt.Errorf("no valid prompt template found")
}
