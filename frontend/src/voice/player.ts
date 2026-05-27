/**
 * Package voice は、VOICEVOX 音声の取得および再生キューの管理を担当する。
 */

import { ScriptItem } from "../types";

/**
 * VoicePlayer は、バックエンドの TTS API を利用して音声を再生するクラスである。
 * 複数のセリフが重ならないよう、キュー（待ち行列）による順次再生を制御する。
 */
export class VoicePlayer {
  private queue: ScriptItem[] = [];
  private isPlaying: boolean = false;
  private baseUrl: string;

  /**
   * 音声の再生が開始された時に呼び出されるコールバック。
   * UIの描画（テキストの表示）を音声と同期させるために使用する。
   */
  public onPlayStart?: (item: ScriptItem) => void;

  /**
   * コンストラクタである。
   * @param baseUrl [string] バックエンドのベースURL (例: http://localhost:8080)
   */
  constructor(baseUrl: string = "http://localhost:8080") {
    this.baseUrl = baseUrl;
  }

  /**
   * enqueue は、新しいセリフを再生待ち行列に追加するメソッドである。
   * 既に再生中でなければ、自動的に再生を開始する。
   * @param item [ScriptItem] AI から受信したセリフデータ
   */
  public enqueue(item: ScriptItem): void {
    console.log(`音声をキューに追加: [${item.speaker}] ${item.text}`);
    this.queue.push(item);
    if (!this.isPlaying) {
      this.processQueue();
    }
  }

  /**
   * processQueue は、キュー内のセリフを順番に処理する内部メソッドである。
   */
  private async processQueue(): Promise<void> {
    if (this.queue.length === 0) {
      this.isPlaying = false;
      return;
    }

    this.isPlaying = true;
    const item = this.queue.shift();

    if (item) {
      try {
        await this.play(item);
      } catch (error) {
        console.error("音声再生エラー:", error);
        // エラー時もUIを進めるためコールバックを呼ぶ
        if (this.onPlayStart) this.onPlayStart(item);
      }
    }

    // 次のセリフへ
    this.processQueue();
  }

  /**
   * play は、実際にバックエンドから音声を取得して再生する内部メソッドである。
   * 再生が完了するまで Promise を保留する。
   *
   * 内部処理:
   * - 1. [拡張] item.local_audio が指定されている場合は、そのパスから直接音声を取得する。
   * - 2. [標準] それ以外は fetch API を使用して、/api/tts から WAV データを取得する。
   * - 3. [標準] URL.createObjectURL を使用して、Blob を再生可能な URL に変換する。
   * - 4. [標準] HTMLAudioElement を使用して音声を再生し、ended イベントを待機する。
   * @param item [ScriptItem]
   */
  private play(item: ScriptItem): Promise<void> {
    return new Promise((resolve, reject) => {
      // ローカル音源パスがある場合はそれを利用し、なければAPIを叩く
      const audioSource = item.local_audio
        ? item.local_audio
        : (() => {
            const params = new URLSearchParams({
              text: item.text,
              speaker: item.speaker_id.toString(),
            });
            return `${this.baseUrl}/api/tts?${params}`;
          })();

      fetch(audioSource)
        .then(async (response) => {
          if (!response.ok)
            throw new Error(`音声取得エラー: ${response.statusText}`);
          return response.blob();
        })
        .then((blob) => {
          const url = URL.createObjectURL(blob);
          const audio = new Audio(url);

          // 再生終了時の処理
          audio.onended = () => {
            URL.revokeObjectURL(url); // メモリ解放
            // [最適化] セリフ間の「間」をさらに 0.1秒へ短縮し、極限までテンポを上げる
            setTimeout(resolve, 100);
          };

          audio.onerror = (e) => {
            URL.revokeObjectURL(url);
            reject(e);
          };

          // 再生開始
          console.log(
            `再生中...: ${item.text} (Source: ${item.local_audio ? "Local" : "API"})`,
          );

          // UIと同期させるため、再生開始直前にコールバックを呼ぶ
          if (this.onPlayStart) {
            this.onPlayStart(item);
          }

          audio.play().catch(reject);
        })
        .catch(reject);
    });
  }
}
