import { socketManager } from "@/utils/class/SocketManager";
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

    // 設定驗證失敗時跳轉到登入頁面
    socketManager.setOnAuthError(() => {
      // 清除 localStorage 中的 auth 資料
      localStorage.removeItem("auth-storage");
      window.location.href = "/login";
    });

    console.log("token: ", token);
    console.log("name: ", name);
    socketManager.connect(
      `ws://localhost:5555/game/ws?token=${token}&name=${name}`,
    );
    this.scene.start("PreloadScene");
  }
}
