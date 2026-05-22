/**
 * Package ui は、設定サイドバーの開閉および管理を担当する。
 */

/**
 * SidebarUI は、設定画面の表示・非表示を制御するクラスである。
 */
export class SidebarUI {
  private sidebar: HTMLElement;
  private openBtn: HTMLElement;
  private closeBtn: HTMLElement;

  /**
   * コンストラクタである。
   *
   * 内部処理:
   * - 1. [標準] `getElementById()` で必要な要素を取得する。
   * - 2. [内部] `setupListeners()` を呼び出し、クリックイベントを登録する。
   */
  constructor() {
    this.sidebar = this.getEl("settings-sidebar");
    this.openBtn = this.getEl("open-settings");
    this.closeBtn = this.getEl("close-settings");

    this.setupListeners();
  }

  /**
   * setupListeners は、ボタンクリック時のイベントリスナーを設定する内部メソッドである。
   */
  private setupListeners(): void {
    // 開くボタン
    this.openBtn.addEventListener("click", () => {
      this.sidebar.classList.remove("hidden");
    });

    // 閉じるボタン
    this.closeBtn.addEventListener("click", () => {
      this.sidebar.classList.add("hidden");
    });
  }

  /**
   * getEl は、ID から要素を安全に取得するユーティリティメソッドである。
   */
  private getEl(id: string): HTMLElement {
    const el = document.getElementById(id);
    if (!el) throw new Error(`Element #${id} not found`);
    return el;
  }
}
