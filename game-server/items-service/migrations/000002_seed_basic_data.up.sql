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

-- Insert Sample Weapons
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
     100, 200, 0.30, 'sword', 'Legendary dragon slayer sword');

-- Insert Sample Armors
INSERT INTO armors (id, type_id, rarity_id, defense_rating, durability, magic_resistance, armor_slot, description)
VALUES
    ('880e8400-e29b-41d4-a716-446655440001',
     '550e8400-e29b-41d4-a716-446655440002',
     '660e8400-e29b-41d4-a716-446655440001',
     20, 100, 10, 'chest', 'Basic leather armor'),

    ('880e8400-e29b-41d4-a716-446655440002',
     '550e8400-e29b-41d4-a716-446655440002',
     '660e8400-e29b-41d4-a716-446655440004',
     80, 180, 50, 'chest', 'Epic mithril plate armor');

-- Insert Sample Consumables
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
     0, 50, 0, 99, 'Small mana potion');

-- Insert Sample Item Templates
INSERT INTO item_templates (id, item_name, item_code, type_id, rarity_id, item_type, item_id, icon_url, is_tradeable, is_droppable, required_level, base_sell_price, base_buy_price)
VALUES
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
     '/icons/consumables/mana_small.png', true, true, 1, 5, 10);
