package fetch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"audit-workflow/internal/config"
	"audit-workflow/internal/httpclient"
)

type loginResp struct {
	Data struct {
		AccessToken string `json:"access_token"`
	} `json:"data"`
}

type listResp struct {
	Data struct {
		Data []map[string]any `json:"data"`
	} `json:"data"`
}

func Run(cfg *config.RootConfig) error {
	cl := httpclient.New(cfg.Yuheng.VerifySSL, cfg.Yuheng.TimeoutS)

	fmt.Print("[Login] 尝试登录 ")
	tok, err := login(cl, cfg)
	if err != nil || tok == "" {
		fmt.Println("失败")
		return err
	}
	fmt.Println("OK")

	outFile := cfg.PendingAuditsPath()
	outDir := filepath.Dir(outFile)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	f, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer f.Close()

	pageNo := 1
	pageSize := cfg.Yuheng.ListPageSize
	total := 0

	for {
		fmt.Printf("[Fetch] List page %d ", pageNo)
		items, err := fetchList(cl, cfg, tok, pageNo, pageSize)
		if err != nil {
			fmt.Println("失败", err)
			break
		}
		if len(items) == 0 {
			fmt.Println("空")
			break
		}
		fmt.Printf("获取到 %d 个ID，开始获取详情...\n", len(items))

		for _, it := range items {
			idVal, ok := it["id"].(float64)
			if !ok {
				continue
			}
			id := int(idVal)
			detail, err := fetchDetail(cl, cfg, tok, id)
			if err != nil || detail == nil {
				fmt.Printf("  [Warn] 无法获取 ID %d 的详情\n", id)
				continue
			}

			dataToSave := map[string]any{
				"name":             firstString(detail["name"], detail["title"]),
				"description":      firstString(detail["description"], detail["desc"]),
				"xray_poc_content": firstString(detail["xray_poc_content"], detail["poc"]),
				"req_pkg":          firstString(detail["req_pkg"]),
				"resp_pkg":         firstString(detail["resp_pkg"]),
				"type":             firstString(detail["type"], "HTTP"),
				"_raw":             detail,
			}

			rec := map[string]any{
				"id":         id,
				"data":       dataToSave,
				"fetched_at": utcISO(),
			}
			b, _ := json.Marshal(rec)
			f.Write(b)
			f.WriteString("\n")
			total++
		}

		if len(items) < pageSize {
			break
		}
		pageNo++
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("[Success] 写入 %d 条到 %s\n", total, filepath.Base(outFile))
	return nil
}

func login(cl *httpclient.Client, cfg *config.RootConfig) (string, error) {
	fullURL, err := resolveURL(cfg.Yuheng.BaseURL, "/api/login")
	if err != nil {
		return "", err
	}
	body := map[string]any{
		"username": cfg.Yuheng.Username,
		"password": cfg.Yuheng.Password,
	}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, fullURL, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	var out loginResp
	code, err := cl.DoJSON(req, &out)
	if err != nil {
		return "", err
	}
	if code != 200 {
		return "", fmt.Errorf("http %d", code)
	}
	return out.Data.AccessToken, nil
}

func fetchList(cl *httpclient.Client, cfg *config.RootConfig, token string, pageNo, pageSize int) ([]map[string]any, error) {
	endpoint := cfg.Yuheng.ListEndpoint
	if endpoint == "" {
		endpoint = "/api/lines/operation"
	}
	endpoint = normalizePath(endpoint)
	fullURL, err := resolveURL(cfg.Yuheng.BaseURL, endpoint)
	if err != nil {
		return nil, err
	}

	filters := map[string]any{}
	for k, v := range cfg.Yuheng.ListFilters {
		filters[k] = v
	}
	if _, ok := filters["review_status"]; !ok {
		filters["review_status"] = "待审核"
	}
	if _, ok := filters["type"]; !ok {
		filters["type"] = "HTTP"
	}
	for k, v := range cfg.Yuheng.ListTimeFields {
		if val, ok := filters[k]; ok && v != "" {
			delete(filters, k)
			filters[v] = val
		}
	}

	method := strings.ToUpper(cfg.Yuheng.ListMethod)
	if method != http.MethodPost {
		method = http.MethodGet
	}

	sendStyle := strings.ToLower(cfg.Yuheng.ListSendStyle)
	if sendStyle != "json" {
		sendStyle = "query"
	}

	var req *http.Request

	if method == http.MethodGet || sendStyle == "query" {
		values := url.Values{}
		for k, v := range filters {
			if v == nil {
				continue
			}
			values.Set(k, fmt.Sprint(v))
		}
		values.Set("page_no", fmt.Sprint(pageNo))
		values.Set("page_size", fmt.Sprint(pageSize))
		u, err := url.Parse(fullURL)
		if err != nil {
			return nil, err
		}
		u.RawQuery = values.Encode()
		req, _ = http.NewRequest(http.MethodGet, u.String(), nil)
	} else {
		body := map[string]any{
			"filters":   filters,
			"page_no":   pageNo,
			"page_size": pageSize,
		}
		b, _ := json.Marshal(body)
		req, _ = http.NewRequest(http.MethodPost, fullURL, bytes.NewReader(b))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Cookie", "AccessToken="+token+";")

	debug := strings.ToLower(os.Getenv("FETCH_DEBUG"))
	if debug == "1" || debug == "true" || debug == "yes" {
		if req.Method == http.MethodGet {
			fmt.Printf("[FetchDebug] %s %s\n", req.Method, req.URL.String())
		} else {
			fmt.Printf("[FetchDebug] %s %s\n", req.Method, fullURL)
		}
	}

	var out listResp
	code, err := cl.DoJSON(req, &out)
	if err != nil {
		return nil, err
	}
	if code != 200 {
		return nil, fmt.Errorf("http %d", code)
	}
	return out.Data.Data, nil
}

func fetchDetail(cl *httpclient.Client, cfg *config.RootConfig, token string, id int) (map[string]any, error) {
	fullURL, err := resolveURL(cfg.Yuheng.BaseURL, fmt.Sprintf("/api/operation_side/audit/lines/%d", id))
	if err != nil {
		return nil, err
	}
	req, _ := http.NewRequest(http.MethodGet, fullURL, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Cookie", "AccessToken="+token+";")
	var raw struct {
		Data map[string]any `json:"data"`
	}
	code, err := cl.DoJSON(req, &raw)
	if err != nil {
		return nil, err
	}
	if code != 200 {
		return nil, fmt.Errorf("http %d", code)
	}
	return raw.Data, nil
}

func firstString(vals ...any) string {
	for _, v := range vals {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return ""
}

func utcISO() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z")
}

func normalizePath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	if strings.HasPrefix(p, "/") {
		return p
	}
	return "/" + p
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
