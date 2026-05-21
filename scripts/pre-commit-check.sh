#!/bin/bash
# AI Radio Room - Pre-commit Hook
# 1. ブランチ名の命名規則チェック
# 2. ビルド、フォーマット、静的解析の強制実行
# ※ 常にプロジェクトルートから実行されることを前提とする。

echo "🔍 コミット前の品質チェックを開始する..."

# --- ブランチ名チェック ---
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
BRANCH_PATTERN="^(feat|fix|docs|style|refactor|chore|test|revert)/issue-[0-9]+-.+$"

if [[ "$CURRENT_BRANCH" == "main" || "$CURRENT_BRANCH" == "develop" ]]; then
    echo "⚠️  注意: $CURRENT_BRANCH ブランチへの直接コミットを許可しているが、極力 feature ブランチを使用すること。"
elif [[ ! $CURRENT_BRANCH =~ $BRANCH_PATTERN ]]; then
    echo "❌ エラー: ブランチ名 '$CURRENT_BRANCH' が規約に違反している。"
    echo "規約: (prefix)/issue-[番号]-概要"
    echo "例: feat/issue-1-db-schema"
    exit 1
fi

# --- Backend (Go) のチェック ---
echo "📦 Checking Backend (Go)..."
go fmt ./backend/...
go vet ./backend/...
if ! go build -o /dev/null ./backend/cmd/server/main.go; then
    echo "❌ Backend のビルドに失敗したため、コミットを中断する。"
    exit 1
fi

# --- Frontend (TS) のチェック ---
if [ -d "frontend/node_modules" ]; then
    echo "📦 Checking Frontend (TS)..."
    pnpm --filter frontend run format
    pnpm --filter frontend run lint
    if ! pnpm --filter frontend run build; then
        echo "❌ Frontend のビルドに失敗したため、コミットを中断する。"
        exit 1
    fi
else
    echo "⚠️  Frontend の node_modules が見つからないため、TSのチェックをスキップする。"
fi

echo "✅ すべての品質チェックに合格した。"
exit 0
