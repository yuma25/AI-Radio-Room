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

## Documentation
- すべてのコード要素（パッケージ、構造体、インターフェース、関数など）に、JavaDoc形式に近い詳細な日本語コメントを必須とする。
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
    - **パッケージ**: パッケージ全体の役割に加え、含まれる主要な構造体、インターフェース、関数、およびそれらの相互関係（どの構造体がどの主要な機能を担い、他の要素とどう連携するかなど）を網羅的に記載する。
    - **インポート**: インポートされている各外部パッケージについて、本ファイル内のどの構造体、インターフェース、関数で使用されているか、およびその依存関係の目的を詳細に記載する。
    - **構造体・インターフェース**: 各フィールドやメソッドが何を表し、どのような役割を持つかを詳細に記載する。
    - **処理の詳細**: 内部ロジックのステップを明記する。
    - **SDKと自作ロジックの区別**: `()` を含む呼び出しに対し、それが外部SDK（パッケージ名）のものか、内部実装のものかを明確に区別して記載する。
    - **詳細なメタ情報**: すべての `()` を含む呼び出しに対し、所属パッケージ、引数、戻り値、役割を明記する。
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
