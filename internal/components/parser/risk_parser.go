package parser

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func ParseStructuredJSON(text string) (int, map[string]any, error) {
	res := map[string]any{}
	if err := json.Unmarshal([]byte(text), &res); err == nil {
		score := extractScoreFromMap(res)
		return score, res, nil
	}
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start >= 0 && end > start {
		sub := text[start : end+1]
		if err := json.Unmarshal([]byte(sub), &res); err == nil {
			score := extractScoreFromMap(res)
			return score, res, nil
		}
	}
	return -1, nil, fmt.Errorf("no structured json found")
}

func ApplyStructuredFields(base map[string]any, structured map[string]any) {
	if structured == nil {
		return
	}
	if v, ok := structured["tactic_name"]; ok {
		base["tactic_name"] = v
	}
	if v, ok := structured["technique_name"]; ok {
		base["technique_name"] = v
	}
	if v, ok := structured["sub_technique_name"]; ok {
		base["sub_technique_name"] = v
	}
	if v, ok := structured["product_feedback"]; ok {
		base["product_feedback"] = v
	}
	if v, ok := structured["eval_description"]; ok {
		base["eval_description"] = v
	}
	if v, ok := structured["devices"]; ok {
		base["devices"] = v
	}
	if v, ok := structured["attack_result"]; ok {
		base["attack_result"] = v
	}
	if v, ok := structured["community_tags"]; ok {
		base["community_tags"] = v
	}
	if v, ok := structured["serial_number"]; ok {
		base["serial_number"] = v
	}
	if existing, ok := base["suggestion"]; ok {
		if s, ok2 := existing.(string); ok2 && strings.TrimSpace(s) != "" {
		} else {
			if v, ok3 := structured["suggestion"]; ok3 {
				base["suggestion"] = v
			}
		}
	} else {
		if v, ok3 := structured["suggestion"]; ok3 {
			base["suggestion"] = v
		}
	}
	if v, ok := structured["extra_fields"]; ok {
		base["extra_fields"] = v
	}
	if v, ok := structured["level_id"]; ok {
		n := NormalizeNumber(v)
		if n >= 0 {
			base["level_id"] = n
		}
	}
}

func ClampScore(v int) int {
	if v < 0 {
		return 0
	}
	if v > 10 {
		return 10
	}
	return v
}

func NormalizeNumber(v any) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case float32:
		return int(t)
	case int:
		return t
	case int64:
		return int(t)
	case json.Number:
		i, err := t.Int64()
		if err != nil {
			return -1
		}
		return int(i)
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return -1
		}
		n, err := strconv.Atoi(s)
		if err != nil {
			return -1
		}
		return n
	default:
		return -1
	}
}

func extractScoreFromMap(m map[string]any) int {
	v, ok := m["risk_score"]
	if !ok {
		return -1
	}
	n := NormalizeNumber(v)
	if n < 0 {
		return -1
	}
	return ClampScore(n)
}
