/**
 * Package main は、フロントエンドアプリケーションのエントリポイントである。
 * 各 UI コンポーネントを初期化し、グローバルなイベントを管理する。
 */
import "./style.css";
import { ChatUI } from "./ui/chat";
import { SidebarUI } from "./ui/sidebar";

/**
 * アプリケーションの初期化シーケンス。
 */
document.addEventListener("DOMContentLoaded", () => {
  try {
    // 1. [内部] 各 UI クラスのインスタンス化
    const chat = new ChatUI("message-list");
    new SidebarUI();

    // 2. [内部] 放送開始ボタンのイベント設定
    const startBtn = document.getElementById("start-broadcast");
    const themeInput = document.getElementById(
      "theme-input",
    ) as HTMLInputElement;

    if (startBtn && themeInput) {
      startBtn.addEventListener("click", () => {
        const theme = themeInput.value || "AIとの共生";
        console.log(`🎙️ 放送開始リクエスト: テーマ「${theme}」`);

        // モックメッセージの追加（動作確認用）
        chat.clear();
        chat.addMessage({
          speaker: "mc",
          text: `「${theme}」をテーマに放送を開始します！`,
          emotion: "excited",
        });

        setTimeout(() => {
          chat.addMessage({
            speaker: "guest-a",
            text: "ワクワクするのだ！楽しみなのだ！",
            emotion: "excited",
          });
        }, 1000);

        setTimeout(() => {
          chat.addMessage({
            speaker: "guest-b",
            text: "現実的な視点からも議論しましょう。",
            emotion: "calm",
          });
        }, 2500);
      });
    }

    console.log("✅ AI Radio Room Frontend initialized.");
  } catch (error) {
    console.error("❌ Failed to initialize application:", error);
  }
});
