package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	// 1. .env ファイルの読み込み
	// (エラーになっても、環境変数として設定されている可能性があるので続行)
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Println(".env ファイルが見つかりません。環境変数から直接読み込みます。")
	}

	// 2. Gemini APIキーの確認 (疎通テスト)
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY が設定されていません。")
	}
	fmt.Printf("✅ Gemini API Key Loaded: [ %s... ]\n", apiKey[:5])

	// 3. データベース接続テスト
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL が設定されていません。")
	}

	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("❌ データベースへの接続に失敗しました: %v\n", err)
	}
	defer conn.Close(context.Background())

	var greeting string
	err = conn.QueryRow(context.Background(), "select 'Database Connected!'").Scan(&greeting)
	if err != nil {
		log.Fatalf("❌ クエリの実行に失敗しました: %v\n", err)
	}

	fmt.Printf("✅ %s\n", greeting)
	fmt.Println("--------------------------------------------------")
	fmt.Println("サーバーを起動する準備が整いました。")
}
