package tushare

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientCallUsesExplicitFieldsAndDoesNotExposeTokenOnAPIError(t *testing.T) {
	var got Request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(Response{Code: 1001, Msg: "bad token"})
	}))
	t.Cleanup(server.Close)

	client := NewClient(server.URL, "secret-token")
	_, err := client.Call(context.Background(), "daily", map[string]any{"ts_code": "600519.SH"}, DailyFields())
	if err == nil {
		t.Fatalf("Call() error = nil, want error")
	}
	if strings.Contains(err.Error(), "secret-token") {
		t.Fatalf("error exposes token: %v", err)
	}
	if got.Fields != strings.Join(DailyFields(), ",") {
		t.Fatalf("fields = %q, want explicit daily fields", got.Fields)
	}
	if got.Token != "secret-token" {
		t.Fatalf("token not sent")
	}
}

func TestRowsMapsFieldsToItems(t *testing.T) {
	got := Rows(Response{
		Data: Data{
			Fields: []string{"ts_code", "name", "missing"},
			Items:  rawItems([]any{"600519.SH", "Kweichow Moutai"}),
		},
	})

	if len(got) != 1 {
		t.Fatalf("len(rows) = %d, want 1", len(got))
	}
	if got[0]["ts_code"] != "600519.SH" || got[0]["missing"] != "" {
		t.Fatalf("row = %+v, want mapped row with empty missing field", got[0])
	}
}

func TestRowsConvertsMixedTushareItemTypes(t *testing.T) {
	got := Rows(Response{
		Data: Data{
			Fields: []string{"exchange", "cal_date", "is_open", "pretrade_date", "empty"},
			Items:  rawItems([]any{"SSE", 20260630, 1, nil, true}),
		},
	})

	if len(got) != 1 {
		t.Fatalf("len(rows) = %d, want 1", len(got))
	}
	if got[0]["exchange"] != "SSE" || got[0]["cal_date"] != "20260630" || got[0]["is_open"] != "1" || got[0]["pretrade_date"] != "" || got[0]["empty"] != "true" {
		t.Fatalf("row = %+v, want mixed values converted to strings", got[0])
	}
}

func rawItems(items ...[]any) [][]json.RawMessage {
	rows := make([][]json.RawMessage, 0, len(items))
	for _, item := range items {
		row := make([]json.RawMessage, 0, len(item))
		for _, value := range item {
			raw, err := json.Marshal(value)
			if err != nil {
				panic(err)
			}
			row = append(row, raw)
		}
		rows = append(rows, row)
	}
	return rows
}
