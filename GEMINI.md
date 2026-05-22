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
    - **パッケージ**: パッケージ全体の役割に加え、含まれる主要な構造体、インターフェース、関数、およびそれらの相互関係を網羅的に記載する。
    - **インポート**: 各外部パッケージの使用箇所と導入目的を詳細に記載する。
    - **構造体・インターフェース**: 各フィールドやメソッドが何を表し、どのような役割を持つかを詳細に記載する。
    - **メソッド・関数**: すべての関数宣言に対し、詳細な引数・戻り値・技術的ステップを明記する。内部での重要なメソッド呼び出しに対しても、その意図と役割を記載する。
    - **SDKと自作ロジックの区別**: 外部SDK（パッケージ名）の呼び出しと、内部実装を明確に区別して記載する。

- **用語集（グロッサリ）の管理**:
  - アルファベットの略称や専門用語が登場した際は、必ずその意味を解説する。
  - プロジェクト全体で一貫した理解を保つため、`docs/glossary.md` を作成・随時更新する。
  - 新規機能の実装やドキュメント作成時に新しい用語を導入した場合は、この用語集への追記を必須とする。

- **システム・リファレンス（言語化されたフロー）の義務化**:
  - コミット時や大きな機能追加時には、必ず `docs/system_reference.md` を作成・更新する。
  - これは `docs/system_flow.md` （シーケンス図）の言語化版であり、アプリ全体の一貫した処理の流れを、具体的なメソッド名・引数・戻り値とともに詳細に解説するものである。
  - 構造体がどのインターフェースを満たしているか（実装関係）や、継承・委譲の関係についても明記する。

## Commit Standards
- **AIによるコミット時の必須確認事項**:
  - すべてのメソッド・関数に対する詳細コメント記載が完了していること。
  - `docs/system_reference.md` が最新のコード状態を反映していること。
  - `scripts/pre-commit-check.sh` がパスしていること。
- **Git Hooks**:
  - `pre-commit`: ブランチ名規約 `(feat|fix|...)/issue-[番号]-概要` を検証し、ビルドを強制する。
  - `commit-msg`: プレフィックス `feat:`, `fix:`, `docs:`, `style:`, `refactor:`, `chore:` を必須化する。

## AI Acknowledgement
This project is developed with the active collaboration of AI agents.
- **Gemini-CLI[bot]** (`gemini-cli[bot]@users.noreply.github.com`): Acting as a primary developer and architect.
- **GitHub-Copilot[bot]** (`github-copilot[bot]@users.noreply.github.com`): Providing development support and suggestions.
- **AI-Driven Development**: All code is reviewed and refined by AI agents.
