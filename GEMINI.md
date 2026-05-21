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
