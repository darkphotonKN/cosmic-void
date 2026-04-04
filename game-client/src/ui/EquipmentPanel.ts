import {
  ItemState,
  EquipmentSlot,
  EquippedItems,
  getItemType,
  getValidSlotsForItem,
  getSlotDisplayName,
} from '@/types/gameState';

// Slot layout configuration: label position and slot box position
interface SlotLayout {
  slot: EquipmentSlot;
  label: string;
  x: number; // offset from panel center
  y: number; // offset from top of equipment area
}

// Row data for manual hit testing
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

const PANEL_WIDTH = 360;
const EQUIP_AREA_HEIGHT = 260;
const INVENTORY_ROW_HEIGHT = 28;
const MAX_VISIBLE_INVENTORY = 6;
const PANEL_PADDING = 16;
const SLOT_BOX_W = 120;
const SLOT_BOX_H = 22;

// Colors
const COLOR_CYAN = 0x00f0ff;
const COLOR_BG = 0x0a0a12;
const COLOR_SLOT_EMPTY = 0x112233;
const COLOR_WEAPON = 0xff4466;
const COLOR_ARMOR = 0x44aaff;
const COLOR_CONSUMABLE = 0x44ff88;
const COLOR_RING = 0xdd88ff;

function getSlotColor(slot: EquipmentSlot): number {
  if (slot === 'weapon') return COLOR_WEAPON;
  if (slot === 'ring_1' || slot === 'ring_2') return COLOR_RING;
  if (slot.startsWith('consumable')) return COLOR_CONSUMABLE;
  return COLOR_ARMOR;
}

// Equipment slot layout — body silhouette arrangement
const SLOT_LAYOUT: SlotLayout[] = [
  // Left column: weapon + consumables
  { slot: 'weapon', label: 'WEAPON', x: -70, y: 0 },
  { slot: 'consumable_1', label: 'CONS 1', x: -70, y: 30 },
  { slot: 'consumable_2', label: 'CONS 2', x: -70, y: 60 },
  { slot: 'consumable_3', label: 'CONS 3', x: -70, y: 90 },
  // Right column: armor
  { slot: 'head', label: 'HEAD', x: 70, y: 0 },
  { slot: 'body', label: 'BODY', x: 70, y: 30 },
  { slot: 'hands', label: 'HANDS', x: 70, y: 60 },
  { slot: 'feet', label: 'FEET', x: 70, y: 90 },
  // Bottom: rings
  { slot: 'ring_1', label: 'RING 1', x: -70, y: 130 },
  { slot: 'ring_2', label: 'RING 2', x: 70, y: 130 },
];

export class EquipmentPanel {
  private scene: Phaser.Scene;
  private container?: Phaser.GameObjects.Container;
  private visible = false;

  // State
  private equipped: EquippedItems = {
    weapon: null, head: null, body: null, hands: null, feet: null,
    ring_1: null, ring_2: null, consumable_1: null, consumable_2: null, consumable_3: null,
  };
  private inventory: ItemState[] = [];

  // Hit areas
  private slotHitAreas: SlotHitArea[] = [];
  private inventoryHitAreas: InventoryRowHitArea[] = [];

  // Slot graphics/text references for updates
  private slotGraphics: Map<EquipmentSlot, Phaser.GameObjects.Graphics> = new Map();
  private slotTexts: Map<EquipmentSlot, Phaser.GameObjects.Text> = new Map();
  private slotLabels: Map<EquipmentSlot, Phaser.GameObjects.Text> = new Map();

  // Inventory row references
  private invRowGraphics: Phaser.GameObjects.Graphics[] = [];
  private invRowTexts: Phaser.GameObjects.Text[] = [];
  private invEmptyText?: Phaser.GameObjects.Text;

  // Context menu
  private contextMenu?: Phaser.GameObjects.Container;

  // Hover state
  private hoveredSlot?: EquipmentSlot;
  private hoveredInvIndex = -1;

  // Tooltip
  private tooltip?: Phaser.GameObjects.Container;

  // Callbacks
  onEquip?: (item: ItemState, slot: EquipmentSlot) => void;
  onUnequip?: (item: ItemState, slot: EquipmentSlot) => void;

  // Pointer listener refs for cleanup
  private pointerMoveHandler?: (p: Phaser.Input.Pointer) => void;
  private pointerDownHandler?: (p: Phaser.Input.Pointer) => void;

  constructor(scene: Phaser.Scene) {
    this.scene = scene;
  }

  toggle(): void {
    if (this.visible) {
      this.hide();
    } else {
      this.show();
    }
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
    if (this.visible) {
      this.rebuildInventoryRows();
    }
  }

  updateEquipment(equipped: EquippedItems): void {
    this.equipped = equipped;
    if (this.visible) {
      this.refreshSlots();
    }
  }

  destroy(): void {
    this.hide();
  }

  // === BUILD ===

  private build(): void {
    const cam = this.scene.cameras.main;
    const centerX = cam.width / 2;
    const centerY = cam.height / 2;

    const inventoryHeight = INVENTORY_ROW_HEIGHT * MAX_VISIBLE_INVENTORY + 40; // 40 for title+padding
    const panelHeight = EQUIP_AREA_HEIGHT + inventoryHeight + 40; // 40 for top title + hint

    const children: Phaser.GameObjects.GameObject[] = [];

    // Background
    const bg = this.scene.add.graphics();
    bg.fillStyle(COLOR_BG, 0.92);
    bg.fillRoundedRect(-PANEL_WIDTH / 2, -panelHeight / 2, PANEL_WIDTH, panelHeight, 8);
    bg.lineStyle(1, COLOR_CYAN, 0.6);
    bg.strokeRoundedRect(-PANEL_WIDTH / 2, -panelHeight / 2, PANEL_WIDTH, panelHeight, 8);
    children.push(bg);

    // Title
    const title = this.scene.add.text(0, -panelHeight / 2 + 16, 'EQUIPMENT', {
      fontSize: '18px',
      color: '#00f0ff',
      letterSpacing: 6,
    });
    title.setOrigin(0.5);
    children.push(title);

    // Equipment area separator
    const equipTop = -panelHeight / 2 + 42;

    // Build equipment slots
    this.slotHitAreas = [];
    this.slotGraphics.clear();
    this.slotTexts.clear();
    this.slotLabels.clear();

    for (const layout of SLOT_LAYOUT) {
      const slotX = layout.x;
      const slotY = equipTop + layout.y;

      // Label
      const label = this.scene.add.text(slotX - SLOT_BOX_W / 2, slotY - 1, layout.label, {
        fontSize: '9px',
        color: '#445566',
        letterSpacing: 2,
      });
      children.push(label);
      this.slotLabels.set(layout.slot, label);

      // Slot box
      const slotGfx = this.scene.add.graphics();
      children.push(slotGfx);
      this.slotGraphics.set(layout.slot, slotGfx);

      // Slot item text
      const slotText = this.scene.add.text(slotX, slotY + SLOT_BOX_H / 2 + 10, '—', {
        fontSize: '12px',
        color: '#334455',
      });
      slotText.setOrigin(0.5);
      children.push(slotText);
      this.slotTexts.set(layout.slot, slotText);

      // Hit area (will be set after container position is known)
      this.slotHitAreas.push({
        slot: layout.slot,
        rect: { x: 0, y: 0, w: SLOT_BOX_W, h: SLOT_BOX_H },
        item: null,
      });
    }

    // Separator between equipment and inventory
    const sepY = equipTop + 160;
    const sep = this.scene.add.graphics();
    sep.lineStyle(1, COLOR_CYAN, 0.15);
    sep.lineBetween(-PANEL_WIDTH / 2 + PANEL_PADDING, sepY, PANEL_WIDTH / 2 - PANEL_PADDING, sepY);
    children.push(sep);

    // Inventory title
    const invTitle = this.scene.add.text(0, sepY + 8, 'INVENTORY', {
      fontSize: '14px',
      color: '#00f0ff',
      letterSpacing: 4,
    });
    invTitle.setOrigin(0.5);
    children.push(invTitle);

    // Hint
    const hint = this.scene.add.text(0, panelHeight / 2 - 18, 'RIGHT-CLICK  //  I Close', {
      fontSize: '11px',
      color: '#334455',
    });
    hint.setOrigin(0.5);
    children.push(hint);

    // Create container
    this.container = this.scene.add.container(centerX, centerY, children);
    this.container.setDepth(1100);
    this.container.setScrollFactor(0);

    // Now calculate screen-space hit areas
    this.updateHitAreas();

    // Refresh slot visuals
    this.refreshSlots();

    // Build inventory rows
    this.rebuildInventoryRows();

    // Register input handlers
    this.pointerMoveHandler = (p: Phaser.Input.Pointer) => this.handlePointerMove(p);
    this.pointerDownHandler = (p: Phaser.Input.Pointer) => this.handlePointerDown(p);
    this.scene.input.on('pointermove', this.pointerMoveHandler);
    this.scene.input.on('pointerdown', this.pointerDownHandler);
  }

  private cleanup(): void {
    if (this.pointerMoveHandler) {
      this.scene.input.off('pointermove', this.pointerMoveHandler);
      this.pointerMoveHandler = undefined;
    }
    if (this.pointerDownHandler) {
      this.scene.input.off('pointerdown', this.pointerDownHandler);
      this.pointerDownHandler = undefined;
    }
    this.slotGraphics.clear();
    this.slotTexts.clear();
    this.slotLabels.clear();
    this.invRowGraphics = [];
    this.invRowTexts = [];
    this.invEmptyText = undefined;
    this.slotHitAreas = [];
    this.inventoryHitAreas = [];
    this.hoveredSlot = undefined;
    this.hoveredInvIndex = -1;
    if (this.container) {
      this.container.destroy();
      this.container = undefined;
    }
  }

  // === HIT AREAS ===

  private updateHitAreas(): void {
    if (!this.container) return;
    const cx = this.container.x;
    const cy = this.container.y;
    const panelHeight = this.getPanelHeight();
    const equipTop = -panelHeight / 2 + 42;

    for (let i = 0; i < SLOT_LAYOUT.length; i++) {
      const layout = SLOT_LAYOUT[i];
      const screenX = cx + layout.x - SLOT_BOX_W / 2;
      const screenY = cy + equipTop + layout.y;
      this.slotHitAreas[i].rect = {
        x: screenX,
        y: screenY,
        w: SLOT_BOX_W,
        h: SLOT_BOX_H,
      };
    }
  }

  private getPanelHeight(): number {
    const inventoryHeight = INVENTORY_ROW_HEIGHT * MAX_VISIBLE_INVENTORY + 40;
    return EQUIP_AREA_HEIGHT + inventoryHeight + 40;
  }

  // === SLOT RENDERING ===

  private refreshSlots(): void {
    for (const layout of SLOT_LAYOUT) {
      const item = this.equipped[layout.slot];
      const gfx = this.slotGraphics.get(layout.slot);
      const text = this.slotTexts.get(layout.slot);
      const hitArea = this.slotHitAreas.find(h => h.slot === layout.slot);

      if (hitArea) hitArea.item = item;
      if (!gfx || !text) continue;

      const slotX = layout.x - SLOT_BOX_W / 2;
      const panelHeight = this.getPanelHeight();
      const equipTop = -panelHeight / 2 + 42;
      const slotY = equipTop + layout.y;

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
        gfx.fillStyle(COLOR_SLOT_EMPTY, 0.3);
        gfx.fillRoundedRect(slotX, slotY, SLOT_BOX_W, SLOT_BOX_H, 4);
        gfx.lineStyle(1, COLOR_CYAN, 0.06);
        gfx.strokeRoundedRect(slotX, slotY, SLOT_BOX_W, SLOT_BOX_H, 4);
        text.setText('—');
        text.setColor('#334455');
      }
    }
  }

  // === INVENTORY ROWS ===

  private rebuildInventoryRows(): void {
    if (!this.container) return;

    // Clean existing
    for (const g of this.invRowGraphics) g.destroy();
    for (const t of this.invRowTexts) t.destroy();
    if (this.invEmptyText) { this.invEmptyText.destroy(); this.invEmptyText = undefined; }
    this.invRowGraphics = [];
    this.invRowTexts = [];
    this.inventoryHitAreas = [];

    const panelHeight = this.getPanelHeight();
    const equipTop = -panelHeight / 2 + 42;
    const sepY = equipTop + 160;
    const invStartY = sepY + 28;
    const rowWidth = PANEL_WIDTH - PANEL_PADDING * 2;

    if (this.inventory.length === 0) {
      this.invEmptyText = this.scene.add.text(0, invStartY + 40, '(Empty)', {
        fontSize: '13px',
        color: '#334455',
      });
      this.invEmptyText.setOrigin(0.5);
      this.container.add(this.invEmptyText);
      return;
    }

    const cx = this.container.x;
    const cy = this.container.y;

    const visibleItems = this.inventory.slice(0, MAX_VISIBLE_INVENTORY);
    for (let i = 0; i < visibleItems.length; i++) {
      const item = visibleItems[i];
      const rowTop = invStartY + i * INVENTORY_ROW_HEIGHT;
      const rowCenterY = rowTop + INVENTORY_ROW_HEIGHT / 2;

      // Row background
      const rowBg = this.scene.add.graphics();
      const bgAlpha = i % 2 === 0 ? 0.25 : 0.15;
      rowBg.fillStyle(0x112233, bgAlpha);
      rowBg.fillRoundedRect(-rowWidth / 2, rowTop, rowWidth, INVENTORY_ROW_HEIGHT, 4);
      rowBg.lineStyle(1, COLOR_CYAN, 0.06);
      rowBg.lineBetween(-rowWidth / 2 + 8, rowTop + INVENTORY_ROW_HEIGHT, rowWidth / 2 - 8, rowTop + INVENTORY_ROW_HEIGHT);
      this.container.add(rowBg);
      this.invRowGraphics.push(rowBg);

      // Item text
      const label = this.scene.add.text(0, rowCenterY, this.formatItemLine(item), {
        fontSize: '13px',
        color: '#ccdde8',
      });
      label.setOrigin(0.5);
      this.container.add(label);
      this.invRowTexts.push(label);

      // Screen-space hit area
      this.inventoryHitAreas.push({
        rect: {
          x: cx - rowWidth / 2,
          y: cy + rowTop,
          w: rowWidth,
          h: INVENTORY_ROW_HEIGHT,
        },
        item,
      });
    }

    // Overflow indicator
    if (this.inventory.length > MAX_VISIBLE_INVENTORY) {
      const moreY = invStartY + MAX_VISIBLE_INVENTORY * INVENTORY_ROW_HEIGHT + 4;
      const moreText = this.scene.add.text(0, moreY, `+${this.inventory.length - MAX_VISIBLE_INVENTORY} more...`, {
        fontSize: '11px',
        color: '#445566',
      });
      moreText.setOrigin(0.5);
      this.container.add(moreText);
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

  // === INPUT HANDLING ===

  private handlePointerMove(pointer: Phaser.Input.Pointer): void {
    if (!this.visible || this.contextMenu) return;

    // Check equipment slots
    let foundSlot: EquipmentSlot | undefined;
    for (const hit of this.slotHitAreas) {
      if (this.pointInRect(pointer.x, pointer.y, hit.rect)) {
        foundSlot = hit.slot;
        break;
      }
    }

    if (foundSlot !== this.hoveredSlot) {
      // Unhover previous slot
      if (this.hoveredSlot) {
        this.setSlotHover(this.hoveredSlot, false);
        this.hideTooltip();
      }
      this.hoveredSlot = foundSlot;
      if (foundSlot) {
        this.setSlotHover(foundSlot, true);
        const item = this.equipped[foundSlot];
        if (item) this.showTooltip(item, pointer.x, pointer.y);
      }
    } else if (foundSlot) {
      this.moveTooltip(pointer.x, pointer.y);
    }

    // Check inventory rows
    let foundInvIdx = -1;
    if (!foundSlot) {
      for (let i = 0; i < this.inventoryHitAreas.length; i++) {
        if (this.pointInRect(pointer.x, pointer.y, this.inventoryHitAreas[i].rect)) {
          foundInvIdx = i;
          break;
        }
      }
    }

    if (foundInvIdx !== this.hoveredInvIndex) {
      if (this.hoveredInvIndex !== -1) {
        this.setInvRowHover(this.hoveredInvIndex, false);
        this.hideTooltip();
      }
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

    // If context menu is open, check if click is on a menu option
    if (this.contextMenu) {
      if (this.handleContextMenuClick(pointer)) return;
      this.dismissContextMenu();
      return;
    }

    // Right-click only for context menus
    if (pointer.rightButtonDown()) {
      // Check equipment slots
      for (const hit of this.slotHitAreas) {
        if (this.pointInRect(pointer.x, pointer.y, hit.rect) && hit.item) {
          this.showUnequipMenu(hit.item, hit.slot, pointer.x, pointer.y);
          return;
        }
      }

      // Check inventory rows
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
      action: () => {
        this.equipItem(item, slot);
        this.dismissContextMenu();
      },
    }));

    this.buildContextMenu(options, screenX, screenY);
  }

  private showUnequipMenu(item: ItemState, slot: EquipmentSlot, screenX: number, screenY: number): void {
    this.dismissContextMenu();
    this.hideTooltip();

    const options: ContextMenuOption[] = [{
      label: 'Unequip',
      action: () => {
        this.unequipItem(item, slot);
        this.dismissContextMenu();
      },
    }];

    this.buildContextMenu(options, screenX, screenY);
  }

  private buildContextMenu(options: ContextMenuOption[], screenX: number, screenY: number): void {
    const menuWidth = 180;
    const rowHeight = 28;
    const menuHeight = options.length * rowHeight + 8;

    const children: Phaser.GameObjects.GameObject[] = [];

    // Background
    const bg = this.scene.add.graphics();
    bg.fillStyle(0x080810, 0.95);
    bg.fillRoundedRect(0, 0, menuWidth, menuHeight, 6);
    bg.lineStyle(1, COLOR_CYAN, 0.4);
    bg.strokeRoundedRect(0, 0, menuWidth, menuHeight, 6);
    children.push(bg);

    // Options
    for (let i = 0; i < options.length; i++) {
      const optY = 4 + i * rowHeight;
      const optBg = this.scene.add.graphics();
      children.push(optBg);

      const label = this.scene.add.text(12, optY + rowHeight / 2, options[i].label, {
        fontSize: '12px',
        color: '#ccdde8',
      });
      label.setOrigin(0, 0.5);
      children.push(label);
    }

    // Position with screen edge clamping
    let x = screenX;
    let y = screenY;
    const cam = this.scene.cameras.main;
    if (x + menuWidth > cam.width) x = cam.width - menuWidth - 4;
    if (y + menuHeight > cam.height) y = cam.height - menuHeight - 4;

    this.contextMenu = this.scene.add.container(x, y, children);
    this.contextMenu.setDepth(2100);
    this.contextMenu.setScrollFactor(0);

    // Store options for hit testing
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

    if (localX < 0 || localX > menuWidth || localY < 0 || localY > options.length * rowHeight + 8) {
      return false;
    }

    const index = Math.floor((localY - 4) / rowHeight);
    if (index >= 0 && index < options.length) {
      options[index].action();
      return true;
    }

    return false;
  }

  dismissContextMenu(): void {
    if (this.contextMenu) {
      this.contextMenu.destroy();
      this.contextMenu = undefined;
    }
  }

  // === EQUIP/UNEQUIP ACTIONS ===

  private equipItem(item: ItemState, slot: EquipmentSlot): void {
    // Move currently equipped item back to inventory if slot is occupied
    const existing = this.equipped[slot];
    if (existing) {
      this.inventory.push(existing);
    }

    // Remove from inventory
    this.inventory = this.inventory.filter(i => i.entity_id !== item.entity_id);

    // Equip
    this.equipped[slot] = item;

    // Refresh
    this.refreshSlots();
    this.rebuildInventoryRows();

    // Callback
    this.onEquip?.(item, slot);
  }

  private unequipItem(item: ItemState, slot: EquipmentSlot): void {
    // Move to inventory
    this.inventory.push(item);

    // Clear slot
    this.equipped[slot] = null;

    // Refresh
    this.refreshSlots();
    this.rebuildInventoryRows();

    // Callback
    this.onUnequip?.(item, slot);
  }

  // === HOVER VISUALS ===

  private setSlotHover(slot: EquipmentSlot, hovered: boolean): void {
    const gfx = this.slotGraphics.get(slot);
    const layout = SLOT_LAYOUT.find(l => l.slot === slot);
    if (!gfx || !layout) return;

    const slotX = layout.x - SLOT_BOX_W / 2;
    const panelHeight = this.getPanelHeight();
    const equipTop = -panelHeight / 2 + 42;
    const slotY = equipTop + layout.y;
    const item = this.equipped[slot];

    gfx.clear();
    if (hovered) {
      const color = item ? getSlotColor(slot) : COLOR_CYAN;
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
      gfx.fillStyle(COLOR_SLOT_EMPTY, 0.3);
      gfx.fillRoundedRect(slotX, slotY, SLOT_BOX_W, SLOT_BOX_H, 4);
      gfx.lineStyle(1, COLOR_CYAN, 0.06);
      gfx.strokeRoundedRect(slotX, slotY, SLOT_BOX_W, SLOT_BOX_H, 4);
    }
  }

  private setInvRowHover(index: number, hovered: boolean): void {
    const gfx = this.invRowGraphics[index];
    const text = this.invRowTexts[index];
    if (!gfx || !text) return;

    const rowWidth = PANEL_WIDTH - PANEL_PADDING * 2;
    const panelHeight = this.getPanelHeight();
    const equipTop = -panelHeight / 2 + 42;
    const sepY = equipTop + 160;
    const invStartY = sepY + 28;
    const rowTop = invStartY + index * INVENTORY_ROW_HEIGHT;

    gfx.clear();
    if (hovered) {
      gfx.fillStyle(COLOR_CYAN, 0.08);
      gfx.fillRoundedRect(-rowWidth / 2, rowTop, rowWidth, INVENTORY_ROW_HEIGHT, 4);
      gfx.lineStyle(1, COLOR_CYAN, 0.2);
      gfx.strokeRoundedRect(-rowWidth / 2, rowTop, rowWidth, INVENTORY_ROW_HEIGHT, 4);
      text.setColor('#00f0ff');
    } else {
      const bgAlpha = index % 2 === 0 ? 0.25 : 0.15;
      gfx.fillStyle(0x112233, bgAlpha);
      gfx.fillRoundedRect(-rowWidth / 2, rowTop, rowWidth, INVENTORY_ROW_HEIGHT, 4);
      gfx.lineStyle(1, COLOR_CYAN, 0.06);
      gfx.lineBetween(-rowWidth / 2 + 8, rowTop + INVENTORY_ROW_HEIGHT, rowWidth / 2 - 8, rowTop + INVENTORY_ROW_HEIGHT);
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

    // Name
    const name = this.scene.add.text(padding, curY, item.name, {
      fontSize: '14px',
      color: '#00f0ff',
      fontStyle: 'bold',
    });
    children.push(name);
    curY += 20;

    // Type badge
    const type = getItemType(item);
    const typeColors: Record<string, string> = { weapon: '#ff4466', armor: '#44aaff', consumable: '#44ff88', unknown: '#556677' };
    const typeLabel = type.toUpperCase();
    const typeText = this.scene.add.text(padding, curY, typeLabel, {
      fontSize: '10px',
      color: typeColors[type] || '#556677',
      letterSpacing: 2,
    });
    children.push(typeText);
    curY += 18;

    // Stats
    if (item.attack_power) { this.addTooltipStat(children, padding, curY, tooltipWidth, 'ATK', `${item.attack_power}`, '#ff4466'); curY += 18; }
    if (item.critical_rate) { this.addTooltipStat(children, padding, curY, tooltipWidth, 'CRIT', `${Math.round(item.critical_rate)}%`, '#ffaa33'); curY += 18; }
    if (item.defense_rating) { this.addTooltipStat(children, padding, curY, tooltipWidth, 'DEF', `${item.defense_rating}`, '#44aaff'); curY += 18; }
    if (item.healing_amount) { this.addTooltipStat(children, padding, curY, tooltipWidth, 'HEAL', `+${item.healing_amount} HP`, '#44ff88'); curY += 18; }
    if (item.mana_amount) { this.addTooltipStat(children, padding, curY, tooltipWidth, 'MANA', `+${item.mana_amount} MP`, '#aa88ff'); curY += 18; }

    if (item.description) {
      curY += 4;
      const desc = this.scene.add.text(padding, curY, item.description, {
        fontSize: '11px',
        color: '#667788',
        wordWrap: { width: tooltipWidth - padding * 2 },
        lineSpacing: 3,
      });
      children.push(desc);
      curY += desc.height;
    }

    const tooltipHeight = curY + padding;

    // Background
    const bg = this.scene.add.graphics();
    const borderColor = parseInt((typeColors[type] || '#00f0ff').slice(1), 16);
    bg.fillStyle(0x080810, 0.95);
    bg.fillRoundedRect(0, 0, tooltipWidth, tooltipHeight, 6);
    bg.lineStyle(1, borderColor, 0.4);
    bg.strokeRoundedRect(0, 0, tooltipWidth, tooltipHeight, 6);
    children.unshift(bg);

    // Position
    let x = screenX + 14;
    let y = screenY - 10;
    const cam = this.scene.cameras.main;
    if (x + tooltipWidth > cam.width) x = screenX - tooltipWidth - 8;
    if (y + tooltipHeight > cam.height) y = screenY - tooltipHeight - 8;

    this.tooltip = this.scene.add.container(x, y, children);
    this.tooltip.setDepth(2200);
    this.tooltip.setScrollFactor(0);
  }

  private addTooltipStat(children: Phaser.GameObjects.GameObject[], padding: number, y: number, width: number, label: string, value: string, color: string): void {
    const lbl = this.scene.add.text(padding, y, label, { fontSize: '11px', color: '#556677', letterSpacing: 2 });
    const val = this.scene.add.text(width - padding, y, value, { fontSize: '12px', color });
    val.setOrigin(1, 0);
    children.push(lbl, val);
  }

  private moveTooltip(screenX: number, screenY: number): void {
    if (!this.tooltip) return;
    this.tooltip.setPosition(screenX + 14, screenY - 10);
  }

  private hideTooltip(): void {
    if (this.tooltip) {
      this.tooltip.destroy();
      this.tooltip = undefined;
    }
  }

  // === UTIL ===

  private pointInRect(px: number, py: number, rect: { x: number; y: number; w: number; h: number }): boolean {
    return px >= rect.x && px <= rect.x + rect.w && py >= rect.y && py <= rect.y + rect.h;
  }
}
