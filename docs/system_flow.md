# AI Radio Room システム処理フロー

このドキュメントは、AI Radio Room における主要な処理の流れをシーケンス図として可視化したものである。

## 1. サーバー起動・初期化シーケンス
システムがリクエストを受け付け可能になるまでの基盤構築フロー。

```mermaid
%%{init: {'theme': 'dark', 'themeVariables': { 'primaryColor': '#333', 'primaryTextColor': '#fff', 'primaryBorderColor': '#777', 'lineColor': '#aaa', 'secondaryColor': '#444', 'tertiaryColor': '#222'}}}%%
sequenceDiagram
    autonumber
    
    box rgb(30, 50, 80) External / Config
        participant Env as OS Environment / .env
    end
    
    box rgb(60, 60, 60) Backend Process
        participant Main as cmd/server/main.go
        participant AI as internal/ai (Gemini)
        participant VV as internal/voicevox (Client)
        participant DB as internal/db (Repository)
    end
    
    box rgb(30, 60, 30) Infrastructure
        participant PG as PostgreSQL (pgxpool)
    end

    Main->>Main: [内部] loadEnv()
    Env-->>Main: 変数展開
    
    Main->>AI: [内部] NewGeminiClient(ctx)
    AI-->>Main: クライアント返却
    
    Main->>VV: [内部] NewClient(url)
    VV-->>Main: クライアント返却 (Singleflight内包)
    
    Main->>PG: [SDK] pgxpool.New(ctx, url)
    PG-->>Main: 接続プール確立
    
    Main->>DB: [内部] NewRepository(pool)
    DB-->>Main: リポジトリ返却
    
    Note right of Main: ポート :8080 で待機開始
```

## 2. 放送開始・複数フェーズ台本生成
AI が「導入・本題・結び」の 3 段階で台本を書き、逐次配信するフロー。

```mermaid
%%{init: {'theme': 'dark', 'themeVariables': { 'primaryColor': '#333', 'primaryTextColor': '#fff', 'primaryBorderColor': '#777', 'lineColor': '#aaa', 'secondaryColor': '#444', 'tertiaryColor': '#222'}}}%%
sequenceDiagram
    autonumber

    box rgb(40, 40, 40) Client (Frontend)
        participant UI as Studio UI (Monitor)
    end

    box rgb(60, 60, 60) Backend (Go)
        participant Handler as api/broadcast.go
        participant AI as internal/ai
        participant VV as internal/voicevox
    end

    UI->>Handler: POST /api/broadcasts (theme)
    
    loop Phase: 導入 -> 本題 -> 結び
        Handler->>AI: GenerateRadioScript(config + History)
        AI-->>Handler: []ScriptItem (text + audio_text)
        
        par 並列音声合成 (Goルーチン)
            Handler->>VV: GenerateWav(audio_text)
            Note over VV: Singleflight で重複リクエストを集約
            VV-->>Handler: 合成完了 (キャッシュ済)
        and 逐次データ配信 (SSE)
            loop 各セリフ
                Handler->>UI: SSE data: {JSON}
                Note right of UI: 受信後、音声プリフェッチ開始
            end
        end
        
        Handler->>Handler: 歴史(History)を更新して次のフェーズへ
    end

    Handler->>UI: event: close
```

## 3. 音声・UI 同期再生フロー
フロントエンドで音声とアバター演出、テロップを完全に一致させるフロー。

```mermaid
%%{init: {'theme': 'dark', 'themeVariables': { 'primaryColor': '#333', 'primaryTextColor': '#fff', 'primaryBorderColor': '#777', 'lineColor': '#aaa', 'secondaryColor': '#444', 'tertiaryColor': '#222'}}}%%
sequenceDiagram
    autonumber

    box rgb(40, 40, 40) Frontend Logic
        participant Player as VoicePlayer.ts
        participant Monitor as MonitorUI.ts
    end

    box rgb(60, 60, 60) Backend
        participant TTS as /api/tts
    end

    Note over Player: SSE 受信 & キュー登録
    
    rect rgb(50, 50, 50)
    Note over Player: シーケンシャル・プリフェッチ
    Player->>TTS: GET /api/tts (audio_text)
    TTS-->>Player: WAV (Blob)
    end

    Note over Player: 再生準備完了
    Player->>Player: Audio.play()
    
    critical 再生開始コールバック
        Player->>Monitor: onPlayStart(item)
        Monitor->>Monitor: アバターを「speaking」に強調
        Monitor->>Monitor: テロップをタイプライター表示 (text)
    end

    Player->>Monitor: 再生終了後に強調解除
```

## 4. データベース永続化
放送履歴と発言ログの記録。

```mermaid
%%{init: {'theme': 'dark', 'themeVariables': { 'primaryColor': '#333', 'primaryTextColor': '#fff', 'primaryBorderColor': '#777', 'lineColor': '#aaa', 'secondaryColor': '#444', 'tertiaryColor': '#222'}}}%%
sequenceDiagram
    autonumber

    box rgb(60, 60, 60) Backend
        participant Repo as internal/db (Repository)
    end

    box rgb(30, 60, 30) Storage
        participant DB as PostgreSQL
    end

    Note over Repo, DB: 放送セッション開始
    Repo->>DB: INSERT INTO broadcasts (title)
    DB-->>Repo: UUID 発行

    loop 配信中
        Repo->>DB: INSERT INTO transcripts (ID, Text, AudioText...)
    end

    Note over Repo, DB: 放送終了
    Repo->>DB: UPDATE broadcasts SET ended_at = NOW
```
