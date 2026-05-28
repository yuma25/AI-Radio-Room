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

  // 先行取得用のシーケンシャルキュー
  private preFetchQueue: ScriptItem[] = [];
  private isPreFetching: boolean = false;

  /**
   * 音声の再生が開始された時に呼び出されるコールバック。
   * UIの描画（テキストの表示）を音声と同期させるために使用する。
   */
  public onPlayStart?: (item: ScriptItem) => void;

  /**
   * キュー内の全ての音声再生が完了した時に呼び出されるコールバック。
   */
  public onQueueEmpty?: () => void;

  /**
   * 現在音声再生中かどうかを返す。
   */
  public getIsPlaying(): boolean {
    return this.isPlaying;
  }

  /**
   * コンストラクタである。
   * @param baseUrl [string] バックエンドのベースURL (例: http://localhost:8080)
   */
  constructor(baseUrl: string = "http://localhost:8080") {
    this.baseUrl = baseUrl;
  }

  /**
   * enqueue は、新しいセリフを再生待ち行列に追加するメソッドである。
   * 追加した瞬間に音声の先行取得（Pre-fetch）キューにも登録する。
   * @param item [ScriptItem] AI から受信したセリフデータ
   */
  public enqueue(item: ScriptItem): void {
    console.log(
      `音声をキューに追加: [${item.speaker}] ${item.text.slice(0, 15)}...`,
    );
    this.queue.push(item);

    // [最適化] 並列リクエストによるネットワークのパンクを防ぐため、1つずつダウンロードする
    this.enqueuePreFetch(item);

    if (!this.isPlaying) {
      this.processQueue();
    }
  }

  /**
   * 先行取得キューに追加し、処理を開始する
   */
  private enqueuePreFetch(item: ScriptItem): void {
    this.preFetchQueue.push(item);
    if (!this.isPreFetching) {
      this.processPreFetchQueue();
    }
  }

  /**
   * 1つずつ順番に取得する内部メソッド
   */
  private async processPreFetchQueue(): Promise<void> {
    if (this.preFetchQueue.length === 0) {
      this.isPreFetching = false;
      return;
    }

    this.isPreFetching = true;
    const item = this.preFetchQueue.shift();

    if (item) {
      await this.preFetch(item);
    }

    // 次のアイテムを処理
    this.processPreFetchQueue();
  }

  /**
   * preFetch は、音声データを事前にダウンロードして Blob URL を作成する内部メソッドである。
   * @param item [ScriptItem]
   */
  private async preFetch(item: ScriptItem): Promise<void> {
    // すでにローカル音源があるか、取得済みの場合は何もしない
    if (item.local_audio || item.audioBlobUrl) return;

    try {
      const params = new URLSearchParams({
        text: item.audio_text,
        speaker: item.speaker_id.toString(),
      });
      const response = await fetch(`${this.baseUrl}/api/tts?${params}`);
      if (!response.ok) return;

      const blob = await response.blob();
      item.audioBlobUrl = URL.createObjectURL(blob);
      console.log(`先行取得完了: ${item.text.slice(0, 10)}...`);
    } catch (e) {
      console.warn("先行取得失敗:", e);
    }
  }

  /**
   * processQueue は、キュー内のセリフを順番に処理する内部メソッドである。
   */
  private async processQueue(): Promise<void> {
    if (this.queue.length === 0) {
      this.isPlaying = false;
      if (this.onQueueEmpty) {
        this.onQueueEmpty();
      }
      return;
    }

    this.isPlaying = true;
    const item = this.queue.shift();

    if (item) {
      try {
        await this.play(item);
      } catch (error) {
        console.error("音声再生エラー:", error);
        if (this.onPlayStart) this.onPlayStart(item);
      }
    }

    this.processQueue();
  }

  /**
   * play は、音声を再生する内部メソッドである。
   * 先行取得済みのデータがあればそれを使用し、なければその場で取得する。
   * @param item [ScriptItem]
   */
  private play(item: ScriptItem): Promise<void> {
    return new Promise((resolve, reject) => {
      // 1. ローカル音源パス
      // 2. 先行取得済みの Blob URL
      // 3. その場で API を叩く
      const getAudioUrl = async (): Promise<string> => {
        if (item.local_audio) return item.local_audio;
        if (item.audioBlobUrl) return item.audioBlobUrl;

        console.log(`リアルタイム取得: ${item.text.slice(0, 10)}...`);
        const params = new URLSearchParams({
          text: item.text,
          speaker: item.speaker_id.toString(),
        });
        const response = await fetch(`${this.baseUrl}/api/tts?${params}`);
        const blob = await response.blob();
        return URL.createObjectURL(blob);
      };

      getAudioUrl()
        .then((url) => {
          const audio = new Audio(url);

          audio.onended = () => {
            // 先行取得した URL の場合は、ここでメモリ解放
            if (url.startsWith("blob:")) {
              URL.revokeObjectURL(url);
            }
            // [極限最適化] セリフ間の「間」を 0.1秒にし、沈黙を排除
            setTimeout(resolve, 100);
          };

          audio.onerror = (e) => reject(e);

          if (this.onPlayStart) {
            this.onPlayStart(item);
          }

          audio.play().catch(reject);
        })
        .catch(reject);
    });
  }
}
