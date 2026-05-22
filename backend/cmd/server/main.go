/**
 * Package main は、AI Radio Room のバックエンドサーバーのエントリポイントである。
 * このパッケージはサーバーの起動準備と、各基盤コンポーネント間のオーケストレーションを担当する。
 *
 * 【パッケージの構成要素】
 * - 1. `main`: 全体の起動フロー（設定ロード -> 基盤チェック -> サーバー開始）を管理する関数。
 * - 2. `loadEnv`: `godotenv` を使用した環境変数の展開。
 * - 3. `checkGeminiAPIKey`: AI 連携用の設定バリデーション。
 * - 4. `connectAndTestDB`: `pgx` を使用したデータベース接続確認。
 *
 * 【要素間の関係】
 * - `main` 関数はすべての補助関数を同期的に呼び出し、一つでも失敗した場合はプロセスを異常終了（log.Fatalf）させる。
 * - 外部パッケージ（pgx, godotenv, os 等）は、それぞれ「DB接続」「設定管理」「環境変数アクセス」という明確な役割でインポートされている。
 */
package main

import (
	"context" // 使用箇所: connectAndTestDB。目的: データベースへの疎通確認リクエストのタイムアウト管理のため。
	"fmt"     // 使用箇所: main, checkGeminiAPIKey。目的: 標準出力への進捗表示およびエラーメッセージ構築のため。
	"log"     // 使用箇所: main。目的: 致命的なエラーが発生した際にエラー内容を記録し、プロセスを異常終了させるため。
	"os"      // 使用箇所: checkGeminiAPIKey, connectAndTestDB。目的: 環境変数から API キーや DB 接続情報を取得するため。

	"github.com/jackc/pgx/v5"  // 使用箇所: connectAndTestDB。目的: PostgreSQL へのネットワーク接続を確立し、疎通テストを行うため。
	"github.com/joho/godotenv" // 使用箇所: loadEnv。目的: ローカル開発環境の `.env` ファイルを解析し、環境変数としてロードするため。
)

/**
 * メインメソッドである。
 *
 * 内部処理:
 * - 1. `loadEnv()` を呼び出す。
 *   - 本パッケージ内の補助関数であり、環境変数のロードを行う。
 * - 2. `checkGeminiAPIKey()` を呼び出す。
 *   - 本パッケージ内の補助関数であり、AI 設定のバリデーションを行う。
 * - 3. `connectAndTestDB()` を呼び出す。
 *   - 本パッケージ内の補助関数であり、データベースの死活監視を行う。
 * - 4. `log.Fatalf(format string, v ...any)` を使用する。
 *   - `log` パッケージの関数であり、エラーメッセージを標準エラー出力した後に `os.Exit(1)` を呼び出して終了させる役割を持つ。
 * - 5. `fmt.Println(a ...any) (n int, err error)` を使用する。
 *   - `fmt` パッケージの関数であり、最終的なステータスを標準出力に書き出す役割を持つ。
 * @param なし
 * @return なし
 */
func main() {
	// --- 1. 設定ファイルの読み込み ---
	loadEnv()

	// --- 2. Gemini APIキーの確認 ---
	if err := checkGeminiAPIKey(); err != nil {
		log.Fatalf("❌ API設定エラー: %v", err)
	}

	// --- 3. データベース接続テスト ---
	if err := connectAndTestDB(); err != nil {
		log.Fatalf("❌ データベース接続エラー: %v", err)
	}

	fmt.Println("--------------------------------------------------")
	fmt.Println("🎉 すべての基盤チェックが成功した。サーバーを起動する準備が整っている。")
}

/**
 * .env ファイルの読み込みメソッドである。
 * プロジェクトルートまたは現在のディレクトリから .env を探し、プロセスにロードする。
 *
 * 内部処理:
 * - 1. `godotenv.Load(filenames ...string) error` を使用する。
 *   - カレントディレクトリの `.env` を試行し、失敗した場合は 1 階層上の `../.env`（プロジェクトルート）を試行する。
 * - 2. `log.Println(v ...any)` を使用する。
 *   - どちらのパスでもファイルが見つからない場合にのみ情報を出力する。
 * @param なし
 * @return なし
 */
func loadEnv() {
	// まずカレントディレクトリを試す
	if err := godotenv.Load(".env"); err == nil {
		return
	}
	// 失敗した場合はプロジェクトルート（一つ上）を試す
	if err := godotenv.Load("../.env"); err == nil {
		return
	}
	log.Println("情報: .env ファイルが見つからなかった。OSの環境変数を使用する。")
}

/**
 * Gemini API キーの検証メソッドである。
 *
 * 内部処理:
 * - 1. `os.Getenv(key string) string` を使用する。
 *   - `os` パッケージの関数であり、環境変数 `GEMINI_API_KEY` の値を取得する。
 * - 2. `fmt.Errorf(format string, a ...any) error` を使用する。
 *   - `fmt` パッケージの関数であり、キーが存在しない場合にエラーオブジェクトを作成して返す。
 * - 3. `fmt.Printf(format string, a ...any) (n int, err error)` を使用する。
 *   - `fmt` パッケージの関数であり、成功時にキーのマスク済みプレビューを出力する。
 * @param なし
 * @return err [error] キーが未設定の場合はエラーを、正常な場合は nil を返す。
 */
func checkGeminiAPIKey() error {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("GEMINI_API_KEY が環境変数に設定されていない")
	}
	fmt.Printf("✅ Gemini APIキーの読み込みに成功した: [ %s... ]\n", apiKey[:5])
	return nil
}

/**
 * データベース接続確認メソッドである。
 *
 * 内部処理:
 * - 1. `os.Getenv(key string) string` を使用して、接続文字列 `DATABASE_URL` を取得する。
 * - 2. `context.Background() context.Context` を使用する。
 *   - `context` パッケージの関数であり、接続操作のルートとなる空のコンテキストを生成する。
 * - 3. `pgx.Connect(ctx context.Context, connString string) (*pgx.Conn, error)` を使用する。
 *   - `github.com/jackc/pgx/v5` パッケージの関数であり、指定された URL で PostgreSQL との TCP 接続を確立する。
 * - 4. `defer (*pgx.Conn).Close(ctx context.Context) error` を使用する。
 *   - `pgx` パッケージのメソッドであり、関数のライフサイクル終了時に DB 接続を確実に切断する。
 * - 5. `(*pgx.Conn).QueryRow(ctx context.Context, sql string, args ...any) pgx.Row` を使用する。
 *   - `pgx` パッケージのメソッドであり、疎通確認用の SELECT クエリを実行する。
 * - 6. `(pgx.Row).Scan(dest ...any) error` を使用する。
 *   - `pgx` パッケージのメソッドであり、クエリの実行結果を変数に詰め込む。
 * @param なし
 * @return err [error] 接続失敗、またはクエリ実行失敗時にエラーを返す。
 */
func connectAndTestDB() error {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return fmt.Errorf("DATABASE_URL が環境変数に設定されていない")
	}

	ctx := context.Background()

	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("データベースへの接続に失敗した: %w", err)
	}
	defer conn.Close(ctx)

	var greeting string
	err = conn.QueryRow(ctx, "SELECT 'PostgreSQL への接続が正常に確立された'").Scan(&greeting)
	if err != nil {
		return fmt.Errorf("疎通確認クエリの実行に失敗した: %w", err)
	}

	fmt.Printf("✅ %s\n", greeting)
	return nil
}
