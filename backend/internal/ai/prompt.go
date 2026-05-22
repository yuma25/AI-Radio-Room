/**
 * Package ai は、Google Gemini API を使用した台本生成ロジックを提供するパッケージである。
 * このファイルでは、AI への指示文（システムプロンプト）の動的構築と、設定用の構造体を定義する。
 *
 * 【パッケージの構成要素】
 * - 1. `Character`: 登場人物の人格や役割を定義する構造体。
 * - 2. `GeneratorConfig`: ラジオ番組全体のテーマや構成ルールをまとめる設定構造体。
 * - 3. `BuildSystemPrompt`: `GeneratorConfig` を解析し、AI が理解可能な自然言語プロンプトに変換する関数。
 * - 4. `GetDefaultConfig`: 標準的なラジオ番組設定を生成する補助関数。
 *
 * 【要素間の関係】
 * - `GeneratorConfig` は複数の `Character` を含み、`BuildSystemPrompt` はこれらすべてを組み合わせて一つの巨大な指示文を作成する。
 * - この指示文は、同パッケージ内の `GeminiClient` によって API リクエストのメタデータとして使用される。
 */
package ai

import (
	"fmt"     // 使用箇所: BuildSystemPrompt。目的: キャラクター設定やルールを文字列としてフォーマットするため。
	"strings" // 使用箇所: BuildSystemPrompt。目的: strings.Builder を使用して効率的に巨大なプロンプト文字列を構築するため。
)

/**
 * Character は、ラジオ番組に登場する個別のキャラクター設定を保持する構造体である。
 * 各キャラクターの性格、口癖、役割を AI に伝えることで、パーソナリティの再現性を高める。
 */
type Character struct {
	Role        string `json:"role"`        // システム上の役割識別名 (例: mc, guest-a, guest-b)
	Name        string `json:"name"`        // 放送中に呼ばれる表示名
	Personality string `json:"personality"` // 性格描写や特定の語尾（例：「〜なのだ」）などの特徴
	Description string `json:"description"` // 番組内での立ち位置や具体的な役割（例：話題の振り役、反対意見担当）
}

/**
 * GeneratorConfig は、AI 台本生成の挙動を詳細にカスタマイズするための設定構造体である。
 * これを変更することで、番組の雰囲気や会話の密度を動的に制御できる。
 */
type GeneratorConfig struct {
	Theme      string      `json:"theme"`      // 番組で議論するメインの話題
	Characters []Character `json:"characters"` // 番組に出演するキャラクターのリスト
	Rules      []string    `json:"rules"`      // AI に守らせる具体的な構成上の制約や追加ルール
}

/**
 * BuildSystemPrompt は、GeneratorConfig の内容に基づいて AI への指示文（システムプロンプト）を動的に構築する。
 *
 * 内部処理:
 * - 1. [標準] `strings.Builder` を使用し、ヒープ割り当てを抑えつつ効率的に文字列を結合する。
 * - 2. [内部] 設定されたキャラクター情報をループで展開し、詳細な性格設定テキストを作成する。
 * - 3. [内部] 番組構成ルールを箇条書きで追加する。
 * - 4. [内部] 最終的な JSON 出力フォーマットの厳密な定義を末尾に付与する。
 * @param config [GeneratorConfig] 台本生成の各種設定
 * @return prompt [string] 構築されたシステムプロンプト文字列
 */
func BuildSystemPrompt(config GeneratorConfig) string {
	var b strings.Builder

	// strings.Builder の WriteString メソッドを使用してプロンプトを構築していく。
	b.WriteString("あなたはプロのラジオ構成作家です。以下のキャラクターによるラジオ番組の台本を生成してください。\n\n")

	b.WriteString("【キャラクター設定】\n")
	for i, c := range config.Characters {
		// fmt.Sprintf() を使用して、各キャラクターの情報をフォーマットする。
		b.WriteString(fmt.Sprintf("%d. %s (%s):\n", i+1, c.Name, c.Role))
		b.WriteString(fmt.Sprintf("   - 性格: %s\n", c.Personality))
		b.WriteString(fmt.Sprintf("   - 役割: %s\n", c.Description))
	}

	b.WriteString("\n【番組構成のルール】\n")
	for _, rule := range config.Rules {
		b.WriteString(fmt.Sprintf("- %s\n", rule))
	}

	b.WriteString("\n【出力形式（厳守）】\n")
	b.WriteString("以下の JSON 構造の配列のみを返してください。余計な解説やマークダウン記法は不要です。\n")
	b.WriteString("[\n  {\n    \"speaker\": \"キャラクターのRole名\",\n    \"text\": \"セリフの内容\",\n    \"emotion\": \"normal\" | \"excited\" | \"calm\"\n  }\n]\n")

	return b.String()
}

/**
 * GetDefaultConfig は、標準的なラジオ設定（メタン・ずんだもん・つむぎ）を返す初期化関数である。
 *
 * 内部処理:
 * - 1. [内部] 静的に定義された初期設定を `GeneratorConfig` 構造体に詰めて返す。
 * @param なし
 * @return config [GeneratorConfig] 標準のキャラクター設定を含む構造体
 */
func GetDefaultConfig() GeneratorConfig {
	return GeneratorConfig{
		Theme: "AIとの共生",
		Characters: []Character{
			{
				Role:        "mc",
				Name:        "メタン",
				Personality: "明るくハキハキしており、進行とタイムキープを重視する。",
				Description: "オープニングとエンディング、話題の振り、まとめを担当。",
			},
			{
				Role:        "guest-a",
				Name:        "ずんだもん",
				Personality: "好奇心旺盛で新しいものが大好き。語尾に「なのだ」を付ける。",
				Description: "常に前向きな意見を出し、未来の可能性を熱く語る。",
			},
			{
				Role:        "guest-b",
				Name:        "つむぎ",
				Personality: "冷静沈着でデータ重視. やや保守的な視点を持つ。",
				Description: "Aの意見に対し、コストやリスク、現実的な課題を指摘する。",
			},
		},
		Rules: []string{
			"指定された「テーマ」に基づいて、3人が掛け合いを行う。",
			"会話は自然な流れ（MCの挨拶 -> ゲスト紹介 -> 本題の議論 -> MCのまとめ）にする。",
			"1回のリクエストで、20〜30行程度の密度の高い発言リストを生成する。",
		},
	}
}
