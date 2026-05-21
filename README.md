# AI Radio Room (AIラジオルーム)

3人のAIパーソナリティによるリアルタイム・ラジオ番組配信アプリ。

## プロジェクト構造 (Project Structure)

本プロジェクトはモノレポ構成を採用しています。

- `/frontend`: TypeScript + Vite (pnpm) - スタジオUIおよび音声合成
- `/backend`: Go (Standard Library) - 番組進行・Gemini API連携
- `/docs`: 設計ドキュメント・仕様書

## チーム開発のための共通ルール

### 1. パッケージ管理
- フロントエンドは必ず **pnpm** を使用してください。

### 2. ブランチ戦略
- `main`: プロダクション用（安定版）
- `develop`: 開発用ブランチ
- `feature/*`: 機能追加
- `fix/*`: バグ修正

### 3. コミットメッセージ
- プレフィックスを付ける（例: `feat:`, `fix:`, `docs:`, `style:`, `refactor:`）

## セットアップ手順

詳細なコマンド一覧は [docs/commands.md](./docs/commands.md) を参照してください。

### 1. データベースの起動
```bash
docker compose up -d db
```

### 2. バックエンドの起動
```bash
cd backend
go run cmd/server/main.go
```

### 3. フロントエンドの起動
```bash
cd frontend
pnpm dev
```
