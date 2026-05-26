# AI Radio Room システム・リファレンス

本ドキュメントは、AI Radio Room バックエンドシステムの全容を言語化したものである。システムフロー図 (`docs/system_flow.md`) と対応しており、データ定義および具体的な処理の連鎖を詳細に解説する。

---

## 1. バックエンドの処理フロー（全体像）

システムにおける主要な動作フローは、大きく「初期化」「AI生成」「音声合成」「DB保存」のフェーズに分かれる。

### 1.1 初期化・準備フェーズ
外部環境のロードと各エンジン（AI, DB, VOICEVOX）のクライアントを準備する。

1. **環境変数のロード**:
    - `godotenv.Load()` を使用し、APIキー、DB接続URL、VOICEVOX接続URLをロードする。
2. **AI クライアントの初期化**:
    - `ai.NewGeminiClient(ctx)` を呼び出し、Gemini API との通信インターフェースを確立する。
3. **VOICEVOX クライアントの初期化**:
    - `voicevox.NewClient(baseURL)` を呼び出し、Docker コンテナ上の音声合成エンジンとの通信インターフェースを確立する。
4. **DB 接続とリポジトリ初期化**:
    - `pgx.Connect()` で接続後、`db.NewRepository(conn)` でデータ永続化層を準備する。

### 1.2 台本生成とストリーミングフェーズ
1. **放送開始リクエスト**:
    - `POST /api/broadcasts` を受け取り、`repo.CreateBroadcast` でセッションを開始する。
2. **AI 台本生成**:
    - `aiClient.GenerateRadioScript` を呼び出し、Gemini から 1 放送分のセリフリスト (`[]ScriptItem`) を一括取得する。
3. **SSE ストリーミング配信**:
    - 各 `ScriptItem` をループで処理し、以下の情報を `data: ` フィールドに含めて SSE で送信する。
    - **`speaker_id`**: `ai.GetSpeakerID(role, emotion)` を呼び出し、役割と感情に基づいた VOICEVOX スタイルIDを付与する。
4. **DB 保存**:
    - 配信と並行して `repo.SaveTranscript` を実行し、発言ログを永続化する。

### 1.3 音声合成フェーズ (TTS) と並列先行キャッシュ
フロントエンドからのリクエストに応じて、リアルタイムに音声を生成する。音声再生のラグを無くすため、バックエンド内で強力な並列キャッシュ機構を備えている。

1. **並列先行合成**:
    - `handleStartBroadcast` にて AI 台本が完成した直後、全セリフに対する音声合成を Goルーチンを用いて一斉に開始する。
    - エンジンへの負荷を考慮し、セマフォ（チャネル）によって同時実行数を制限している。
2. **キャッシュ機構**:
    - 生成された WAV データは、`voicevox.VoiceCache` (メモリ上のスレッドセーフなマップ) に `話者ID:テキスト` をキーとして一時保存される。
3. **音声合成リクエスト (TTS)**:
    - `GET /api/tts?text=...&speaker=...` を受け取る。
4. **VOICEVOX 連携 (またはキャッシュヒット)**:
    - `vvClient.GenerateWav` を呼び出す。
    - **キャッシュヒット時**: 既に先行合成が完了している場合は、API 通信を行わずにメモリから即座に WAV データを返す。
    - **キャッシュミス時**: 内部で `audio_query` と `synthesis` の 2 段階の API リクエストを VOICEVOX エンジンに送信し、WAV を生成・キャッシュして返す。
5. **音声データの返却**:
    - 生成または取得された WAV バイナリを `Content-Type: audio/wav` としてレスポンスする。

---

## 2. API サーバーと通信プロトコル (`backend/internal/api`)

### 2.1 エンドポイント定義
- **`POST /api/broadcasts`**: 放送開始と台本の SSE ストリーミング配信。
- **`GET /api/tts`**: テキストから音声バイナリ(WAV)への変換。
- **`GET /api/broadcasts`**: 過去の放送履歴一覧の取得（実装予定）。

### 2.2 技術的特記事項 (SSE)
- **バッファ制御**: 各セリフ送信後に `(http.Flusher).Flush()` を呼び出し、即時性を確保する。
- **非同期配信**: デモ用の強制的な遅延（`time.Sleep`）を排除し、SSE のテキスト配信は DB 保存と同時に最高速で完了させる。
- **終了通知**: 全セリフ送信完了後、`event: close` を送信してストリームを閉じる。先行合成の Goルーチンがバックグラウンドで処理を継続している場合でも、フロントエンドへの配信はブロックしない。

---

## 3. 共通技術規約（エンジニアリング・スタンダード）

### 3.1 Context の取り扱い
- Go の `context.Context` をすべての I/O 処理の第一引数とする。
- クライアントの切断を検知し、重い AI 処理や音声生成処理を即座にキャンセルするために `r.Context()` を末端まで伝播させる。

### 3.2 エラーハンドリング
- **Fail-Fast**: 起動時の接続失敗は `log.Fatalf` でプロセスを即座に停止させる。
- **継続性**: 放送中の部分的な保存失敗等はログ出力に留め、ユーザーへの配信を優先する。

---

## 4. データ定義詳細

### 4.1 AI 連携モデル (`backend/internal/ai`)

#### `ScriptItem` 構造体
- **`Speaker`** [string]: 役割名 (mc, guest-a, guest-b)。
- **`SpeakerID`** [int]: VOICEVOX のスタイルID（2: めたん, 3: ずんだもん等）。
- **`Text`** [string]: 発言本文。
- **`Emotion`** [string]: 感情トーン (normal, excited, calm)。

#### 話者マッピング (`ai.GetSpeakerID`)
- `mc` + `normal` -> 2 (めたん)
- `guest-a` + `normal` -> 3 (ずんだもん)
- `guest-b` -> 8 (春日部つむぎ)
- 感情に応じて「ツンツン（excited）」や「ささやき（calm）」のスタイルIDを動的に選択する。

### 4.2 データベースモデル (`backend/internal/db`)
- `broadcasts`: ID, タイトル, 開始・終了日時。
- `transcripts`: ID, 放送ID, 話者名, 内容, 感情, 作成日時。
