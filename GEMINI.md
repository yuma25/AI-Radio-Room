# AI Radio Room Project Instructions

## Tech Stack (Confirmed)
- **Frontend**: TypeScript (Vanilla), Vite, pnpm
- **Backend**: Go (Standard library + chi), Gemini API
- **Database**: PostgreSQL
- **Environment**: Docker & Docker Compose
- **Communication**: SSE (Server-Sent Events)

## Engineering Standards
- **Pre-commit Checks**: NEVER commit code that fails Lint, Format, or Build.
  - Frontend: `pnpm run lint`, `pnpm run format`, `pnpm run build`
  - Backend: `go fmt ./...`, `go vet ./...`, `go build ./...`
- **Cross-Platform**: Support both WSL2 (Linux) and Windows native development.
- **Environment Variables**: Use `.env` for secrets (API keys, DB passwords). Never commit `.env`.
- **Database**: Use Docker for PostgreSQL (`docker compose up -d db`). Run App logic natively for performance.
- **AI API**: Use Google Gemini API via official Go SDK. Read API keys from env vars.
- **Documentation**: すべてのコード要素（パッケージ、構造体、関数など）に、JavaDoc形式に近い詳細な日本語コメントを必須とする。
  - **記述形式**:
    ```go
    /**
     * [概要]
     * @param 引数名 [型] 説明
     * @return 戻り値名 [型] 説明
     */
    ```
  - **文体**: 日本語コメントの語尾は「である」調（常体）に統一する。
  - **記載内容**:
    - 処理の詳細なステップ。
    - すべての `()` を含む関数・メソッド呼び出しに対し、所属パッケージ、引数、戻り値、役割を明記する。
    - 引数や戻り値が「なし」の場合も、その旨を明記する。

## Commit Standards
- **AIによるコミット時の必須確認事項**:
  - すべての `()` に対するコメント記載が漏れなく完了していること。
  - `scripts/pre-commit-check.sh` がパスしていること。
- **Git Hooks**:
  - `pre-commit`: ブランチ名規約 `(feat|fix|...)/issue-[番号]-概要` を検証し、ビルドを強制する。
  - `commit-msg`: プレフィックス `feat:`, `fix:`, `docs:`, `style:`, `refactor:`, `chore:` を必須化する。

## AI Acknowledgement
This project is developed with the active collaboration of AI agents.
- **Gemini-CLI[bot]** (`gemini-cli[bot]@users.noreply.github.com`): Acting as a primary developer and architect.
- **GitHub-Copilot[bot]** (`github-copilot[bot]@users.noreply.github.com`): Providing development support and suggestions.
- **AI-Driven Development**: All code is reviewed and refined by AI agents.
- **AI-Driven Development**: All code is reviewed and refined by AI agents.
