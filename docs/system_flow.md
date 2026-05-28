# 🌊 システム処理フロー (System Flow)

本ドキュメントは、AI Radio Room における主要な処理の流れをシーケンス図として可視化したものである。

---

## 🏗️ 1. サーバー起動・初期化シーケンス
> システムがリクエストを受け付け可能になるまでの基盤構築フロー。

```mermaid
%%{init: {'theme': 'base', 'themeVariables': { 'primaryColor': '#eef2ff', 'primaryTextColor': '#312e81', 'primaryBorderColor': '#4338ca', 'lineColor': '#6366f1', 'secondaryColor': '#f0fdf4', 'tertiaryColor': '#fff'}}}%%
sequenceDiagram
    autonumber
    
    box #f8fafa "External / Config"
        participant Env as OS Environment / .env
    end
    
    box #eef2ff "Backend Process"
        participant Main as cmd/server/main.go
        participant AI as internal/ai (Gemini)
        participant VV as internal/voicevox (Client)
        participant DB as internal/db (Repository)
    end
    
    box #f0fdf4 "Infrastructure"
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
    
    Note right of Main: 🚀 ポート :8080 で待機開始
```

---

## 🎙️ 2. 放送開始・複数フェーズ台本生成
> AI が「導入・本題・結び」の 3 段階で台本を書き、逐次配信するフロー。

```mermaid
%%{init: {'theme': 'base', 'themeVariables': { 'primaryColor': '#fff7ed', 'primaryTextColor': '#9a3412', 'primaryBorderColor': '#ea580c', 'lineColor': '#f97316'}}}%%
sequenceDiagram
    autonumber

    box #fafafa "Client (Frontend)"
        participant UI as Studio UI (Monitor)
    end

    box #fff7ed "Backend (Go)"
        participant Handler as api/broadcast.go
        participant AI as internal/ai
        participant VV as internal/voicevox
    end

    UI->>Handler: POST /api/broadcasts (theme)
    
    loop Phase: 導入 -> 本題 -> 結び
        Handler->>AI: GenerateRadioScript(History)
        AI-->>Handler: []ScriptItem (text + audio_text)
        
        par 並列音声合成 (Goルーチン)
            Handler->>VV: GenerateWav(audio_text)
            Note over VV: Singleflight で集約
            VV-->>Handler: 合成完了
        and 逐次データ配信 (SSE)
            loop 各セリフ
                Handler->>UI: SSE data: {JSON}
                Note right of UI: 受信後、音声プリフェッチ
            end
        end
        
        Handler->>Handler: History を更新して次のフェーズへ
    end

    Handler->>UI: event: close
```

---

## 🎧 3. 音声・UI 同期再生フロー
> フロントエンドで音声とアバター演出、テロップを完全に一致させるフロー。

```mermaid
%%{init: {'theme': 'base', 'themeVariables': { 'primaryColor': '#f5f3ff', 'primaryTextColor': '#5b21b6', 'primaryBorderColor': '#7c3aed', 'lineColor': '#8b5cf6'}}}%%
sequenceDiagram
    autonumber

    box #fafafa "Frontend Logic"
        participant Player as VoicePlayer.ts
        participant Monitor as MonitorUI.ts
    end

    box #f5f3ff "Backend"
        participant TTS as /api/tts
    end

    Note over Player: SSE 受信 & キュー登録
    
    rect #ede9fe
    Note over Player: シーケンシャル・プリフェッチ
    Player->>TTS: GET /api/tts (audio_text)
    TTS-->>Player: WAV (Blob)
    end

    Note over Player: 再生準備完了
    Player->>Player: Audio.play()
    
    critical ⚡ 再生開始コールバック
        Player->>Monitor: onPlayStart(item)
        Monitor->>Monitor: アバターを「speaking」に強調
        Monitor->>Monitor: テロップをタイプライター表示
    end

    Player->>Monitor: 再生終了後に強調解除
```

---

## 🗄️ 4. データベース永続化
> 放送履歴と発言ログの記録。

```mermaid
%%{init: {'theme': 'base', 'themeVariables': { 'primaryColor': '#f0fdfa', 'primaryTextColor': '#115e59', 'primaryBorderColor': '#0d9488', 'lineColor': '#14b8a6'}}}%%
sequenceDiagram
    autonumber

    box #ffffff "Backend"
        participant Repo as internal/db (Repository)
    end

    box #f0fdfa "Storage"
        participant DB as PostgreSQL
    end

    Note over Repo, DB: 放送セッション開始
    Repo->>DB: INSERT INTO broadcasts
    DB-->>Repo: UUID 発行

    loop 配信中
        Repo->>DB: INSERT INTO transcripts
    end

    Note over Repo, DB: 放送終了
    Repo->>DB: UPDATE SET ended_at = NOW
```
