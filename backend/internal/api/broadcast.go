/**
 * Package api 内の放送関連ハンドラーを定義するファイルである。
 */
package api

import (
	"encoding/json"
	"fmt"
	"net/http"

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
 * handleStartBroadcast は、新しい放送を開始し、台本を複数フェーズに分けて生成・配信するハンドラーである。
 * フェーズごとに AI を呼び出すことで、会話の質とトータルの長さを大幅に向上させている。
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

	// キャッシュのクリア（新しい放送のため）
	s.vvClient.Cache.Clear()

	// SSE 設定
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "ストリーミングがサポートされていない", http.StatusInternalServerError)
		return
	}

	// --- 複数フェーズによる生成ループ ---
	phases := []struct {
		name        string
		instruction string
	}{
		{
			name:        "導入",
			instruction: "【フェーズ1: 導入】指定されたテーマについて、MCがリスナーに向けて軽く紹介し、ゲストの二人に振ってください。挨拶は既に終わっている前提で、すぐに本題へのフリから開始してください。",
		},
		{
			name:        "本題",
			instruction: "【フェーズ2: 本題】テーマについてさらに深く、具体的なエピソードや独自の視点を交えて3人で賑やかに雑談してください。お互いの意見に対してリアクションを大きく取り、会話を盛り上げてください。",
		},
		{
			name:        "結び",
			instruction: "【フェーズ3: 結び】楽しい時間はあっという間です。これまでの会話を振り返りつつ、最後は全員で「お相手は、四国めたんと、」「ずんだもんなのだ！」「春日部つむぎでした。バイバーイ！」のように明るく番組を締めくってください。これが最後のセリフです。",
		},
	}

	config := ai.GetDefaultConfig()
	config.Theme = req.Theme
	var fullHistory []ai.ScriptItem

	for _, phase := range phases {
		fmt.Printf("--- Phase: %s 開始 ---\n", phase.name)
		config.PhaseInstruction = phase.instruction
		config.History = fullHistory // 前のフェーズまでの会話を渡す

		// AI 台本生成 (現在のフェーズ分)
		script, err := s.aiClient.GenerateRadioScript(ctx, config)
		if err != nil {
			fmt.Printf("警告: フェーズ %s の生成に失敗した: %v\n", phase.name, err)
			continue
		}

		if len(script) == 0 {
			continue
		}

		// 音声合成用のメタデータ設定
		for i := range script {
			item := &script[i]
			item.SpeakerID = ai.GetSpeakerID(item.Speaker, item.Emotion)
		}

		// 逐次保存と配信 (音声合成はフロントエンドの逐次プリフェッチに任せる)
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

		// 履歴を更新（次のフェーズに引き継ぐ）
		fullHistory = append(fullHistory, script...)
	}

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
