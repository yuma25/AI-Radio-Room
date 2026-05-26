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

    // 音声の再生開始に合わせてUIにテキストを描画する（同期処理）
    voice.onPlayStart = (item: ScriptItem) => {
      chat.addMessage(item);
    };

    // --- BGM機能の初期化 ---
    // ※今回はダミーのフリーBGM URLを使用する。本番では public/audio/bgm.mp3 等に置き換える。
    const bgmAudio = new Audio(
      "https://cdn.pixabay.com/download/audio/2022/05/27/audio_1808fbf07a.mp3?filename=lofi-study-112191.mp3",
    );
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

    let currentTheme = "AIとの共生"; // デフォルトテーマ

    if (modalStartBtn && modalThemeInput && themeModal && currentThemeDisplay) {
      modalStartBtn.addEventListener("click", () => {
        const inputVal = modalThemeInput.value.trim();
        if (inputVal) {
          currentTheme = inputVal;
        }
        // テーマをヘッダーに掲示してモーダルを閉じる
        const span = currentThemeDisplay.querySelector("span");
        if (span) span.textContent = currentTheme;
        currentThemeDisplay.classList.remove("hidden");
        themeModal.classList.add("hidden");
      });
    }

    // 3. [内部] 放送開始ボタンのイベント設定
    const startBtn = document.getElementById("start-broadcast");
    const onAirIndicator = document.getElementById("on-air-indicator");

    if (startBtn) {
      startBtn.addEventListener("click", async () => {
        console.log(`🎙️ 放送開始リクエスト: テーマ「${currentTheme}」`);

        // 画面の初期化
        chat.clear();
        startBtn.setAttribute("disabled", "true");
        if (onAirIndicator) {
          onAirIndicator.classList.add("active");
        }

        // BGM再生開始
        bgmAudio
          .play()
          .catch((e) => console.warn("BGM自動再生がブロックされました:", e));

        // [演出] AIの台本生成（初回ラグ）を隠蔽するためのプレトーク（ダミーデータ）をキューに追加
        // バックエンドがAIから台本を受け取るまでの間（数秒〜10秒）、この自己紹介を再生して時間を稼ぐ
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

        try {
          // [標準] fetch API を使用して、バックエンドのストリーミングエンドポイントを呼び出す。
          // NOTE: EventSource は POST をサポートしていないため、fetch と ReadableStream を組み合わせて SSE を受信する。
          const response = await fetch("http://localhost:8080/api/broadcasts", {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
            },
            body: JSON.stringify({ theme: currentTheme }),
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

                  // [内部] 受信したセリフを音声再生キューに追加する。描画は onPlayStart に委譲する。
                  voice.enqueue(item);
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
            speaker_id: 2,
            text: "通信エラーが発生したため、放送を中断します。",
            emotion: "calm",
          });
        } finally {
          startBtn.removeAttribute("disabled");
          if (onAirIndicator) {
            onAirIndicator.classList.remove("active");
          }
          // bgmAudio.pause(); // 放送終了後もBGMを流し続けるかはお好みで
          console.log("🏁 放送ストリームが終了した。");
        }
      });
    }

    console.log("✅ AI Radio Room Frontend initialized.");
  } catch (error) {
    console.error("❌ Failed to initialize application:", error);
  }
});
