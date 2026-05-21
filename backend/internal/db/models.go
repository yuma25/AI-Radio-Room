package db

import (
	"time"

	"github.com/google/uuid"
)

// Broadcast は放送セッションの情報を保持する構造体です。
// DBの broadcasts テーブルに対応しています。
//
// フィールド:
// - ID: 放送を一意に識別するUUID
// - Title: 放送のテーマやタイトル
// - StartedAt: 放送が開始された日時
// - EndedAt: 放送が終了した日時（放送中の場合は nil）
type Broadcast struct {
	ID        uuid.UUID  `json:"id"`
	Title     string     `json:"title"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
}

// Transcript は放送内の個別の発言ログを保持する構造体です。
// DBの transcripts テーブルに対応しています。
//
// フィールド:
// - ID: 発言を一意に識別するUUID
// - BroadcastID: 紐づく放送セッションのID
// - SpeakerID: 発言者（mc, guest-a, guest-b）
// - Text: 発言内容
// - Emotion: 発言時の感情（normal, excited, calm）
// - CreatedAt: 発言が記録された日時
type Transcript struct {
	ID          uuid.UUID `json:"id"`
	BroadcastID uuid.UUID `json:"broadcast_id"`
	SpeakerID   string    `json:"speaker_id"`
	Text        string    `json:"text"`
	Emotion     string    `json:"emotion"`
	CreatedAt   time.Time `json:"created_at"`
}
