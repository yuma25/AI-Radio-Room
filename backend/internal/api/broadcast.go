/**
 * Package api 内の放送関連ハンドラーを定義するファイルである。
 */
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"backend/internal/ai"
	"backend/internal/db"
)

/**
 * broadcastRequest は、放送開始リクエストのボディを表す構造体である。
 */
type broadcastRequest struct {
	Theme string `json:"theme"`
}

/**
 * handleStartBroadcast は、新しい放送を開始し、台本を生成して SSE で配信するハンドラーである。
 * 並列先行合成キャッシュにより、音声再生のラグを解消する。
 */
func (s *Server) handleStartBroadcast(w http.ResponseWriter, r *http.Request) {
	var req broadcastRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "無効なリクエストボディである", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// 1. 放送セッションの作成
	broadcastID, err := s.repo.CreateBroadcast(ctx, req.Theme)
	if err != nil {
		http.Error(w, fmt.Sprintf("放送の作成に失敗した: %v", err), http.StatusInternalServerError)
		return
	}

	// 2. AI 台本生成
	config := ai.GetDefaultConfig()
	config.Theme = req.Theme
	script, err := s.aiClient.GenerateRadioScript(ctx, config)
	if err != nil {
		http.Error(w, fmt.Sprintf("台本生成に失敗した: %v", err), http.StatusInternalServerError)
		return
	}

	// キャッシュのクリア（新しい放送のため）
	s.vvClient.Cache.Clear()

	// 3. 音声の先行並列合成 (Goルーチン)
	// VOICEVOX エンジンへの負荷を考慮し、セマフォで同時実行数を制限する（例: 3並列）
	semaphore := make(chan struct{}, 3)
	var wg sync.WaitGroup

	for i := range script {
		item := &script[i]
		item.SpeakerID = ai.GetSpeakerID(item.Speaker, item.Emotion)

		wg.Add(1)
		go func(text string, speakerID int) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			_, _ = s.vvClient.GenerateWav(ctx, text, speakerID)
		}(item.Text, item.SpeakerID)
	}

	// 4. SSE 設定
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "ストリーミングがサポートされていない", http.StatusInternalServerError)
		return
	}

	// 5. 逐次保存と配信 (音声合成を待たずに即座に開始)
	for _, item := range script {
		transcript := db.Transcript{
			BroadcastID: broadcastID,
			SpeakerID:   item.Speaker,
			Text:        item.Text,
			Emotion:     item.Emotion,
		}
		if err := s.repo.SaveTranscript(ctx, &transcript); err != nil {
			fmt.Printf("警告: 発言の保存に失敗した: %v\n", err)
		}

		data, _ := json.Marshal(item)
		fmt.Fprintf(w, "data: %s\n\n", string(data))
		flusher.Flush()
	}

	wg.Wait()

	if err := s.repo.EndBroadcast(ctx, broadcastID); err != nil {
		fmt.Printf("警告: 放送終了処理に失敗した: %v\n", err)
	}

	fmt.Fprintf(w, "event: close\ndata: 放送が終了した\n\n")
	flusher.Flush()
}

/**
 * handleListBroadcasts は、過去の放送履歴一覧を JSON で返すハンドラーである。
 */
func (s *Server) handleListBroadcasts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]string{})
}
