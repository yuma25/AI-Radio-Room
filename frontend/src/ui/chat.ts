/**
 * Package ui は、チャット画面の描画およびメッセージ管理を担当する。
 */
import { ScriptItem } from "../types";

/**
 * ChatUI は、メッセージリストの操作をカプセル化するクラスである。
 */
export class ChatUI {
  private messageList: HTMLElement;

  /**
   * コンストラクタである。
   * @param elementId [string] メッセージを格納するコンテナの ID
   */
  constructor(elementId: string) {
    const el = document.getElementById(elementId);
    if (!el) throw new Error(`Element with id ${elementId} not found`);
    this.messageList = el;
  }

  /**
   * addMessage は、新しいセリフを画面に追加し、最下部へスクロールするメソッドである。
   * タイプライター風のアニメーションで1文字ずつテキストを表示する。
   *
   * 内部処理:
   * - 1. [内部] 引数から HTML 要素を生成し、適切な CSS クラス (mc, guest-a, guest-b) を付与する。
   * - 2. [標準] `appendChild()` を使用して DOM に追加する。
   * - 3. [内部] `setInterval` を用いて、テキストを1文字ずつ追加表示する。
   * @param item [ScriptItem] AI から受信したセリフデータ
   */
  public addMessage(item: ScriptItem): void {
    const messageDiv = document.createElement("div");
    messageDiv.classList.add("message", item.speaker);

    const nameSpan = document.createElement("span");
    nameSpan.classList.add("speaker-name");
    nameSpan.textContent = item.speaker.toUpperCase();

    const textP = document.createElement("p");
    textP.classList.add("cursor"); // 点滅カーソルを追加

    messageDiv.appendChild(nameSpan);
    messageDiv.appendChild(textP);

    this.messageList.appendChild(messageDiv);
    this.scrollToBottom();

    // タイプライターアニメーション
    const fullText = item.text;
    let currentIndex = 0;

    const typingInterval = setInterval(() => {
      textP.textContent = fullText.slice(0, currentIndex + 1);
      currentIndex++;
      this.scrollToBottom(); // 文字が追加されるたびにスクロールを追従させる

      if (currentIndex >= fullText.length) {
        clearInterval(typingInterval);
        textP.classList.remove("cursor"); // アニメーション終了時にカーソルを消す
      }
    }, 50); // 1文字あたりの表示速度 (ミリ秒)
  }

  /**
   * clear は、表示されているすべてのメッセージを削除するメソッドである。
   */
  public clear(): void {
    this.messageList.innerHTML = "";
  }

  /**
   * scrollToBottom は、メッセージコンテナを一番下までスクロールさせる内部メソッドである。
   */
  private scrollToBottom(): void {
    const container = this.messageList.parentElement;
    if (container) {
      container.scrollTo({
        top: container.scrollHeight,
        behavior: "smooth",
      });
    }
  }
}
