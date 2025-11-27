import { createStarfield } from "@/utils/Background";
import { Player } from "@/utils/class/Player";
import Phaser from "phaser";
import { io, Socket } from "socket.io-client";

export class GameScene extends Phaser.Scene {
  private score: number = 0;
  private scoreText!: Phaser.GameObjects.Text;
  private me!: Player;
  private players!: Player[];
  private enemy!: Player;
  private cursors!: Phaser.Types.Input.Keyboard.CursorKeys;
  private speed: number = 300;
  private keyW!: Phaser.Input.Keyboard.Key;
  private keyA!: Phaser.Input.Keyboard.Key;
  private keyS!: Phaser.Input.Keyboard.Key;
  private keyD!: Phaser.Input.Keyboard.Key;
  private socket!: Socket;
  private isWaitingForServer!: boolean;
  // fog of war
  private fogOfWar!: Phaser.GameObjects.Graphics;
  private visionCircle!: Phaser.GameObjects.Graphics;

  constructor() {
    super({ key: "GameScene" });
  }

  preload() {
    // 載入 spritesheet（注意要設定每個幀的大小）
    this.load.spritesheet(
      "dude",
      "https://labs.phaser.io/assets/sprites/dude.png",
      { frameWidth: 32, frameHeight: 48 },
    );
  }

  create(): void {
    // connect websoket
    this.socket = io("http://localhost:3001", {
      transports: ["websocket"], // 強制使用 WebSocket
      reconnection: true, // 自動重連
      reconnectionDelay: 1000,
    });

    // 迷霧 start
    const width = this.cameras.main.width;
    const height = this.cameras.main.height;

    // 建立視野圓圈（用於遮罩）
    this.visionCircle = this.add.graphics();
    this.visionCircle.fillStyle(0xffffff);
    this.visionCircle.fillCircle(400, 300, 100);

    // 建立黑色迷霧層（半透明）
    this.fogOfWar = this.add.graphics();
    this.fogOfWar.fillStyle(0x000000, 0.85); // 第二個參數是透明度，0-1
    this.fogOfWar.fillRect(0, 0, width, height);
    this.fogOfWar.setDepth(100);

    // 使用 GeometryMask，invertAlpha 讓圓圈內透明
    const mask = new Phaser.Display.Masks.GeometryMask(this, this.visionCircle);
    mask.invertAlpha = true;
    this.fogOfWar.setMask(mask);
    // 迷霧 end

    // 星空背景
    createStarfield(this);

    // Score display
    this.scoreText = this.add.text(16, 16, "Score: 0", {
      fontSize: "18px",
      color: "#ffffff",
    });

    // this.physics.add.sprite(400, 300, "dude");
    this.me = new Player(this, 400, 300, "dude", "dsfgsdarfgsd", "Nick");
    this.me.setCollideWorldBounds(true);
    this.physics.add.existing(this.me);

    this.enemy = new Player(this, 200, 200, "dude", "dsfgsdarfgsd", "John"); // Assuming 'dude' texture can be used for enemy
    this.enemy.setTint(0xff0000); // Make enemy red
    this.enemy.setCollideWorldBounds(true);
    // 讓敵人只在可視範圍內顯示
    const enemyMask = new Phaser.Display.Masks.GeometryMask(
      this,
      this.visionCircle,
    );
    this.enemy.setMask(enemyMask);

    this.physics.add.collider(
      this.me,
      this.enemy,
      (obj1, obj2) => {
        const player = obj1 as Phaser.Physics.Arcade.Sprite;
        const enemy = obj2 as Phaser.Physics.Arcade.Sprite;
        this.handlePlayerEnemyCollision(player, enemy);
      },
      undefined,
      this,
    );

    // 建立動畫：向左走
    this.anims.create({
      key: "left",
      frames: this.anims.generateFrameNumbers("dude", { start: 0, end: 3 }),
      frameRate: 10,
      repeat: -1, // -1 表示無限循環
    });

    // 建立動畫：靜止面向前方
    this.anims.create({
      key: "turn",
      frames: [{ key: "dude", frame: 4 }],
      frameRate: 20,
    });

    // 建立動畫：向右走
    this.anims.create({
      key: "right",
      frames: this.anims.generateFrameNumbers("dude", { start: 5, end: 8 }),
      frameRate: 10,
      repeat: -1,
    });
    // 設置鍵盤輸入
    this.cursors = this.input.keyboard!.createCursorKeys();

    // 創建 WASD 鍵
    this.keyW = this.input.keyboard!.addKey(Phaser.Input.Keyboard.KeyCodes.W);
    this.keyA = this.input.keyboard!.addKey(Phaser.Input.Keyboard.KeyCodes.A);
    this.keyS = this.input.keyboard!.addKey(Phaser.Input.Keyboard.KeyCodes.S);
    this.keyD = this.input.keyboard!.addKey(Phaser.Input.Keyboard.KeyCodes.D);
    // ESC to return to menu
    this.input.keyboard?.on("keydown-ESC", () => {
      this.scene.start("MainMenuScene");
    });

    // websocket
    this.socket = io("port", {
      transports: ["websocket"], // 強制使用 WebSocket
      reconnection: true, // 自動重連
      reconnectionDelay: 1000,
    });
    // self
    this.socket.on("moveConfirmed", (data: { x: number; y: number }) => {
      // 伺服器確認後才移動
      this.me.setVelocity(data.x, data.y);
      this.isWaitingForServer = false;
    });
    // other plays
    this.socket.on("playerMoved", (data) => {
      console.log("收到玩家移動:", data.id, data.x, data.y);
      // 每次只會收到「一個玩家」的位置

      // const targetPlayer = this.players.get(data.id);
      // if (otherPlayer) {
      //   otherPlayer.setPosition(data.x, data.y);
      // }
    });
  }

  update() {
    // 更新迷霧可視範圍
    const radius = 100;
    this.visionCircle.clear();
    // this.visionCircle.fillStyle(0xffffff);
    this.visionCircle.fillCircle(this.me.x, this.me.y, radius);

    this.me.setVelocity(0);

    // websock sample
    // if (!this.isWaitingForServer) {
    //   if (this.cursors.left.isDown || this.keyA.isDown) {
    //     this.socket.emit("Move", {
    //       x: 0,
    //       y: -this.speed,
    //     });
    //   }
    // }
    // frontend move sample
    if (this.cursors.left.isDown || this.keyA.isDown) {
      this.me.setVelocityX(-this.speed);
      this.me.anims.play("left", true);
      this.socket.emit("Move", {
        x: -this.speed,
        y: 0,
      });
    } else if (this.cursors.right.isDown || this.keyD.isDown) {
      this.me.setVelocityX(this.speed);
      this.me.anims.play("right", true);
      this.socket.emit("Move", {
        x: this.speed,
        y: 0,
      });
    } else {
      this.me.setVelocityX(0);
      this.me.anims.play("turn");
    }

    if (this.cursors.up.isDown || this.keyW.isDown) {
      this.me.setVelocityY(-this.speed);
      this.socket.emit("Move", {
        x: 0,
        y: this.speed,
      });
    } else if (this.cursors.down.isDown || this.keyS.isDown) {
      this.me.setVelocityY(this.speed);
      this.socket.emit("Move", {
        x: 0,
        y: -this.speed,
      });
    }

    const distance = Phaser.Math.Distance.Between(
      this.me.x,
      this.me.y,
      this.enemy.x,
      this.enemy.y,
    );

    const container = document.getElementById("game-container");

    if (container) {
      if (distance < 100) {
        // 敵人靠近,邊框變紅
        container.className = "danger";
      } else {
        // 安全距離,邊框變綠
        container.className = "safe";
      }
    }
  }

  handlePlayerEnemyCollision(
    player: Phaser.Physics.Arcade.Sprite,
    enemy: Phaser.Physics.Arcade.Sprite,
  ): void {
    console.log("Player-Enemy collision!");
    // this.gameOver();
  }

  addScore(points: number): void {
    this.score += points;
    this.scoreText.setText(`Score: ${this.score}`);
  }

  gameOver(): void {
    this.scene.start("GameOverScene", { score: this.score });
  }
}
