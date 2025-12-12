import { socketManager } from "@/utils/class/SocketManager";
import { v4 as uuidv4 } from "uuid";
import Phaser from "phaser";

export class BootScene extends Phaser.Scene {
  constructor() {
    super({ key: "BootScene" });
  }

  preload(): void {
    // Load minimal assets needed for preloader
  }

  create(): void {
    // const uuid = uuidv4();
    // TODO: replace login token
    const uuid = uuidv4();
    const name = [...Array(Math.floor(Math.random() * 5) + 4)]
      .map(() => {
        const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ";
        return chars[Math.floor(Math.random() * chars.length)];
      })
      .join("");
    console.log("uuid: ", uuid);
    console.log("name: ", name);
    socketManager.connect(
      `ws://localhost:5555/game/ws?token=${uuid}&name=${name}`,
    );
    this.scene.start("PreloadScene");
  }
}
