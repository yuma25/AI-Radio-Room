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
	Theme            string       `json:"theme"`             // 番組で議論するメインの話題
	Characters       []Character  `json:"characters"`        // 番組に出演するキャラクターのリスト
	Rules            []string     `json:"rules"`             // AI に守らせる具体的な構成上の制約や追加ルール
	History          []ScriptItem `json:"history"`           // 以前のフェーズで生成された会話履歴
	PhaseInstruction string       `json:"phase_instruction"` // 現在のフェーズ（導入、本題、結び等）に特化した指示文
}

/**
 * BuildSystemPrompt は、GeneratorConfig の内容に基づいて AI への指示文（システムプロンプト）を動的に構築する。
 *
 * 内部処理:
 * - 1. [標準] `strings.Builder` を使用し、ヒープ割り当てを抑えつつ効率的に文字列を結合する。
 * - 2. [内部] 設定されたキャラクター情報をループで展開し、詳細な性格設定テキストを作成する。
 * - 3. [内部] 番組構成ルールを箇条書きで追加する。
 * - 4. [内部] 会話履歴（History）が存在する場合、文脈維持のためにプロンプトに組み込む。
 * - 5. [内部] 現在のフェーズに特化した指示（PhaseInstruction）を追加する。
 * - 6. [内部] 最終的な JSON 出力フォーマットの厳密な定義を末尾に付与する。
 * @param config [GeneratorConfig] 台本生成の各種設定
 * @return prompt [string] 構築されたシステムプロンプト文字列
 */
func BuildSystemPrompt(config GeneratorConfig) string {
	var b strings.Builder

	b.WriteString("あなたはプロのラジオ構成作家です。以下のキャラクター設定とルールに基づき、自然なラジオ番組の台本を生成してください。\n\n")

	b.WriteString("【キャラクター設定】\n")
	for i, c := range config.Characters {
		b.WriteString(fmt.Sprintf("%d. %s (%s):\n", i+1, c.Name, c.Role))
		b.WriteString(fmt.Sprintf("   - 性格: %s\n", c.Personality))
		b.WriteString(fmt.Sprintf("   - 役割: %s\n", c.Description))
	}

	b.WriteString("\n【共通ルール】\n")
	for _, rule := range config.Rules {
		b.WriteString(fmt.Sprintf("- %s\n", rule))
	}

	if len(config.History) > 0 {
		b.WriteString("\n【これまでの会話履歴】\n")
		b.WriteString("以下の会話の続きから自然に繋げてください。同じ挨拶やフリを繰り返さないよう注意してください。\n")
		for _, item := range config.History {
			b.WriteString(fmt.Sprintf("[%s]: %s\n", item.Speaker, item.Text))
		}
	}

	if config.PhaseInstruction != "" {
		b.WriteString("\n【現在のフェーズに対する具体的な指示】\n")
		b.WriteString(fmt.Sprintf("!!!重要!!!: %s\n", config.PhaseInstruction))
	}

	b.WriteString("\n【出力形式（厳守）】\n")
	b.WriteString("以下の JSON 構造の配列のみを返してください。余計な解説やマークダウン記法は不要です。\n")
	b.WriteString("[\n  {\n    \"speaker\": \"キャラクターのRole名\",\n    \"text\": \"画面表示用のセリフ（漢字を適切に使い、読みやすく）\",\n    \"audio_text\": \"音声合成用のセリフ（読み間違いを防ぐため、英単語・数字・難読漢字はすべてカタカナやひらがなにする）\",\n    \"emotion\": \"normal\" | \"excited\" | \"calm\"\n  }\n]\n")

	return b.String()
}

/**
 * GetDefaultConfig は、標準的なラジオ設定（四国めたん・ずんだもん・春日部つむぎ）を返す初期化関数である。
 *
 * 内部処理:
 * - 1. [内部] 静的に定義された初期設定を `GeneratorConfig` 構造体に詰めて返す。
 * @param なし
 * @return config [GeneratorConfig] 標準のキャラクター設定を含む構造体
 */
func GetDefaultConfig() GeneratorConfig {
	return GeneratorConfig{
		Theme: "最近の面白かった出来事",
		Characters: []Character{
			{
				Role:        "mc",
				Name:        "四国めたん",
				Personality: "明るくて親しみやすいお姉さん。よく笑い、周りの話を上手く引き出す。",
				Description: "話題を振りつつ、ゲストの話に共感したり軽くツッコミを入れたりする。",
			},
			{
				Role:        "guest-a",
				Name:        "ずんだもん",
				Personality: "無邪気でちょっとお調子者。語尾に「なのだ」を付ける。感情が表に出やすい。",
				Description: "テーマに対して素直な感想や、ちょっとズレたボケを担当する。",
			},
			{
				Role:        "guest-b",
				Name:        "春日部つむぎ",
				Personality: "マイペースでマイワールドを持っている。少し毒舌なところもあるが、基本は優しい。",
				Description: "ずんだもんのボケに冷静にツッコミを入れたり、自分の日常あるあるを語る。",
			},
		},
		Rules: []string{
			"指定された「テーマ」に基づいて、3人が掛け合いを行うラジオ番組として進行すること。",
			"専門家のような難しい議論や堅苦しい説明は絶対に行わないでください。あくまで「友達同士が放課後やカフェで雑談しているような、日常的でゆるい雰囲気」を心がけてください。",
			"会話の中で、具体的な「選手名」「人物名」「商品名」「地名」などの固有名詞を積極的に出し、リアリティと説得力を持たせること。",
			"難しい専門用語は使わず、誰にでもわかる簡単な言葉や、身近な「あるある」エピソードを交えて話すこと。",
			"各キャラクターの発言は、短すぎず長すぎない、1セリフあたり40〜80文字程度の自然な長さにすること。",
			"【超重要】VOICEVOXでの読み間違いを避けるため、audio_text フィールドには以下のルールを適用すること：",
			"- 英単語はすべてカタカナにする（例: AI -> エーアイ, LLM -> エルエルエム）。",
			"- 数字は読み方に合わせてひらがなにする（例: 1つ -> ひとつ, 2024年 -> にせんにじゅうよねん）。",
			"- 読み間違いやすい漢字や固有名詞はすべてひらがなかカタカナにする。",
			"- 適切な位置に読点（、）を打ち、自然なイントネーションにする。",
			"原稿を読んでいるような硬い表現は避け、言い淀み（「えっと」「その…」「うーん」など）や相槌（「へえ！」「なるほど」「あー、わかる！」）、笑い声（「ふふっ」「あはは！」）を積極的に交え、極めて人間らしいリアルな口語表現にすること。",
			"【重要】感情の起伏（喜怒哀楽）をはっきりと表現し、発言内容に合わせて emotion フィールド（normal, excited, calm）を適切に変化させること。冗談を言い合って笑う場面（excited）を多めに入れ、楽しい空間を作ること。",
			"深刻な対立やディベートは避け、お互いに共感したり、軽くツッコミを入れたりしながら、和気あいあいとトークを広げること。",
			"1回のリクエストで、15〜20行程度のテンポの良い発言リスト（JSON配列）を生成すること。",
		},
	}
}
