/**
 * Package ai は、Google Gemini API を使用した台本生成ロジックを提供するパッケージである。
 *
 * 【パッケージの構成要素】
 * - 1. `GeminiClient`: Gemini API との物理的な通信を管理する中心的な構造体。
 * - 2. `GeneratorConfig`: キャラクター設定や番組ルールを保持する、台本生成の入力となる設定構造体。
 * - 3. `ScriptItem`: 生成された台本の 1 セリフを表すデータ構造体。
 * - 4. `BuildSystemPrompt`: 設定から AI へのプロンプトを動的に生成する関数。
 *
 * 【要素間の関係】
 * - `GeminiClient` は、`GeneratorConfig` を受け取り、`BuildSystemPrompt` を内部で呼び出して最終的なプロンプトを構築する。
 * - 生成されたレスポンスは `ScriptItem` の配列としてパースされ、上位レイヤーに返される。
 * - 各外部パッケージ（genai, option 等）は、`GeminiClient` の初期化とコンテンツ生成のためにインポートされている。
 */
package ai

import (
	"context"       // 使用箇所: GeminiClient の各メソッド。目的: API 通信のタイムアウトやキャンセル管理のため。
	"encoding/json" // 使用箇所: GenerateRadioScript メソッド。目的: AI からの JSON レスポンスを構造体にデコードするため。
	"fmt"           // 使用箇所: 各メソッド。目的: 文字列のフォーマットやエラーメッセージの構築のため。
	"os"            // 使用箇所: NewGeminiClient。目的: 環境変数から GEMINI_API_KEY を読み込むため。

	"github.com/google/generative-ai-go/genai" // 使用箇所: GeminiClient 構造体、メソッド。目的: Gemini モデルの操作と生成リクエストの送信のため。
	"google.golang.org/api/option"             // 使用箇所: NewGeminiClient。目的: API キーによる認証オプションを指定するため。
)

/**
 * ScriptItem は、AI が生成した個別の発言データを表す構造体である。
 */
type ScriptItem struct {
	Speaker   string `json:"speaker"`    // 発言者の役割名 (mc, guest-a, guest-b)
	SpeakerID int    `json:"speaker_id"` // VOICEVOX の話者スタイルID
	Text      string `json:"text"`       // AI が生成したセリフの本文
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
 *   - `os` パッケージの関数であり、環境変数から指定したキー（GEMINI_API_KEY）の値を取得する。
 * - 2. `option.WithAPIKey(apiKey string) option.ClientOption` を使用する。
 *   - `google.golang.org/api/option` パッケージの関数であり、API キーを認証情報として設定するオブジェクトを生成する。
 * - 3. `genai.NewClient(ctx context.Context, opts ...option.ClientOption) (*genai.Client, error)` を使用する。
 *   - `github.com/google/generative-ai-go/genai` パッケージの関数であり、設定されたオプションを用いて API クライアントを初期化する。
 * @param ctx [context.Context] 接続操作用のコンテキスト
 * @return client [*GeminiClient] 初期化されたクライアントのポインタ
 * @return err [error] 初期化に失敗した場合のエラー
 */
func NewGeminiClient(ctx context.Context) (*GeminiClient, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY が環境変数に設定されていない")
	}

	// [SDK] option.WithAPIKey を使用して API キーを設定する。
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("Gemini クライアントの作成に失敗した: %w", err)
	}

	return &GeminiClient{client: client}, nil
}

/**
 * GenerateRadioScript は、指定された設定に基づいてラジオの台本を生成するメソッドである。
 *
 * 内部処理:
 * - 1. `(*genai.Client).GenerativeModel(name string) *genai.Model` を使用する。
 *   - `genai` パッケージのメソッドであり、指定したモデル名（gemini-2.5-flash）に基づいてモデルインスタンスを生成する。
 * - 2. `BuildSystemPrompt(config GeneratorConfig) string` を呼び出す。
 *   - `ai` パッケージ（内部）の関数であり、設定から AI への指示文を動的に組み立てる。
 * - 3. `fmt.Sprintf(format string, a ...any) string` を使用する。
 *   - `fmt` パッケージの関数であり、システムプロンプトとテーマを結合する。
 * - 4. `genai.Text(string)` を使用する。
 *   - `genai` パッケージの型キャストであり、文字列を SDK が受け入れ可能な Part インターフェースに変換する。
 * - 5. `(*genai.Model).GenerateContent(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error)` を使用する。
 *   - `genai` パッケージのメソッドであり、Gemini API に生成リクエストを送信する。
 * - 6. 文字列レスポンスに対し、マークダウンのコードブロック（```json）をトリミングする。
 * - 7. `json.Unmarshal(data []byte, v any) error` を使用する。
 *   - `encoding/json` パッケージの関数であり、クリーンアップされた JSON 文字列を構造体に変換する。
 * @param ctx [context.Context] 通信用のコンテキスト
 * @param config [GeneratorConfig] 台本生成のカスタマイズ設定
 * @return script [[]ScriptItem] 生成されたセリフのリスト
 * @return err [error] 生成またはパースに失敗した場合のエラー
 */
func (g *GeminiClient) GenerateRadioScript(ctx context.Context, config GeneratorConfig) ([]ScriptItem, error) {
	// [SDK] gemini-2.5-flash モデルを使用する。
	model := g.client.GenerativeModel("gemini-2.5-flash")

	// [SDK] JSON 出力を強制するための設定。
	model.ResponseMIMEType = "application/json"

	// [内部] 動的なプロンプトの構築。
	systemPrompt := BuildSystemPrompt(config)
	promptText := fmt.Sprintf("%s\n\nテーマ: %s", systemPrompt, config.Theme)

	// [SDK] genai.Text(string) は、プレーンテキストを genai.Part に変換する。
	resp, err := model.GenerateContent(ctx, genai.Text(promptText))
	if err != nil {
		return nil, fmt.Errorf("コンテンツ生成に失敗した: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("AI からの応答が空である")
	}

	var responseText string
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			responseText += string(text)
		}
	}

	// [内部] マークダウン除去。
	cleanJSON := responseText
	if len(cleanJSON) > 7 && cleanJSON[:7] == "```json" {
		cleanJSON = cleanJSON[7:]
	}
	if len(cleanJSON) > 3 && cleanJSON[len(cleanJSON)-3:] == "```" {
		cleanJSON = cleanJSON[:len(cleanJSON)-3]
	}

	// [標準] JSON パース。
	var script []ScriptItem
	if err := json.Unmarshal([]byte(cleanJSON), &script); err != nil {
		return nil, fmt.Errorf("JSON のパースに失敗した: %w\nRaw Response: %s", err, responseText)
	}

	return script, nil
}

/**
 * Close は、Gemini クライアントのリソースを解放するメソッドである。
 *
 * 内部処理:
 * - 1. `(*genai.Client).Close() error` を呼び出す。
 *   - `genai` パッケージのメソッドであり、内部で保持している API とのネットワーク接続やリソースを安全に閉鎖する。
 * @param なし
 * @return err [error] 閉鎖に失敗した場合のエラー
 */
func (g *GeminiClient) Close() error {
	return g.client.Close()
}
