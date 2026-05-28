# AI Radio Room (AIラジオルーム)

3人の個性豊かなAIキャラクターによるリアルタイム・ラジオ番組生成・配信アプリケーション。

## 🎙️ プロジェクト概要

AI Radio Room は、Google Gemini API と VOICEVOX を活用し、まるで本物のラジオ番組のような掛け合いをリアルタイムに生成・放送するスタジオアプリです。

単なるテキスト生成に留まらず、キャラクターの性格、番組独自のルール、テーマを自由に変更できる「動的構成エンジン」を搭載。現在は、3段階の構成（導入・本題・結び）による深みのある対談と、喋っているキャラクターが動く没入感のあるスタジオUIを提供しています。

## ✨ 主要機能

- **複数フェーズ台本生成**: 「導入」「本題」「結び」の3段階でAIを呼び出し、文脈を維持しながら長尺で自然な番組構成を実現。
- **デュアルテキスト戦略**: 画面用の「綺麗な文字（漢字混じり）」と、音声用の「正確な読み（カナ）」を同時に生成し、VOICEVOX の読み間違いを徹底排除。
- **ラジオブース UI**: 喋っているキャラクターがアニメーションし、リアルタイムテロップ（字幕）が表示されるライブモニター風の画面。
- **SSE ストリーミング**: 生成されたセリフを一行ずつリアルタイムにブラウザへ配信。
- **高性能同期再生**: シングルフライト（負荷集約）とシーケンシャル・プリフェッチ（順序取得）により、途切れることのない滑らかな放送。

## 🛠️ 技術スタック

### Backend
- **Go (Standard Library + chi)**
  - 並行処理（Goroutine）を活かした高速なストリーミング配信。
- **Google Gemini API (Gemini 2.5 Flash)**
  - 構造化出力（JSON）に特化した、高速・高精度な台本生成。
- **PostgreSQL (pgxpool)**
  - 放送履歴と発言ログの永続化。接続プールによる安定したアクセス。
- **VOICEVOX Engine**
  - 個性豊かなキャラクター音声の合成。

### Frontend
- **TypeScript (Vanilla)**
  - フレームワークに依存しない、軽量かつ型安全な UI 制御。
- **Vite**
  - 爆速な開発環境とビルド。
- **Neumorphism & Glassmorphism**
  - 柔らかさと透明感を活かした、モダンなオーディオプレイヤー風デザイン。

## 🚀 セットアップと実行

### 1. 準備 (Prerequisites)
- [Docker](https://www.docker.com/) & Docker Compose
- [Go](https://go.dev/) 1.25.0 以上
- [Node.js](https://nodejs.org/) & [pnpm](https://pnpm.io/)
- Google Gemini API キー ([Google AI Studio](https://aistudio.google.com/) で取得)
- [VOICEVOX Engine](https://voicevox.hiroshiba.jp/) (ローカルまたは Docker で起動)

### 2. 環境設定
`.env.example` をコピーして `.env` を作成し、必要な値を設定してください。
```bash
cp .env.example .env
# GEMINI_API_KEY=あなたのキー
# DATABASE_URL=postgres://user:pass@localhost:5432/radio_room
```

### 3. 初回セットアップ
```bash
pnpm setup
```

### 4. 開発用サーバーの起動
以下のコマンドで、DB、バックエンド、フロントエンドがすべて並列で起動します。
```bash
pnpm run dev
```

## ⌨️ 主要コマンド

| コマンド | 内容 |
| :--- | :--- |
| `pnpm run dev` | **推奨**: 全てのコンポーネントを一括起動 |
| `pnpm check` | **コミット前必須**: Lint/Format/Buildを一括実行 |
| `pnpm setup` | 依存関係のインストールと整理 |
| `pnpm db` | データベース(Docker)のみを起動 |
| `pnpm db:stop` | データベースを停止 |

## 📐 ドキュメント

- [システムリファレンス](./docs/system_reference.md): メソッドレベルの詳細な処理フロー
- [システムフロー図](./docs/system_flow.md): シーケンス図によるデータの流れの可視化
- [用語集](./docs/glossary.md): プロジェクト特有の技術用語定義
- [開発規約 (GEMINI.md)](./GEMINI.md): AI との共同開発における必須ルール

## 📜 ライセンス
This project is developed with the active collaboration of AI agents.
See [GEMINI.md](./GEMINI.md) for detailed acknowledgements.
