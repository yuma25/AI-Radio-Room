/**
 * Package api の追加ハンドラーを定義する。
 */
package api

import (
	"fmt"
	"net/http"
	"strconv"
)

/*
*
  - handleTTS は、テキストから音声データを生成して返すハンドラーである。

...

  - 内部処理:
  - - 1. Query Parameter から `text` と `speaker` (ID) を取得する。
  - - 2. `voicevoxClient.GenerateWav(ctx, text, speakerID)` を呼び出し、WAV データを取得する。
  - - 3. レスポンスヘッダーに `Content-Type: audio/wav` を設定し、バイナリデータを書き出す。
*/
func (s *Server) handleTTS(w http.ResponseWriter, r *http.Request) {
	text := r.URL.Query().Get("text")
	speakerStr := r.URL.Query().Get("speaker")

	if text == "" || speakerStr == "" {
		http.Error(w, "text または speaker パラメータが不足している", http.StatusBadRequest)
		return
	}

	speakerID, err := strconv.Atoi(speakerStr)
	if err != nil {
		http.Error(w, "speaker ID が無効である", http.StatusBadRequest)
		return
	}

	// [内部] VOICEVOX クライアントを使用して音声を生成する。
	wav, err := s.vvClient.GenerateWav(r.Context(), text, speakerID)
	if err != nil {
		http.Error(w, fmt.Sprintf("音声合成失敗: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "audio/wav")
	w.Write(wav)
}
