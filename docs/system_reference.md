# AI Radio Room システム・リファレンス

本ドキュメントは、AI Radio Room バックエンドシステムの全容を言語化したものである。システムフロー図 (`docs/system_flow.md`) と対応しており、データ定義および具体的な処理の連鎖を詳細に解説する。

---

## 1. バックエンドの処理フロー（全体像）

システムにおける主要な動作フローは、大きく「初期化」「AI生成」「DB保存」の 3 つのフェーズに分かれる。以下に `backend/cmd/test_ai/main.go` を基準とした具体的な流れを示す。

### 1.1 初期化・準備フェーズ
外部環境のロードと AI 連携用のインターフェースを準備する。

1. **環境変数のロード**:
    - `godotenv.Load(".env")` (パッケージ: `github.com/joho/godotenv`) を呼び出す。
    - プロジェクトルートの `.env` ファイルから API キーや DB 接続情報をプロセスに読み込む。
2. **AI クライアントの初期化**:
    - `ai.NewGeminiClient(ctx)` を呼び出す。
    - 内部で `genai.NewClient(ctx, option.WithAPIKey(apiKey))` が実行され、Gemini API との通信インターフェース (`*ai.GeminiClient`) が確立される。
3. **設定の構築**:
    - `ai.GetDefaultConfig()` を呼び出す。
    - AI に渡すテーマやキャラクター設定 (`ai.GeneratorConfig`) の雛形を取得し、必要に応じて動的にルール（「爆笑する」等）を追加する。

### 1.2 台本生成フェーズ (AI へのリクエスト)
構築された設定に基づき、Gemini API を通じて対話を生成する。

1. **生成処理の開始**:
    - `client.GenerateRadioScript(ctx, config)` を呼び出す。
2. **プロンプト（指示文）の構築**:
    - 内部で `ai.BuildSystemPrompt(config)` が呼び出される。
    - `strings.Builder` を用いてキャラクター設定やルールが結合され、巨大な「プロンプト文字列」が完成する。
3. **API へのリクエスト送信**:
    - `(*genai.Model).GenerateContent(ctx, genai.Text(promptText))` が呼び出され、Google の AI サーバーにリクエストが送信される。
4. **結果のパース（解析）**:
    - API から返ってきた JSON 文字列から Markdown 記法を除去する。
    - `json.Unmarshal()` を使用して、Go の構造体スライス `[]ai.ScriptItem` に変換する。

### 1.3 データベース保存フェーズ (リポジトリ操作)
本番サーバーでは、台本生成と並行して以下の順序で実行される。

1. **DB リポジトリの初期化**:
    - `pgx.Connect(ctx, dbURL)` で物理接続を確立後、`db.NewRepository(conn)` でリポジトリインスタンスを生成する。
2. **放送セッションの作成**:
    - `repo.CreateBroadcast(ctx, "テーマ名")` を呼び出す。
    - `broadcasts` テーブルに新規レコードを作成し、一意な `broadcast_id` (UUID) を取得する。
3. **セリフの逐次保存と配信**:
    - 生成された `[]ai.ScriptItem` をループで回し、1 行ごとに `repo.SaveTranscript(ctx, &transcript)` を呼び出して DB へ保存すると同時に、SSE を通じてクライアントへ送出する。
4. **放送セッションの終了**:
    - 全てのセリフの処理が完了後、`repo.EndBroadcast(ctx, broadcast_id)` を呼び出し、該当放送の終了時刻 (`ended_at`) を更新する。

---

## 2. API サーバーと通信プロトコル (`backend/internal/api`)

### 2.1 ルーティング構造
`go-chi/chi` ルーターを使用し、以下のエンドポイントを提供する。

- **`POST /api/broadcasts`**:
    - **役割**: 放送開始リクエストを受け付け、台本生成とリアルタイム配信を開始する。
    - **入力**: `{"theme": "テーマ名"}` (JSON)
    - **出力**: `text/event-stream` (SSE)
- **`GET /api/broadcasts`**:
    - **役割**: 過去の放送履歴一覧を取得する（※実装予定）。

### 2.2 SSE (Server-Sent Events) 配信フロー
放送開始 API におけるリアルタイム配信の詳細は以下の通りである。

1. **レスポンスヘッダーの設定**:
    - `Content-Type: text/event-stream`
    - `Cache-Control: no-cache`
    - `Connection: keep-alive`
2. **データの送出**:
    - AI が生成した台本を 1 セリフごとに JSON 化し、`data: {"speaker": "...", "text": "...", "emotion": "..."}\n\n` の形式で書き出す。
3. **バッファのフラッシュ**:
    - `(http.Flusher).Flush()` を呼び出し、サーバー側のバッファに溜め込まず即座にクライアントへデータを送信する。
4. **終了合図**:
    - 全セリフの送信完了後、`event: close` カスタムイベントを送信し、クライアント側にストリームの終了を通知する。

---

## 3. 共通技術規約（エンジニアリング・スタンダード）

本プロジェクトでは、コードの品質と保守性を維持するため、以下の技術規約を遵守する。

### 3.1 Context の取り扱い
Go の `context.Context` パッケージを、すべての I/O 操作およびバックグラウンド処理の基盤として使用する。

- **役割**: リクエストのキャンセル伝播、タイムアウト制御、およびスコープ内の値（トレースID等）を伝播させる。
- **伝播ルール**: 
    - メインスレッド（起動時）: `context.Background()` を生成し、初期化関数へ渡す。
    - HTTP ハンドラー: `r.Context()` を取得し、そこから派生する処理（DB, AI）へ第一引数として渡す。これにより、クライアントが接続を切断した際に、重い AI 生成処理や DB クエリを即座に中断させ、リソースを保護することができる。
- **命名規則**: 常に第一引数とし、変数名は `ctx` に統一する。

### 3.2 エラーハンドリングとロギング
- **Fail-Fast 原則**: 起動フェーズにおいて、DB 接続失敗や API キーの欠落など、システムが正常に動作し得ない致命的なエラーが発生した場合は、`log.Fatalf` を用いて即座にプロセスを終了させる。
- **回復可能なエラー**: API 実行中の単一セリフの保存失敗などは、エラーログを出力しつつ処理（配信）を継続させる。
- **日本語コメントの義務化**: すべての複雑なロジックには、技術的ステップとその意図を日本語（常体）で詳細に記載する。

---

## 4. データ定義詳細


### 4.1 AI 連携モデル (`backend/internal/ai`)

#### `Character` 構造体
- **`Role`** [string]: 内部的な役割名 (mc, guest-a, guest-b)。
- **`Name`** [string]: 放送内で呼ばれる名前。
- **`Personality`** [string]: 性格描写や口癖。
- **`Description`** [string]: 番組内での立ち位置や具体的な役割。

#### `GeneratorConfig` 構造体
- **`Theme`** [string]: 番組のメインテーマ。
- **`Characters`** [[]Character]: 登場させるキャラクター設定のリスト。
- **`Rules`** [[]string]: 構成上の制約や追加ルール。

#### `ScriptItem` 構造体
- **`Speaker`** [string]: 発言者の役割名。
- **`Text`** [string]: 発言本文。
- **`Emotion`** [string]: 感情トーン (normal, excited, calm)。

### 4.2 データベースモデル (`backend/internal/db`)

#### `Broadcast` 構造体
- **`ID`** [uuid.UUID]: 一意な識別子（プライマリキー）。
- **`Title`** [string]: 放送のタイトル。
- **`StartedAt`** [time.Time]: 開始日時。
- **`EndedAt`** [*time.Time]: 終了日時。

#### `Transcript` 構造体
- **`ID`** [uuid.UUID]: 発言の一意識別子。
- **`BroadcastID`** [uuid.UUID]: 紐づく放送の ID（外部キー）。
- **`SpeakerID`** [string]: 発言者の役割名。
- **`Text`** [string]: 発言内容。
- **`Emotion`** [string]: 発言時の感情。
- **`CreatedAt`** [time.Time]: 記録日時。
