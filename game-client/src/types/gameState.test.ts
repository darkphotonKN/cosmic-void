import { describe, it, expect } from 'vitest';
import { getItemType, getValidSlotsForItem, getSlotDisplayName, ItemState, EquipmentSlot } from './gameState';

function makeItem(overrides: Partial<ItemState> = {}): ItemState {
  return {
    item_id: 'template-1',
    entity_id: 'entity-1',
    name: 'Test Item',
    quantity: 1,
    ...overrides,
  };
}

describe('getItemType', () => {
  it('should return weapon when item has attack_power', () => {
    expect(getItemType(makeItem({ attack_power: 10 }))).toBe('weapon');
  });

  it('should return weapon when item has weapon_type', () => {
    expect(getItemType(makeItem({ weapon_type: 'sword' }))).toBe('weapon');
  });

  it('should return armor when item has defense_rating', () => {
    expect(getItemType(makeItem({ defense_rating: 5 }))).toBe('armor');
  });

  it('should return armor when item has armor_slot', () => {
    expect(getItemType(makeItem({ armor_slot: 'chest' }))).toBe('armor');
  });

  it('should return armor when defense_rating is 0', () => {
    expect(getItemType(makeItem({ defense_rating: 0, armor_slot: 'head' }))).toBe('armor');
  });

  it('should return consumable when item has healing_amount', () => {
    expect(getItemType(makeItem({ healing_amount: 50 }))).toBe('consumable');
  });

  it('should return consumable when item has mana_amount', () => {
    expect(getItemType(makeItem({ mana_amount: 30 }))).toBe('consumable');
  });

  it('should return unknown for item with no type fields', () => {
    expect(getItemType(makeItem())).toBe('unknown');
  });
});

describe('getValidSlotsForItem', () => {
  it('should return [weapon] for weapon items', () => {
    expect(getValidSlotsForItem(makeItem({ attack_power: 10 }))).toEqual(['weapon']);
  });

  it('should return [head] for armor with armor_slot=head', () => {
    expect(getValidSlotsForItem(makeItem({ defense_rating: 5, armor_slot: 'head' }))).toEqual(['head']);
  });

  it('should return [body] for armor with armor_slot=chest', () => {
    expect(getValidSlotsForItem(makeItem({ defense_rating: 5, armor_slot: 'chest' }))).toEqual(['body']);
  });

  it('should return [hands] for armor with armor_slot=gloves', () => {
    expect(getValidSlotsForItem(makeItem({ defense_rating: 5, armor_slot: 'gloves' }))).toEqual(['hands']);
  });

  it('should return [feet] for armor with armor_slot=legs', () => {
    expect(getValidSlotsForItem(makeItem({ defense_rating: 5, armor_slot: 'legs' }))).toEqual(['feet']);
  });

  it('should return empty array for armor with unrecognized armor_slot', () => {
    expect(getValidSlotsForItem(makeItem({ defense_rating: 5, armor_slot: 'unknown_slot' }))).toEqual([]);
  });

  it('should return all consumable slots for consumable items', () => {
    expect(getValidSlotsForItem(makeItem({ healing_amount: 50 }))).toEqual(['consumable_1', 'consumable_2', 'consumable_3']);
  });

  it('should return empty array for unknown items', () => {
    expect(getValidSlotsForItem(makeItem())).toEqual([]);
  });

  it('should return empty array for armor with no armor_slot', () => {
    expect(getValidSlotsForItem(makeItem({ defense_rating: 0, armor_slot: undefined }))).toEqual([]);
  });
});

describe('getSlotDisplayName', () => {
  const cases: [EquipmentSlot, string][] = [
    ['weapon', 'Weapon'],
    ['head', 'Head'],
    ['body', 'Body'],
    ['hands', 'Hands'],
    ['feet', 'Feet'],
    ['ring_1', 'Ring 1'],
    ['ring_2', 'Ring 2'],
    ['consumable_1', 'Consumable 1'],
    ['consumable_2', 'Consumable 2'],
    ['consumable_3', 'Consumable 3'],
  ];

  cases.forEach(([slot, expected]) => {
    it(`should return "${expected}" for slot "${slot}"`, () => {
      expect(getSlotDisplayName(slot)).toBe(expected);
    });
  });
});
