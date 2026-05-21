package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Repository はデータベース操作の実装を保持する構造体です。
type Repository struct {
	conn *pgx.Conn
}

// NewRepository は新しい Repository インスタンスを作成して返します。
// 
// 引数:
// - conn: 確立済みの pgx データベース接続
// 戻り値:
// - *Repository: 作成されたリポジトリのポインタ
func NewRepository(conn *pgx.Conn) *Repository {
	return &Repository{conn: conn}
}

// CreateBroadcast は新しい放送セッションをデータベースに登録します。
// 
// 引数:
// - ctx: コンテキスト（タイムアウトやキャンセル管理用）
// - title: 放送のテーマ名
// 戻り値:
// - uuid.UUID: 生成された放送ID
// - error: 失敗時のエラー情報
func (r *Repository) CreateBroadcast(ctx context.Context, title string) (uuid.UUID, error) {
	var id uuid.UUID
	query := `INSERT INTO broadcasts (title) VALUES ($1) RETURNING id`
	err := r.conn.QueryRow(ctx, query, title).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create broadcast: %w", err)
	}
	return id, nil
}

// EndBroadcast は放送セッションの終了時刻を現在時刻で更新します。
// 
// 引数:
// - ctx: コンテキスト
// - id: 終了させる放送セッションのID
// 戻り値:
// - error: 失敗時のエラー情報
func (r *Repository) EndBroadcast(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE broadcasts SET ended_at = NOW() WHERE id = $1`
	_, err := r.conn.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to end broadcast: %w", err)
	}
	return nil
}

// SaveTranscript はAIキャラクターの発言ログをデータベースに保存します。
// 成功時、引数で渡された Transcript 構造体の ID と CreatedAt フィールドがDBの値で更新されます。
// 
// 引数:
// - ctx: コンテキスト
// - t: 保存する発言データ（ポインタ）
// 戻り値:
// - error: 失敗時のエラー情報
func (r *Repository) SaveTranscript(ctx context.Context, t *Transcript) error {
	query := `
		INSERT INTO transcripts (broadcast_id, speaker_id, text, emotion)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	err := r.conn.QueryRow(ctx, query, t.BroadcastID, t.SpeakerID, t.Text, t.Emotion).Scan(&t.ID, &t.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to save transcript: %w", err)
	}
	return nil
}

// GetTranscriptsByBroadcastID は特定の放送セッションに紐づくすべての発言ログを取得します。
// 結果は発言日時（created_at）の昇順でソートされます。
// 
// 引数:
// - ctx: コンテキスト
// - broadcastID: 取得対象の放送セッションID
// 戻り値:
// - []Transcript: 取得された発言ログのリスト
// - error: 失敗時のエラー情報
func (r *Repository) GetTranscriptsByBroadcastID(ctx context.Context, broadcastID uuid.UUID) ([]Transcript, error) {
	query := `
		SELECT id, broadcast_id, speaker_id, text, emotion, created_at
		FROM transcripts
		WHERE broadcast_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.conn.Query(ctx, query, broadcastID)
	if err != nil {
		return nil, fmt.Errorf("failed to query transcripts: %w", err)
	}
	defer rows.Close()

	var transcripts []Transcript
	for rows.Next() {
		var t Transcript
		err := rows.Scan(&t.ID, &t.BroadcastID, &t.SpeakerID, &t.Text, &t.Emotion, &t.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transcript: %w", err)
		}
		transcripts = append(transcripts, t)
	}

	return transcripts, nil
}
