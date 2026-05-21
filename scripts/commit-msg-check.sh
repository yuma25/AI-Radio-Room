#!/bin/bash
# AI Radio Room - Commit Message Hook
# コミットメッセージのプレフィックスを強制する。

MSG_FILE=$1
MESSAGE=$(cat "$MSG_FILE")
PATTERN="^(feat|fix|docs|style|refactor|chore|test|revert)(\(.+\))?: .+$"

if [[ ! "$MESSAGE" =~ $PATTERN ]]; then
    echo "❌ エラー: コミットメッセージの形式が正しくない。"
    echo "規約: [prefix]: [内容]"
    echo "使用可能な prefix: feat, fix, docs, style, refactor, chore, test, revert"
    echo "例: feat: Gemini API連携の実装"
    exit 1
fi

exit 0
