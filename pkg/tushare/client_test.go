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
			Items:  [][]string{{"600519.SH", "Kweichow Moutai"}},
		},
	})

	if len(got) != 1 {
		t.Fatalf("len(rows) = %d, want 1", len(got))
	}
	if got[0]["ts_code"] != "600519.SH" || got[0]["missing"] != "" {
		t.Fatalf("row = %+v, want mapped row with empty missing field", got[0])
	}
}
