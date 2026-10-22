# AI Radio Room (双方向AIラジオ生放送システム)

Google Gemini API による対話推論とローカルの VOICEVOX Engine による音声合成を組み合わせた、自律型の双方向AIラジオ生放送Webアプリケーションです。

---

## 1. ディレクトリ構成

フロントエンド関連・バックエンド関連のファイルがそれぞれのフォルダ内に完全に独立して自己完結する構成になっています。

```text
AI-Radio-Room/
├── .env                              # プロジェクト共通の環境変数設定
├── .env.example                      # 環境変数テンプレート
├── .gitignore                        # Git除外設定
├── docker-compose.yml                # VOICEVOX / Backend / Frontend 一括起動設定
├── README.md                         # 本ドキュメント
│
├── backend/                          # 【バックエンド】Go 1.24 / クリーンアーキテクチャ全8層
│   ├── cmd/
│   │   └── server/
│   │       └── main.go               # 【第8層: Main】エントリーポイント
│   ├── internal/
│   │   ├── domain/
│   │   │   ├── model/                # 【第1層: Domain Model】中核ルール・エンティティ
│   │   │   └── repository/           # 【第2層: Domain Repository Interface】契約書
│   │   ├── usecase/                  # 【第3層: UseCase】業務シナリオ
│   │   ├── adapter/
│   │   │   ├── gateway/              # 【第4層: Adapter Output / Gateway】外部通信・マッピング
│   │   │   └── handler/              # 【第5層: Adapter Input / Handler】リクエスト受付・DTO変換
│   │   ├── infrastructure/           # 【第6層: Infrastructure】HTTPサーバー・ルーティング
│   │   └── di/                       # 【第7層: DI Container】依存性注入・配線
│   ├── .dockerignore
│   ├── Dockerfile                    # Go マルチステージビルド
│   └── go.mod                        # Go モジュール設定
│
└── frontend/                         # 【フロントエンド】React 19 / TypeScript / Vite / 全8層
    ├── src/
    │   ├── domain/
    │   │   ├── model/                # 【第1層: Domain Model】純粋TS型・不変条件
    │   │   └── repository/           # 【第2層: Domain Repository Port】外部通信契約
    │   ├── usecase/                  # 【第3層: UseCase】業務手順シナリオ
    │   ├── adapter/
    │   │   ├── gateway/              # 【第4層: Adapter Output / Gateway】API通信・マッピング
    │   │   └── presenter/            # 【第6層: Adapter Input / Presenter】カスタムフック・状態管理
    │   ├── infrastructure/           # 【第5層: Infrastructure】HTTP通信クライアント基盤
    │   ├── presentation/
    │   │   └── components/           # 【第7層: Presentation / View】受動的UI描画
    │   ├── di/                       # 【第8層: Application Setup / DI Container】
    │   ├── App.tsx                   # メインコンポーネント
    │   ├── main.tsx                  # ブラウザDOMマウント
    │   └── index.css                 # スタイルシート
    ├── .dockerignore
    ├── Dockerfile                    # Node ビルド ➔ Nginx 配信
    ├── index.html                    # HTMLテンプレート
    ├── nginx.conf                    # Nginx設定（SPAルーティング・SSE対応）
    ├── node_modules/                 # フロント専用依存パッケージ
    ├── package.json                  # フロント専用依存定義
    ├── pnpm-lock.yaml                # ロックファイル
    ├── postcss.config.js             # PostCSS設定
    ├── tailwind.config.js            # Tailwind CSS設定
    ├── tsconfig.json                 # TypeScript設定（ルート）
    ├── tsconfig.app.json             # TypeScript設定（アプリ）
    ├── tsconfig.node.json            # TypeScript設定（ビルドツール）
    └── vite.config.ts                # Vite設定（プロキシ・エイリアス）
```

---

## 2. アーキテクチャのレイヤ構造と役割

### バックエンド（Go）
- **【第1層: Domain Model】**: 中核データ（エンティティや値オブジェクト）と絶対のビジネスルール。SQL/JSON/HTTPの依存厳禁。
- **【第2層: Domain Repository Interface】**: UseCaseが必要とするデータアクセスのメソッド定義（契約）。
- **【第3層: UseCase】**: 業務シナリオ（段取り）を制御。HTTPやSQLを直接触らない。
- **【第4層: Adapter Output / Gateway】**: 第2層のインターフェースを実装し、外部API/DBと通信・型変換。
- **【第5層: Adapter Input / Handler】**: HTTPリクエストを受け付け、DTOに変換してUseCaseを呼び出す。
- **【第6層: Infrastructure】**: Webサーバー起動、ルーティング、ミドルウェア等の物理インフラ。
- **【第7層: DI Container】**: 起動時に各レイヤをインスタンス化して依存関係を配線・注入する。
- **【第8層: Main】**: 環境変数を読み込み、サーバー起動・グレースフルシャットダウンを制御する電源スイッチ。

### フロントエンド（React / TypeScript）
- **【第1層: Domain Model】**: 純粋なTypeScriptによる中核データ型と不変条件ルール（React/DOM依存厳禁）。
- **【第2層: Domain Repository Port】**: 外部通信・データ取得の契約インターフェース。
- **【第3層: UseCase】**: 画面描画に依存しない業務手順（シナリオ）制御（Hooks/JSX厳禁）。
- **【第4層: Adapter Output / Gateway】**: 第2層のポートを実装し、外部APIと通信してドメイン型へ変換。
- **【第5層: Infrastructure】**: 生の fetch 制御、認証ヘッダー付与、タイムアウト等の通信基盤。
- **【第6層: Adapter Input / Presenter】**: カスタムフック。画面状態（State）を管理し、Viewからの操作を受けてUseCaseを呼ぶ。
- **【第7層: Presentation / View】**: 受動的なコンポーネント（JSXマークアップとUI描画に専念）。
- **【第8層: Application Setup / DI Container】**: 各レイヤを組み立てて結合し、ブラウザDOMにマウントする。

---

## 3. 環境変数設定

ルートディレクトリの `.env.example` をコピーして `.env` を作成します。

```bash
cp .env.example .env
```

| 変数名 | 説明 | 例 |
| :--- | :--- | :--- |
| `PORT` | バックエンドの待受ポート | `8080` |
| `ENV` | 動作環境 | `development` |
| `VOICEVOX_URL` | VOICEVOX Engine の接続先URL | `http://localhost:50021` |
| `GEMINI_API_KEY` | Google Gemini API キー | `AIzaSy...` |
| `GEMINI_MODEL` | 使用するGeminiモデル | `gemini-2.5-flash` |
| `VITE_API_URL` | フロントエンドが接続するバックエンドURL | `http://localhost:8080` |

---

## 4. 起動手順

### 開発環境での個別起動

1. **VOICEVOX Engine の起動 (Docker)**:
   ```bash
   docker compose up -d voicevox
   ```
2. **バックエンドの起動 (Go)**:
   ```bash
   cd backend
   go run cmd/server/main.go
   ```
3. **フロントエンドの起動 (Vite)**:
   ```bash
   cd frontend
   pnpm run dev
   ```

### Docker Compose による一括起動
```bash
docker compose up -d
```
