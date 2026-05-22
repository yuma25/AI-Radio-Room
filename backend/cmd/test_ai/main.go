/**
 * Package main は、Gemini AI による台本生成ロジックの単体テスト用エントリポイントである。
 *
 * 【パッケージの構成要素】
 * - 1. `main`: AI クライアントの初期化、動的な設定変更、および生成結果の表示を行うテスト関数。
 *
 * 【要素間の関係】
 * - `main` 関数は `internal/ai` パッケージの `GeminiClient` および `GeneratorConfig` に依存している。
 * - `ai.NewGeminiClient` でクライアントを生成し、`ai.GetDefaultConfig` で取得した設定を加工して `GenerateRadioScript` に渡すフローとなっている。
 * - インポートされた `godotenv` はテスト実行時の認証情報ロードのために使用される。
 */
package main

import (
	"context" // 使用箇所: main。目的: API リクエストのライフサイクル管理のため。
	"fmt"     // 使用箇所: main。目的: 生成された台本をターミナルに整形して表示するため。
	"log"     // 使用箇所: main。目的: クライアント初期化や生成に失敗した際に、エラー詳細を出力して異常終了させるため。
	"os"      // 使用箇所: main。目的: コマンド引数からテーマを受け取るため。

	"backend/internal/ai"      // 使用箇所: main。目的: 自作の AI 生成ロジックおよび設定モデルを利用するため。
	"github.com/joho/godotenv" // 使用箇所: loadEnv。目的: テスト実行時に `.env` から API キーを読み込むため。
)

/**
 * メインメソッドである。
 *
 * 内部処理:
 * - 1. `loadEnv()` を呼び出し、環境変数をロードする。
 * - 2. `context.Background() context.Context` を使用する。
 *   - `context` パッケージの関数であり、AI API 呼び出しの親となるコンテキストを生成する。
 * - 3. `ai.NewGeminiClient(ctx context.Context) (*ai.GeminiClient, error)` を使用する。
 *   - `ai` パッケージ（自作）の関数であり、認証情報を用いて Gemini クライアントを初期化する。
 * - 4. `defer (*ai.GeminiClient).Close() error` を使用する。
 *   - `ai` パッケージ（自作）のメソッドであり、テスト終了時にリソースを解放する。
 * - 5. `ai.GetDefaultConfig() ai.GeneratorConfig` を使用する。
 *   - `ai` パッケージ（自作）の関数であり、デフォルトのキャラクター設定セットを取得する。
 * - 6. `(*ai.GeminiClient).GenerateRadioScript(ctx context.Context, config ai.GeneratorConfig) ([]ai.ScriptItem, error)` を使用する。
 *   - `ai` パッケージ（自作）のメソッドであり、設定に基づき AI に台本生成を実行させる。
 * - 7. `fmt.Printf(format string, a ...any)` および `fmt.Println(a ...any)` を使用する。
 *   - `fmt` パッケージの関数であり、生成された台本の各セリフをアイコン付きで整形して出力する。
 * - 8. `log.Fatalf(format string, v ...any)` を使用する。
 *   - `log` パッケージの関数であり、回復不能なエラー時に詳細を記録してプログラムを強制終了させる。
 * @param なし
 * @return なし
 */
func main() {
	// --- 1. 設定のロード ---
	loadEnv()

	ctx := context.Background()

	// --- 2. AI クライアントの初期化 ---
	client, err := ai.NewGeminiClient(ctx)
	if err != nil {
		log.Fatalf("❌ クライアント初期化失敗: %v", err)
	}
	defer client.Close()

	// --- 3. 動的な設定の構築 ---
	config := ai.GetDefaultConfig()

	if len(os.Args) > 1 {
		config.Theme = os.Args[1]
	}

	config.Rules = append(config.Rules, "各キャラクターは必ず1回は爆笑すること。")

	fmt.Printf("🎙️ 動的設定で台本を生成中...（テーマ: %s）\n", config.Theme)
	fmt.Printf("📝 現在のルール数: %d\n", len(config.Rules))

	// --- 4. 台本生成の実行 ---
	script, err := client.GenerateRadioScript(ctx, config)
	if err != nil {
		log.Fatalf("❌ 生成失敗: %v", err)
	}

	// --- 5. 結果の表示 ---
	fmt.Println("--------------------------------------------------")
	fmt.Printf("🎭 生成された台本（全 %d 行）:\n", len(script))
	for i, item := range script {
		var icon string
		switch item.Speaker {
		case "mc":
			icon = "🎤"
		case "guest-a":
			icon = "📢"
		case "guest-b":
			icon = "💡"
		default:
			icon = "👤"
		}
		fmt.Printf("[%02d] %s %-8s: %-30s [%s]\n", i+1, icon, item.Speaker, item.Text, item.Emotion)
	}
	fmt.Println("--------------------------------------------------")
	fmt.Println("✅ AI生成テストが正常に完了した。")
}

/**
 * .env ファイルの読み込みメソッドである。
 *
 * 内部処理:
 * - 1. `godotenv.Load(filenames ...string) error` を使用する。
 *   - カレントディレクトリ、または 1 階層上のディレクトリから `.env` を探し、環境変数としてロードする。
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
