/**
 * Package voicevox 内のキャッシュ機能を定義する。
 */
package voicevox

import (
	"fmt"
	"sync"
)

/**
 * VoiceCache は、生成された音声データ(WAV)を一時的に保持するスレッドセーフなキャッシュ構造体である。
 */
type VoiceCache struct {
	mu    sync.RWMutex
	cache map[string][]byte // キー: セリフ内容と話者IDの組み合わせ, 値: WAVデータ
}

/**
 * NewVoiceCache は、新しい VoiceCache インスタンスを生成する。
 */
func NewVoiceCache() *VoiceCache {
	return &VoiceCache{
		cache: make(map[string][]byte),
	}
}

/**
 * Get は、キャッシュからデータを取得する。
 */
func (c *VoiceCache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	data, ok := c.cache[key]
	return data, ok
}

/**
 * Set は、キャッシュにデータを保存する。
 */
func (c *VoiceCache) Set(key string, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[key] = data
}

/**
 * Clear は、キャッシュを空にする。
 */
func (c *VoiceCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string][]byte)
}

/**
 * GenerateKey は、テキストと話者IDから一意なキャッシュキーを生成する。
 */
func GenerateKey(text string, speakerID int) string {
	return fmt.Sprintf("%d:%s", speakerID, text)
}
