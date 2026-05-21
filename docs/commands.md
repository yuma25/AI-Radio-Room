# 開発コマンドリファレンス (Command Reference)

このドキュメントでは、AI Radio Roomの開発で使用する主要なコマンドをまとめています。

## 1. 環境構築 (Initial Setup)

### リポジトリのクローンと依存関係のインストール
```bash
# フロントエンドの準備
cd frontend
pnpm install

# バックエンドの準備
cd ../backend
go mod tidy
```

### 設定ファイルの作成
`.env.example` をコピーして `.env` を作成し、中身を編集してください。
```bash
cp .env.example .env
```

## 2. データベース管理 (Database)

データベースはDockerを使用して起動します。

| アクション | コマンド |
| :--- | :--- |
| **起動 (バックグラウンド)** | `docker compose up -d db` |
| **停止** | `docker compose stop db` |
| **削除 (データも消去)** | `docker compose down -v` |
| **ログの確認** | `docker compose logs -f db` |

## 3. バックエンド開発 (Backend / Go)

`backend` ディレクトリ内で実行します。

| アクション | コマンド |
| :--- | :--- |
| **疎通テスト実行** | `go run backend/cmd/server/main.go` |
| **依存関係の整理** | `cd backend && go mod tidy && cd ..` |
| **テストの実行** | `go test ./backend/...` |

## 4. フロントエンド開発 (Frontend / TypeScript)

| アクション | コマンド |
| :--- | :--- |
| **開発サーバー起動** | `pnpm --filter frontend dev` |
| **ビルド (本番用)** | `pnpm --filter frontend build` |
| **型チェック** | `pnpm --filter frontend exec tsc` |

## 5. コミット前の品質チェック (Pre-commit Checks)

「動かないコード」や「汚いコード」を共通ブランチに入れないため、コミット前に必ず以下のコマンドを実行してください。

### フロントエンド
```bash
cd frontend
pnpm run format  # コード整形
pnpm run lint    # 静的解析
pnpm run build   # ビルド確認
```

### バックエンド
```bash
cd backend
go fmt ./...     # コード整形
go vet ./...     # 静的解析
go build ./...   # ビルド確認
```

## 6. チーム開発のワークフロー

新しい機能を追加する際の推奨フロー：

1. `main` ブランチを最新にする: `git checkout main && git pull`
2. 作業用ブランチを作成: `git checkout -b feat/your-feature-name`
3. DBを起動: `docker compose up -d db`
4. バックエンドとフロントエンドをそれぞれ起動して開発
5. 完了したらコミットしてプルリクエストを作成
