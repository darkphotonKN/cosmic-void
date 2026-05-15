import { socketManager } from "@/utils/class/SocketManager";
import { getWsBaseUrl } from "@/utils/wsUrl";
import Phaser from "phaser";

export class BootScene extends Phaser.Scene {
  constructor() {
    super({ key: "BootScene" });
  }

  preload(): void {
    // Load minimal assets needed for preloader
  }

  create(): void {
    // 直接從 localStorage 讀取 auth 資料
    let token = "";
    let name = "Guest";

    try {
      const authStorage = localStorage.getItem("auth-storage");
      if (authStorage) {
        const parsed = JSON.parse(authStorage);
        token = parsed.state?.accessToken || "";
        name = parsed.state?.memberInfo?.name || "Guest";
      }
    } catch (e) {
      console.error("Failed to parse auth storage:", e);
    }

    // 沒有 token 直接跳轉到登入頁面
    if (!token) {
      window.location.href = "/login";
      return;
    }

    // Handle auth errors (close code 4001 from server)
    socketManager.setOnAuthError(() => {
      localStorage.removeItem("auth-storage");
      window.location.href = "/login";
    });

    // Handle generic server-side message errors. Most are fatal session issues
    // (e.g. operator already queued from another tab/session) — kick out and force re-auth.
    socketManager.setOnMessageError(({ message, error }) => {
      const reason =
        message ||
        error ||
        "Operator session conflict detected. You have been disconnected.";
      sessionStorage.setItem("auth-error-message", reason);
      socketManager.disconnect();
      localStorage.removeItem("auth-storage");
      window.location.href = "/login";
    });

    console.log("token: ", token);
    console.log("name: ", name);
    socketManager.connect(
      `${getWsBaseUrl()}/game/ws?token=${token}&name=${name}`,
    );
    this.scene.start("PreloadScene");
  }
}
