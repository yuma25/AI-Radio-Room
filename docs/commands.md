# 開発コマンドリファレンス (Command Reference)

このプロジェクトでは、OS（Windows/Linux）を問わず共通のコマンドを使用できるよう、プロジェクトルートに `package.json` を配置しています。

## 1. クイックスタート (Quick Start)

### 初回セットアップ
```bash
pnpm setup
```

### 開発用サーバーの起動
以下のコマンドを実行するだけで、DB、バックエンド、フロントエンドが全て並列で起動します。
```bash
pnpm run dev
```

## 2. コマンド一覧 (Command List)

| カテゴリ | コマンド | 内容 |
| :--- | :--- | :--- |
| **開発** | `pnpm run dev` | **推奨**: DB、バックエンド(Air)、フロントエンドを一括起動 |
| **データベース** | `pnpm db` | DBをバックグラウンドで起動 |
| | `pnpm db:stop` | DBを停止 |
| | `pnpm db:down` | DBを削除（データも消去） |
| **個別起動** | `pnpm dev:back` | Go バックエンドサーバーの起動 (Air) |
| | `pnpm dev:front` | Vite フロントエンドサーバーの起動 |
| **品質管理** | `pnpm check` | **コミット前必須**。Lint/Format/Buildを一括実行 |
| **環境構築** | `pnpm setup` | 依存関係のインストールと整理 |

## 3. 従来の個別コマンド (Manual Commands)

特定のディレクトリ内で個別に実行したい場合に使用します。

### バックエンド (backend/)
- `go mod tidy`: 依存関係の整理
- `go test ./...`: テスト実行

### フロントエンド (frontend/)
- `pnpm install`: パッケージインストール
- `pnpm run build`: 本番ビルド

## 4. チーム開発のワークフロー

1. `main` ブランチを最新にする: `git checkout main && git pull`
2. 作業用ブランチを作成: `git checkout -b (feat|fix)/issue-[番号]-概要`
3. 開発開始: `pnpm run dev`
4. 完了後のチェック: `pnpm check`
5. コミット & プルリクエスト作成
