# AI Radio Room 設計仕様書 (Professional Edition)

## 1. プロジェクト概要
3人のAIキャラクターによるリアルタイム・ラジオ番組配信アプリ。
放送履歴の永続化、VOICEVOXによる高品質音声、および放送終了後の自動アーカイブ機能を備える。

## 2. 通信プロトコルとインターフェース (Interface & Protocols)

### 2.1 APIプロトコル
- **REST (JSON over HTTP/2)**: 
  - 放送開始/終了コマンド、お便り投稿（将来拡張）などのリソース操作に使用。
  - **冪等性 (Idempotency)**: 同一の開始リクエストが連打されても、放送セッションが重複しないようサーバー側で排他制御を行う。
- **SSE (Server-Sent Events)**:
  - サーバーからブラウザへのセリフデータのリアルタイム配信に使用。
  - 軽量かつHTTP標準のため、Webブラウザでのデバッグと実装が容易。

### 2.2 セキュリティ
- **CORS**: フロントエンド(ポート5173)とバックエンド(8080)の分離に対応し、許可されたオリジンのみアクセスを許可。
- **秘密情報の保護**: APIキーやDB接続情報は `.env` で管理し、本番環境では環境変数インジェクションを使用。

## 3. システムアーキテクチャ (Architecture)

### 3.1 構造設計
- **モジュラー・モノリス (Modular Monolith)**:
  - 開発のシンプルさを維持しつつ、`internal/db`, `internal/ai`, `internal/archive` 等のパッケージに役割を分離。
  - 各モジュールはインターフェースを介して通信し、将来的なマイクロサービス化が可能な疎結合な設計。
- **ディレクトリ構造**: 
  - `cmd/`: エントリポイント
  - `internal/`: 非公開ビジネスロジック (Repository, Service)
  - `migrations/`: DBマイグレーションファイル

### 3.2 オブザーバビリティ (Observability)
- **構造化ロギング**: Goの標準機能または軽量ライブラリを使用し、エラーの追跡が容易なログを出力。
- **ヘルスチェック**: `/health` エンドポイントを設け、DBやVOICEVOXエンジンとの接続状態を監視可能にする。

## 4. データと状態管理 (Data Management)

### 4.1 データベース
- **PostgreSQL**: ACID特性を重視し、放送履歴の整合性を担保。
- **一貫性モデル**: **強一致性 (Strong Consistency)** を採用。放送中に生成されたセリフは即座にDBへ保存され、クライアントへ配信される。

### 4.2 スキーマ設計 (Schema Draft)
- `broadcasts`: 放送メタデータ (ID, テーマ, 開始/終了日時)
- `transcripts`: 発言ログ (ID, 放送ID, 話者, 内容, 感情, タイムスタンプ)

## 5. 音響演出と合成 (Audio Engine)

### 5.1 VOICEVOX 連携
- Dockerコンテナとして稼働する VOICEVOX Engine を使用。
- **キャラクター割り当て**:
  - MC: 四国めたん, Guest A: ずんだもん, Guest B: 春日部つむぎ 等。

### 5.2 BGM / ジングル
- `frontend/public/audio/` 内のアセットを Web Audio API で制御。
- 放送開始時のOPジングル再生と、背後でのループBGM（音量自動調整）。

## 6. 放送アーカイブ (Archiving)
- **Markdown出力**: 放送終了信号をトリガーに、DBから全データを抽出し `.md` ファイルを生成。
- **保存先**: `backend/archives/`
- **目的**: 放送内容のテキスト保存、印刷、および将来的なRAG（AIの記憶）への活用。

## 7. 開発・運用 (Infra & CI/CD)
- **Docker Compose**: 全環境（DB, VOICEVOX）を一括管理。
- **品質管理ゲート**: コミット前に Lint (ESLint, go vet), Format (Prettier, go fmt), Build の通過を必須とする。
- **CI/CD**: GitHub Actions を利用した自動ビルドとバリデーションを想定。

## 8. ロードマップ
1. **Phase 1: Foundation**: 環境構築、DBマイグレーション、疎通確認。
2. **Phase 2: Backend Logic**: Gemini API連携、台本生成ロジック。
3. **Phase 3: Audio Engine**: VOICEVOX連携、フロントエンド音声再生。
4. **Phase 4: Frontend UI**: スタジオUI構築、SSE接続による同期。
5. **Phase 5: Archive & BGM**: Markdown出力、音響演出の追加。
