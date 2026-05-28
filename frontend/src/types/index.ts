/**
 * AI Radio Room フロントエンド共通型定義
 *
 * このファイルでは、バックエンド (Go) と整合性を保つためのデータ構造を定義する。
 */

/**
 * Character は、ラジオ番組に登場する個別のキャラクター設定である。
 */
export interface Character {
  role: "mc" | "guest-a" | "guest-b"; // 役割名
  name: string; // 表示名
  personality: string; // 性格
  description: string; // 役割詳細
}

/**
 * GeneratorConfig は、台本生成の入力パラメータである。
 */
export interface GeneratorConfig {
  theme: string; // 放送テーマ
  characters: Character[]; // キャラクターリスト
  rules: string[]; // 構成ルール
}

/**
 * ScriptItem は、AI から受信する 1 セリフのデータである。
 */
export interface ScriptItem {
  speaker: string; // 発言者の役割 (mc, guest-a, guest-b)
  speaker_id: number; // VOICEVOX の話者スタイルID
  text: string; // 画面表示用のセリフ（漢字混じり）
  audio_text: string; // 音声合成用のセリフ（読み仮名主体）
  emotion: "normal" | "excited" | "calm"; // 感情
  local_audio?: string; // [拡張] 事前生成されたローカル音声ファイルへのパス
  audioBlobUrl?: string; // [拡張] 先行取得（Pre-fetch）された音声データの Blob URL
}
