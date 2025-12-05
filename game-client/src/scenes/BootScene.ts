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
    socketManager.connect(
      `ws://localhost:5555/game/ws?token=00000000-0000-0000-0000-000000000017&name=Kranti`,
    );
    this.scene.start("PreloadScene");
  }
}
