/**
 * Package main は、Gemini AI による台本生成ロジックの単体テスト用エントリポイントである。
 */
package main

import (
	"context"
	"fmt"
	"log"

	"backend/internal/ai"
	"github.com/joho/godotenv"
)

/**
 * メインメソッドである。
 * 実際に Gemini API を呼び出し、台本が JSON として正しく生成・パースされるかを確認する。
 * @param なし
 * @return なし
 */
func main() {
	// .env の読み込み。
	if err := godotenv.Load(".env"); err != nil {
		log.Println("情報: .env が見つからないため環境変数を使用する。")
	}

	ctx := context.Background()

	// 1. クライアントの初期化。
	// ai.NewGeminiClient(ctx context.Context) を呼び出す。
	client, err := ai.NewGeminiClient(ctx)
	if err != nil {
		log.Fatalf("❌ クライアント初期化失敗: %v", err)
	}
	// defer 文でクライアントのリソースを解放する。
	defer client.Close()

	// 2. 台本生成のテスト。
	fmt.Println("🎙️ 台本を生成中...（テーマ: AIとの共生）")
	// client.GenerateRadioScript(ctx context.Context, theme string) を呼び出す。
	script, err := client.GenerateRadioScript(ctx, "AIとの共生")
	if err != nil {
		log.Fatalf("❌ 生成失敗: %v", err)
	}

	// 3. 結果の表示。
	fmt.Println("--------------------------------------------------")
	fmt.Printf("🎭 生成された台本（全 %d 行）:\n", len(script))
	for i, item := range script {
		icon := "🎤"
		if item.Speaker == "guest-a" {
			icon = "📢"
		} else if item.Speaker == "guest-b" {
			icon = "💡"
		}
		fmt.Printf("[%02d] %s %-8s: %-30s [%s]\n", i+1, icon, item.Speaker, item.Text, item.Emotion)
	}
	fmt.Println("--------------------------------------------------")
	fmt.Println("✅ AI生成テストが正常に完了した。")
}
