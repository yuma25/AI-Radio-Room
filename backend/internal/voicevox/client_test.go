package voicevox

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

/**
 * TestNewClient は、NewClient が正しく構造体を初期化するかをテストする。
 */
func TestNewClient(t *testing.T) {
	baseURL := "http://localhost:50021"
	client := NewClient(baseURL)

	if client.baseURL != baseURL {
		t.Errorf("Expected baseURL %s, got %s", baseURL, client.baseURL)
	}
	if client.http == nil {
		t.Error("Expected http client to be initialized, got nil")
	}
}

/**
 * TestGenerateWav は、モックサーバーを使用して音声合成フローをテストする。
 */
func TestGenerateWav(t *testing.T) {
	// モックサーバーの作成
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/audio_query":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"dummy": "query"}`))
		case "/synthesis":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("dummy wav data"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	ctx := context.Background()

	wav, err := client.GenerateWav(ctx, "テスト", 2)
	if err != nil {
		t.Fatalf("GenerateWav failed: %v", err)
	}

	if string(wav) != "dummy wav data" {
		t.Errorf("Expected 'dummy wav data', got %s", string(wav))
	}
}
