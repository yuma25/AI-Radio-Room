/**
 * Package db は、PostgreSQL データベースとの接続管理および永続化層（リポジトリ）を提供するパッケージである。
 *
 * 【パッケージの構成要素】
 * - 1. `Repository`: PostgreSQL への物理接続を保持し、SQL クエリの実行をカプセル化する中心的な構造体。
 *
 * 【要素間の関係】
 * - `Repository` はインポートされた `pgx/v5` パッケージに依存しており、すべてのメソッドでデータベース接続（`conn`）を利用する。
 * - `CreateBroadcast` や `SaveTranscript` といったメソッドは、`uuid` パッケージを利用して一意識別子を扱う。
 * - 各操作は `context` パッケージを利用してタイムアウトやキャンセルの制御が行われる。
 */
package db

import (
	"context" // 使用箇所: Repository の全メソッド。目的: DB 操作のタイムアウトおよびキャンセル管理のため。
	"fmt"     // 使用箇所: 各メソッドのエラーハンドリング。目的: エラーメッセージの装飾とラップのため。

	"github.com/google/uuid"          // 使用箇所: CreateBroadcast, GetTranscriptsByBroadcastID。目的: UUID 型の引数および戻り値を扱うため。
	"github.com/jackc/pgx/v5/pgxpool" // 使用箇所: Repository 構造体、NewRepository 関数。目的: PostgreSQL への低レベルな接続およびクエリ実行のため。
)

/**
 * Repository は、データベース操作の実装を保持する構造体である。
 */
type Repository struct {
	conn *pgxpool.Pool // [SDK] PostgreSQL へのアクティブな接続オブジェクト
}

/**
 * NewRepository は、新しい Repository インスタンスを作成して返す。
 *
 * 内部処理:
 * - 1. 引数として渡された `*pgx.Conn` を構造体のフィールドに格納して返す。
 * @param conn [*pgx.Conn] 確立済みのデータベース接続
 * @return repo [*Repository] 作成されたリポジトリのポインタ
 */
func NewRepository(conn *pgxpool.Pool) *Repository {
	return &Repository{conn: conn}
}

/**
 * CreateBroadcast は、新しい放送セッションをデータベースに登録する。
 *
 * 内部処理:
 * - 1. `INSERT INTO broadcasts ...` という SQL 文を文字列として定義する。
 * - 2. `(*pgx.Conn).QueryRow(ctx context.Context, sql string, args ...any) pgx.Row` を使用する。
 *   - `pgx` パッケージのメソッドであり、指定されたコンテキストと引数を用いて SQL を実行し、単一の行（RETURNING id）を返す。
 * - 3. `(pgx.Row).Scan(dest ...any) error` を使用する。
 *   - `pgx` パッケージのメソッドであり、QueryRow で取得した結果のカラム値を Go の変数（id）にコピーする。
 * @param ctx [context.Context] 操作用のコンテキスト
 * @param title [string] 放送のテーマ名
 * @return id [uuid.UUID] 生成された放送ID
 * @return err [error] 失敗時のエラー情報
 */
func (r *Repository) CreateBroadcast(ctx context.Context, title string) (uuid.UUID, error) {
	var id uuid.UUID
	query := `INSERT INTO broadcasts (title) VALUES ($1) RETURNING id`
	err := r.conn.QueryRow(ctx, query, title).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create broadcast: %w", err)
	}
	return id, nil
}

/**
 * EndBroadcast は、放送セッションの終了時刻を現在時刻で更新する。
 *
 * 内部処理:
 * - 1. `UPDATE broadcasts SET ended_at = NOW() ...` という SQL 文を文字列として定義する。
 * - 2. `(*pgx.Conn).Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)` を使用する。
 *   - `pgx` パッケージのメソッドであり、結果を返さない（または件数のみを返す）更新クエリを実行する。
 * @param ctx [context.Context] コンテキスト
 * @param id [uuid.UUID] 終了させる放送セッションのID
 * @return err [error] 失敗時のエラー情報
 */
func (r *Repository) EndBroadcast(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE broadcasts SET ended_at = NOW() WHERE id = $1`
	_, err := r.conn.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to end broadcast: %w", err)
	}
	return nil
}

/**
 * SaveTranscript は、AIキャラクターの発言ログをデータベースに保存する。
 *
 * 内部処理:
 * - 1. `INSERT INTO transcripts ...` という SQL 文を文字列として定義する。
 * - 2. `(*pgx.Conn).QueryRow(ctx context.Context, sql string, args ...any) pgx.Row` を使用する。
 *   - `pgx` パッケージのメソッドであり、発言データの挿入と同時に自動生成された `id` と `created_at` を取得する。
 * - 3. `(pgx.Row).Scan(dest ...any) error` を使用する。
 *   - `pgx` パッケージのメソッドであり、取得した値を引数の構造体（t）の各フィールドにコピーする。
 * @param ctx [context.Context] コンテキスト
 * @param t [*Transcript] 保存する発言データ（ポインタ）
 * @return err [error] 失敗時のエラー情報
 */
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

/**
 * GetTranscriptsByBroadcastID は、特定の放送セッションに紐づくすべての発言ログを取得する。
 *
 * 内部処理:
 * - 1. `SELECT ... FROM transcripts ...` という SQL 文を文字列として定義する。
 * - 2. `(*pgx.Conn).Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)` を使用する。
 *   - `pgx` パッケージのメソッドであり、複数行の結果を返す検索クエリを実行する。
 * - 3. `defer (pgx.Rows).Close()` を使用する。
 *   - `pgx` パッケージのメソッドであり、関数終了時に結果セットのリソースを確実に解放する。
 * - 4. `(pgx.Rows).Next() bool` を使用する。
 *   - `pgx` パッケージのメソッドであり、結果セット内の次のレコードが存在するかを確認する。
 * - 5. `(pgx.Rows).Scan(dest ...any) error` を使用する。
 *   - `pgx` パッケージのメソッドであり、現在のレコードの値を構造体変数にコピーする。
 * @param ctx [context.Context] コンテキスト
 * @param broadcastID [uuid.UUID] 取得対象の放送セッションID
 * @return transcripts [[]Transcript] 取得された発言ログのリスト
 * @return err [error] 失敗時のエラー情報
 */
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
