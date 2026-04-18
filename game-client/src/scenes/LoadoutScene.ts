import Phaser from 'phaser';
import { EquipmentPanel } from '@/ui/EquipmentPanel';
import { ItemState, EquippedItems, EquipmentSlot } from '@/types/gameState';
import { apiClient } from '@/utils/api';

// Map client-side EquipmentSlot UI names → backend canonical slot names.
// Client uses body/hands/feet (visual body regions), backend uses
// chest/gloves/legs (matching armor_slot CHECK constraint).
const SLOT_TO_BACKEND: Record<EquipmentSlot, string> = {
  weapon: 'weapon',
  head: 'head',
  body: 'chest',
  hands: 'gloves',
  feet: 'legs',
  ring_1: 'ring_1',
  ring_2: 'ring_2',
  consumable_1: 'consumable_1',
  consumable_2: 'consumable_2',
  consumable_3: 'consumable_3',
};
function toBackendSlot(slot: EquipmentSlot): string {
  return SLOT_TO_BACKEND[slot];
}

export class LoadoutScene extends Phaser.Scene {
  private equipmentPanel?: EquipmentPanel;
  private equipped: EquippedItems = {
    weapon: null, head: null, body: null, hands: null, feet: null,
    ring_1: null, ring_2: null, consumable_1: null, consumable_2: null, consumable_3: null,
  };
  private inventory: ItemState[] = [];
  private isLoading = false;
  private loadingOverlay?: Phaser.GameObjects.Container;

  constructor() {
    super({ key: 'LoadoutScene' });
  }

  create(): void {
    // Make all text created in this scene sharp on high-DPI displays.
    // Canvas is 1080x720 but Retina renders at 2x — text normally gets
    // hardware-upscaled and blurs. setResolution tells Phaser to rasterize
    // text at devicePixelRatio internally, then display at logical size.
    const origAddText = this.add.text.bind(this.add);
    this.add.text = ((x: number, y: number, text: string | string[], style?: Phaser.Types.GameObjects.Text.TextStyle) => {
      const t = origAddText(x, y, text, style);
      t.setResolution(window.devicePixelRatio || 1);
      return t;
    }) as typeof this.add.text;

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

    // E to equip/unequip hovered item
    this.input.keyboard?.on('keydown-E', () => {
      if (this.isLoading) return;
      this.equipmentPanel?.handleEquipKey();
    });

    // Create equipment panel
    this.equipmentPanel = new EquipmentPanel(this);
    this.equipmentPanel.onEquip = (item, slot) => {
      if (this.isLoading) return;
      this.equipped[slot] = item;
      this.inventory = this.inventory.filter(i => i.entity_id !== item.entity_id);
      this.withLoading(() => apiClient.updateLoadout(toBackendSlot(slot), item.entity_id));
    };
    this.equipmentPanel.onUnequip = (item, slot) => {
      if (this.isLoading) return;
      this.equipped[slot] = null;
      this.inventory.push(item);
      this.withLoading(() => apiClient.updateLoadout(toBackendSlot(slot), null));
    };

    this.equipmentPanel.show();

    // Fetch data from backend
    this.loadData();

    // Disable default right-click context menu
    this.input.mouse?.disableContextMenu();
  }

  private async loadData(): Promise<void> {
    this.showLoadingOverlay();
    try {
      const [instancesRes, loadoutRes] = await Promise.all([
        apiClient.getItemInstances(),
        apiClient.getLoadout(),
      ]);

      // NOTE: proto JSON tags for ItemInstance are snake_case (attack_power,
      // weapon_type, armor_slot, ...) — NOT camelCase. Do not use item.attackPower.
      const allItems: ItemState[] = (instancesRes.result?.items || []).map((item: any) => ({
        item_id: item.template_id || '',
        entity_id: item.id || '',
        name: item.name || '',
        quantity: 1,
        attack_power: item.attack_power || undefined,
        critical_rate: item.critical_rate || undefined,
        weapon_type: item.weapon_type || undefined,
        defense_rating: item.defense_rating || undefined,
        armor_slot: item.armor_slot || undefined,
        healing_amount: item.healing_amount || undefined,
        mana_amount: item.mana_amount || undefined,
        description: item.description || undefined,
      }));

      // Build equipped items from loadout response
      const loadout = loadoutRes.result || {};
      const equippedIds = new Set<string>();

      const findAndEquip = (id: string, slot: EquipmentSlot) => {
        if (!id) return;
        const item = allItems.find(i => i.entity_id === id);
        if (item) {
          this.equipped[slot] = item;
          equippedIds.add(id);
        }
      };

      findAndEquip(loadout.weaponId, 'weapon');
      findAndEquip(loadout.headId, 'head');
      findAndEquip(loadout.chestId, 'body');
      findAndEquip(loadout.glovesId, 'hands');
      findAndEquip(loadout.legsId, 'feet');
      findAndEquip(loadout.ring1Id, 'ring_1');
      findAndEquip(loadout.ring2Id, 'ring_2');
      findAndEquip(loadout.consumable1Id, 'consumable_1');
      findAndEquip(loadout.consumable2Id, 'consumable_2');
      findAndEquip(loadout.consumable3Id, 'consumable_3');

      // Remaining items go to inventory
      this.inventory = allItems.filter(i => !equippedIds.has(i.entity_id));

      // Update UI
      if (this.equipmentPanel) {
        this.equipmentPanel.updateEquipment(this.equipped);
        this.equipmentPanel.updateInventory(this.inventory);
      }
    } catch (err) {
      console.error('Failed to load loadout data:', err);
    } finally {
      this.hideLoadingOverlay();
    }
  }

  private async withLoading(fn: () => Promise<any>): Promise<void> {
    this.showLoadingOverlay();
    try {
      await fn();
    } catch (err) {
      console.error('Loadout update failed:', err);
    } finally {
      this.hideLoadingOverlay();
    }
  }

  private showLoadingOverlay(): void {
    this.isLoading = true;
    const { width, height } = this.cameras.main;

    const bg = this.add.rectangle(width / 2, height / 2, width, height, 0x000000, 0.4);
    bg.setInteractive(); // blocks clicks through

    const text = this.add.text(width / 2, height / 2, 'UPDATING...', {
      fontSize: '16px',
      color: '#00f0ff',
      letterSpacing: 4,
    });
    text.setOrigin(0.5);

    this.loadingOverlay = this.add.container(0, 0, [bg, text]);
    this.loadingOverlay.setDepth(2000);
  }

  private hideLoadingOverlay(): void {
    this.isLoading = false;
    if (this.loadingOverlay) {
      this.loadingOverlay.destroy();
      this.loadingOverlay = undefined;
    }
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
