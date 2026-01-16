package submit

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"audit-workflow/internal/components/tools/taxonomy"
	"audit-workflow/internal/config"
	"audit-workflow/internal/httpclient"
)

type riskRecord struct {
	ID   any            `json:"id"`
	Data map[string]any `json:"data"`
}

func Run(cfg *config.RootConfig) error {
	taxPath := "ATT&CK.csv"
	if _, err := os.Stat(taxPath); os.IsNotExist(err) {
		taxPath = "../ATT&CK.csv"
	}
	if err := taxonomy.Load(taxPath); err != nil {
		fmt.Printf("[Warning] Failed to load taxonomy from %s: %v\n", taxPath, err)
	}

	inputFile := "data/pending_audits_results.jsonl"
	f, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	cl := httpclient.New(cfg.Yuheng.VerifySSL, cfg.Yuheng.TimeoutS)

	fmt.Print("[Login] Authenticating... ")
	tok, err := login(cl, cfg)
	if err != nil || tok == "" {
		fmt.Println("Failed")
		return err
	}
	fmt.Println("OK")

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	success := 0
	fail := 0

	var records []riskRecord
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var rec riskRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			continue
		}
		records = append(records, rec)
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	fmt.Printf("[Info] Found %d records to process\n", len(records))

	for _, rec := range records {
		data := rec.Data
		if data == nil {
			continue
		}
		rawDetail, _ := data["_raw"].(map[string]any)
		if rawDetail == nil {
			fmt.Printf("[Skip] ID %v: Missing raw detail or score\n", rec.ID)
			continue
		}
		score := normalizeScore(data["risk_score"])
		if score < 0 {
			fmt.Printf("[Skip] ID %v: Missing valid score\n", rec.ID)
			continue
		}

		if v, ok := data["eval_description"]; ok {
			rawDetail["eval_description"] = v
		}
		rawDetail["attack_result"] = "成功"

		if v, ok := data["level_id"]; ok {
			rawDetail["level_id"] = v
		}

		if v, ok := rawDetail["devices"]; ok {
			if !isValidDevices(v) {
				delete(rawDetail, "devices")
				fmt.Printf("[Warning] ID %v: Invalid devices format in _raw, dropped\n", rec.ID)
			}
		}

		tName := firstString(data["tactic_name"])
		teName := firstString(data["technique_name"])
		subName := firstString(data["sub_technique_name"])

		fmt.Printf("[Analysis] ID %v: AI Suggestion -> Tactic: '%s', Technique: '%s', Sub: '%s'\n", rec.ID, tName, teName, subName)

		if tName != "" {
			tid, teid, subid, found := taxonomy.LookupIDs(tName, teName, subName)
			if !found {
				if tid2, ok := taxonomy.LookupTacticID(tName); ok {
					tid = tid2
					teid = 0
					subid = 0
					teName = ""
					subName = ""
					found = true
					fmt.Printf("[Warning] ID %v: Technique '%s' not found, falling back to Tactic '%s'\n", rec.ID, firstString(data["technique_name"]), tName)
				} else {
					fmt.Printf("[Warning] ID %v: Tactic '%s' not found in taxonomy\n", rec.ID, tName)
				}
			}

			if found {
				rawDetail["tactics"] = []map[string]any{{
					"tactic_id":          tid,
					"tactic_name":        tName,
					"technique_id":       teid,
					"technique_name":     teName,
					"sub_technique_id":   subid,
					"sub_technique_name": subName,
				}}
			}
		}

		if v, ok := data["community_tags"]; ok {
			rawDetail["community_tags"] = v
		}
		if v, ok := data["serial_number"]; ok {
			rawDetail["serial_number"] = v
		}
		if v, ok := data["product_feedback"]; ok {
			rawDetail["product_feedback"] = v
		}

		suggestion := firstString(data["suggestion"])

		fmt.Printf("[Submit] Processing ID %v (Score: %d)... ", rec.ID, score)
		if submitReview(cl, cfg, tok, rawDetail, score, suggestion) {
			fmt.Println("Success")
			success++
		} else {
			fmt.Println("Failed")
			fail++
		}
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("[Summary] Success: %d, Failed: %d\n", success, fail)
	return nil
}

func login(cl *httpclient.Client, cfg *config.RootConfig) (string, error) {
	fullURL, err := resolveURL(cfg.Yuheng.BaseURL, "/api/login")
	if err != nil {
		return "", err
	}
	body := map[string]any{"username": cfg.Yuheng.Username, "password": cfg.Yuheng.Password}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, fullURL, bytesReader(b))
	req.Header.Set("Content-Type", "application/json")
	var out struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	code, err := cl.DoJSON(req, &out)
	if err != nil {
		return "", err
	}
	if code != 200 {
		return "", fmt.Errorf("http %d", code)
	}
	return out.Data.AccessToken, nil
}

func submitReview(cl *httpclient.Client, cfg *config.RootConfig, token string, editData map[string]any, score int, suggestion string) bool {
	fullURL, err := resolveURL(cfg.Yuheng.BaseURL, fmt.Sprintf("/api/operation_side/lines/%v/review", editData["id"]))
	if err != nil {
		fmt.Printf("[Error] Submit failed: %v\n", err)
		return false
	}

	reviewData := map[string]any{
		"id":                       editData["id"],
		"tools":                    editData["tools"],
		"devices":                  editData["devices"],
		"endpoint_json":            editData["endpoint_json"],
		"tactics":                  editData["tactics"],
		"labels":                   editData["labels"],
		"name":                     editData["name"],
		"type":                     editData["type"],
		"attribute_classification": editData["attribute_classification"],
		"attack_type_id":           editData["attack_type_id"],
		"vul_name":                 editData["vul_name"],
		"description":              editData["description"],
		"req_char":                 editData["req_char"],
		"resp_char":                editData["resp_char"],
		"cve_id":                   editData["cve_id"],
		"cnvd_id":                  editData["cnvd_id"],
		"code":                     editData["code"],
		"reference_link":           editData["reference_link"],
		"screenshot_of_proof":      editData["screenshot_of_proof"],
		"req_pkg":                  editData["req_pkg"],
		"resp_pkg":                 editData["resp_pkg"],
		"asset_type_id":            editData["asset_type_id"],
		"asset_id":                 editData["asset_id"],
		"category_id":              editData["category_id"],
		"eval_points":              editData["eval_points"],
		"pcap":                     editData["pcap"],
		"pml":                      editData["pml"],
		"csv":                      editData["csv"],
		"strace":                   editData["strace"],
		"jar":                      editData["jar"],
		"external_ip":              editData["external_ip"],
		"external_port":            editData["external_port"],
		"reverse_shell":            editData["reverse_shell"],
		"reverse_shell_ip":         editData["reverse_shell_ip"],
		"reverse_shell_port":       editData["reverse_shell_port"],
		"eval_description":         editData["eval_description"],
		"suggestion":               suggestion,
		"score":                    score,
		"level_id":                 editData["level_id"],
		"windows_pid":              editData["windows_pid"],
		"subject":                  editData["subject"],
		"body":                     editData["body"],
		"phishing_web":             editData["phishing_web"],
		"phishing_url_path":        editData["phishing_url_path"],
		"phishing_thumbnail":       editData["phishing_thumbnail"],
		"phishing_attachment":      editData["phishing_attachment"],
		"malware":                  editData["malware"],
		"attack_result":            editData["attack_result"],
		"product_feedback":         editData["product_feedback"],
		"chaitin_number":           editData["chaitin_number"],
		"pcap_name":                editData["pcap_name"],
		"pml_name":                 editData["pml_name"],
		"csv_name":                 editData["csv_name"],
		"strace_name":              editData["strace_name"],
		"jar_name":                 editData["jar_name"],
		"attack_type_name":         editData["attack_type_name"],
		"category_name":            editData["category_name"],
		"asset_type_name":          editData["asset_type_name"],
		"asset_name":               editData["asset_name"],
		"level_name":               editData["level_name"],
		"result":                   "通过",
		"content":                  "",
	}

	b, _ := json.Marshal(reviewData)
	req, _ := http.NewRequest(http.MethodPut, fullURL, bytesReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Cookie", "AccessToken="+token+";")

	var out map[string]any
	code, err := cl.DoJSON(req, &out)
	if err != nil {
		fmt.Printf("[Error] Submit failed: %v\n", err)
		return false
	}
	if code != 200 {
		fmt.Printf("[Error] Submit failed: status=%d, msg=%v\n", code, out)
		return false
	}
	return true
}

func resolveURL(base, ref string) (string, error) {
	b := strings.TrimSpace(base)
	if b == "" {
		return "", fmt.Errorf("empty base_url")
	}
	bu, err := url.Parse(b)
	if err != nil {
		return "", err
	}
	ru, err := url.Parse(strings.TrimSpace(ref))
	if err != nil {
		return "", err
	}
	return bu.ResolveReference(ru).String(), nil
}

type bytesReaderWrapper struct {
	pos int
	b   []byte
}

func bytesReader(b []byte) *bytesReaderWrapper {
	return &bytesReaderWrapper{b: b}
}

func (r *bytesReaderWrapper) Read(p []byte) (int, error) {
	if r.pos >= len(r.b) {
		return 0, io.EOF
	}
	n := copy(p, r.b[r.pos:])
	r.pos += n
	return n, nil
}

func normalizeScore(v any) int {
	switch t := v.(type) {
	case float64:
		return clampScoreInt(int(t))
	case float32:
		return clampScoreInt(int(t))
	case int:
		return clampScoreInt(t)
	case int64:
		return clampScoreInt(int(t))
	case json.Number:
		i, err := t.Int64()
		if err != nil {
			return -1
		}
		return clampScoreInt(int(i))
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return -1
		}
		n, err := strconv.Atoi(s)
		if err != nil {
			return -1
		}
		return clampScoreInt(n)
	default:
		return -1
	}
}

func clampScoreInt(v int) int {
	if v < 1 {
		return -1
	}
	if v > 10 {
		return 10
	}
	return v
}

func firstString(v any) string {
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	return ""
}

func isValidDevices(v any) bool {
	if arr, ok := v.([]any); ok {
		if len(arr) == 0 {
			return true
		}
		_, isMap := arr[0].(map[string]any)
		return isMap
	}
	return false
}
