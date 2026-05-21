/**
 * Package ai は、Gemini API を使用した台本生成ロジックを提供するパッケージである。
 * このファイルでは、API クライアントの初期化および AI との対話処理を実装する。
 */
package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

/**
 * ScriptItem は、AI が生成した個別の発言データを表す構造体である。
 */
type ScriptItem struct {
	Speaker string `json:"speaker"` // 発言者 (mc, guest-a, guest-b)
	Text    string `json:"text"`    // 発言内容
	Emotion string `json:"emotion"` // 感情トーン (normal, excited, calm)
}

/**
 * GeminiClient は、Google Gemini API との通信を管理する構造体である。
 */
type GeminiClient struct {
	client *genai.Client
}

/**
 * NewGeminiClient は、新しい GeminiClient のインスタンスを作成して返す。
 *
 * 内部処理:
 * - 1. `os.Getenv(key string) string` を使用し、環境変数から API キーを取得する。
 * - 2. `genai.NewClient(ctx context.Context, opts ...option.ClientOption) (*genai.Client, error)` を使用する。
 *   - `genai` パッケージの関数であり、API キーを認証情報としてクライアントを初期化する。
 * @param ctx [context.Context] 接続操作用のコンテキスト
 * @return client [*GeminiClient] 初期化されたクライアントのポインタ
 * @return err [error] 初期化に失敗した場合のエラー
 */
func NewGeminiClient(ctx context.Context) (*GeminiClient, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY が環境変数に設定されていない")
	}

	// option.WithAPIKey(apiKey string) ClientOption を使用して API キーを設定する。
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("Gemini クライアントの作成に失敗した: %w", err)
	}

	return &GeminiClient{client: client}, nil
}

/**
 * GenerateRadioScript は、指定されたテーマに基づいてラジオの台本を生成するメソッドである。
 *
 * 内部処理:
 * - 1. `(*genai.Client).GenerativeModel(name string) *genai.Model` を使用する。
 *   - 指定したモデル名（gemini-1.5-flash）に基づいてモデルインスタンスを生成する。
 * - 2. モデルの設定（ResponseMIMEType）を "application/json" に指定し、構造化出力を強制する。
 * - 3. `SystemPrompt` とユーザ入力を結合し、プロンプトを構築する。
 * - 4. `(*genai.Model).GenerateContent(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error)` を使用する。
 *   - Gemini API にリクエストを送信し、生成結果を取得する。
 * - 5. `json.Unmarshal(data []byte, v any) error` を使用する。
 *   - 文字列レスポンスを `[]ScriptItem` 構造体に変換する。
 * @param ctx [context.Context] 通信用のコンテキスト
 * @param theme [string] ラジオ番組のテーマ
 * @return script [[]ScriptItem] 生成されたセリフのリスト
 * @return err [error] 生成またはパースに失敗した場合のエラー
 */
func (g *GeminiClient) GenerateRadioScript(ctx context.Context, theme string) ([]ScriptItem, error) {
	// gemini-1.5-flash モデルを使用する。
	model := g.client.GenerativeModel("gemini-1.5-flash")

	// JSON 出力を強制するための設定（ライブラリのバージョン互換性のため一時的に無効化）。
	// model.ResponseMIMEType = "application/json"

	// プロンプトの構築。
	promptText := fmt.Sprintf("%s\n\nテーマ: %s", SystemPrompt, theme)

	// genai.Text(string) は、プレーンテキストを genai.Part インターフェースに変換する。
	resp, err := model.GenerateContent(ctx, genai.Text(promptText))
	if err != nil {
		return nil, fmt.Errorf("コンテンツ生成に失敗した: %w", err)
	}

	// 候補（Candidates）が空でないことを確認する。
	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("AI からの応答が空である")
	}

	// 生成されたコンテンツを抽出する。
	var responseText string
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			responseText += string(text)
		}
	}

	// AI が Markdown のコードブロック（```json ... ```）で返してくる場合があるため、
	// 不要な囲み文字を除去する。
	// strings パッケージを使用してトリミングを行う。
	cleanJSON := responseText
	if len(cleanJSON) > 7 && cleanJSON[:7] == "```json" {
		// 冒頭の ```json を削除
		cleanJSON = cleanJSON[7:]
	}
	if len(cleanJSON) > 3 && cleanJSON[len(cleanJSON)-3:] == "```" {
		// 末尾の ``` を削除
		cleanJSON = cleanJSON[:len(cleanJSON)-3]
	}

	// JSON のパース。
	var script []ScriptItem
	// json.Unmarshal() を使用して、クリーンアップされた JSON 文字列を構造体に変換する。
	if err := json.Unmarshal([]byte(cleanJSON), &script); err != nil {
		return nil, fmt.Errorf("JSON のパースに失敗した: %w\nRaw Response: %s", err, responseText)
	}

	return script, nil
}

/**
 * Close は、Gemini クライアントのリソースを解放するメソッドである。
 * 内部処理:
 * - `(*genai.Client).Close() error` を呼び出して、内部の接続を閉じる。
 * @param なし
 * @return err [error] 閉鎖に失敗した場合のエラー
 */
func (g *GeminiClient) Close() error {
	return g.client.Close()
}
