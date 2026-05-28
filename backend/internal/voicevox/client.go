/**
 * Package voicevox は、VOICEVOX Engine ととの通信を担当するパッケージである。
 */
package voicevox

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"golang.org/x/sync/singleflight"
)

/**
 * Client は、VOICEVOX API との物理的な通信を管理する構造体である。
 */
type Client struct {
	baseURL string             // VOICEVOX Engine のベースURL (例: http://localhost:50021)
	http    *http.Client       // 再利用可能な HTTP クライアント
	Cache   *VoiceCache        // 音声データのキャッシュ
	sfGroup singleflight.Group // 同一テキストの同時リクエストを 1 つにまとめるためのグループ
}

/**
 * NewClient は、新しい VOICEVOX クライアントのインスタンスを作成して返す。
 * @param baseURL [string] エンジンのベースURL
 * @return [*Client] クライアントのポインタ
 */
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{},
		Cache:   NewVoiceCache(),
	}
}

/**
 * GenerateWav は、指定されたテキストと話者IDから音声データ(WAV)を生成する。
 * キャッシュが存在する場合はそれを返し、存在しない場合は新たに生成してキャッシュに保存する。
 *
 * 内部処理:
 * - 1. キャッシュを確認し、存在すれば即座に返す。
 * - 2. singleflight を使用し、全く同じテキスト+話者のリクエストが同時に複数来た場合、1 つの処理に集約する。
 * - 3. `audio_query` エンドポイントを呼び出し、音声合成用のパラメータを取得する。
 * - 4. 取得したパラメータを用いて `synthesis` エンドポイントを呼び出し、WAV データを生成する。
 * - 5. 生成されたデータをキャッシュに保存する。
 * @param ctx [context.Context] コンテキスト
 * @param text [string] 合成するテキスト
 * @param speakerID [int] 話者ID
 * @return ([]byte, error) WAVデータ
 */
func (c *Client) GenerateWav(ctx context.Context, text string, speakerID int) ([]byte, error) {
	key := GenerateKey(text, speakerID)

	// --- 1. キャッシュチェック ---
	if data, ok := c.Cache.Get(key); ok {
		return data, nil
	}

	// --- 2. 同時リクエストの集約 (singleflight) ---
	// 重複したリクエスト（バックグラウンド生成とフロントエンドからの取得など）を 1 つにまとめる。
	v, err, _ := c.sfGroup.Do(key, func() (interface{}, error) {
		// Do 内部で再度キャッシュを確認する（先行したリクエストが完了している可能性があるため）
		if data, ok := c.Cache.Get(key); ok {
			return data, nil
		}

		// AudioQuery の生成
		query, err := c.createAudioQuery(ctx, text, speakerID)
		if err != nil {
			return nil, fmt.Errorf("audio_query 生成失敗: %w", err)
		}

		// 音声合成 (Synthesis)
		wav, err := c.synthesis(ctx, query, speakerID)
		if err != nil {
			return nil, fmt.Errorf("synthesis 実行失敗: %w", err)
		}

		// キャッシュに保存
		c.Cache.Set(key, wav)

		return wav, nil
	})

	if err != nil {
		return nil, err
	}

	return v.([]byte), nil
}

/**
 * createAudioQuery は、テキストから音声合成パラメータ(JSON)を取得する内部メソッドである。
 */
func (c *Client) createAudioQuery(ctx context.Context, text string, speakerID int) ([]byte, error) {
	params := url.Values{}
	params.Add("text", text)
	params.Add("speaker", fmt.Sprintf("%d", speakerID))

	endpoint := fmt.Sprintf("%s/audio_query?%s", c.baseURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status error: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

/**
 * synthesis は、パラメータを元に WAV データを生成する内部メソッドである。
 */
func (c *Client) synthesis(ctx context.Context, query []byte, speakerID int) ([]byte, error) {
	params := url.Values{}
	params.Add("speaker", fmt.Sprintf("%d", speakerID))

	endpoint := fmt.Sprintf("%s/synthesis?%s", c.baseURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(query))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status error: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
