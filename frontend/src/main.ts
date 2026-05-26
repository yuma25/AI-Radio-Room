/**
 * Package main は、フロントエンドアプリケーションのエントリポイントである。
 * 各 UI コンポーネントを初期化し、バックエンド API とのリアルタイム通信を管理する。
 */
import "./style.css";
import { ChatUI } from "./ui/chat";
import { SidebarUI } from "./ui/sidebar";
import { ScriptItem } from "./types";

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
      startBtn.addEventListener("click", async () => {
        const theme = themeInput.value || "AIとの共生";
        console.log(`🎙️ 放送開始リクエスト: テーマ「${theme}」`);

        // 画面の初期化
        chat.clear();
        startBtn.setAttribute("disabled", "true");

        try {
          // [標準] fetch API を使用して、バックエンドのストリーミングエンドポイントを呼び出す。
          // NOTE: EventSource は POST をサポートしていないため、fetch と ReadableStream を組み合わせて SSE を受信する。
          const response = await fetch("http://localhost:8080/api/broadcasts", {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
            },
            body: JSON.stringify({ theme }),
          });

          if (!response.ok || !response.body) {
            throw new Error(`サーバーエラー: ${response.statusText}`);
          }

          // [標準] ReadableStream を使用した逐次読み込み。
          const reader = response.body.getReader();
          const decoder = new TextDecoder();
          let buffer = "";

          while (true) {
            const { value, done } = await reader.read();
            if (done) break;

            // チャンクをテキストにデコードし、バッファに追加
            buffer += decoder.decode(value, { stream: true });

            // SSE フォーマット "data: {...}\n\n" をパースする。
            const lines = buffer.split("\n\n");
            buffer = lines.pop() || ""; // 完了していない行をバッファに残す

            for (const line of lines) {
              if (line.startsWith("data: ")) {
                try {
                  const jsonStr = line.replace("data: ", "").trim();
                  if (jsonStr === "") continue;

                  const item: ScriptItem = JSON.parse(jsonStr);
                  console.log("📥 受信データ:", item);

                  // [内部] 受信したセリフを UI に反映する。
                  chat.addMessage(item);
                } catch (e) {
                  console.warn("JSON パース失敗:", e, line);
                }
              }
            }
          }
        } catch (error) {
          console.error("❌ 通信エラー:", error);
          chat.addMessage({
            speaker: "mc",
            text: "通信エラーが発生したため、放送を中断します。",
            emotion: "calm",
          });
        } finally {
          startBtn.removeAttribute("disabled");
          console.log("🏁 放送ストリームが終了した。");
        }
      });
    }

    console.log("✅ AI Radio Room Frontend initialized.");
  } catch (error) {
    console.error("❌ Failed to initialize application:", error);
  }
});
