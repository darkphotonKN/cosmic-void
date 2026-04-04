import {
  ItemState,
  EquipmentSlot,
  EquippedItems,
  getItemType,
  getValidSlotsForItem,
  getSlotDisplayName,
} from '@/types/gameState';

interface SlotLayout {
  slot: EquipmentSlot;
  label: string;
  x: number;
  y: number;
}

interface SlotHitArea {
  slot: EquipmentSlot;
  rect: { x: number; y: number; w: number; h: number };
  item: ItemState | null;
}

interface InventoryRowHitArea {
  rect: { x: number; y: number; w: number; h: number };
  item: ItemState;
}

interface ContextMenuOption {
  label: string;
  action: () => void;
}

// Panel dimensions
const EQUIP_W = 380;
const EQUIP_H = 340;
const INV_W = 320;
const INV_H = 340;
const GAP = 20;
const SLOT_BOX_W = 110;
const SLOT_BOX_H = 34;
const INV_ROW_H = 30;
const MAX_VISIBLE_INV = 8;
const PADDING = 16;

// Colors
const C_CYAN = 0x00f0ff;
const C_BG = 0x0a0a12;
const C_SLOT_EMPTY = 0x112233;
const C_WEAPON = 0xff4466;
const C_ARMOR = 0x44aaff;
const C_CONSUMABLE = 0x44ff88;
const C_RING = 0xdd88ff;

function getSlotColor(slot: EquipmentSlot): number {
  if (slot === 'weapon') return C_WEAPON;
  if (slot === 'ring_1' || slot === 'ring_2') return C_RING;
  if (slot.startsWith('consumable')) return C_CONSUMABLE;
  return C_ARMOR;
}

// Body silhouette layout:
//   [WEAPON]   [HEAD]
//   [HANDS]    [BODY]    [RING 1]
//              [FEET]    [RING 2]
//   [CON 1]   [CON 2]   [CON 3]
const SLOT_LAYOUT: SlotLayout[] = [
  // Row 1: weapon left, head center
  { slot: 'weapon', label: 'WEAPON', x: -110, y: 0 },
  { slot: 'head', label: 'HEAD', x: 0, y: 0 },
  // Row 2: hands left (arms), body center, ring right
  { slot: 'hands', label: 'HANDS', x: -110, y: 50 },
  { slot: 'body', label: 'BODY', x: 0, y: 50 },
  { slot: 'ring_1', label: 'RING 1', x: 110, y: 50 },
  // Row 3: feet center, ring right
  { slot: 'feet', label: 'FEET', x: 0, y: 100 },
  { slot: 'ring_2', label: 'RING 2', x: 110, y: 100 },
  // Row 4: consumables across bottom
  { slot: 'consumable_1', label: 'CONS 1', x: -110, y: 160 },
  { slot: 'consumable_2', label: 'CONS 2', x: 0, y: 160 },
  { slot: 'consumable_3', label: 'CONS 3', x: 110, y: 160 },
];

export class EquipmentPanel {
  private scene: Phaser.Scene;
  private equipContainer?: Phaser.GameObjects.Container;
  private invContainer?: Phaser.GameObjects.Container;
  private visible = false;

  private equipped: EquippedItems = {
    weapon: null, head: null, body: null, hands: null, feet: null,
    ring_1: null, ring_2: null, consumable_1: null, consumable_2: null, consumable_3: null,
  };
  private inventory: ItemState[] = [];

  private slotHitAreas: SlotHitArea[] = [];
  private inventoryHitAreas: InventoryRowHitArea[] = [];

  private slotGraphics: Map<EquipmentSlot, Phaser.GameObjects.Graphics> = new Map();
  private slotTexts: Map<EquipmentSlot, Phaser.GameObjects.Text> = new Map();

  private invRowGraphics: Phaser.GameObjects.Graphics[] = [];
  private invRowTexts: Phaser.GameObjects.Text[] = [];
  private invEmptyText?: Phaser.GameObjects.Text;

  private contextMenu?: Phaser.GameObjects.Container;
  private hoveredSlot?: EquipmentSlot;
  private hoveredInvIndex = -1;
  private tooltip?: Phaser.GameObjects.Container;

  onEquip?: (item: ItemState, slot: EquipmentSlot) => void;
  onUnequip?: (item: ItemState, slot: EquipmentSlot) => void;

  private pointerMoveHandler?: (p: Phaser.Input.Pointer) => void;
  private pointerDownHandler?: (p: Phaser.Input.Pointer) => void;

  constructor(scene: Phaser.Scene) {
    this.scene = scene;
  }

  toggle(): void {
    this.visible ? this.hide() : this.show();
  }

  show(): void {
    if (this.visible) return;
    this.visible = true;
    this.build();
  }

  hide(): void {
    if (!this.visible) return;
    this.visible = false;
    this.dismissContextMenu();
    this.hideTooltip();
    this.cleanup();
  }

  isVisible(): boolean {
    return this.visible;
  }

  updateInventory(items: ItemState[]): void {
    this.inventory = items;
    if (this.visible) this.rebuildInventoryRows();
  }

  updateEquipment(equipped: EquippedItems): void {
    this.equipped = equipped;
    if (this.visible) this.refreshSlots();
  }

  destroy(): void {
    this.hide();
  }

  // === BUILD TWO SEPARATE PANELS ===

  private build(): void {
    const cam = this.scene.cameras.main;
    const totalW = EQUIP_W + GAP + INV_W;
    const leftX = (cam.width - totalW) / 2 + EQUIP_W / 2;
    const rightX = leftX + EQUIP_W / 2 + GAP + INV_W / 2;
    const centerY = cam.height / 2;

    // --- EQUIPMENT PANEL (left) ---
    const eqChildren: Phaser.GameObjects.GameObject[] = [];
    const eqBg = this.scene.add.graphics();
    eqBg.fillStyle(C_BG, 0.92);
    eqBg.fillRoundedRect(-EQUIP_W / 2, -EQUIP_H / 2, EQUIP_W, EQUIP_H, 8);
    eqBg.lineStyle(1, C_CYAN, 0.5);
    eqBg.strokeRoundedRect(-EQUIP_W / 2, -EQUIP_H / 2, EQUIP_W, EQUIP_H, 8);
    eqChildren.push(eqBg);

    const eqTitle = this.scene.add.text(0, -EQUIP_H / 2 + 16, 'EQUIPMENT', {
      fontSize: '16px', color: '#00f0ff', letterSpacing: 5,
    });
    eqTitle.setOrigin(0.5);
    eqChildren.push(eqTitle);

    // Slots
    this.slotHitAreas = [];
    this.slotGraphics.clear();
    this.slotTexts.clear();
    const slotsTop = -EQUIP_H / 2 + 48;

    for (const layout of SLOT_LAYOUT) {
      const slotX = layout.x;
      const slotY = slotsTop + layout.y;

      const label = this.scene.add.text(slotX, slotY + 2, layout.label, {
        fontSize: '9px', color: '#445566', letterSpacing: 2,
      });
      label.setOrigin(0.5, 0);
      eqChildren.push(label);

      const slotGfx = this.scene.add.graphics();
      eqChildren.push(slotGfx);
      this.slotGraphics.set(layout.slot, slotGfx);

      const slotText = this.scene.add.text(slotX, slotY + SLOT_BOX_H - 4, '—', {
        fontSize: '11px', color: '#334455',
      });
      slotText.setOrigin(0.5);
      eqChildren.push(slotText);
      this.slotTexts.set(layout.slot, slotText);

      this.slotHitAreas.push({ slot: layout.slot, rect: { x: 0, y: 0, w: SLOT_BOX_W, h: SLOT_BOX_H }, item: null });
    }

    const eqHint = this.scene.add.text(0, EQUIP_H / 2 - 16, 'RIGHT-CLICK TO UNEQUIP', {
      fontSize: '9px', color: '#334455', letterSpacing: 2,
    });
    eqHint.setOrigin(0.5);
    eqChildren.push(eqHint);

    this.equipContainer = this.scene.add.container(leftX, centerY, eqChildren);
    this.equipContainer.setDepth(1100);
    this.equipContainer.setScrollFactor(0);

    // --- INVENTORY PANEL (right) ---
    const invChildren: Phaser.GameObjects.GameObject[] = [];
    const invBg = this.scene.add.graphics();
    invBg.fillStyle(C_BG, 0.92);
    invBg.fillRoundedRect(-INV_W / 2, -INV_H / 2, INV_W, INV_H, 8);
    invBg.lineStyle(1, C_CYAN, 0.5);
    invBg.strokeRoundedRect(-INV_W / 2, -INV_H / 2, INV_W, INV_H, 8);
    invChildren.push(invBg);

    const invTitle = this.scene.add.text(0, -INV_H / 2 + 16, 'INVENTORY', {
      fontSize: '16px', color: '#00f0ff', letterSpacing: 5,
    });
    invTitle.setOrigin(0.5);
    invChildren.push(invTitle);

    const invHint = this.scene.add.text(0, INV_H / 2 - 16, 'RIGHT-CLICK TO EQUIP  //  I CLOSE', {
      fontSize: '9px', color: '#334455', letterSpacing: 2,
    });
    invHint.setOrigin(0.5);
    invChildren.push(invHint);

    this.invContainer = this.scene.add.container(rightX, centerY, invChildren);
    this.invContainer.setDepth(1100);
    this.invContainer.setScrollFactor(0);

    // Calculate hit areas
    this.updateSlotHitAreas();
    this.refreshSlots();
    this.rebuildInventoryRows();

    // Input
    this.pointerMoveHandler = (p: Phaser.Input.Pointer) => this.handlePointerMove(p);
    this.pointerDownHandler = (p: Phaser.Input.Pointer) => this.handlePointerDown(p);
    this.scene.input.on('pointermove', this.pointerMoveHandler);
    this.scene.input.on('pointerdown', this.pointerDownHandler);
  }

  private cleanup(): void {
    if (this.pointerMoveHandler) { this.scene.input.off('pointermove', this.pointerMoveHandler); this.pointerMoveHandler = undefined; }
    if (this.pointerDownHandler) { this.scene.input.off('pointerdown', this.pointerDownHandler); this.pointerDownHandler = undefined; }
    this.slotGraphics.clear();
    this.slotTexts.clear();
    this.invRowGraphics = [];
    this.invRowTexts = [];
    this.invEmptyText = undefined;
    this.slotHitAreas = [];
    this.inventoryHitAreas = [];
    this.hoveredSlot = undefined;
    this.hoveredInvIndex = -1;
    if (this.equipContainer) { this.equipContainer.destroy(); this.equipContainer = undefined; }
    if (this.invContainer) { this.invContainer.destroy(); this.invContainer = undefined; }
  }

  // === HIT AREAS ===

  private updateSlotHitAreas(): void {
    if (!this.equipContainer) return;
    const cx = this.equipContainer.x;
    const cy = this.equipContainer.y;
    const slotsTop = -EQUIP_H / 2 + 48;

    for (let i = 0; i < SLOT_LAYOUT.length; i++) {
      const layout = SLOT_LAYOUT[i];
      this.slotHitAreas[i].rect = {
        x: cx + layout.x - SLOT_BOX_W / 2,
        y: cy + slotsTop + layout.y,
        w: SLOT_BOX_W,
        h: SLOT_BOX_H,
      };
    }
  }

  // === SLOT RENDERING ===

  private refreshSlots(): void {
    const slotsTop = -EQUIP_H / 2 + 48;
    for (const layout of SLOT_LAYOUT) {
      const item = this.equipped[layout.slot];
      const gfx = this.slotGraphics.get(layout.slot);
      const text = this.slotTexts.get(layout.slot);
      const hitArea = this.slotHitAreas.find(h => h.slot === layout.slot);
      if (hitArea) hitArea.item = item;
      if (!gfx || !text) continue;

      const slotX = layout.x - SLOT_BOX_W / 2;
      const slotY = slotsTop + layout.y;

      gfx.clear();
      if (item) {
        const color = getSlotColor(layout.slot);
        gfx.fillStyle(color, 0.08);
        gfx.fillRoundedRect(slotX, slotY, SLOT_BOX_W, SLOT_BOX_H, 4);
        gfx.lineStyle(1, color, 0.3);
        gfx.strokeRoundedRect(slotX, slotY, SLOT_BOX_W, SLOT_BOX_H, 4);
        text.setText(item.name);
        text.setColor('#ccdde8');
      } else {
        gfx.fillStyle(C_SLOT_EMPTY, 0.3);
        gfx.fillRoundedRect(slotX, slotY, SLOT_BOX_W, SLOT_BOX_H, 4);
        gfx.lineStyle(1, C_CYAN, 0.06);
        gfx.strokeRoundedRect(slotX, slotY, SLOT_BOX_W, SLOT_BOX_H, 4);
        text.setText('—');
        text.setColor('#334455');
      }
    }
  }

  // === INVENTORY ROWS ===

  private rebuildInventoryRows(): void {
    if (!this.invContainer) return;

    for (const g of this.invRowGraphics) g.destroy();
    for (const t of this.invRowTexts) t.destroy();
    if (this.invEmptyText) { this.invEmptyText.destroy(); this.invEmptyText = undefined; }
    this.invRowGraphics = [];
    this.invRowTexts = [];
    this.inventoryHitAreas = [];

    const rowsTop = -INV_H / 2 + 42;
    const rowWidth = INV_W - PADDING * 2;

    if (this.inventory.length === 0) {
      this.invEmptyText = this.scene.add.text(0, 0, '(Empty)', { fontSize: '13px', color: '#334455' });
      this.invEmptyText.setOrigin(0.5);
      this.invContainer.add(this.invEmptyText);
      return;
    }

    const cx = this.invContainer.x;
    const cy = this.invContainer.y;
    const visibleItems = this.inventory.slice(0, MAX_VISIBLE_INV);

    for (let i = 0; i < visibleItems.length; i++) {
      const item = visibleItems[i];
      const rowTop = rowsTop + i * INV_ROW_H;

      const rowBg = this.scene.add.graphics();
      const bgAlpha = i % 2 === 0 ? 0.25 : 0.15;
      rowBg.fillStyle(0x112233, bgAlpha);
      rowBg.fillRoundedRect(-rowWidth / 2, rowTop, rowWidth, INV_ROW_H, 4);
      rowBg.lineStyle(1, C_CYAN, 0.06);
      rowBg.lineBetween(-rowWidth / 2 + 8, rowTop + INV_ROW_H, rowWidth / 2 - 8, rowTop + INV_ROW_H);
      this.invContainer.add(rowBg);
      this.invRowGraphics.push(rowBg);

      const label = this.scene.add.text(0, rowTop + INV_ROW_H / 2, this.formatItemLine(item), {
        fontSize: '13px', color: '#ccdde8',
      });
      label.setOrigin(0.5);
      this.invContainer.add(label);
      this.invRowTexts.push(label);

      this.inventoryHitAreas.push({
        rect: { x: cx - rowWidth / 2, y: cy + rowTop, w: rowWidth, h: INV_ROW_H },
        item,
      });
    }

    if (this.inventory.length > MAX_VISIBLE_INV) {
      const moreY = rowsTop + MAX_VISIBLE_INV * INV_ROW_H + 4;
      const moreText = this.scene.add.text(0, moreY, `+${this.inventory.length - MAX_VISIBLE_INV} more...`, {
        fontSize: '11px', color: '#445566',
      });
      moreText.setOrigin(0.5);
      this.invContainer.add(moreText);
      this.invRowTexts.push(moreText);
    }
  }

  private formatItemLine(item: ItemState): string {
    const type = getItemType(item);
    switch (type) {
      case 'weapon': return item.attack_power ? `${item.name}  ATK ${item.attack_power}` : item.name;
      case 'armor': return item.defense_rating ? `${item.name}  DEF ${item.defense_rating}` : item.name;
      case 'consumable': {
        if (item.healing_amount) return `${item.name}  +${item.healing_amount} HP`;
        if (item.mana_amount) return `${item.name}  +${item.mana_amount} MP`;
        return item.name;
      }
      default: return item.quantity > 1 ? `${item.name} x${item.quantity}` : item.name;
    }
  }

  // === INPUT ===

  private handlePointerMove(pointer: Phaser.Input.Pointer): void {
    if (!this.visible || this.contextMenu) return;

    let foundSlot: EquipmentSlot | undefined;
    for (const hit of this.slotHitAreas) {
      if (this.pointInRect(pointer.x, pointer.y, hit.rect)) { foundSlot = hit.slot; break; }
    }

    if (foundSlot !== this.hoveredSlot) {
      if (this.hoveredSlot) { this.setSlotHover(this.hoveredSlot, false); this.hideTooltip(); }
      this.hoveredSlot = foundSlot;
      if (foundSlot) {
        this.setSlotHover(foundSlot, true);
        const item = this.equipped[foundSlot];
        if (item) this.showTooltip(item, pointer.x, pointer.y);
      }
    } else if (foundSlot) {
      this.moveTooltip(pointer.x, pointer.y);
    }

    let foundInvIdx = -1;
    if (!foundSlot) {
      for (let i = 0; i < this.inventoryHitAreas.length; i++) {
        if (this.pointInRect(pointer.x, pointer.y, this.inventoryHitAreas[i].rect)) { foundInvIdx = i; break; }
      }
    }

    if (foundInvIdx !== this.hoveredInvIndex) {
      if (this.hoveredInvIndex !== -1) { this.setInvRowHover(this.hoveredInvIndex, false); this.hideTooltip(); }
      this.hoveredInvIndex = foundInvIdx;
      if (foundInvIdx !== -1) {
        this.setInvRowHover(foundInvIdx, true);
        this.showTooltip(this.inventoryHitAreas[foundInvIdx].item, pointer.x, pointer.y);
      }
    } else if (foundInvIdx !== -1) {
      this.moveTooltip(pointer.x, pointer.y);
    }
  }

  private handlePointerDown(pointer: Phaser.Input.Pointer): void {
    if (!this.visible) return;

    if (this.contextMenu) {
      if (this.handleContextMenuClick(pointer)) return;
      this.dismissContextMenu();
      return;
    }

    if (pointer.rightButtonDown()) {
      for (const hit of this.slotHitAreas) {
        if (this.pointInRect(pointer.x, pointer.y, hit.rect) && hit.item) {
          this.showUnequipMenu(hit.item, hit.slot, pointer.x, pointer.y);
          return;
        }
      }
      for (const hitArea of this.inventoryHitAreas) {
        if (this.pointInRect(pointer.x, pointer.y, hitArea.rect)) {
          this.showEquipMenu(hitArea.item, pointer.x, pointer.y);
          return;
        }
      }
    }
  }

  // === CONTEXT MENU ===

  private showEquipMenu(item: ItemState, screenX: number, screenY: number): void {
    this.dismissContextMenu();
    this.hideTooltip();
    const validSlots = getValidSlotsForItem(item);
    if (validSlots.length === 0) return;

    const options: ContextMenuOption[] = validSlots.map(slot => ({
      label: `Equip to ${getSlotDisplayName(slot)}`,
      action: () => { this.equipItem(item, slot); this.dismissContextMenu(); },
    }));
    this.buildContextMenu(options, screenX, screenY);
  }

  private showUnequipMenu(item: ItemState, slot: EquipmentSlot, screenX: number, screenY: number): void {
    this.dismissContextMenu();
    this.hideTooltip();
    this.buildContextMenu([{
      label: 'Unequip',
      action: () => { this.unequipItem(item, slot); this.dismissContextMenu(); },
    }], screenX, screenY);
  }

  private buildContextMenu(options: ContextMenuOption[], screenX: number, screenY: number): void {
    const menuWidth = 200;
    const rowHeight = 32;
    const menuHeight = options.length * rowHeight + 8;
    const children: Phaser.GameObjects.GameObject[] = [];

    const bg = this.scene.add.graphics();
    bg.fillStyle(0x080810, 0.95);
    bg.fillRoundedRect(0, 0, menuWidth, menuHeight, 6);
    bg.lineStyle(1, C_CYAN, 0.4);
    bg.strokeRoundedRect(0, 0, menuWidth, menuHeight, 6);
    children.push(bg);

    for (let i = 0; i < options.length; i++) {
      const optY = 4 + i * rowHeight;
      const optBg = this.scene.add.graphics();
      children.push(optBg);
      const label = this.scene.add.text(14, optY + rowHeight / 2, options[i].label, {
        fontSize: '13px', color: '#ccdde8',
      });
      label.setOrigin(0, 0.5);
      children.push(label);
    }

    let x = screenX;
    let y = screenY;
    const cam = this.scene.cameras.main;
    if (x + menuWidth > cam.width) x = cam.width - menuWidth - 4;
    if (y + menuHeight > cam.height) y = cam.height - menuHeight - 4;

    this.contextMenu = this.scene.add.container(x, y, children);
    this.contextMenu.setDepth(2100);
    this.contextMenu.setScrollFactor(0);
    (this.contextMenu as any)._menuOptions = options;
    (this.contextMenu as any)._menuRowHeight = rowHeight;
    (this.contextMenu as any)._menuWidth = menuWidth;
  }

  private handleContextMenuClick(pointer: Phaser.Input.Pointer): boolean {
    if (!this.contextMenu) return false;
    const menu = this.contextMenu;
    const options = (menu as any)._menuOptions as ContextMenuOption[];
    const rowHeight = (menu as any)._menuRowHeight as number;
    const menuWidth = (menu as any)._menuWidth as number;
    const localX = pointer.x - menu.x;
    const localY = pointer.y - menu.y;
    if (localX < 0 || localX > menuWidth || localY < 0 || localY > options.length * rowHeight + 8) return false;
    const index = Math.floor((localY - 4) / rowHeight);
    if (index >= 0 && index < options.length) { options[index].action(); return true; }
    return false;
  }

  dismissContextMenu(): void {
    if (this.contextMenu) { this.contextMenu.destroy(); this.contextMenu = undefined; }
  }

  // === EQUIP/UNEQUIP ===

  private equipItem(item: ItemState, slot: EquipmentSlot): void {
    const existing = this.equipped[slot];
    if (existing) this.inventory.push(existing);
    this.inventory = this.inventory.filter(i => i.entity_id !== item.entity_id);
    this.equipped[slot] = item;
    this.refreshSlots();
    this.rebuildInventoryRows();
    this.onEquip?.(item, slot);
  }

  private unequipItem(item: ItemState, slot: EquipmentSlot): void {
    this.inventory.push(item);
    this.equipped[slot] = null;
    this.refreshSlots();
    this.rebuildInventoryRows();
    this.onUnequip?.(item, slot);
  }

  // === HOVER ===

  private setSlotHover(slot: EquipmentSlot, hovered: boolean): void {
    const gfx = this.slotGraphics.get(slot);
    const layout = SLOT_LAYOUT.find(l => l.slot === slot);
    if (!gfx || !layout) return;
    const slotX = layout.x - SLOT_BOX_W / 2;
    const slotsTop = -EQUIP_H / 2 + 48;
    const slotY = slotsTop + layout.y;
    const item = this.equipped[slot];

    gfx.clear();
    if (hovered) {
      const color = item ? getSlotColor(slot) : C_CYAN;
      gfx.fillStyle(color, 0.15);
      gfx.fillRoundedRect(slotX, slotY, SLOT_BOX_W, SLOT_BOX_H, 4);
      gfx.lineStyle(1, color, 0.5);
      gfx.strokeRoundedRect(slotX, slotY, SLOT_BOX_W, SLOT_BOX_H, 4);
    } else if (item) {
      const color = getSlotColor(slot);
      gfx.fillStyle(color, 0.08);
      gfx.fillRoundedRect(slotX, slotY, SLOT_BOX_W, SLOT_BOX_H, 4);
      gfx.lineStyle(1, color, 0.3);
      gfx.strokeRoundedRect(slotX, slotY, SLOT_BOX_W, SLOT_BOX_H, 4);
    } else {
      gfx.fillStyle(C_SLOT_EMPTY, 0.3);
      gfx.fillRoundedRect(slotX, slotY, SLOT_BOX_W, SLOT_BOX_H, 4);
      gfx.lineStyle(1, C_CYAN, 0.06);
      gfx.strokeRoundedRect(slotX, slotY, SLOT_BOX_W, SLOT_BOX_H, 4);
    }
  }

  private setInvRowHover(index: number, hovered: boolean): void {
    const gfx = this.invRowGraphics[index];
    const text = this.invRowTexts[index];
    if (!gfx || !text) return;
    const rowWidth = INV_W - PADDING * 2;
    const rowsTop = -INV_H / 2 + 42;
    const rowTop = rowsTop + index * INV_ROW_H;

    gfx.clear();
    if (hovered) {
      gfx.fillStyle(C_CYAN, 0.08);
      gfx.fillRoundedRect(-rowWidth / 2, rowTop, rowWidth, INV_ROW_H, 4);
      gfx.lineStyle(1, C_CYAN, 0.2);
      gfx.strokeRoundedRect(-rowWidth / 2, rowTop, rowWidth, INV_ROW_H, 4);
      text.setColor('#00f0ff');
    } else {
      const bgAlpha = index % 2 === 0 ? 0.25 : 0.15;
      gfx.fillStyle(0x112233, bgAlpha);
      gfx.fillRoundedRect(-rowWidth / 2, rowTop, rowWidth, INV_ROW_H, 4);
      gfx.lineStyle(1, C_CYAN, 0.06);
      gfx.lineBetween(-rowWidth / 2 + 8, rowTop + INV_ROW_H, rowWidth / 2 - 8, rowTop + INV_ROW_H);
      text.setColor('#ccdde8');
    }
  }

  // === TOOLTIP ===

  private showTooltip(item: ItemState, screenX: number, screenY: number): void {
    this.hideTooltip();
    const padding = 12;
    const tooltipWidth = 200;
    const children: Phaser.GameObjects.GameObject[] = [];
    let curY = padding;

    const name = this.scene.add.text(padding, curY, item.name, { fontSize: '14px', color: '#00f0ff', fontStyle: 'bold' });
    children.push(name);
    curY += 20;

    const type = getItemType(item);
    const typeColors: Record<string, string> = { weapon: '#ff4466', armor: '#44aaff', consumable: '#44ff88', unknown: '#556677' };
    const typeText = this.scene.add.text(padding, curY, type.toUpperCase(), { fontSize: '10px', color: typeColors[type] || '#556677', letterSpacing: 2 });
    children.push(typeText);
    curY += 18;

    if (item.attack_power) { this.addStat(children, padding, curY, tooltipWidth, 'ATK', `${item.attack_power}`, '#ff4466'); curY += 18; }
    if (item.critical_rate) { this.addStat(children, padding, curY, tooltipWidth, 'CRIT', `${Math.round(item.critical_rate)}%`, '#ffaa33'); curY += 18; }
    if (item.defense_rating) { this.addStat(children, padding, curY, tooltipWidth, 'DEF', `${item.defense_rating}`, '#44aaff'); curY += 18; }
    if (item.healing_amount) { this.addStat(children, padding, curY, tooltipWidth, 'HEAL', `+${item.healing_amount} HP`, '#44ff88'); curY += 18; }
    if (item.mana_amount) { this.addStat(children, padding, curY, tooltipWidth, 'MANA', `+${item.mana_amount} MP`, '#aa88ff'); curY += 18; }

    if (item.description) {
      curY += 4;
      const desc = this.scene.add.text(padding, curY, item.description, {
        fontSize: '11px', color: '#667788', wordWrap: { width: tooltipWidth - padding * 2 }, lineSpacing: 3,
      });
      children.push(desc);
      curY += desc.height;
    }

    const tooltipHeight = curY + padding;
    const bg = this.scene.add.graphics();
    const borderColor = parseInt((typeColors[type] || '#00f0ff').slice(1), 16);
    bg.fillStyle(0x080810, 0.95);
    bg.fillRoundedRect(0, 0, tooltipWidth, tooltipHeight, 6);
    bg.lineStyle(1, borderColor, 0.4);
    bg.strokeRoundedRect(0, 0, tooltipWidth, tooltipHeight, 6);
    children.unshift(bg);

    let x = screenX + 14;
    let y = screenY - 10;
    const cam = this.scene.cameras.main;
    if (x + tooltipWidth > cam.width) x = screenX - tooltipWidth - 8;
    if (y + tooltipHeight > cam.height) y = screenY - tooltipHeight - 8;

    this.tooltip = this.scene.add.container(x, y, children);
    this.tooltip.setDepth(2200);
    this.tooltip.setScrollFactor(0);
  }

  private addStat(children: Phaser.GameObjects.GameObject[], padding: number, y: number, width: number, label: string, value: string, color: string): void {
    children.push(
      this.scene.add.text(padding, y, label, { fontSize: '11px', color: '#556677', letterSpacing: 2 }),
      this.scene.add.text(width - padding, y, value, { fontSize: '12px', color }).setOrigin(1, 0),
    );
  }

  private moveTooltip(screenX: number, screenY: number): void {
    if (this.tooltip) this.tooltip.setPosition(screenX + 14, screenY - 10);
  }

  private hideTooltip(): void {
    if (this.tooltip) { this.tooltip.destroy(); this.tooltip = undefined; }
  }

  // === UTIL ===

  private pointInRect(px: number, py: number, rect: { x: number; y: number; w: number; h: number }): boolean {
    return px >= rect.x && px <= rect.x + rect.w && py >= rect.y && py <= rect.y + rect.h;
  }
}
