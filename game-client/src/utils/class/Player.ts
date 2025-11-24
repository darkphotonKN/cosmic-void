import Phaser from "phaser";

export class Player extends Phaser.Physics.Arcade.Sprite {
  public id: string; // ✅ 自訂屬性：資料庫 ID
  public username: string;
  public health: number;
  public score: number;

  constructor(
    scene: Phaser.Scene,
    x: number,
    y: number,
    texture: string,
    id: string,
    username: string,
  ) {
    // 呼叫父類別的 constructor
    super(scene, x, y, texture);

    // 設定自訂屬性
    this.id = id;
    this.username = username;
    this.health = 100;
    this.score = 0;

    // 將自己加入場景
    scene.add.existing(this);
    scene.physics.add.existing(this);

    // 設定物理屬性
    this.setCollideWorldBounds(true);
    this.setBounce(0.2);
  }

  // 自訂方法
  takeDamage(amount: number): void {
    this.health -= amount;
    console.log(`玩家 ${this.username} (ID: ${this.id}) 受到 ${amount} 傷害`);

    if (this.health <= 0) {
      this.die();
    } else {
      // 閃紅色表示受傷
      this.setTint(0xff0000);
      this.scene.time.delayedCall(200, () => {
        this.clearTint();
      });
    }
  }

  addScore(points: number): void {
    this.score += points;
    console.log(`玩家 ${this.username} 得分: ${this.score}`);
  }

  die(): void {
    console.log(`玩家 ${this.username} 死亡`);
    this.setTint(0x000000);
    this.setAlpha(0.5);
    this.disableBody(true, false); // 停用物理但保留顯示
  }

  respawn(x: number, y: number): void {
    this.health = 100;
    this.setPosition(x, y);
    this.clearTint();
    this.setAlpha(1);
    this.enableBody(true, x, y, true, true);
  }
}
