# GitHub Issue 分割案 (Issue Roadmap)

チーム開発においてコンフリクトを避け、AI（Copilot/Gemini）を最大限活用するためのIssue定義です。

---

## Issue #1: 【Backend】DBスキーマの作成とデータアクセス層の実装
**担当領域**: データベースとGoの接続部分
**編集許可ファイル**:
- `backend/migrations/001_init.sql`
- `backend/internal/db/repository.go`
- `backend/internal/db/models.go`

**📝 AIへの指示用プロンプト**:
```text
[Task]
PostgreSQLのテーブル定義と、Go言語からのデータ操作（CRUD）インターフェースを実装してください。
docs/design.md に定義されている `broadcasts` と `transcripts` の2つのテーブルを作成します。

[Requirements]
1. `backend/migrations/001_init.sql` を作成し、2つのテーブルの CREATE TABLE 文を記述する。
2. `backend/internal/db/models.go` を作成し、DBの各カラムに対応するGoの構造体（struct）を定義する。
3. `backend/internal/db/repository.go` を作成し、`pgx` を使って「放送の新規作成」「セリフの保存」「放送IDに紐づく全セリフの取得」を行う関数を実装する。

[Constraints]
- main.go や他のディレクトリのファイルは絶対に修正しないこと。
- Lint (go vet) と Format (go fmt) が通るコードであること。
```

---

## Issue #2: 【Backend】Gemini APIによる台本自動生成ロジックの実装
**担当領域**: AIへのプロンプト送信とJSONパース
**編集許可ファイル**:
- `backend/internal/ai/gemini.go`
- `backend/internal/ai/prompt.go`

**📝 AIへの指示用プロンプト**:
```text
[Task]
Google Generative AI (Gemini API) を使用して、3人のキャラクターによるラジオ番組の台本を生成するGoのパッケージを実装してください。

[Requirements]
1. `backend/internal/ai/gemini.go` を作成し、Gemini APIクライアントの初期化処理を書く。
2. 引数として「テーマ（文字列）」を受け取り、Geminiにリクエストを送る関数を作成する。
3. Geminiからのレスポンスを、指定されたJSONスキーマにパースしてGoの構造体（スライス）として返す。
4. `backend/internal/ai/prompt.go` にシステムプロンプトの定数を定義する。
```

---

## Issue #3: 【Frontend】VOICEVOX 音声合成エンジンのクライアント実装
**担当領域**: フロントエンドからVOICEVOX APIを叩き、音声を鳴らす部分
**編集許可ファイル**:
- `frontend/src/voice.ts`
- `frontend/src/types.ts`

**📝 AIへの指示用プロンプト**:
```text
[Task]
フロントエンド(TypeScript)から、ローカルの VOICEVOX Engine (http://localhost:50021) と通信し、音声を再生するモジュールを実装してください。

[Requirements]
1. `frontend/src/types.ts` を作成し、セリフデータの型を定義する。
2. `frontend/src/voice.ts` を作成し、`playVoice` 関数を実装する。
3. WAVデータを取得し、ブラウザで音声を再生。終了を検知するPromiseを返すこと。
```

---

## Issue #4: 【Backend】Markdown アーカイブ生成機能の実装
**担当領域**: 放送終了後に、テキストファイル（.md）を出力する部分
**編集許可ファイル**:
- `backend/internal/archive/markdown.go`
- `backend/archives/` (ディレクトリ)

**📝 AIへの指示用プロンプト**:
```text
[Task]
放送のセリフデータを受け取り、整形されたMarkdownファイルを指定ディレクトリに出力するGoのパッケージを実装してください。
出力形式は docs/design.md の仕様に従ってください。
```

---

## Issue #5: 【Frontend】UI制御とSSEキューイングの統合
**担当領域**: UI描画、SSE受信、音声再生の同期
**編集許可ファイル**:
- `frontend/src/ui.ts`
- `frontend/src/main.ts`

**📝 AIへの指示用プロンプト**:
```text
[Task]
SSEを受信して画面(ui.ts)と音声(voice.ts)を同期させる司令塔を実装してください。
メッセージをキューに溜め、音声の終了を待ってから次の発言を表示・再生する排他制御を行ってください。
```
