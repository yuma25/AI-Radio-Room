/**
 * Package main は、フロントエンドアプリケーションのエントリポイントである。
 * 各 UI コンポーネントを初期化し、バックエンド API とのリアルタイム通信を管理する。
 */
import "./style.css";
import { MonitorUI } from "./ui/monitor";
// import { SidebarUI } from "./ui/sidebar"; // TODO: UI刷新に伴い一時無効化
import { VoicePlayer } from "./voice/player";
import { ScriptItem } from "./types";

/**
 * アプリケーションの初期化シーケンス。
 */
document.addEventListener("DOMContentLoaded", () => {
  try {
    // 1. [内部] 各 UI クラスおよび音声プレイヤーの初期化
    const monitor = new MonitorUI();
    const voice = new VoicePlayer(); // デフォルトで http://localhost:8080 を使用
    // new SidebarUI(); // TODO: サイドバーUIがHTMLから削除されているため無効化

    // FOUC対策: JSの準備が整ったらbodyをフェードインさせる
    document.body.classList.add("ready");

    // 音声の再生開始に合わせてUIを更新する（同期処理）
    voice.onPlayStart = (item: ScriptItem) => {
      monitor.update(item);
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

    const backToHomeBtn = document.getElementById("back-to-home-btn");
    const modalBgmToggle = document.getElementById(
      "modal-bgm-toggle",
    ) as HTMLInputElement;

    let currentTheme = "AIとの共生"; // デフォルトテーマ
    let isStreamActive = false; // ストリームが動作中かどうかのフラグ

    /**
     * 放送が終了したか確認し、終了していれば「戻る」ボタンを表示する。
     */
    const showBackBtnIfFinished = () => {
      if (!isStreamActive && !voice.getIsPlaying()) {
        console.log("🏁 放送が終了しました。戻るボタンを表示します。");
        if (backToHomeBtn) {
          backToHomeBtn.classList.remove("hidden");
        }
        if (onAirIndicator) {
          onAirIndicator.classList.remove("active");
        }
      }
    };

    /**
     * UIをホーム状態に戻す内部関数。
     * 「ホームに戻る」ボタンがクリックされた時に実行される。
     */
    const resetToHome = () => {
      console.log("🔄 ホームに戻ります。");
      if (themeModal) themeModal.classList.remove("hidden");
      if (modalStartBtn) modalStartBtn.removeAttribute("disabled");
      if (currentThemeDisplay) currentThemeDisplay.classList.add("hidden");
      if (backToHomeBtn) backToHomeBtn.classList.add("hidden");

      // BGMを停止
      bgmAudio.pause();
      bgmAudio.currentTime = 0;
    };

    if (backToHomeBtn) {
      backToHomeBtn.addEventListener("click", resetToHome);
    }

    // 音声再生が全て完了した時の処理を登録
    voice.onQueueEmpty = () => {
      showBackBtnIfFinished();
    };

    if (modalStartBtn && modalThemeInput && themeModal && currentThemeDisplay) {
      modalStartBtn.addEventListener("click", () => {
        const inputVal = modalThemeInput.value.trim();
        if (inputVal) {
          currentTheme = inputVal;
        }

        // --- UIの初期化 ---
        // モーダルを即座に隠す（classList.add("hidden") が確実に効くようにする）
        themeModal.classList.add("hidden");
        monitor.clear();
        modalStartBtn.setAttribute("disabled", "true");
        if (backToHomeBtn) backToHomeBtn.classList.add("hidden");
        isStreamActive = true;

        // テーマをヘッダーに掲示
        const span = currentThemeDisplay.querySelector("span");
        if (span) span.textContent = currentTheme;
        currentThemeDisplay.classList.remove("hidden");

        console.log(`🎙️ 放送準備完了: テーマ「${currentTheme}」`);

        // BGM再生開始（チェックボックスがONの場合のみ）
        if (modalBgmToggle && modalBgmToggle.checked) {
          bgmAudio
            .play()
            .catch((e) => console.warn("BGM自動再生がブロックされました:", e));
        }

        // [演出] カウントダウンを 10秒に設定。この間に裏側で全ての生成・合成を終わらせる。
        if (countdownOverlay && countdownNumber) {
          countdownOverlay.classList.remove("hidden");
          let count = 10;
          countdownNumber.textContent = count.toString();

          // 1. [並列処理A] バックエンドへメイン台本の生成リクエストを開始
          const fetchPromise = fetch("http://localhost:8080/api/broadcasts", {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
            },
            body: JSON.stringify({ theme: currentTheme }),
          });

          // 2. [並列処理B] テーマ発表セリフ（動的パーツ）の音声合成を、カウントダウンを待たずに今すぐ開始させる
          const themeTalkItem: ScriptItem = {
            speaker: "mc",
            speaker_id: 6,
            text: `本日のトークテーマは「${currentTheme}」です！今日もこの3人で、熱く語り合っていきましょう！`,
            audio_text: `ほんじつの トークテーマは 「${currentTheme}」です！ きょうも この さんにんで、 あつく かたりあって いきましょう！`,
            emotion: "excited",
          };
          const themeTTSParams = new URLSearchParams({
            text: themeTalkItem.audio_text,
            speaker: themeTalkItem.speaker_id.toString(),
          });
          // 先行して Fetch を開始し、ブラウザのキャッシュに載せる
          fetch(`http://localhost:8080/api/tts?${themeTTSParams}`).catch((e) =>
            console.warn("Pre-fetch failed:", e),
          );

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

              // カウントダウン終了後、事前生成されたローカル音源を使用して即座に自己紹介を開始
              const preTalks: ScriptItem[] = [
                {
                  speaker: "mc",
                  speaker_id: 2,
                  text: "皆さんこんにちは！AI Radio Studio の時間がやってまいりました！メインパーソナリティの、四国めたんです。",
                  audio_text:
                    "みなさん こんにちは！ エーアイ ラジオ スタジオの じかんが やってまいりました！ メインパーソナリティの、 しこくめたんです。",
                  emotion: "normal",
                  local_audio: "/audio/intro_metan.wav",
                },
                {
                  speaker: "guest-a",
                  speaker_id: 3,
                  text: "みんなよろしくなのだ！アシスタントのずんだもんなのだ！",
                  audio_text:
                    "みんな よろしくなのだ！ アシスタントの ずんだもんなのだ！",
                  emotion: "normal",
                  local_audio: "/audio/intro_zunda.wav",
                },
                {
                  speaker: "guest-b",
                  speaker_id: 8,
                  text: "同じくアシスタントの、春日部つむぎです。本日もよろしくお願いします。",
                  audio_text:
                    "おなじく アシスタントの、 かすかべつむぎです。 ほんじつも よろしく おねがいします。",
                  emotion: "normal",
                  local_audio: "/audio/intro_tsumugi.wav",
                },
                themeTalkItem, // 先行して合成リクエストを送った動的パーツ
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
                  monitor.update({
                    speaker: "mc",
                    speaker_id: 2,
                    text: "通信エラーが発生したため、放送を中断します。",
                    audio_text:
                      "つうしんエラーが はっせいしたため、 ほうそうを ちゅうだんします。",
                    emotion: "calm",
                  });
                })
                .finally(() => {
                  isStreamActive = false;
                  console.log("🏁 放送ストリームが終了した。");
                  showBackBtnIfFinished();
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
