-- Insert Item Types
INSERT INTO item_types (id, type_code, name, description, is_active, sort_order)
VALUES
    ('550e8400-e29b-41d4-a716-446655440001', 'weapon', 'Weapon', 'Weapons for combat', true, 1),
    ('550e8400-e29b-41d4-a716-446655440002', 'armor', 'Armor', 'Protective gear', true, 2),
    ('550e8400-e29b-41d4-a716-446655440003', 'consumable', 'Consumable', 'Items that can be consumed', true, 3);

-- Insert Item Rarities
INSERT INTO item_rarities (id, rarity_code, rarity_name, color_hex, drop_rate_multiplier, sort_order)
VALUES
    ('660e8400-e29b-41d4-a716-446655440001', 'common', 'Common', '#FFFFFF', 1.00, 1),
    ('660e8400-e29b-41d4-a716-446655440002', 'uncommon', 'Uncommon', '#00FF00', 0.75, 2),
    ('660e8400-e29b-41d4-a716-446655440003', 'rare', 'Rare', '#0000FF', 0.50, 3),
    ('660e8400-e29b-41d4-a716-446655440004', 'epic', 'Epic', '#9400D3', 0.25, 4),
    ('660e8400-e29b-41d4-a716-446655440005', 'legendary', 'Legendary', '#FFD700', 0.10, 5);

-- Insert Sample Weapons (7 items)
INSERT INTO weapons (id, type_id, rarity_id, attack_power, durability, critical_rate, weapon_type, description)
VALUES
    ('770e8400-e29b-41d4-a716-446655440001',
     '550e8400-e29b-41d4-a716-446655440001',
     '660e8400-e29b-41d4-a716-446655440001',
     10, 100, 0.05, 'sword', 'Basic iron sword'),

    ('770e8400-e29b-41d4-a716-446655440002',
     '550e8400-e29b-41d4-a716-446655440001',
     '660e8400-e29b-41d4-a716-446655440003',
     50, 150, 0.15, 'sword', 'Rare steel sword with increased critical rate'),

    ('770e8400-e29b-41d4-a716-446655440003',
     '550e8400-e29b-41d4-a716-446655440001',
     '660e8400-e29b-41d4-a716-446655440005',
     100, 200, 0.30, 'sword', 'Legendary dragon slayer sword'),

    ('770e8400-e29b-41d4-a716-446655440004',
     '550e8400-e29b-41d4-a716-446655440001',
     '660e8400-e29b-41d4-a716-446655440002',
     25, 120, 0.08, 'axe', 'Uncommon wooden axe'),

    ('770e8400-e29b-41d4-a716-446655440005',
     '550e8400-e29b-41d4-a716-446655440001',
     '660e8400-e29b-41d4-a716-446655440003',
     45, 140, 0.12, 'bow', 'Rare elven bow'),

    ('770e8400-e29b-41d4-a716-446655440006',
     '550e8400-e29b-41d4-a716-446655440001',
     '660e8400-e29b-41d4-a716-446655440004',
     70, 180, 0.20, 'sword', 'Epic shadow blade'),

    ('770e8400-e29b-41d4-a716-446655440007',
     '550e8400-e29b-41d4-a716-446655440001',
     '660e8400-e29b-41d4-a716-446655440002',
     15, 110, 0.06, 'axe', 'Uncommon iron axe');

-- Insert Sample Armors (6 items)
INSERT INTO armors (id, type_id, rarity_id, defense_rating, durability, magic_resistance, armor_slot, description)
VALUES
    ('880e8400-e29b-41d4-a716-446655440001',
     '550e8400-e29b-41d4-a716-446655440002',
     '660e8400-e29b-41d4-a716-446655440001',
     20, 100, 10, 'chest', 'Basic leather armor'),

    ('880e8400-e29b-41d4-a716-446655440002',
     '550e8400-e29b-41d4-a716-446655440002',
     '660e8400-e29b-41d4-a716-446655440004',
     80, 180, 50, 'chest', 'Epic mithril plate armor'),

    ('880e8400-e29b-41d4-a716-446655440003',
     '550e8400-e29b-41d4-a716-446655440002',
     '660e8400-e29b-41d4-a716-446655440001',
     15, 80, 5, 'head', 'Basic wooden helm'),

    ('880e8400-e29b-41d4-a716-446655440004',
     '550e8400-e29b-41d4-a716-446655440002',
     '660e8400-e29b-41d4-a716-446655440002',
     30, 120, 15, 'legs', 'Uncommon chainmail leggings'),

    ('880e8400-e29b-41d4-a716-446655440005',
     '550e8400-e29b-41d4-a716-446655440002',
     '660e8400-e29b-41d4-a716-446655440003',
     50, 150, 30, 'shield', 'Rare tower shield'),

    ('880e8400-e29b-41d4-a716-446655440006',
     '550e8400-e29b-41d4-a716-446655440002',
     '660e8400-e29b-41d4-a716-446655440005',
     120, 250, 80, 'chest', 'Legendary divine plate');

-- Insert Sample Consumables (7 items)
INSERT INTO consumables (id, type_id, rarity_id, healing_amount, mana_amount, buff_duration, max_stack_size, description)
VALUES
    ('990e8400-e29b-41d4-a716-446655440001',
     '550e8400-e29b-41d4-a716-446655440003',
     '660e8400-e29b-41d4-a716-446655440001',
     50, 0, 0, 99, 'Small health potion'),

    ('990e8400-e29b-41d4-a716-446655440002',
     '550e8400-e29b-41d4-a716-446655440003',
     '660e8400-e29b-41d4-a716-446655440002',
     100, 0, 0, 99, 'Medium health potion'),

    ('990e8400-e29b-41d4-a716-446655440003',
     '550e8400-e29b-41d4-a716-446655440003',
     '660e8400-e29b-41d4-a716-446655440001',
     0, 50, 0, 99, 'Small mana potion'),

    ('990e8400-e29b-41d4-a716-446655440004',
     '550e8400-e29b-41d4-a716-446655440003',
     '660e8400-e29b-41d4-a716-446655440003',
     200, 0, 0, 50, 'Large health potion'),

    ('990e8400-e29b-41d4-a716-446655440005',
     '550e8400-e29b-41d4-a716-446655440003',
     '660e8400-e29b-41d4-a716-446655440002',
     0, 100, 0, 99, 'Medium mana potion'),

    ('990e8400-e29b-41d4-a716-446655440006',
     '550e8400-e29b-41d4-a716-446655440003',
     '660e8400-e29b-41d4-a716-446655440004',
     500, 500, 60, 10, 'Epic elixir of restoration'),

    ('990e8400-e29b-41d4-a716-446655440007',
     '550e8400-e29b-41d4-a716-446655440003',
     '660e8400-e29b-41d4-a716-446655440001',
     0, 0, 30, 99, 'Attack buff scroll');

-- Insert Sample Item Templates (20 items)
INSERT INTO item_templates (id, item_name, item_code, type_id, rarity_id, item_type, item_id, icon_url, is_tradeable, is_droppable, required_level, base_sell_price, base_buy_price)
VALUES
    -- Weapons (7)
    ('aa0e8400-e29b-41d4-a716-446655440001',
     'Iron Sword', 'IRON_SWORD_001',
     '550e8400-e29b-41d4-a716-446655440001',
     '660e8400-e29b-41d4-a716-446655440001',
     'weapon', '770e8400-e29b-41d4-a716-446655440001',
     '/icons/weapons/iron_sword.png', true, true, 1, 10, 20),

    ('aa0e8400-e29b-41d4-a716-446655440002',
     'Steel Sword', 'STEEL_SWORD_001',
     '550e8400-e29b-41d4-a716-446655440001',
     '660e8400-e29b-41d4-a716-446655440003',
     'weapon', '770e8400-e29b-41d4-a716-446655440002',
     '/icons/weapons/steel_sword.png', true, true, 10, 100, 200),

    ('aa0e8400-e29b-41d4-a716-446655440003',
     'Dragon Slayer', 'DRAGON_SLAYER_001',
     '550e8400-e29b-41d4-a716-446655440001',
     '660e8400-e29b-41d4-a716-446655440005',
     'weapon', '770e8400-e29b-41d4-a716-446655440003',
     '/icons/weapons/dragon_slayer.png', true, false, 50, 5000, 10000),

    ('aa0e8400-e29b-41d4-a716-446655440009',
     'Wooden Axe', 'WOODEN_AXE_001',
     '550e8400-e29b-41d4-a716-446655440001',
     '660e8400-e29b-41d4-a716-446655440002',
     'weapon', '770e8400-e29b-41d4-a716-446655440004',
     '/icons/weapons/wooden_axe.png', true, true, 1, 15, 30),

    ('aa0e8400-e29b-41d4-a716-446655440010',
     'Elven Bow', 'ELVEN_BOW_001',
     '550e8400-e29b-41d4-a716-446655440001',
     '660e8400-e29b-41d4-a716-446655440003',
     'weapon', '770e8400-e29b-41d4-a716-446655440005',
     '/icons/weapons/elven_bow.png', true, true, 15, 150, 300),

    ('aa0e8400-e29b-41d4-a716-446655440011',
     'Shadow Blade', 'SHADOW_BLADE_001',
     '550e8400-e29b-41d4-a716-446655440001',
     '660e8400-e29b-41d4-a716-446655440004',
     'weapon', '770e8400-e29b-41d4-a716-446655440006',
     '/icons/weapons/shadow_blade.png', true, true, 25, 800, 1600),

    ('aa0e8400-e29b-41d4-a716-446655440012',
     'Iron Axe', 'IRON_AXE_001',
     '550e8400-e29b-41d4-a716-446655440001',
     '660e8400-e29b-41d4-a716-446655440002',
     'weapon', '770e8400-e29b-41d4-a716-446655440007',
     '/icons/weapons/iron_axe.png', true, true, 5, 30, 60),

    -- Armors (6)
    ('aa0e8400-e29b-41d4-a716-446655440004',
     'Leather Armor', 'LEATHER_ARMOR_001',
     '550e8400-e29b-41d4-a716-446655440002',
     '660e8400-e29b-41d4-a716-446655440001',
     'armor', '880e8400-e29b-41d4-a716-446655440001',
     '/icons/armors/leather_armor.png', true, true, 1, 15, 30),

    ('aa0e8400-e29b-41d4-a716-446655440005',
     'Mithril Plate', 'MITHRIL_PLATE_001',
     '550e8400-e29b-41d4-a716-446655440002',
     '660e8400-e29b-41d4-a716-446655440004',
     'armor', '880e8400-e29b-41d4-a716-446655440002',
     '/icons/armors/mithril_plate.png', true, true, 30, 2000, 4000),

    ('aa0e8400-e29b-41d4-a716-446655440013',
     'Wooden Helm', 'WOODEN_HELM_001',
     '550e8400-e29b-41d4-a716-446655440002',
     '660e8400-e29b-41d4-a716-446655440001',
     'armor', '880e8400-e29b-41d4-a716-446655440003',
     '/icons/armors/wooden_helm.png', true, true, 1, 10, 20),

    ('aa0e8400-e29b-41d4-a716-446655440014',
     'Chainmail Leggings', 'CHAINMAIL_LEGS_001',
     '550e8400-e29b-41d4-a716-446655440002',
     '660e8400-e29b-41d4-a716-446655440002',
     'armor', '880e8400-e29b-41d4-a716-446655440004',
     '/icons/armors/chainmail_legs.png', true, true, 10, 60, 120),

    ('aa0e8400-e29b-41d4-a716-446655440015',
     'Tower Shield', 'TOWER_SHIELD_001',
     '550e8400-e29b-41d4-a716-446655440002',
     '660e8400-e29b-41d4-a716-446655440003',
     'armor', '880e8400-e29b-41d4-a716-446655440005',
     '/icons/armors/tower_shield.png', true, true, 20, 300, 600),

    ('aa0e8400-e29b-41d4-a716-446655440016',
     'Divine Plate', 'DIVINE_PLATE_001',
     '550e8400-e29b-41d4-a716-446655440002',
     '660e8400-e29b-41d4-a716-446655440005',
     'armor', '880e8400-e29b-41d4-a716-446655440006',
     '/icons/armors/divine_plate.png', true, false, 45, 8000, 16000),

    -- Consumables (7)
    ('aa0e8400-e29b-41d4-a716-446655440006',
     'Small Health Potion', 'HEALTH_POT_SMALL',
     '550e8400-e29b-41d4-a716-446655440003',
     '660e8400-e29b-41d4-a716-446655440001',
     'consumable', '990e8400-e29b-41d4-a716-446655440001',
     '/icons/consumables/health_small.png', true, true, 1, 5, 10),

    ('aa0e8400-e29b-41d4-a716-446655440007',
     'Medium Health Potion', 'HEALTH_POT_MEDIUM',
     '550e8400-e29b-41d4-a716-446655440003',
     '660e8400-e29b-41d4-a716-446655440002',
     'consumable', '990e8400-e29b-41d4-a716-446655440002',
     '/icons/consumables/health_medium.png', true, true, 5, 20, 40),

    ('aa0e8400-e29b-41d4-a716-446655440008',
     'Small Mana Potion', 'MANA_POT_SMALL',
     '550e8400-e29b-41d4-a716-446655440003',
     '660e8400-e29b-41d4-a716-446655440001',
     'consumable', '990e8400-e29b-41d4-a716-446655440003',
     '/icons/consumables/mana_small.png', true, true, 1, 5, 10),

    ('aa0e8400-e29b-41d4-a716-446655440017',
     'Large Health Potion', 'HEALTH_POT_LARGE',
     '550e8400-e29b-41d4-a716-446655440003',
     '660e8400-e29b-41d4-a716-446655440003',
     'consumable', '990e8400-e29b-41d4-a716-446655440004',
     '/icons/consumables/health_large.png', true, true, 15, 50, 100),

    ('aa0e8400-e29b-41d4-a716-446655440018',
     'Medium Mana Potion', 'MANA_POT_MEDIUM',
     '550e8400-e29b-41d4-a716-446655440003',
     '660e8400-e29b-41d4-a716-446655440002',
     'consumable', '990e8400-e29b-41d4-a716-446655440005',
     '/icons/consumables/mana_medium.png', true, true, 10, 30, 60),

    ('aa0e8400-e29b-41d4-a716-446655440019',
     'Elixir of Restoration', 'ELIXIR_RESTORE_001',
     '550e8400-e29b-41d4-a716-446655440003',
     '660e8400-e29b-41d4-a716-446655440004',
     'consumable', '990e8400-e29b-41d4-a716-446655440006',
     '/icons/consumables/elixir_restore.png', true, true, 30, 500, 1000),

    ('aa0e8400-e29b-41d4-a716-446655440020',
     'Attack Buff Scroll', 'ATTACK_SCROLL_001',
     '550e8400-e29b-41d4-a716-446655440003',
     '660e8400-e29b-41d4-a716-446655440001',
     'consumable', '990e8400-e29b-41d4-a716-446655440007',
     '/icons/consumables/attack_scroll.png', true, true, 1, 8, 15);
