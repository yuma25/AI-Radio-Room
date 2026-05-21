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
	// loadEnv() を呼び出し、.env ファイルから環境変数をプロセスにロードする。
	loadEnv()

	// --- 2. Gemini APIキーの確認 ---
	// checkGeminiAPIKey() を呼び出し、戻り値である error 構造体を確認する。
	if err := checkGeminiAPIKey(); err != nil {
		// log.Fatalf(format string, v ...any) は、log パッケージの関数である。
		// ログを出力した直後に os.Exit(1) を呼び出し、プログラムを異常終了させる役割を持つ。
		log.Fatalf("❌ API設定エラー: %v", err)
	}

	// --- 3. データベース接続テスト ---
	// connectAndTestDB() を呼び出し、戻り値である error 構造体を確認する。
	if err := connectAndTestDB(); err != nil {
		// log.Fatalf() により、接続失敗時はここでプログラムを停止させる。
		log.Fatalf("❌ データベース接続エラー: %v", err)
	}

	// fmt.Println(a ...any) は、fmt パッケージの関数である。
	// 標準出力に指定された内容を表示し、最後に改行を追加する。
	fmt.Println("--------------------------------------------------")
	fmt.Println("🎉 すべての基盤チェックが成功した。サーバーを起動する準備が整っている。")
}

/**
 * .env ファイルの読み込みメソッドである。
 * プロジェクトのルートディレクトリに存在する .env ファイルをプロセスにロードする。
 *
 * 内部処理:
 * - 1. `godotenv.Load(filenames ...string) error` を使用する。
 *   - `godotenv` パッケージの関数であり、指定されたファイルを読み込み、`os.Setenv` を介して環境変数に展開する。
 *   - 常にプロジェクトルートでの実行を前提とし、直下の ".env" のみを対象とする。
 * @param なし
 * @return なし
 */
func loadEnv() {
	// godotenv.Load() を呼び出し、ルート直下の設定ファイルを読み込む。
	if err := godotenv.Load(".env"); err != nil {
		// log.Println(v ...any) は、log パッケージの関数である。
		// 標準エラー出力にタイムスタンプと共にログを記録する。
		log.Println("情報: .env ファイルが見つからなかった。OSの環境変数を使用する。")
	}
}

/**
 * Gemini API キーの検証メソッドである。
 * OS環境変数に API キーが正しく設定されているかを確認する。
 *
 * 内部処理:
 * - 1. `os.Getenv(key string) string` を使用する。
 *   - `os` パッケージの関数であり、OSの環境変数から指定したキーの値を取得する。値が存在しない場合は空文字列を返す。
 * - 2. `fmt.Printf(format string, a ...any)` を使用する。
 *   - 書式指定に従って標準出力に文字列を表示する。
 * @param なし
 * @return err [error] キーが未設定の場合は fmt.Errorf で生成したエラーを、設定されている場合は nil を返す。
 */
func checkGeminiAPIKey() error {
	// os.Getenv() を呼び出し、GEMINI_API_KEY の値を取得する。
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		// fmt.Errorf(format string, a ...any) は、fmt パッケージの関数である。
		// 書式に従ったメッセージを持つ、新しい error 型のインスタンスを生成して返す。
		return fmt.Errorf("GEMINI_API_KEY が環境変数に設定されていない")
	}
	// fmt.Printf() を呼び出し、安全のためにキーの冒頭部分のみを出力する。
	fmt.Printf("✅ Gemini APIキーの読み込みに成功した: [ %s... ]\n", apiKey[:5])
	return nil
}

/**
 * データベース接続確認メソッドである。
 * 環境変数 DATABASE_URL を使用して PostgreSQL への疎通テストを行う。
 *
 * 内部処理:
 * - 1. `os.Getenv(key string) string` を使用し、接続文字列を取得する。
 * - 2. `context.Background() Context` を使用する。
 *   - `context` パッケージの関数であり、空のコンテキスト（親コンテキスト）を生成する。
 * - 3. `pgx.Connect(ctx context.Context, connString string) (*pgx.Conn, error)` を使用する。
 *   - `pgx` パッケージの関数であり、指定された接続文字列で DB への物理的なネットワーク接続を確立し、`*pgx.Conn` 構造体を返す。
 * - 4. `(*pgx.Conn).Close(ctx context.Context) error` を使用する。
 *   - `*pgx.Conn` 構造体のメソッドであり、確立された DB 接続を安全に切断する。
 * - 5. `(*pgx.Conn).QueryRow(ctx context.Context, sql string, args ...any) Row` を使用する。
 *   - `*pgx.Conn` 構造体のメソッドであり、最大1行の結果を返す SQL クエリを実行し、`Row` インターフェースを返す。
 * - 6. `(Row).Scan(dest ...any) error` を使用する。
 *   - `Row` インターフェースのメソッドであり、クエリ結果のカラムの値を指定した変数にコピーする。
 * @param なし
 * @return err [error] 接続失敗、またはクエリ実行失敗時にエラーを返す。
 */
func connectAndTestDB() error {
	// os.Getenv() を呼び出し、DATABASE_URL を取得する。
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return fmt.Errorf("DATABASE_URL が環境変数に設定されていない")
	}

	// context.Background() を呼び出し、空のコンテキストを生成する。
	ctx := context.Background()

	// pgx.Connect() を呼び出し、データベース接続オブジェクト（*pgx.Conn）を生成する。
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		// %w フォーマットを使用して、元のエラーをラップして返す。
		return fmt.Errorf("データベースへの接続に失敗した: %w", err)
	}
	// defer 文を用いて、関数終了時に conn.Close() メソッドを確実に呼び出し、接続を閉じる。
	defer conn.Close(ctx)

	var greeting string
	// conn.QueryRow() を呼び出してクエリを実行し、戻り値である Row オブジェクトの Scan() メソッドを呼び出す。
	err = conn.QueryRow(ctx, "SELECT 'PostgreSQL への接続が正常に確立された'").Scan(&greeting)
	if err != nil {
		return fmt.Errorf("疎通確認クエリの実行に失敗した: %w", err)
	}

	// fmt.Printf() で接続成功メッセージを表示する。
	fmt.Printf("✅ %s\n", greeting)
	return nil
}
