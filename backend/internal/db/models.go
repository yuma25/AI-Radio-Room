/**
 * Package db は、PostgreSQL データベースとの接続管理および永続化層（リポジトリ）を提供するパッケージである。
 *
 * 【パッケージの構成要素】
 * - 1. `Broadcast`: 放送セッション（テーマ、開始・終了時刻）を表現するデータモデル。
 * - 2. `Transcript`: 放送内の 1 つの発言データを表現するデータモデル。
 * - 3. `Repository`: 上記モデルを DB に対して保存・取得するための永続化ロジック。
 *
 * 【要素間の関係】
 * - `Repository` は `pgx` ドライバを介して SQL クエリを実行し、DB レコードを `Broadcast` / `Transcript` 構造体に変換（スキャン）する。
 * - `Transcript` は常に特定の `Broadcast.ID` を外部キーとして持ち、特定の放送回に紐づく形で管理される。
 * - 外部の `time` パッケージは時刻記録に、`uuid` パッケージは主キーの識別に使用される。
 */
package db

import (
	"time" // 使用箇所: Broadcast, Transcript 構造体。目的: 放送開始・終了時間、および発言時刻を扱うため。

	"github.com/google/uuid" // 使用箇所: 各構造体の ID フィールド、Repository。目的: データベース内で一意な識別子（UUID）を生成・識別するため。
)

/**
 * Broadcast は、一つの放送セッションの情報を保持する構造体である。
 * データベースの `broadcasts` テーブルの 1 レコードに対応する。
 */
type Broadcast struct {
	ID        uuid.UUID  `json:"id"`                 // 放送を一意に識別するUUID。プライマリキーとして使用される。
	Title     string     `json:"title"`              // 放送のテーマやタイトル。ユーザーの入力に基づいて設定される。
	StartedAt time.Time  `json:"started_at"`         // 放送が開始された日時。DB 側で挿入時に自動決定される。
	EndedAt   *time.Time `json:"ended_at,omitempty"` // 放送が終了した日時。終了メソッドが呼ばれるまでは nil となる。
}

/**
 * Transcript は、放送内での個別の発言データを保持する構造体である。
 * データベースの `transcripts` テーブルの 1 レコードに対応する。
 */
type Transcript struct {
	ID          uuid.UUID `json:"id"`           // 発言を一意に識別するUUID。プライマリキーとして使用される。
	BroadcastID uuid.UUID `json:"broadcast_id"` // 紐づく放送セッションのID。`broadcasts.id` への外部キー制約を持つ。
	SpeakerID   string    `json:"speaker_id"`   // 発言したキャラクターの役割名 (mc, guest-a, guest-b) を格納する。
	Text        string    `json:"text"`         // 発言内容のテキスト本文。AI によって生成された文字列が格納される。
	Emotion     string    `json:"emotion"`      // 発言時の感情トーン。フロントエンドでの表情切り替え等に使用される。
	CreatedAt   time.Time `json:"created_at"`   // 発言が記録された日時。順序制御のために DB 側で自動設定される。
}
