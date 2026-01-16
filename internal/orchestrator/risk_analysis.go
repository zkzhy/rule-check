package orchestrator

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	modelcomp "audit-workflow/internal/components/model"
	"audit-workflow/internal/components/parser"
	promptcomp "audit-workflow/internal/components/prompt"
	"audit-workflow/internal/components/tools/taxonomy"
	"audit-workflow/internal/config"
	"audit-workflow/internal/types"
)

type RiskAnalysisOptions struct {
	Resume bool
}

func RunRiskAnalysis(ctx context.Context, cfg *config.RootConfig) error {
	return RunRiskAnalysisWithOptions(ctx, cfg, RiskAnalysisOptions{})
}

func RunRiskAnalysisWithOptions(ctx context.Context, cfg *config.RootConfig, opt RiskAnalysisOptions) error {
	if err := os.MkdirAll("data", 0o755); err != nil {
		return fmt.Errorf("create data dir failed: %w", err)
	}
	inFile := cfg.Paths.OutputFile
	if inFile == "" {
		inFile = "data/pending_audits.jsonl"
	}
	outResultsFile := "data/pending_audits_results.jsonl"

	items, err := loadPendingRecords(inFile)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		fmt.Printf("[Info] No items found in %s\n", filepath.Base(inFile))
		return nil
	}

	csvPath := resolveATTCKCSVPath(cfg)
	if csvPath == "" {
		return fmt.Errorf("ATT&CK.csv not found (set ai.attck.csv_path or place it at ./ATT&CK.csv or ../ATT&CK.csv)")
	}
	if err := taxonomy.Load(csvPath); err != nil {
		return fmt.Errorf("load ATT&CK.csv failed: %w", err)
	}

	tacticCandidates := buildTacticCandidates(cfg)
	if len(tacticCandidates) == 0 {
		return fmt.Errorf("no tactic candidates available")
	}
	tacticCandidatesJSON, _ := json.Marshal(tacticCandidates)

	tmpl, err := promptcomp.BuildRiskTemplate(cfg)
	if err != nil {
		return fmt.Errorf("load prompt template failed: %w", err)
	}
	tacticTmpl := promptcomp.BuildATTCKTacticTemplate()

	processed := map[string]bool{}
	if opt.Resume {
		processed, err = loadProcessedIDs(outResultsFile)
		if err != nil {
			return err
		}
	}

	wfResults, err := openResultsFile(outResultsFile, opt.Resume)
	if err != nil {
		return fmt.Errorf("open results file failed: %w", err)
	}
	defer wfResults.Close()

	debug := strings.ToLower(os.Getenv("AI_DEBUG"))
	debugMode := debug == "1" || debug == "true" || debug == "yes"
	toProcess := make([]types.PendingRecord, 0, len(items))
	toProcessIdx := make([]int, 0, len(items))
	for idx, rec := range items {
		if opt.Resume && processed[fmt.Sprint(rec.ID)] {
			continue
		}
		toProcess = append(toProcess, rec)
		toProcessIdx = append(toProcessIdx, idx)
	}

	fmt.Printf("[Info] Starting risk analysis for %d items using %s (model: %s, concurrency: %d)...\n", len(toProcess), cfg.AI.Provider, cfg.AI.Model, cfg.AI.Concurrency)

	limiter := newLLMLimiter(cfg.AI.RateLimitQPS)
	if limiter != nil {
		defer limiter.Close()
	}

	type job struct {
		idx int
		rec types.PendingRecord
	}
	type result struct {
		idx   int
		id    any
		wrote bool
		line  []byte
		log   string
	}
	type workerProcessor func(context.Context, int, types.PendingRecord) result

	initWorker := func() (workerProcessor, error) {
		chatModel, err := modelcomp.NewChatModel(ctx, cfg)
		if err != nil {
			return nil, fmt.Errorf("init chat model failed: %w", err)
		}

		waitLLM := func(ctx context.Context) error {
			if limiter != nil {
				return limiter.Wait(ctx)
			}
			t := time.NewTimer(100 * time.Millisecond)
			defer t.Stop()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-t.C:
				return nil
			}
		}

		return func(ctx context.Context, idx int, rec types.PendingRecord) result {
			total := len(items)
			data := rec.Data
			if data == nil {
				data = map[string]any{}
			}

			contextText := buildTrimmedContext(cfg, data)

			tacticMsgs, err := tacticTmpl.Format(ctx, map[string]any{
				"context":           contextText,
				"tactic_candidates": string(tacticCandidatesJSON),
			})
			if err != nil {
				return result{idx: idx, id: rec.ID, wrote: false, log: fmt.Sprintf("[%d/%d] ID: %v -> Prompt Format Error: %v", idx+1, total, rec.ID, err)}
			}

			if debugMode && idx == 0 {
				fmt.Println("=== DEBUG PROMPT BEGIN ===")
				if len(tacticMsgs) > 0 {
					fmt.Println(truncate(tacticMsgs[0].Content, 500) + "...")
				}
				fmt.Println("=== DEBUG PROMPT END ===")
			}

			if err := waitLLM(ctx); err != nil {
				return result{idx: idx, id: rec.ID, wrote: false, log: fmt.Sprintf("[%d/%d] ID: %v -> Error: %v", idx+1, total, rec.ID, err)}
			}
			tacticResp, err := chatModel.Generate(ctx, tacticMsgs)
			selectedTactic := ""
			if err == nil {
				selectedTactic = parseJSONStringField(tacticResp.Content, "tactic_name")
			}
			selectedTactic = strings.TrimSpace(selectedTactic)
			if !isInList(selectedTactic, tacticCandidates) {
				selectedTactic = tacticCandidates[0]
			}

			techCands := taxonomy.GenerateTechniqueCandidates(
				selectedTactic,
				contextText,
				cfg.AI.ATTCK.TechniqueTopK,
				cfg.AI.ATTCK.SubMaxPerTechnique,
			)
			techCandsText := taxonomy.FormatTechniqueCandidates(selectedTactic, techCands, cfg.AI.ATTCK.CandidateMaxRunes)
			allowedTech, allowedSub := buildAllowedFromCandidates(techCands)

			msgs, err := tmpl.Format(ctx, map[string]any{
				"context":              contextText,
				"tactic_name_selected": selectedTactic,
				"technique_candidates": techCandsText,
			})
			if err != nil {
				return result{idx: idx, id: rec.ID, wrote: false, log: fmt.Sprintf("[%d/%d] ID: %v -> Prompt Format Error: %v", idx+1, total, rec.ID, err)}
			}

			var promptText string
			if len(msgs) > 0 {
				promptText = msgs[0].Content
			}
			if debugMode && idx == 0 {
				fmt.Println("=== DEBUG PROMPT2 BEGIN ===")
				fmt.Println(truncate(promptText, 500) + "...")
				fmt.Println("=== DEBUG PROMPT2 END ===")
			}

			if err := waitLLM(ctx); err != nil {
				return result{idx: idx, id: rec.ID, wrote: false, log: fmt.Sprintf("[%d/%d] ID: %v -> Error: %v", idx+1, total, rec.ID, err)}
			}
			respMsg, err := chatModel.Generate(ctx, msgs)

			var rawResponse string
			var score int
			var logLine string

			if err != nil {
				score = -1
				logLine = fmt.Sprintf("[%d/%d] ID: %v -> Error: %v", idx+1, total, rec.ID, err)
			} else {
				rawResponse = respMsg.Content
				if debugMode && idx == 0 {
					fmt.Println("=== DEBUG RESPONSE BEGIN ===")
					fmt.Println(rawResponse)
					fmt.Println("=== DEBUG RESPONSE END ===")
				}
				structuredScore, structuredData, parseErr := parser.ParseStructuredJSON(rawResponse)
				if parseErr != nil {
					score = parseScore(rawResponse)
					logLine = fmt.Sprintf("[%d/%d] ID: %v -> Score(text): %d", idx+1, total, rec.ID, score)
				} else {
					if structuredScore >= 0 {
						score = structuredScore
					}
					parser.ApplyStructuredFields(data, structuredData)
					sanitizeATTCKSelection(data, selectedTactic, allowedTech, allowedSub)
					logLine = fmt.Sprintf("[%d/%d] ID: %v -> Score(json): %d", idx+1, total, rec.ID, score)
				}
			}

			newData := map[string]any{}
			for k, v := range data {
				newData[k] = v
			}

			if existing, ok := newData["risk_score"]; ok {
				num := parser.NormalizeNumber(existing)
				if num >= 0 {
					newData["risk_score"] = parser.ClampScore(num)
				}
			} else if score >= 0 {
				newData["risk_score"] = parser.ClampScore(score)
			}

			resultsRec := map[string]any{"id": rec.ID, "generated_at": utcISO(), "data": newData}
			bResults, _ := json.Marshal(resultsRec)
			return result{idx: idx, id: rec.ID, wrote: true, line: bResults, log: logLine}
		}, nil
	}

	workers := cfg.AI.Concurrency
	if workers <= 0 {
		workers = 1
	}
	if workers > len(toProcess) && len(toProcess) > 0 {
		workers = len(toProcess)
	}

	jobsCh := make(chan job)
	resultsCh := make(chan result)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		processor, err := initWorker()
		if err != nil {
			return err
		}
		wg.Add(1)
		go func(p workerProcessor) {
			defer wg.Done()
			for j := range jobsCh {
				r := p(ctx, j.idx, j.rec)
				select {
				case <-ctx.Done():
					return
				case resultsCh <- r:
				}
			}
		}(processor)
	}

	go func() {
		for i, rec := range toProcess {
			select {
			case <-ctx.Done():
				close(jobsCh)
				return
			case jobsCh <- job{idx: toProcessIdx[i], rec: rec}:
			}
		}
		close(jobsCh)
	}()

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	written := 0
	for r := range resultsCh {
		if strings.TrimSpace(r.log) != "" {
			fmt.Println(r.log)
		}
		if r.wrote && len(r.line) > 0 {
			_, _ = wfResults.Write(r.line)
			_, _ = wfResults.WriteString("\n")
			written++
		}
	}

	fmt.Printf("[Success] Completed. %d records processed, results written to %s\n", written, filepath.Base(outResultsFile))
	return nil
}

type llmLimiter struct {
	tokens <-chan struct{}
	stop   func()
}

func newLLMLimiter(qps int) *llmLimiter {
	if qps <= 0 {
		return nil
	}
	tokens := make(chan struct{}, qps)
	done := make(chan struct{})
	interval := time.Second / time.Duration(qps)
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				select {
				case tokens <- struct{}{}:
				default:
				}
			}
		}
	}()
	return &llmLimiter{
		tokens: tokens,
		stop:   func() { close(done) },
	}
}

func (l *llmLimiter) Wait(ctx context.Context) error {
	if l == nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-l.tokens:
		return nil
	}
}

func (l *llmLimiter) Close() {
	if l == nil || l.stop == nil {
		return
	}
	l.stop()
}

func loadPendingRecords(path string) ([]types.PendingRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var items []types.PendingRecord
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var rec types.PendingRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			continue
		}
		items = append(items, rec)
	}
	if err := scanner.Err(); err != nil {
		return items, err
	}
	return items, nil
}

func openResultsFile(path string, resume bool) (*os.File, error) {
	if resume {
		return os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	}
	return os.Create(path)
}

func loadProcessedIDs(path string) (map[string]bool, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]bool{}, nil
		}
		return nil, err
	}
	defer f.Close()

	ids := map[string]bool{}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var rec struct {
			ID any `json:"id"`
		}
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			continue
		}
		if strings.TrimSpace(fmt.Sprint(rec.ID)) != "" {
			ids[fmt.Sprint(rec.ID)] = true
		}
	}
	if err := scanner.Err(); err != nil {
		return ids, err
	}
	return ids, nil
}

func parseScore(text string) int {
	re := regexp.MustCompile(`\d+`)
	match := re.FindString(text)
	if match == "" {
		return -1
	}
	val, err := strconv.Atoi(match)
	if err != nil {
		return -1
	}
	if val < 1 {
		return 1
	}
	if val > 10 {
		return 10
	}
	return val
}

func firstString(v any) string {
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	return ""
}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}

func buildTrimmedContext(cfg *config.RootConfig, data map[string]any) string {
	if cfg == nil {
		return ""
	}
	var b strings.Builder
	remaining := cfg.AI.Context.TotalMaxRunes
	if remaining <= 0 {
		return ""
	}

	appendSection := func(title, content string, maxRunes int) {
		content = strings.TrimSpace(content)
		if content == "" || remaining <= 0 {
			return
		}

		if maxRunes > 0 {
			content = truncate(content, maxRunes)
		}

		sepRunes := 0
		if b.Len() > 0 {
			sepRunes = 2
		}
		if remaining <= sepRunes {
			return
		}
		remainingAfterSep := remaining - sepRunes

		sectionHeader := title + "："
		headerRunes := len([]rune(sectionHeader))
		if headerRunes >= remainingAfterSep {
			return
		}

		contentRunes := []rune(content)
		maxContent := remainingAfterSep - headerRunes
		if maxContent <= 0 {
			return
		}
		if len(contentRunes) > maxContent {
			content = string(contentRunes[:maxContent])
		}

		if sepRunes > 0 {
			b.WriteString("\n\n")
			remaining -= 2
		}
		b.WriteString(sectionHeader)
		b.WriteString(content)
		remaining -= headerRunes + len([]rune(content))
	}

	appendSection("漏洞名称", firstString(data["name"]), cfg.AI.Context.NameMaxRunes)
	appendSection("漏洞描述", firstString(data["description"]), cfg.AI.Context.DescriptionMaxRunes)
	appendSection("PoC/证据", firstString(data["xray_poc_content"]), cfg.AI.Context.POCMaxRunes)
	appendSection("请求包", firstString(data["req_pkg"]), cfg.AI.Context.ReqMaxRunes)
	appendSection("响应包", firstString(data["resp_pkg"]), cfg.AI.Context.RespMaxRunes)

	return b.String()
}

func resolveATTCKCSVPath(cfg *config.RootConfig) string {
	var candidates []string
	if cfg != nil && strings.TrimSpace(cfg.AI.ATTCK.CSVPath) != "" {
		candidates = append(candidates, strings.TrimSpace(cfg.AI.ATTCK.CSVPath))
	}
	candidates = append(candidates,
		"ATT&CK.csv",
		"../ATT&CK.csv",
	)

	for _, p := range candidates {
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func buildTacticCandidates(cfg *config.RootConfig) []string {
	all := taxonomy.ListTactics()
	if cfg == nil || len(cfg.AI.ATTCK.TacticAllowlist) == 0 {
		return all
	}
	allowed := make([]string, 0, len(cfg.AI.ATTCK.TacticAllowlist))
	for _, tName := range cfg.AI.ATTCK.TacticAllowlist {
		tName = strings.TrimSpace(tName)
		if tName == "" {
			continue
		}
		if isInList(tName, all) {
			allowed = append(allowed, tName)
		}
	}
	if len(allowed) == 0 {
		return all
	}
	return allowed
}

func isInList(v string, list []string) bool {
	v = strings.TrimSpace(v)
	if v == "" {
		return false
	}
	for _, it := range list {
		if v == strings.TrimSpace(it) {
			return true
		}
	}
	return false
}

func parseJSONStringField(text, field string) string {
	start := strings.IndexByte(text, '{')
	end := strings.LastIndexByte(text, '}')
	if start >= 0 && end > start {
		var m map[string]any
		if err := json.Unmarshal([]byte(text[start:end+1]), &m); err == nil {
			if v, ok := m[field]; ok {
				if s, ok := v.(string); ok {
					return s
				}
			}
		}
	}

	re := regexp.MustCompile(fmt.Sprintf(`"%s"\s*:\s*"([^"]*)"`, regexp.QuoteMeta(field)))
	if match := re.FindStringSubmatch(text); len(match) == 2 {
		return match[1]
	}
	return ""
}

func buildAllowedFromCandidates(cands []taxonomy.TechniqueCandidate) (map[string]bool, map[string]map[string]bool) {
	allowedTech := make(map[string]bool)
	allowedSub := make(map[string]map[string]bool)
	for _, c := range cands {
		tn := strings.TrimSpace(c.TechniqueName)
		if tn == "" {
			continue
		}
		allowedTech[tn] = true
		if _, ok := allowedSub[tn]; !ok {
			allowedSub[tn] = make(map[string]bool)
		}
		for _, s := range c.SubNames {
			sn := strings.TrimSpace(s)
			if sn == "" {
				continue
			}
			allowedSub[tn][sn] = true
		}
	}
	return allowedTech, allowedSub
}

func sanitizeATTCKSelection(data map[string]any, selectedTactic string, allowedTech map[string]bool, allowedSub map[string]map[string]bool) {
	if data == nil {
		return
	}

	data["tactic_name"] = selectedTactic

	tn := strings.TrimSpace(firstString(data["technique_name"]))
	sn := strings.TrimSpace(firstString(data["sub_technique_name"]))
	if tn == "" {
		data["technique_name"] = ""
		data["sub_technique_name"] = ""
		return
	}

	if !allowedTech[tn] {
		data["technique_name"] = ""
		data["sub_technique_name"] = ""
		return
	}

	if sn == "" {
		data["sub_technique_name"] = ""
		return
	}
	if subs, ok := allowedSub[tn]; !ok || !subs[sn] {
		data["sub_technique_name"] = ""
	}
}

func utcISO() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z")
}
