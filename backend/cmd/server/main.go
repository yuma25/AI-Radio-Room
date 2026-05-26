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

	"backend/internal/ai"       // 使用箇所: main。目的: AI クライアントの初期化のため。
	"backend/internal/api"      // 使用箇所: main。目的: API サーバーの生成とルーティングのため。
	"backend/internal/db"       // 使用箇所: main。目的: リポジトリの生成のため。
	"backend/internal/voicevox" // 使用箇所: main。目的: VOICEVOX クライアントの初期化のため。

	"github.com/jackc/pgx/v5"  // 使用箇所: connectDB。目的: PostgreSQL への接続のため。
	"github.com/joho/godotenv" // 使用箇所: loadEnv。目的: .env のロードのため。
)

/**
 * メインメソッドである。
 * プロジェクトの起動シーケンスを制御し、すべての基盤コンポーネントが正常に準備されたことを確認した上でサーバーを開始する。
 *
 * 内部処理:
 * - 1. `loadEnv()` を呼び出し、設定をロードする。
 * - 2. `context.Background()` を呼び出し、アプリケーションの基底となるコンテキストを生成する。
 *      - [技術的背景]: Go における Context は、リクエストのキャンセル、タイムアウト、およびスコープ内の値（トレースID等）を伝播させるための標準的な手段である。
 *      - ここで使用する `Background()` は、サーバーの生存期間全体にわたる「根（Root）」としての役割を持ち、初期化中の SDK 呼び出しやコネクションの管理に利用される。
 * - 3. `ai.NewGeminiClient(ctx)` を呼び出し、Google Gemini API との接続インターフェースを初期化する。
 *      - SDK 内部でのネットワーク接続確立や認証情報の検証に、生成した基底コンテキストを渡す。
 * - 4. `connectDB(ctx)` を呼び出し、PostgreSQL データベースへの物理接続を確立する。
 *      - 接続タイムアウトなどの制御をコンテキスト経由で行う。
 * - 5. `db.NewRepository(conn)` を呼び出し、データベース接続オブジェクト（conn）をラップしたリポジトリを生成する。
 *      - これにより、上位レイレイヤーは SQL の詳細を意識せずにデータ操作が可能となる（依存性の注入）。
 * - 6. `api.NewServer(aiClient, repo, vvClient)` を呼び出し、ビジネスロジックと HTTP インターフェースを橋渡しする API サーバー構造体を生成する。
 * - 7. `http.ListenAndServe(addr, handler)` を呼び出し、指定されたアドレス（:8080）で HTTP サーバーを起動する。
 *      - エラーが発生した場合は、`log.Fatalf` によりプロセスを異常終了させ、システム管理者に通知する。
 * @param なし
 * @return なし
 */
func main() {
	// --- 1. 設定のロード ---
	loadEnv()

	// サーバー全体のライフサイクルを管理するための基底コンテキストを生成する。
	// 初期化フェーズにおいて、各 SDK や DB 接続のタイムアウト・キャンセルの制御に使用される。
	ctx := context.Background()

	// --- 2. 基盤コンポーネントの初期化 ---
	// [SDK] Google Gemini API クライアントを初期化する。
	aiClient, err := ai.NewGeminiClient(ctx)
	if err != nil {
		log.Fatalf("❌ AI クライアント初期化失敗: %v", err)
	}
	// リソースを安全に解放するため、プログラム終了時に Close を実行するよう予約する。
	defer aiClient.Close()

	// VOICEVOX クライアントの初期化。
	vvURL := os.Getenv("VOICEVOX_URL")
	if vvURL == "" {
		vvURL = "http://localhost:50021" // デフォルト値
	}
	vvClient := voicevox.NewClient(vvURL)

	// データベース接続の確立。
	conn, err := connectDB(ctx)
	if err != nil {
		log.Fatalf("❌ データベース接続失敗: %v", err)
	}
	// コネクションのクローズを予約する。
	defer conn.Close(ctx)

	// [内部] データベース操作用のリポジトリを初期化する。
	// これにより、上位レイヤーは SQL の詳細を意識せずにデータ操作が可能となる（依存性の注入）。
	repo := db.NewRepository(conn)

	// --- 3. API サーバーの起動 ---
	// API サーバーに AI クライアント、DB リポジトリ、VOICEVOX クライアントを注入する。
	server := api.NewServer(aiClient, repo, vvClient)
	addr := ":8080"
	fmt.Printf("🚀 AI Radio Room Backend 起動完了 [ %s ]\n", addr)

	// HTTP サーバーの開始。
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
