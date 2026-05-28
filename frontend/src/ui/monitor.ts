/**
 * Package ui は、ラジオブース風のライブモニター画面の描画を担当する。
 */
import { ScriptItem } from "../types";

/**
 * MonitorUI は、放送中の話者可視化とセリフのテロップ表示を管理するクラスである。
 */
export class MonitorUI {
  private subtitleText: HTMLElement;
  private speakerName: HTMLElement;
  private avatars: Map<string, HTMLElement> = new Map();
  private typingInterval: number | null = null;

  /**
   * コンストラクタである。
   */
  constructor() {
    const subtitle = document.getElementById("subtitle-text");
    const name = document.getElementById("active-speaker-name");
    if (!subtitle || !name) throw new Error("Subtitle elements not found");

    this.subtitleText = subtitle;
    this.speakerName = name;

    // 各キャラクターのアバター要素を取得
    ["mc", "guest-a", "guest-b"].forEach((role) => {
      const el = document.getElementById(`avatar-${role}`);
      if (el) this.avatars.set(role, el);
    });
  }

  /**
   * updateSpeaker は、現在喋っているキャラクターを強調し、テロップを更新する。
   * @param item [ScriptItem] AI から受信したセリフデータ
   */
  public update(item: ScriptItem): void {
    // 1. 全てのアバターの強調を解除
    this.avatars.forEach((el) => el.classList.remove("speaking"));

    // 2. 現在の話者を強調
    const activeAvatar = this.avatars.get(item.speaker);
    if (activeAvatar) {
      activeAvatar.classList.add("speaking");
    }

    // 3. 話者名の更新
    const speakerNames: Record<string, string> = {
      mc: "四国めたん",
      "guest-a": "ずんだもん",
      "guest-b": "春日部つむぎ",
    };
    this.speakerName.textContent = speakerNames[item.speaker] || item.speaker;

    // 4. テロップのタイピング表示
    this.typeSubtitle(item.text);
  }

  /**
   * typeSubtitle は、字幕をタイプライター風に表示する。
   */
  private typeSubtitle(text: string): void {
    if (this.typingInterval) {
      clearInterval(this.typingInterval);
    }

    this.subtitleText.textContent = "";
    let index = 0;

    this.typingInterval = window.setInterval(() => {
      this.subtitleText.textContent = text.slice(0, index + 1);
      index++;

      if (index >= text.length) {
        if (this.typingInterval) clearInterval(this.typingInterval);
        this.typingInterval = null;
      }
    }, 40);
  }

  /**
   * clear は、表示を初期化する。
   */
  public clear(): void {
    if (this.typingInterval) clearInterval(this.typingInterval);
    this.subtitleText.textContent = "待機中...";
    this.speakerName.textContent = "-";
    this.avatars.forEach((el) => el.classList.remove("speaking"));
  }
}
