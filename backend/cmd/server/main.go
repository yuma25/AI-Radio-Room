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
	"context"  // 使用箇所: 各初期化処理。目的: DB や AI クライアントのライフサイクル管理のため。
	"fmt"      // 使用箇所: main。目的: 起動メッセージの表示のため。
	"log"      // 使用箇所: main。目的: 致命的エラー時の強制終了のため。
	"net/http" // 使用箇所: main。目的: HTTP サーバーの起動のため。
	"os"       // 使用箇所: connectDB。目的: 環境変数の取得のため。

	"backend/internal/ai"  // 使用箇所: main。目的: AI クライアントの初期化のため。
	"backend/internal/api" // 使用箇所: main。目的: API サーバーの生成とルーティングのため。
	"backend/internal/db"  // 使用箇所: main。目的: リポジトリの生成のため。

	"github.com/jackc/pgx/v5"  // 使用箇所: connectDB。目的: PostgreSQL への接続のため。
	"github.com/joho/godotenv" // 使用箇所: loadEnv。目的: .env のロードのため。
)

/**
 * メインメソッドである。
 *
 * 内部処理:
 * - 1. `loadEnv()` を呼び出し、設定をロードする。
 * - 2. `ai.NewGeminiClient(ctx)` を呼び出し、AI クライアントを初期化する。
 * - 3. `connectDB(ctx)` を呼び出し、データベース接続を確立する。
 * - 4. `db.NewRepository(conn)` を呼び出し、リポジトリを生成する。
 * - 5. `api.NewServer(aiClient, repo)` を呼び出し、API サーバー構造体を生成する。
 * - 6. `http.ListenAndServe(addr, handler)` を呼び出し、8080 ポートで HTTP サーバーを起動する。
 * @param なし
 * @return なし
 */
func main() {
	// --- 1. 設定のロード ---
	loadEnv()

	ctx := context.Background()

	// --- 2. 基盤コンポーネントの初期化 ---
	aiClient, err := ai.NewGeminiClient(ctx)
	if err != nil {
		log.Fatalf("❌ AI クライアント初期化失敗: %v", err)
	}
	defer aiClient.Close()

	conn, err := connectDB(ctx)
	if err != nil {
		log.Fatalf("❌ データベース接続失敗: %v", err)
	}
	defer conn.Close(ctx)

	repo := db.NewRepository(conn)

	// --- 3. API サーバーの起動 ---
	server := api.NewServer(aiClient, repo)
	addr := ":8080"
	fmt.Printf("🚀 AI Radio Room Backend 起動完了 [ %s ]\n", addr)

	if err := http.ListenAndServe(addr, server.Routes()); err != nil {
		log.Fatalf("❌ サーバー停止: %v", err)
	}
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
 * connectDB は、データベースへの接続を確立して返す補助関数である。
 *
 * 内部処理:
 * - 1. 環境変数 `DATABASE_URL` を取得する。
 * - 2. `pgx.Connect(ctx, dbURL)` を呼び出して物理接続を行う。
 * @param ctx [context.Context]
 * @return [*pgx.Conn] 接続オブジェクト
 * @return [error]
 */
func connectDB(ctx context.Context) (*pgx.Conn, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL が未設定である")
	}
	return pgx.Connect(ctx, dbURL)
}
