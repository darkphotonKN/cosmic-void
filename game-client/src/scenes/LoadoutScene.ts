import Phaser from 'phaser';
import { EquipmentPanel } from '@/ui/EquipmentPanel';
import { ItemState, EquippedItems } from '@/types/gameState';

// Placeholder demo items for testing until backend API exists
const DEMO_ITEMS: ItemState[] = [
  { item_id: 'tpl-w1', entity_id: 'demo-w1', name: 'Plasma Blade', quantity: 1, attack_power: 24, critical_rate: 12, weapon_type: 'sword', description: 'A superheated edge that cuts through void matter.' },
  { item_id: 'tpl-w2', entity_id: 'demo-w2', name: 'Arc Repeater', quantity: 1, attack_power: 18, critical_rate: 8, weapon_type: 'rifle', description: 'Rapid-fire energy bolts with modest penetration.' },
  { item_id: 'tpl-a1', entity_id: 'demo-a1', name: 'Neurohelm', quantity: 1, defense_rating: 10, armor_slot: 'head', description: 'Neural-linked helmet with threat overlay.' },
  { item_id: 'tpl-a2', entity_id: 'demo-a2', name: 'Voidweave Vest', quantity: 1, defense_rating: 22, armor_slot: 'chest', description: 'Lightweight composite plating.' },
  { item_id: 'tpl-a3', entity_id: 'demo-a3', name: 'Tac Gauntlets', quantity: 1, defense_rating: 8, armor_slot: 'hands', description: 'Grip-enhanced tactical gloves.' },
  { item_id: 'tpl-a4', entity_id: 'demo-a4', name: 'Mag Boots', quantity: 1, defense_rating: 12, armor_slot: 'feet', description: 'Magnetic adhesion boots for low-grav ops.' },
  { item_id: 'tpl-a5', entity_id: 'demo-a5', name: 'Signal Ring', quantity: 1, defense_rating: 4, armor_slot: 'ring', description: 'Encrypted comms relay ring.' },
  { item_id: 'tpl-a6', entity_id: 'demo-a6', name: 'Null Ring', quantity: 1, defense_rating: 6, armor_slot: 'ring', description: 'Dampens void radiation signatures.' },
  { item_id: 'tpl-c1', entity_id: 'demo-c1', name: 'Stim Pack', quantity: 3, healing_amount: 40, description: 'Emergency field healing compound.' },
  { item_id: 'tpl-c2', entity_id: 'demo-c2', name: 'Mana Cell', quantity: 2, mana_amount: 30, description: 'Compressed energy cell for ability recharge.' },
  { item_id: 'tpl-c3', entity_id: 'demo-c3', name: 'Nano Patch', quantity: 1, healing_amount: 80, description: 'Nanite-infused regeneration patch. Slow but potent.' },
];

export class LoadoutScene extends Phaser.Scene {
  private equipmentPanel?: EquipmentPanel;
  private equipped: EquippedItems = {
    weapon: null, head: null, body: null, hands: null, feet: null,
    ring_1: null, ring_2: null, consumable_1: null, consumable_2: null, consumable_3: null,
  };
  private inventory: ItemState[] = [];

  constructor() {
    super({ key: 'LoadoutScene' });
  }

  create(): void {
    const { width, height } = this.cameras.main;

    // Background
    this.cameras.main.setBackgroundColor('#0a0a12');

    // Star field
    const stars = this.add.graphics();
    for (let i = 0; i < 80; i++) {
      const x = Phaser.Math.Between(0, width);
      const y = Phaser.Math.Between(0, height);
      const size = Math.random() < 0.1 ? 2 : 1;
      const alpha = Phaser.Math.FloatBetween(0.1, 0.4);
      stars.fillStyle(0xffffff, alpha);
      stars.fillRect(x, y, size, size);
    }

    // Subtle grid
    const grid = this.add.graphics();
    grid.lineStyle(1, 0x00f0ff, 0.03);
    for (let x = 0; x <= width; x += 40) grid.lineBetween(x, 0, x, height);
    for (let y = 0; y <= height; y += 40) grid.lineBetween(0, y, width, y);

    // Title
    const title = this.add.text(width / 2, 30, 'LOADOUT', {
      fontSize: '24px',
      color: '#00f0ff',
      letterSpacing: 8,
      fontStyle: 'bold',
    });
    title.setOrigin(0.5);
    title.setShadow(0, 0, '#00f0ff', 6, true, true);

    // Subtitle
    const subtitle = this.add.text(width / 2, 58, 'CONFIGURE YOUR GEAR BEFORE DEPLOYMENT', {
      fontSize: '10px',
      color: '#556677',
      letterSpacing: 4,
    });
    subtitle.setOrigin(0.5);

    // Accent line
    const accent = this.add.graphics();
    accent.lineStyle(1, 0x00f0ff, 0.2);
    accent.lineBetween(width * 0.15, 72, width * 0.85, 72);

    // Back button
    const backBtnX = 70;
    const backBtnY = 30;
    const backBg = this.add.graphics();
    backBg.fillStyle(0x112233, 0.8);
    backBg.fillRoundedRect(backBtnX - 50, backBtnY - 14, 100, 28, 4);
    backBg.lineStyle(1, 0xff00aa, 0.4);
    backBg.strokeRoundedRect(backBtnX - 50, backBtnY - 14, 100, 28, 4);

    const backText = this.add.text(backBtnX, backBtnY, 'BACK', {
      fontSize: '12px',
      color: '#ff00aa',
      letterSpacing: 3,
    });
    backText.setOrigin(0.5);

    const backHit = this.add.rectangle(backBtnX, backBtnY, 100, 28, 0x000000, 0);
    backHit.setInteractive({ useHandCursor: true });
    backHit.on('pointerover', () => {
      backBg.clear();
      backBg.fillStyle(0x221133, 0.9);
      backBg.fillRoundedRect(backBtnX - 50, backBtnY - 14, 100, 28, 4);
      backBg.lineStyle(1, 0xff00aa, 0.7);
      backBg.strokeRoundedRect(backBtnX - 50, backBtnY - 14, 100, 28, 4);
    });
    backHit.on('pointerout', () => {
      backBg.clear();
      backBg.fillStyle(0x112233, 0.8);
      backBg.fillRoundedRect(backBtnX - 50, backBtnY - 14, 100, 28, 4);
      backBg.lineStyle(1, 0xff00aa, 0.4);
      backBg.strokeRoundedRect(backBtnX - 50, backBtnY - 14, 100, 28, 4);
    });
    backHit.on('pointerdown', () => this.returnToMenu());

    // ESC to return
    this.input.keyboard?.on('keydown-ESC', () => this.returnToMenu());

    // Initialize inventory with demo items
    this.inventory = DEMO_ITEMS.map(item => ({ ...item }));

    // Create equipment panel
    this.equipmentPanel = new EquipmentPanel(this);
    this.equipmentPanel.onEquip = (item, slot) => {
      this.equipped[slot] = item;
      this.inventory = this.inventory.filter(i => i.entity_id !== item.entity_id);
    };
    this.equipmentPanel.onUnequip = (item, slot) => {
      this.equipped[slot] = null;
      this.inventory.push(item);
    };

    // Show panel with demo data
    this.equipmentPanel.show();
    this.equipmentPanel.updateInventory(this.inventory);
    this.equipmentPanel.updateEquipment(this.equipped);

    // Disable default right-click context menu
    this.input.mouse?.disableContextMenu();
  }

  private returnToMenu(): void {
    if (this.equipmentPanel) {
      this.equipmentPanel.destroy();
      this.equipmentPanel = undefined;
    }
    this.scene.start('MainMenuScene');
  }

  shutdown(): void {
    if (this.equipmentPanel) {
      this.equipmentPanel.destroy();
      this.equipmentPanel = undefined;
    }
  }
}
