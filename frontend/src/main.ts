import "./style.css";

document.addEventListener("DOMContentLoaded", () => {
  const app = document.querySelector<HTMLDivElement>("#app");
  if (app) {
    app.innerHTML = `
      <div style="padding: 20px; text-align: center;">
        <h1>AI Radio Room</h1>
        <p>Frontend initialized successfully.</p>
      </div>
    `;
  }
});
