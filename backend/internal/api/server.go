/**
 * Package api は、AI Radio Room の外部向け HTTP インターフェースを提供するパッケージである。
 *
 * 【パッケージの構成要素】
 * - 1. `Server`: HTTP ハンドラー間で共有される基盤インスタンス（AI クライアント、DB リポジトリ等）を保持する構造体。
 * - 2. `NewServer`: 各種基盤コンポーネントを注入して Server インスタンスを生成するコンストラクタ。
 * - 3. `Routes`: chi ルーターを構築し、各エンドポイントにハンドラーを紐付けるメソッド。
 *
 * 【要素間の関係】
 * - `Server` は `internal/ai` および `internal/db` パッケージの機能を利用し、ビジネスロジックを HTTP 経由で公開する。
 * - 各ハンドラーメソッド（Broadcast 等）は `Server` のメソッドとして定義され、内部のリソースにアクセスする。
 */
package api

import (
	"net/http" // 使用箇所: Routes, 各ハンドラー。目的: HTTP ステータスコードやレスポンス制御のため。

	"backend/internal/ai"       // 使用箇所: Server 構造体。目的: 台本生成ロジックの呼び出しのため。
	"backend/internal/db"       // 使用箇所: Server 構造体。目的: 放送データの永続化操作のため。
	"backend/internal/voicevox" // 使用箇所: Server 構造体。目的: 音声合成ロジックの呼び出しのため。

	"github.com/go-chi/chi/v5"            // 使用箇所: Routes。目的: パスベースのルーティング制御のため。
	"github.com/go-chi/chi/v5/middleware" // 使用箇所: Routes。目的: 標準的なロギングやリカバリ機能のため。
	"github.com/go-chi/cors"              // 使用箇所: Routes。目的: フロントエンドからのクロスオリジンリクエストを許可するため。
)

/**
 * Server は、API サーバーの状態を管理する構造体である。
 */
type Server struct {
	aiClient *ai.GeminiClient // [SDK] 台本生成用の AI クライアント
	repo     *db.Repository   // [内部] データベース操作用のリポジトリ
	vvClient *voicevox.Client // [内部] 音声合成用の VOICEVOX クライアント
}

/**
 * NewServer は、新しい API Server のポインタを返す。
 *
 * 内部処理:
 * - 1. 引数として受け取った AI クライアント、DB リポジトリ、VV クライアントを構造体に格納して返す。
 * @param aiClient [*ai.GeminiClient] 初期化済みの AI クライアント
 * @param repo [*db.Repository] 初期化済みの DB リポジトリ
 * @param vvClient [*voicevox.Client] 初期化済みの VOICEVOX クライアント
 * @return [*Server] API Server のポインタ
 */
func NewServer(aiClient *ai.GeminiClient, repo *db.Repository, vvClient *voicevox.Client) *Server {
	return &Server{
		aiClient: aiClient,
		repo:     repo,
		vvClient: vvClient,
	}
}

/**
 * Routes は、API のルーティング設定を構築して返す。
 *
 * 内部処理:
 * - 1. `chi.NewRouter()` を使用してルーターを生成する。
 * - 2. `middleware.Logger` や `middleware.Recoverer` を適用し、安定性とデバッグ性を確保する。
 * - 3. `cors.Handler` を使用して、Vite デフォルトの 5173 ポート等からの通信を許可する設定を行う。
 * - 4. `/api/broadcasts` 等のパスに対して、各ハンドラーメソッドをマッピングする。
 * @param なし
 * @return [http.Handler] 設定済みの HTTP ハンドラー
 */
func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()

	// --- 1. ミドルウェアの設定 ---
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// --- 2. CORS の設定 ---
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// --- 3. エンドポイントの定義 ---
	r.Route("/api", func(r chi.Router) {
		r.Post("/broadcasts", s.handleStartBroadcast) // 放送開始
		r.Get("/broadcasts", s.handleListBroadcasts)  // 履歴一覧
		r.Get("/tts", s.handleTTS)                    // 音声合成 (Text-to-Speech)
	})

	return r
}
