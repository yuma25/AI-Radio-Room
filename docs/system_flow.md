# AI Radio Room システム処理フロー

このドキュメントは、バックエンドにおける主要な処理の流れをシーケンス図として可視化したものである。

## 1. サーバー起動・初期化シーケンス
システムがリクエストを受け付け可能になるまでの基盤構築フロー。

```mermaid
%%{init: {'theme': 'dark', 'themeVariables': { 'primaryColor': '#333', 'primaryTextColor': '#fff', 'primaryBorderColor': '#777', 'lineColor': '#aaa', 'secondaryColor': '#444', 'tertiaryColor': '#222'}}}%%
sequenceDiagram
    autonumber
    
    box rgb(30, 50, 80) External / Config
        participant Env as OS Environment
        participant DotEnv as godotenv
    end
    
    box rgb(60, 60, 60) Backend Process
        participant Main as cmd/server/main.go
    end
    
    box rgb(30, 60, 30) Infrastructure
        participant DB as PostgreSQL (pgx)
    end

    Main->>Main: [内部] 起動シーケンス開始
    
    Note over Main, DotEnv: 1. 環境設定の準備
    Main->>DotEnv: [SDK] Load(".env")
    DotEnv-->>Main: 変数をプロセスに展開
    
    Note over Main, Env: 2. 認証情報の検証
    Main->>Env: [標準] Getenv("GEMINI_API_KEY")
    Env-->>Main: APIキー取得
    Main->>Main: [内部] マスク表示・存在チェック
    
    Note over Main, DB: 3. 永続化層の疎通確認
    Main->>DB: [SDK] Connect(DATABASE_URL)
    DB-->>Main: 物理接続確立
    Main->>DB: [内部] SELECT '疎通確認'
    DB-->>Main: クエリ応答成功
    
    Note right of Main: 全基盤の準備完了
```

## 2. 動的台本生成シーケンス
フロントエンドからの要求（GeneratorConfig）に基づいて AI が台本を書き上げるフロー。

```mermaid
%%{init: {'theme': 'dark', 'themeVariables': { 'primaryColor': '#333', 'primaryTextColor': '#fff', 'primaryBorderColor': '#777', 'lineColor': '#aaa', 'secondaryColor': '#444', 'tertiaryColor': '#222'}}}%%
sequenceDiagram
    autonumber

    box rgb(60, 60, 60) Application Logic
        participant App as Server / Test Logic
        participant AI as internal/ai (GeminiClient)
    end

    box rgb(80, 50, 30) AI Cloud Service
        participant SDK as Google genai SDK
        participant API as Gemini 2.5 Flash API
    end

    App->>AI: [内部] NewGeminiClient()
    AI->>SDK: [SDK] NewClient()
    SDK-->>AI: インスタンス返却

    App->>AI: [内部] GenerateRadioScript(config)
    
    Note over AI, SDK: プロンプト・エンジニアリング
    AI->>AI: [内部] BuildSystemPrompt(config)
    
    Note over AI, API: AI との対話実行
    AI->>SDK: [SDK] GenerateContent(prompt)
    SDK->>API: HTTP POST (Secure)
    API-->>SDK: 生成テキスト (JSON形式)
    SDK-->>AI: レスポンス構造体
    
    Note over AI, App: 結果のクレンジングとパース
    AI->>AI: [内部] Markdown 除去・トリミング
    AI->>AI: [標準] JSON Unmarshal
    AI-->>App: []ScriptItem (完全な台本データ)
```

## 3. データベース永続化シーケンス
放送内容を履歴として DB に記録し、整合性を保つフロー。

```mermaid
%%{init: {'theme': 'dark', 'themeVariables': { 'primaryColor': '#333', 'primaryTextColor': '#fff', 'primaryBorderColor': '#777', 'lineColor': '#aaa', 'secondaryColor': '#444', 'tertiaryColor': '#222'}}}%%
sequenceDiagram
    autonumber

    box rgb(60, 60, 60) Backend
        participant App as Broadcast Manager
        participant Repo as internal/db (Repository)
    end

    box rgb(30, 60, 30) Storage
        participant DB as PostgreSQL
    end

    Note over App, DB: 放送セッションの開始
    App->>Repo: [内部] CreateBroadcast(title)
    Repo->>DB: [SDK] INSERT INTO broadcasts
    DB-->>Repo: 新規 UUID 確定
    Repo-->>App: broadcast_id

    Note over App, DB: 逐次データ保存 (セリフごと)
    loop 各セリフの確定時
        App->>Repo: [内部] SaveTranscript(data)
        Repo->>DB: [SDK] INSERT INTO transcripts
        DB-->>Repo: 保存完了 (ID/時刻)
        Repo-->>App: 構造体同期完了
    end

    Note over App, DB: セッション終了
    App->>Repo: [内部] EndBroadcast(id)
    Repo->>DB: [SDK] UPDATE (ended_at = NOW)
    DB-->>Repo: 処理成功
    Repo-->>App: 正常終了
```
