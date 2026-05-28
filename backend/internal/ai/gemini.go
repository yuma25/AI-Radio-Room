/**
 * Package ai は、Google Gemini API を使用した台本生成ロジックを提供するパッケージである。
 */
package ai

import (
	"context"       // 使用箇所: GeminiClient の各メソッド。目的: API 通信のタイムアウトやキャンセル管理のため。
	"encoding/json" // 使用箇所: GenerateRadioScript メソッド。目的: AI からの JSON レスポンスを構造体にデコードするため。
	"fmt"           // 使用箇所: 各メソッド。目的: 文字列のフォーマットやエラーメッセージの構築のため。
	"os"            // 使用箇所: NewGeminiClient。目的: 環境変数から GEMINI_API_KEY を読み込むため。
	"strings"       // 使用箇所: GenerateRadioScript。目的: Markdown コードブロックを除去するため。

	"google.golang.org/genai"
)

/**
 * ScriptItem は、AI が生成した個別の発言データを表す構造体である。
 */
type ScriptItem struct {
	Speaker   string `json:"speaker"`    // 発言者の役割名 (mc, guest-a, guest-b)
	SpeakerID int    `json:"speaker_id"` // VOICEVOX の話者スタイルID
	Text      string `json:"text"`       // 画面表示用のセリフ（漢字混じり）
	AudioText string `json:"audio_text"` // 音声合成用のセリフ（読み仮名・カタカナ主体）
	Emotion   string `json:"emotion"`    // セリフに対応する感情トーン (normal, excited, calm)
}

/**
 * GeminiClient は、Google Gemini API との通信を管理する構造体である。
 * 内部に [SDK] `*genai.Client` を保持し、上位レイヤーに対して高レベルな台本生成インターフェースを提供する。
 */
type GeminiClient struct {
	client *genai.Client // [SDK] 生成AI操作用の SDK クライアントインスタンス
}

/**
 * NewGeminiClient は、新しい GeminiClient のインスタンスを作成して返す。
 *
 * 内部処理:
 * - 1. `os.Getenv(key string) string` を使用する。
 * - 2. `option.WithAPIKey(apiKey string) option.ClientOption` を使用する。
 * - 3. `genai.NewClient(ctx context.Context, opts ...option.ClientOption) (*genai.Client, error)` を使用する。
 * @param ctx [context.Context] 接続操作用のコンテキスト
 * @return client [*GeminiClient] 初期化されたクライアントのポインタ
 * @return err [error] 初期化に失敗した場合のエラー
 */
func NewGeminiClient(ctx context.Context) (*GeminiClient, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY が環境変数に設定されていない")
	}

	// [SDK] 正しい初期化方法: option.WithAPIKey を使用する。
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("Gemini クライアントの作成に失敗した: %w", err)
	}

	return &GeminiClient{client: client}, nil
}

/**
 * GenerateRadioScript は、指定された設定に基づいてラジオの台本を生成するメソッドである。
 * 全ての台本が生成されるのを待ってから一括で返す。
 */
/**
 * GenerateRadioScript は、指定された設定に基づいてラジオの台本を生成するメソッドである。
 * 全ての台本が生成されるのを待ってから一括で返す。
 */
func (g *GeminiClient) GenerateRadioScript(ctx context.Context, config GeneratorConfig) ([]ScriptItem, error) {
	systemPrompt := BuildSystemPrompt(config)
	promptText := fmt.Sprintf("%s\n\nテーマ: %s", systemPrompt, config.Theme)

	resp, err := g.client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		genai.Text(promptText),
		&genai.GenerateContentConfig{
			ResponseMIMEType: "application/json", // 💡 MIME をすべて大文字に修正
		},
	)
	if err != nil {
		return nil, fmt.Errorf("コンテンツ生成に失敗した: %w", err)
	}

	// 💡 Text はメソッドなので末尾に () をつけます
	responseText := resp.Text()
	if responseText == "" {
		return nil, fmt.Errorf("AI からの応答が空である")
	}

	// [内部] Markdown のコードブロック (```json ... ```) が含まれている場合に備え、トリミングする。
	responseText = strings.TrimSpace(responseText)
	if strings.HasPrefix(responseText, "```json") {
		responseText = strings.TrimPrefix(responseText, "```json")
		responseText = strings.TrimSuffix(responseText, "```")
	} else if strings.HasPrefix(responseText, "```") {
		responseText = strings.TrimPrefix(responseText, "```")
		responseText = strings.TrimSuffix(responseText, "```")
	}
	responseText = strings.TrimSpace(responseText)

	// JSON 配列としてパース
	var script []ScriptItem
	if err := json.Unmarshal([]byte(responseText), &script); err != nil {

		return nil, fmt.Errorf("JSON パース失敗: %w\nRaw: %s", err, responseText)
	}

	return script, nil
}

/**
 * GenerateRadioScriptStream は、AI の回答を 1 行ずつストリーミングで処理するためのメソッドである。
 * セリフが生成されるたびに callback 関数を呼び出す。
 */
func (g *GeminiClient) GenerateRadioScriptStream(ctx context.Context, config GeneratorConfig, callback func(ScriptItem)) error {
	// 現在は GenerateRadioScript を利用した擬似的なストリームとして実装。
	script, err := g.GenerateRadioScript(ctx, config)
	if err != nil {
		return err
	}

	for _, item := range script {
		callback(item)
	}

	return nil
}

/**
 * Close は、Gemini クライアントのリソースを解放するメソッドである。
 */
// メソッドを残す場合は、単に nil を返すだけにする
func (g *GeminiClient) Close() error {
	return nil
}
