/**
 * Package api 内の放送関連ハンドラーを定義するファイルである。
 */
package api

import (
	"encoding/json" // 使用箇所: handleStartBroadcast。目的: リクエストボディのパースおよびエラーレスポンスの生成のため。
	"fmt"           // 使用箇所: 各ハンドラー。目的: SSE フォーマットの構築やエラーメッセージの整形のため。
	"net/http"      // 使用箇所: 各ハンドラー。目的: ResponseWriter や Request の操作のため。
	"time"          // 使用箇所: SSE 送信。目的: 送信間隔の調整（デモ用）やタイムスタンプ管理のため。

	"backend/internal/ai" // 使用箇所: handleStartBroadcast。目的: 台本生成リクエストのため。
	"backend/internal/db" // 使用箇所: handleStartBroadcast。目的: 発言ログの構造体利用のため。
)

/**
 * broadcastRequest は、放送開始リクエストのボディを表す構造体である。
 */
type broadcastRequest struct {
	Theme string `json:"theme"` // 放送のメインテーマ
}

/**
 * handleStartBroadcast は、新しい放送を開始し、台本を生成して SSE で配信するハンドラーである。
 *
 * 内部処理:
 * - 1. `json.NewDecoder(r.Body).Decode(v any) error` を使用してリクエストをパースする。
 * - 2. `s.repo.CreateBroadcast(ctx, theme)` を呼び出し、DB に放送レコードを作成する。
 * - 3. `s.aiClient.GenerateRadioScript(ctx, config)` を呼び出し、AI に台本を生成させる。
 * - 4. Response Header に `Content-Type: text/event-stream` を設定し、SSE モードに移行する。
 * - 5. 生成された台本をループで回し、1行ごとに以下を行う：
 *   - DB に保存 (`s.repo.SaveTranscript`)。
 *   - SSE フォーマットでクライアントに書き出し (`fmt.Fprintf(w, "data: %s\n\n", json)`)。
 *   - `(http.Flusher).Flush()` を呼び出し、即座にネットワークへ送出する。
 * - 6. 最後に放送終了処理 (`s.repo.EndBroadcast`) を行う。
 * @param w [http.ResponseWriter] レスポンス書き出し先
 * @param r [*http.Request] リクエスト情報
 * @return なし
 */
func (s *Server) handleStartBroadcast(w http.ResponseWriter, r *http.Request) {
	var req broadcastRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "無効なリクエストボディである", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// --- 1. 放送セッションの作成 ---
	broadcastID, err := s.repo.CreateBroadcast(ctx, req.Theme)
	if err != nil {
		http.Error(w, fmt.Sprintf("放送の作成に失敗した: %v", err), http.StatusInternalServerError)
		return
	}

	// --- 2. AI 台本生成 ---
	config := ai.GetDefaultConfig()
	config.Theme = req.Theme
	script, err := s.aiClient.GenerateRadioScript(ctx, config)
	if err != nil {
		http.Error(w, fmt.Sprintf("台本生成に失敗した: %v", err), http.StatusInternalServerError)
		return
	}

	// --- 3. SSE 設定 ---
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "ストリーミングがサポートされていない", http.StatusInternalServerError)
		return
	}

	// --- 4. 逐次保存と配信 ---
	for _, item := range script {
		// DB 保存
		transcript := db.Transcript{
			BroadcastID: broadcastID,
			SpeakerID:   item.Speaker,
			Text:        item.Text,
			Emotion:     item.Emotion,
		}
		if err := s.repo.SaveTranscript(ctx, &transcript); err != nil {
			fmt.Printf("警告: 発言の保存に失敗した: %v\n", err)
			// 保存失敗しても配信は継続する
		}

		// クライアントへ送信
		data, _ := json.Marshal(item)
		fmt.Fprintf(w, "data: %s\n\n", string(data))
		flusher.Flush()

		// キャラクターの掛け合い感を出すためのウェイト（実運用では不要かもしれないが、デモ効果のため）
		time.Sleep(1500 * time.Millisecond)
	}

	// --- 5. 放送終了 ---
	if err := s.repo.EndBroadcast(ctx, broadcastID); err != nil {
		fmt.Printf("警告: 放送終了処理に失敗した: %v\n", err)
	}

	// SSE 終了の合図
	fmt.Fprintf(w, "event: close\ndata: 放送が終了した\n\n")
	flusher.Flush()
}

/**
 * handleListBroadcasts は、過去の放送履歴一覧を JSON で返すハンドラーである。
 * ※現在はダミーまたは簡易的な実装とする。
 *
 * 内部処理:
 * - 1. 本来はリポジトリから一覧を取得するが、Issue #8 の主目的は SSE のため、一旦空レスポンスまたはエラーを返さない最低限の実装とする。
 * @param w [http.ResponseWriter]
 * @param r [*http.Request]
 * @return なし
 */
func (s *Server) handleListBroadcasts(w http.ResponseWriter, r *http.Request) {
	// TODO: repo に全放送取得メソッドを追加した後に実装する
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]string{})
}
