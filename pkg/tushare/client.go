package tushare

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const defaultTimeout = 10 * time.Second

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

type Option func(*Client)

type Request struct {
	APIName string         `json:"api_name"`
	Token   string         `json:"token"`
	Params  map[string]any `json:"params"`
	Fields  string         `json:"fields"`
}

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data Data   `json:"data"`
}

type Data struct {
	Fields []string            `json:"fields"`
	Items  [][]json.RawMessage `json:"items"`
}

func NewClient(baseURL string, token string, options ...Option) *Client {
	client := &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
	for _, option := range options {
		option(client)
	}
	return client
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(client *Client) {
		if httpClient != nil {
			client.httpClient = httpClient
		}
	}
}

func (c *Client) Call(ctx context.Context, apiName string, params map[string]any, fields []string) (Response, error) {
	reqBody := Request{
		APIName: apiName,
		Token:   c.token,
		Params:  params,
		Fields:  strings.Join(fields, ","),
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return Response{}, fmt.Errorf("encode tushare request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(payload))
	if err != nil {
		return Response{}, fmt.Errorf("build tushare request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("call tushare api %s: %w", apiName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Response{}, fmt.Errorf("call tushare api %s: http status %d", apiName, resp.StatusCode)
	}

	var decoded Response
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return Response{}, fmt.Errorf("decode tushare response for %s: %w", apiName, err)
	}
	if decoded.Code != 0 {
		return Response{}, fmt.Errorf("tushare api %s returned code %d", apiName, decoded.Code)
	}
	return decoded, nil
}

func Rows(resp Response) []map[string]string {
	rows := make([]map[string]string, 0, len(resp.Data.Items))
	for _, item := range resp.Data.Items {
		row := make(map[string]string, len(resp.Data.Fields))
		for i, field := range resp.Data.Fields {
			if i >= len(item) {
				row[field] = ""
				continue
			}
			row[field] = rawValueString(item[i])
		}
		rows = append(rows, row)
	}
	return rows
}

func rawValueString(value json.RawMessage) string {
	if len(value) == 0 || string(value) == "null" {
		return ""
	}
	var text string
	if err := json.Unmarshal(value, &text); err == nil {
		return text
	}
	// Tushare 的 items 单元格有时是数字或布尔值，保留原始 JSON 文本可避免精度和格式变化。
	return strings.TrimSpace(string(value))
}

func StockBasicFields() []string {
	return []string{"ts_code", "symbol", "name", "area", "industry", "list_date", "delist_date", "list_status"}
}

func TradeCalFields() []string {
	return []string{"exchange", "cal_date", "is_open", "pretrade_date"}
}

func DailyFields() []string {
	return []string{"ts_code", "trade_date", "open", "high", "low", "close", "pre_close", "change", "pct_chg", "vol", "amount"}
}
