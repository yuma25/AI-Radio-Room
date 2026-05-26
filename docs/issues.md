# GitHub Issue 分割案 (Issue Roadmap)

GitHub上の Issue 管理と同期させた最新のロードマップです。

---

## Issue #1: 【Backend】DBスキーマの作成とデータアクセス層の実装
**ステータス**: ✅ Closed
**ブランチ名**: `feat/issue-1-db-schema`

---

## Issue #2: 【Backend】Gemini APIによる台本自動生成ロジックの実装
**ステータス**: ✅ Closed
**ブランチ名**: `feat/issue-2-gemini-ai`

---

## Issue #8: 【Backend】HTTP API サーバー基盤と SSE ストリーミングの実装
**ステータス**: 🚀 Next
**ブランチ名**: `feat/issue-8-http-sse-server`
**担当領域**: バックエンド HTTP API サーバー構築、SSE ストリーミングの実装
**編集許可ファイル**:
- `backend/cmd/server/main.go`
- `backend/internal/ai/`
- `backend/internal/db/`

**📝 概要**:
AIロジックとデータベースを統合し、フロントエンドに提供するための HTTP API サーバーを構築します。

**🎯 必須要件**:
1. chi ルーターを使用した API エンドポイントの作成
2. `POST /api/broadcasts` : 放送開始と台本生成リクエストの受付
3. `GET /api/broadcasts` : 過去の放送履歴一覧の取得
4. SSE (Server-Sent Events) による台本のリアルタイム・ストリーミング配信
5. 台本生成と同時にデータベースへの非同期保存

---

## Issue #4: 【Backend】Markdown アーカイブ生成機能（設定情報含む）の実装
**ブランチ名**: `feat/issue-4-archive`

---

## Issue #3: 【Frontend】音声再生・再生キュー管理の実装
**ブランチ名**: `feat/issue-3-voicevox`

---

## Issue #9: 【Backend】VOICEVOX 連携ロジックの実装
**ブランチ名**: `feat/issue-9-voicevox-logic`

---

## Issue #10: 【Frontend】台本生成設定 (GeneratorConfig) UI の実装
**ブランチ名**: `feat/issue-10-generator-config-ui`
