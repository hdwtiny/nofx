package serverchan

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	SendKey string
	HTTP    *http.Client
	BaseURL string
}

func New(sendKey string) *Client {
	return &Client{
		SendKey: strings.TrimSpace(sendKey),
		BaseURL: "https://sctapi.ftqq.com",
		HTTP: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type sendResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Send sends a message via ServerChan (Server酱) new API.
// Endpoint: https://sctapi.ftqq.com/{SendKey}.send
func (c *Client) Send(title, desp string) error {
	if c == nil || c.SendKey == "" {
		return fmt.Errorf("missing send key")
	}
	base := strings.TrimSuffix(strings.TrimSpace(c.BaseURL), "/")
	if base == "" {
		base = "https://sctapi.ftqq.com"
	}
	endpoint := fmt.Sprintf("%s/%s.send", base, url.PathEscape(c.SendKey))
	form := url.Values{}
	form.Set("title", title)
	form.Set("desp", desp)

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("serverchan status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var r sendResp
	if err := json.Unmarshal(body, &r); err != nil {
		// Some deployments may return HTML; still treat as success if 200 and body non-empty.
		if len(body) > 0 {
			return nil
		}
		return err
	}
	if r.Code != 0 {
		return fmt.Errorf("serverchan error code %d: %s", r.Code, r.Message)
	}
	return nil
}

