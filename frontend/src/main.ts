/**
 * Package main は、フロントエンドアプリケーションのエントリポイントである。
 * 各 UI コンポーネントを初期化し、バックエンド API とのリアルタイム通信を管理する。
 */
import "./style.css";
import { ChatUI } from "./ui/chat";
// import { SidebarUI } from "./ui/sidebar"; // TODO: UI刷新に伴い一時無効化
import { VoicePlayer } from "./voice/player";
import { ScriptItem } from "./types";

/**
 * アプリケーションの初期化シーケンス。
 */
document.addEventListener("DOMContentLoaded", () => {
  try {
    // 1. [内部] 各 UI クラスおよび音声プレイヤーの初期化
    const chat = new ChatUI("message-list");
    const voice = new VoicePlayer(); // デフォルトで http://localhost:8080 を使用
    // new SidebarUI(); // TODO: サイドバーUIがHTMLから削除されているため無効化

    // FOUC対策: JSの準備が整ったらbodyをフェードインさせる
    document.body.classList.add("ready");

    // 音声の再生開始に合わせてUIにテキストを描画する（同期処理）
    voice.onPlayStart = (item: ScriptItem) => {
      chat.addMessage(item);
    };

    // --- BGM機能の初期化 ---
    // ローカルに配置したフリー音源（public/audio/bgm.mp3）を使用する。
    const bgmAudio = new Audio("/audio/bgm.mp3");
    bgmAudio.loop = true;
    bgmAudio.volume = 0.1; // 初期音量

    const volumeSlider = document.getElementById(
      "bgm-volume",
    ) as HTMLInputElement;
    if (volumeSlider) {
      volumeSlider.addEventListener("input", (e) => {
        const target = e.target as HTMLInputElement;
        bgmAudio.volume = parseFloat(target.value);
      });
    }

    // 2. モーダルとテーマ設定の処理
    const themeModal = document.getElementById("theme-modal");
    const modalThemeInput = document.getElementById(
      "modal-theme-input",
    ) as HTMLInputElement;
    const modalStartBtn = document.getElementById("modal-start-btn");
    const currentThemeDisplay = document.getElementById(
      "current-theme-display",
    );
    const onAirIndicator = document.getElementById("on-air-indicator");
    const countdownOverlay = document.getElementById("countdown-overlay");
    const countdownNumber = document.getElementById("countdown-number");

    let currentTheme = "AIとの共生"; // デフォルトテーマ

    if (modalStartBtn && modalThemeInput && themeModal && currentThemeDisplay) {
      modalStartBtn.addEventListener("click", () => {
        const inputVal = modalThemeInput.value.trim();
        if (inputVal) {
          currentTheme = inputVal;
        }

        // UIの初期化
        chat.clear();
        modalStartBtn.setAttribute("disabled", "true");

        // テーマをヘッダーに掲示してモーダルを閉じる
        const span = currentThemeDisplay.querySelector("span");
        if (span) span.textContent = currentTheme;
        currentThemeDisplay.classList.remove("hidden");
        themeModal.classList.add("hidden");

        console.log(`🎙️ 放送準備完了: テーマ「${currentTheme}」`);

        // BGM再生開始
        bgmAudio
          .play()
          .catch((e) => console.warn("BGM自動再生がブロックされました:", e));

        // [演出] 10秒のカウントダウンを開始し、その裏でバックエンドに台本生成を依頼する
        if (countdownOverlay && countdownNumber) {
          countdownOverlay.classList.remove("hidden");
          let count = 10;
          countdownNumber.textContent = count.toString();

          // 非同期でバックエンドへリクエストを開始 (awaitしない)
          const fetchPromise = fetch("http://localhost:8080/api/broadcasts", {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
            },
            body: JSON.stringify({ theme: currentTheme }),
          });

          // カウントダウンタイマー
          const timer = setInterval(() => {
            count--;
            if (count > 0) {
              countdownNumber.textContent = count.toString();
            } else {
              clearInterval(timer);
              countdownOverlay.classList.add("hidden");
              if (onAirIndicator) {
                onAirIndicator.classList.add("active");
              }

              // カウントダウン終了後、自然な自己紹介（プレトーク）をキューに追加
              const preTalks: ScriptItem[] = [
                {
                  speaker: "mc",
                  speaker_id: 2, // 四国めたん（ノーマル）
                  text: "皆さんこんにちは！AI Radio Studio の時間がやってまいりました！メインパーソナリティの、四国めたんです。",
                  emotion: "normal",
                },
                {
                  speaker: "guest-a",
                  speaker_id: 3, // ずんだもん（ノーマル）
                  text: "みんなよろしくなのだ！アシスタントのずんだもんなのだ！",
                  emotion: "normal",
                },
                {
                  speaker: "guest-b",
                  speaker_id: 8, // 春日部つむぎ（ノーマル）
                  text: "同じくアシスタントの、春日部つむぎです。本日もよろしくお願いします。",
                  emotion: "normal",
                },
                {
                  speaker: "mc",
                  speaker_id: 6, // 四国めたん（ツンツン/興奮）
                  text: `本日のトークテーマは「${currentTheme}」です！今日もこの3人で、熱く語り合っていきましょう！`,
                  emotion: "excited",
                },
              ];
              preTalks.forEach((item) => voice.enqueue(item));

              // 裏で走らせていた fetch 処理の結果を受け取る
              fetchPromise
                .then(async (response) => {
                  if (!response.ok || !response.body) {
                    throw new Error(`サーバーエラー: ${response.statusText}`);
                  }

                  const reader = response.body.getReader();
                  const decoder = new TextDecoder();
                  let buffer = "";

                  while (true) {
                    const { value, done } = await reader.read();
                    if (done) break;

                    buffer += decoder.decode(value, { stream: true });
                    const lines = buffer.split("\n\n");
                    buffer = lines.pop() || "";

                    for (const line of lines) {
                      if (line.startsWith("data: ")) {
                        try {
                          const jsonStr = line.replace("data: ", "").trim();
                          if (jsonStr === "") continue;

                          const item: ScriptItem = JSON.parse(jsonStr);
                          console.log("📥 受信データ:", item);

                          // 受信したセリフを音声再生キューに追加する
                          voice.enqueue(item);
                        } catch (e) {
                          console.warn("JSON パース失敗:", e, line);
                        }
                      }
                    }
                  }
                })
                .catch((error) => {
                  console.error("❌ 通信エラー:", error);
                  chat.addMessage({
                    speaker: "mc",
                    speaker_id: 2,
                    text: "通信エラーが発生したため、放送を中断します。",
                    emotion: "calm",
                  });
                })
                .finally(() => {
                  if (onAirIndicator) {
                    onAirIndicator.classList.remove("active");
                  }
                  console.log("🏁 放送ストリームが終了した。");
                });
            }
          }, 1000); // 1秒ごとにカウントダウン
        }
      });
    }

    console.log("✅ AI Radio Room Frontend initialized.");
  } catch (error) {
    console.error("❌ Failed to initialize application:", error);
  }
});
