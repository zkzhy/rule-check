package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"crypto/tls"
)

type Client struct {
	inner *http.Client
}

func New(verifySSL bool, timeoutSec float64) *Client {
	tr := &http.Transport{}
	if !verifySSL {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return &Client{
		inner: &http.Client{
			Timeout:   time.Duration(timeoutSec * float64(time.Second)),
			Transport: tr,
		},
	}
}

func (c *Client) DoJSON(req *http.Request, out any) (int, error) {
	resp, err := c.inner.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if out == nil {
		return resp.StatusCode, nil
	}

	var cap limitedCapture
	cap.max = 2048
	dec := json.NewDecoder(io.TeeReader(resp.Body, &cap))
	if err := dec.Decode(out); err != nil {
		ct := resp.Header.Get("Content-Type")
		prefix := bodyPrefix(cap.b, 200)
		if looksLikeHTML(cap.b) {
			return resp.StatusCode, fmt.Errorf("expected JSON but got HTML: status=%d content_type=%q body_prefix=%q: %w", resp.StatusCode, ct, prefix, err)
		}
		return resp.StatusCode, fmt.Errorf("decode json failed: status=%d content_type=%q body_prefix=%q: %w", resp.StatusCode, ct, prefix, err)
	}
	return resp.StatusCode, nil
}

type limitedCapture struct {
	b   []byte
	max int
}

func (c *limitedCapture) Write(p []byte) (int, error) {
	if c.max <= 0 || len(c.b) >= c.max {
		return len(p), nil
	}
	remain := c.max - len(c.b)
	if remain > 0 {
		if len(p) <= remain {
			c.b = append(c.b, p...)
		} else {
			c.b = append(c.b, p[:remain]...)
		}
	}
	return len(p), nil
}

func looksLikeHTML(b []byte) bool {
	s := bytes.TrimSpace(b)
	if len(s) == 0 {
		return false
	}
	if s[0] != '<' {
		return false
	}
	ls := strings.ToLower(string(s[:min(len(s), 200)]))
	return strings.HasPrefix(ls, "<!doctype") || strings.HasPrefix(ls, "<html") || strings.HasPrefix(ls, "<head") || strings.HasPrefix(ls, "<body") || strings.Contains(ls, "<html")
}

func bodyPrefix(b []byte, max int) string {
	s := strings.TrimSpace(string(bytes.TrimSpace(b)))
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
