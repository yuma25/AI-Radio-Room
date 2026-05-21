#!/bin/bash
# AI Radio Room - Pre-commit Hook
# コミット前にビルド、フォーマット、静的解析を強制実行する。

echo "🔍 コミット前の品質チェックを開始する..."

# --- Backend (Go) のチェック ---
echo "📦 Checking Backend (Go)..."
cd backend
go fmt ./...
go vet ./...
if ! go build ./...; then
    echo "❌ Backend のビルドに失敗したため、コミットを中断する。"
    exit 1
fi
cd ..

# --- Frontend (TS) のチェック ---
if [ -d "frontend/node_modules" ]; then
    echo "Frontend のチェックを実行中..."
    cd frontend
    pnpm run format
    pnpm run lint
    if ! pnpm run build; then
        echo "❌ Frontend のビルドに失敗したため、コミットを中断する。"
        exit 1
    fi
    cd ..
else
    echo "⚠️ Frontend の node_modules が見つからないため、TSのチェックをスキップする。"
fi

echo "✅ すべての品質チェックに合格した。"
exit 0
