/**
 * Package main は、AI Radio Room のバックエンドサーバーのエントリポイントである。
 * このパッケージでは、アプリケーションの起動に必要な外部リソース（.env, Gemini API, PostgreSQL）の
 * 疎通確認と初期化シーケンスを管理する。
 */
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

/**
 * メインメソッドである。
 * アプリケーションの起動処理を司り、各設定のロードと接続確認を順番に実行する。
 * @param なし
 * @return なし
 */
func main() {
	// --- 1. 設定ファイルの読み込み ---
	loadEnv()

	// --- 2. Gemini APIキーの確認 ---
	// checkGeminiAPIKey() の戻り値である error 構造体を確認する。
	if err := checkGeminiAPIKey(); err != nil {
		// log.Fatalf(format string, v ...any) は、ログを出力した後に os.Exit(1) を呼び出しプログラムを終了させる。
		log.Fatalf("❌ API設定エラー: %v", err)
	}

	// --- 3. データベース接続テスト ---
	// connectAndTestDB() の戻り値である error 構造体を確認する。
	if err := connectAndTestDB(); err != nil {
		log.Fatalf("❌ データベース接続エラー: %v", err)
	}

	// fmt.Println(a ...any) は標準出力に内容を表示し、最後に改行を追加する。
	fmt.Println("--------------------------------------------------")
	fmt.Println("🎉 すべての基盤チェックが成功した。サーバーを起動する準備が整っている。")
}

/**
 * .env ファイルの探索・読み込みメソッドである。
 * 実行ディレクトリに応じて、適切な階層の .env を探索してプロセスにロードする。
 *
 * 内部処理:
 * - 1. `os.Stat(name string) (FileInfo, error)` を使用する。
 *   - `os` パッケージの関数であり、指定されたパスのファイル情報の取得を試みる。ファイルが存在しない場合はエラーを返す。
 * - 2. `godotenv.Load(filenames ...string) error` を使用する。
 *   - `godotenv` パッケージの関数であり、指定されたファイルを読み込み、`os.Setenv` を介して環境変数に展開する。
 * @param なし
 * @return なし
 */
func loadEnv() {
	// 探索するパスの優先順位リスト
	dotEnvs := []string{"../.env", ".env", "../../.env"}

	loaded := false
	for _, path := range dotEnvs {
		// os.Stat() でファイルの物理的な存在を確認する。
		if _, err := os.Stat(path); err == nil {
			// ファイルが存在する場合のみ、godotenv.Load() でロードを試行する。
			if err := godotenv.Load(path); err == nil {
				loaded = true
				break // 最初に見つかった時点でループを終了する。
			}
		}
	}

	if !loaded {
		// log.Println(v ...any) は標準エラー出力にログを記録する。
		log.Println("情報: 有効な .env ファイルが探索パス内に見つからなかった。OS既定の環境変数を使用する。")
	}
}

/**
 * Gemini API キーの検証メソッドである。
 * OS環境変数に API キーが正しく設定されているかを確認する。
 *
 * 内部処理:
 * - `os.Getenv(key string) string` を使用する。
 *   - `os` パッケージの関数であり、OSの環境変数から指定したキーの値を取得する。値が存在しない場合は空文字列を返す。
 * - `fmt.Printf(format string, a ...any)` を使用する。
 *   - 書式指定に従って標準出力に文字列を表示する。
 * @param なし
 * @return err [error] キーが未設定の場合は fmt.Errorf で生成したエラーを、設定されている場合は nil を返す。
 */
func checkGeminiAPIKey() error {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		// fmt.Errorf(format string, a ...any) は書式に従ったエラーメッセージを持つ error 型の値を生成する。
		return fmt.Errorf("GEMINI_API_KEY が環境変数に設定されていない")
	}
	// 安全のため冒頭5文字のみ出力する。
	fmt.Printf("✅ Gemini APIキーの読み込みに成功した: [ %s... ]\n", apiKey[:5])
	return nil
}

/**
 * データベース接続確認メソッドである。
 * 環境変数 DATABASE_URL を使用して PostgreSQL への疎通テストを行う。
 *
 * 内部処理:
 * - 1. `context.Background() Context` を使用する。
 *   - `context` パッケージの関数であり、空のコンテキスト（親コンテキスト）を生成する。通常、メイン処理の開始時などに使用される。
 * - 2. `pgx.Connect(ctx context.Context, connString string) (*pgx.Conn, error)` を使用する。
 *   - `pgx` パッケージの関数であり、指定された接続文字列を用いて DB への物理的な接続を確立し、`*pgx.Conn` 構造体を返す。
 * - 3. `(*pgx.Conn).Close(ctx context.Context) error` を使用する。
 *   - `*pgx.Conn` 構造体のメソッドであり、確立された DB 接続を安全に切断する。`defer` 文と組み合わせて確実に実行させる。
 * - 4. `(*pgx.Conn).QueryRow(ctx context.Context, sql string, args ...any) Row` を使用する。
 *   - `*pgx.Conn` 構造体のメソッドであり、最大1行の結果を返す SQL クエリを実行し、`Row` インターフェースを返す。
 * - 5. `(Row).Scan(dest ...any) error` を使用する。
 *   - `Row` インターフェースのメソッドであり、クエリ結果のカラムの値を指定した変数にコピーする。
 * @param なし
 * @return err [error] 接続失敗、またはクエリ実行失敗時にエラーを返す。
 */
func connectAndTestDB() error {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return fmt.Errorf("DATABASE_URL が環境変数に設定されていない")
	}

	// context.Background() で空のコンテキストを生成する。
	ctx := context.Background()

	// pgx.Connect() でデータベース接続オブジェクト（*pgx.Conn）を生成する。
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("データベースへの接続に失敗した: %w", err)
	}
	// 関数を抜ける際、conn.Close() を呼び出して接続を確実に閉じる。
	defer conn.Close(ctx)

	var greeting string
	// conn.QueryRow() でクエリを実行し、返された Row オブジェクトの Scan() メソッドで結果を変数に格納する。
	err = conn.QueryRow(ctx, "SELECT 'PostgreSQL への接続が正常に確立された'").Scan(&greeting)
	if err != nil {
		return fmt.Errorf("疎通確認クエリの実行に失敗した: %w", err)
	}

	fmt.Printf("✅ %s\n", greeting)
	return nil
}
