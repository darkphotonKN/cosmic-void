-- Migration 009: Reseed with futuristic space RPG base items + clean rarities
--
-- Armor types (4 material classes × 4 slots = 16):
--   Titanium Alloy  – standard mil-spec plating, balanced stats
--   Nanofiber        – lightweight woven nano-material, low weight / fast
--   Plasma-forged    – energy-hardened exotic plating, highest raw defense
--   Synth-weave      – bio-synthetic adaptive mesh, high energy resistance
--
-- Weapon types (6 melee archetypes):
--   Vibro-blade, Pulse Dagger, Gravity Maul, Photon Lance, Mag-Cleaver, Shock Gauntlet
--
-- Consumables (2 healing stims)

-- Step 1: Clear all existing item + rarity data
TRUNCATE item_templates, weapons, armors, consumables CASCADE;
TRUNCATE item_rarities CASCADE;

-- Step 2: Reseed 4-tier rarities (Common / Rare / Epic / Legendary)
INSERT INTO item_rarities (id, rarity_code, rarity_name, color_hex, drop_rate_multiplier, sort_order) VALUES
    ('660e8400-e29b-41d4-a716-446655440001', 'common',    'Common',    '#B0B0B0', 1.00, 1),
    ('660e8400-e29b-41d4-a716-446655440003', 'rare',      'Rare',      '#3498DB', 0.40, 2),
    ('660e8400-e29b-41d4-a716-446655440004', 'epic',      'Epic',      '#9B59B6', 0.15, 3),
    ('660e8400-e29b-41d4-a716-446655440005', 'legendary', 'Legendary', '#F39C12', 0.05, 4);

-- Step 3: Insert 6 melee weapons (all common rarity)
--   Archetypes drawn from sci-fi RPG staples (Warframe, Destiny, Cyberpunk, Mass Effect)
INSERT INTO weapons (id, rarity_id, attack_power, durability, critical_rate, weapon_type, description) VALUES
    ('77000000-0000-0000-0000-000000000001', '660e8400-e29b-41d4-a716-446655440001',  6, 100, 0.08, 'sword',  'Mono-molecular vibrating combat blade'),
    ('77000000-0000-0000-0000-000000000002', '660e8400-e29b-41d4-a716-446655440001',  3, 100, 0.12, 'knife',  'Compact energy-pulse combat knife'),
    ('77000000-0000-0000-0000-000000000003', '660e8400-e29b-41d4-a716-446655440001',  9, 100, 0.02, 'mace',   'Graviton-charged heavy war hammer'),
    ('77000000-0000-0000-0000-000000000004', '660e8400-e29b-41d4-a716-446655440001',  5, 100, 0.06, 'spear',  'Focused hard-light polearm'),
    ('77000000-0000-0000-0000-000000000005', '660e8400-e29b-41d4-a716-446655440001',  7, 100, 0.04, 'axe',    'Magnetically accelerated cleaving axe'),
    ('77000000-0000-0000-0000-000000000006', '660e8400-e29b-41d4-a716-446655440001',  4, 100, 0.10, 'fist',   'Electrified close-quarters combat gauntlet');

-- Step 4: Insert 16 armors (4 material types × 4 slots)
--   Slots: head, chest, legs, gloves
INSERT INTO armors (id, rarity_id, defense_rating, durability, magic_resistance, armor_slot, description) VALUES
    -- Titanium Alloy (mil-spec balanced plating)
    ('88000000-0000-0000-0000-000000000001', '660e8400-e29b-41d4-a716-446655440001', 3, 100, 1, 'head',   'Standard-issue titanium alloy combat helmet'),
    ('88000000-0000-0000-0000-000000000002', '660e8400-e29b-41d4-a716-446655440001', 6, 100, 2, 'chest',  'Titanium alloy chest plate with ballistic lining'),
    ('88000000-0000-0000-0000-000000000003', '660e8400-e29b-41d4-a716-446655440001', 4, 100, 1, 'legs',   'Reinforced titanium alloy leg guards'),
    ('88000000-0000-0000-0000-000000000004', '660e8400-e29b-41d4-a716-446655440001', 2, 100, 1, 'gloves', 'Articulated titanium alloy combat gauntlets'),
    -- Nanofiber (lightweight woven nano-material)
    ('88000000-0000-0000-0000-000000000005', '660e8400-e29b-41d4-a716-446655440001', 2, 100, 1, 'head',   'Low-profile nanofiber tactical hood'),
    ('88000000-0000-0000-0000-000000000006', '660e8400-e29b-41d4-a716-446655440001', 4, 100, 1, 'chest',  'Flexible nanofiber weave vest'),
    ('88000000-0000-0000-0000-000000000007', '660e8400-e29b-41d4-a716-446655440001', 3, 100, 1, 'legs',   'Nanofiber mesh leggings with joint padding'),
    ('88000000-0000-0000-0000-000000000008', '660e8400-e29b-41d4-a716-446655440001', 1, 100, 0, 'gloves', 'Thin nanofiber utility gloves'),
    -- Plasma-forged (energy-hardened exotic plating)
    ('88000000-0000-0000-0000-000000000009', '660e8400-e29b-41d4-a716-446655440001', 5, 100, 2, 'head',   'Plasma-forged full-face battle visor'),
    ('88000000-0000-0000-0000-000000000010', '660e8400-e29b-41d4-a716-446655440001', 8, 100, 3, 'chest',  'Plasma-forged heavy assault cuirass'),
    ('88000000-0000-0000-0000-000000000011', '660e8400-e29b-41d4-a716-446655440001', 6, 100, 2, 'legs',   'Plasma-forged reinforced greaves'),
    ('88000000-0000-0000-0000-000000000012', '660e8400-e29b-41d4-a716-446655440001', 3, 100, 2, 'gloves', 'Plasma-forged armored fist plates'),
    -- Synth-weave (bio-synthetic adaptive mesh, high energy resistance)
    ('88000000-0000-0000-0000-000000000013', '660e8400-e29b-41d4-a716-446655440001', 2, 100, 3, 'head',   'Synth-weave neural cowl with dampening field'),
    ('88000000-0000-0000-0000-000000000014', '660e8400-e29b-41d4-a716-446655440001', 4, 100, 5, 'chest',  'Synth-weave adaptive torso jacket'),
    ('88000000-0000-0000-0000-000000000015', '660e8400-e29b-41d4-a716-446655440001', 3, 100, 4, 'legs',   'Synth-weave bio-mesh leg wraps'),
    ('88000000-0000-0000-0000-000000000016', '660e8400-e29b-41d4-a716-446655440001', 1, 100, 2, 'gloves', 'Synth-weave reactive grip wraps');

-- Step 5: Insert 2 consumables (heal only, all common)
INSERT INTO consumables (id, rarity_id, healing_amount, mana_amount, buff_duration, max_stack_size, description) VALUES
    ('99000000-0000-0000-0000-000000000001', '660e8400-e29b-41d4-a716-446655440001', 10, 0, 0, 20, 'Basic nano-med stim injection'),
    ('99000000-0000-0000-0000-000000000002', '660e8400-e29b-41d4-a716-446655440001', 25, 0, 0, 10, 'Advanced regenerative stim pack');

-- Step 6: Insert 24 item_templates (6 weapons + 16 armors + 2 consumables)
INSERT INTO item_templates (id, item_name, rarity_id, item_type, item_id, icon_url, required_level, base_sell_price, base_buy_price) VALUES
    -- Weapons
    ('aa000000-0000-0000-0000-000000000001', 'Vibro-blade',      '660e8400-e29b-41d4-a716-446655440001', 'weapon', '77000000-0000-0000-0000-000000000001', '/icons/weapon/vibro_blade.png',      1, 2, 5),
    ('aa000000-0000-0000-0000-000000000002', 'Pulse Dagger',      '660e8400-e29b-41d4-a716-446655440001', 'weapon', '77000000-0000-0000-0000-000000000002', '/icons/weapon/pulse_dagger.png',     1, 1, 3),
    ('aa000000-0000-0000-0000-000000000003', 'Gravity Maul',      '660e8400-e29b-41d4-a716-446655440001', 'weapon', '77000000-0000-0000-0000-000000000003', '/icons/weapon/gravity_maul.png',     1, 3, 7),
    ('aa000000-0000-0000-0000-000000000004', 'Photon Lance',      '660e8400-e29b-41d4-a716-446655440001', 'weapon', '77000000-0000-0000-0000-000000000004', '/icons/weapon/photon_lance.png',     1, 2, 5),
    ('aa000000-0000-0000-0000-000000000005', 'Mag-Cleaver',       '660e8400-e29b-41d4-a716-446655440001', 'weapon', '77000000-0000-0000-0000-000000000005', '/icons/weapon/mag_cleaver.png',      1, 3, 6),
    ('aa000000-0000-0000-0000-000000000006', 'Shock Gauntlet',    '660e8400-e29b-41d4-a716-446655440001', 'weapon', '77000000-0000-0000-0000-000000000006', '/icons/weapon/shock_gauntlet.png',   1, 2, 4),
    -- Titanium Alloy armor set
    ('aa000000-0000-0000-0000-000000000007', 'Titanium Helmet',        '660e8400-e29b-41d4-a716-446655440001', 'armor', '88000000-0000-0000-0000-000000000001', '/icons/armor/titanium_helmet.png',        1, 1, 4),
    ('aa000000-0000-0000-0000-000000000008', 'Titanium Chest Plate',   '660e8400-e29b-41d4-a716-446655440001', 'armor', '88000000-0000-0000-0000-000000000002', '/icons/armor/titanium_chest_plate.png',   1, 3, 6),
    ('aa000000-0000-0000-0000-000000000009', 'Titanium Leg Guards',    '660e8400-e29b-41d4-a716-446655440001', 'armor', '88000000-0000-0000-0000-000000000003', '/icons/armor/titanium_leg_guards.png',    1, 2, 5),
    ('aa000000-0000-0000-0000-000000000010', 'Titanium Gauntlets',     '660e8400-e29b-41d4-a716-446655440001', 'armor', '88000000-0000-0000-0000-000000000004', '/icons/armor/titanium_gauntlets.png',     1, 1, 3),
    -- Nanofiber armor set
    ('aa000000-0000-0000-0000-000000000011', 'Nanofiber Hood',         '660e8400-e29b-41d4-a716-446655440001', 'armor', '88000000-0000-0000-0000-000000000005', '/icons/armor/nanofiber_hood.png',         1, 1, 3),
    ('aa000000-0000-0000-0000-000000000012', 'Nanofiber Vest',         '660e8400-e29b-41d4-a716-446655440001', 'armor', '88000000-0000-0000-0000-000000000006', '/icons/armor/nanofiber_vest.png',         1, 2, 5),
    ('aa000000-0000-0000-0000-000000000013', 'Nanofiber Leggings',     '660e8400-e29b-41d4-a716-446655440001', 'armor', '88000000-0000-0000-0000-000000000007', '/icons/armor/nanofiber_leggings.png',     1, 1, 4),
    ('aa000000-0000-0000-0000-000000000014', 'Nanofiber Gloves',       '660e8400-e29b-41d4-a716-446655440001', 'armor', '88000000-0000-0000-0000-000000000008', '/icons/armor/nanofiber_gloves.png',       1, 1, 2),
    -- Plasma-forged armor set
    ('aa000000-0000-0000-0000-000000000015', 'Plasma-forged Visor',    '660e8400-e29b-41d4-a716-446655440001', 'armor', '88000000-0000-0000-0000-000000000009', '/icons/armor/plasma_forged_visor.png',    1, 2, 5),
    ('aa000000-0000-0000-0000-000000000016', 'Plasma-forged Cuirass',  '660e8400-e29b-41d4-a716-446655440001', 'armor', '88000000-0000-0000-0000-000000000010', '/icons/armor/plasma_forged_cuirass.png',  1, 3, 8),
    ('aa000000-0000-0000-0000-000000000017', 'Plasma-forged Greaves',  '660e8400-e29b-41d4-a716-446655440001', 'armor', '88000000-0000-0000-0000-000000000011', '/icons/armor/plasma_forged_greaves.png',  1, 3, 6),
    ('aa000000-0000-0000-0000-000000000018', 'Plasma-forged Fists',    '660e8400-e29b-41d4-a716-446655440001', 'armor', '88000000-0000-0000-0000-000000000012', '/icons/armor/plasma_forged_fists.png',    1, 2, 4),
    -- Synth-weave armor set
    ('aa000000-0000-0000-0000-000000000019', 'Synth-weave Cowl',       '660e8400-e29b-41d4-a716-446655440001', 'armor', '88000000-0000-0000-0000-000000000013', '/icons/armor/synth_weave_cowl.png',       1, 1, 3),
    ('aa000000-0000-0000-0000-000000000020', 'Synth-weave Jacket',     '660e8400-e29b-41d4-a716-446655440001', 'armor', '88000000-0000-0000-0000-000000000014', '/icons/armor/synth_weave_jacket.png',     1, 2, 5),
    ('aa000000-0000-0000-0000-000000000021', 'Synth-weave Leg Wraps',  '660e8400-e29b-41d4-a716-446655440001', 'armor', '88000000-0000-0000-0000-000000000015', '/icons/armor/synth_weave_leg_wraps.png',  1, 1, 4),
    ('aa000000-0000-0000-0000-000000000022', 'Synth-weave Grip Wraps', '660e8400-e29b-41d4-a716-446655440001', 'armor', '88000000-0000-0000-0000-000000000016', '/icons/armor/synth_weave_grip_wraps.png', 1, 1, 2),
    -- Consumables
    ('aa000000-0000-0000-0000-000000000023', 'Minor Stim Pack',        '660e8400-e29b-41d4-a716-446655440001', 'consumable', '99000000-0000-0000-0000-000000000001', '/icons/consumable/minor_stim_pack.png',   1, 1, 2),
    ('aa000000-0000-0000-0000-000000000024', 'Greater Stim Pack',      '660e8400-e29b-41d4-a716-446655440001', 'consumable', '99000000-0000-0000-0000-000000000002', '/icons/consumable/greater_stim_pack.png', 1, 2, 5);
