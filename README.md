# AI Radio Room (AIラジオルーム)

3人のAIパーソナリティによるリアルタイム・ラジオ番組配信アプリ。

## プロジェクト構造 (Project Structure)

本プロジェクトはモノレポ構成を採用しています。

- `/frontend`: TypeScript + Vite (pnpm) - スタジオUIおよび音声合成
- `/backend`: Go (Standard Library) - 番組進行・Gemini API連携
- `/docs`: 設計ドキュメント・仕様書

## チーム開発のルール

詳細な規約は [GEMINI.md](./GEMINI.md) を参照してください。

### 1. 開発フロー
新たな機能開発やバグ修正を行う際は、以下の手順を遵守してください。
1. **GitHub Issue を作成**: テンプレートを使用して、作業内容と編集対象ファイルを明確にする。
2. **ブランチの作成**: `feat/issue-[番号]-概要` の形式でブランチを作成する。
3. **実装とチェック**: `pre-commit` フックにより自動的に品質チェックが行われる。
4. **プルリクエストの作成**: テンプレートに従い、詳細なコメントと共にレビューを依頼する。

### 2. コミットメッセージ
- プレフィックスを付ける（例: `feat:`, `fix:`, `docs:`, `style:`, `refactor:`）

## セットアップと開発手順

詳細なコマンド一覧は [docs/commands.md](./docs/commands.md) を参照してください。

### 1. 依存関係のインストール (初回のみ)
```bash
pnpm setup
```

### 2. データベースの起動
```bash
pnpm db
```

### 3. アプリケーションの起動 (それぞれ別ターミナル)
- **バックエンド**: `pnpm dev:back`
- **フロントエンド**: `pnpm dev:front`

### 4. コミット前の品質チェック
```bash
pnpm check
```

## 共同開発者 (Contributors)

- [Gemini CLI](https://github.com/google-gemini/gemini-cli) - 開発アシスタント & メンター
- [GitHub Copilot CLI](https://github.com/features/copilot) - 開発支援
