-- Migration 009 down: Clear reseeded items and rarities (cannot restore old data)
TRUNCATE item_templates, weapons, armors, consumables CASCADE;
TRUNCATE item_rarities CASCADE;

-- Restore original 5 rarities from migration 002
INSERT INTO item_rarities (id, rarity_code, rarity_name, color_hex, drop_rate_multiplier, sort_order) VALUES
    ('660e8400-e29b-41d4-a716-446655440001', 'common',    'Common',    '#FFFFFF', 1.00, 1),
    ('660e8400-e29b-41d4-a716-446655440002', 'uncommon',  'Uncommon',  '#00FF00', 0.75, 2),
    ('660e8400-e29b-41d4-a716-446655440003', 'rare',      'Rare',      '#0000FF', 0.50, 3),
    ('660e8400-e29b-41d4-a716-446655440004', 'epic',      'Epic',      '#9400D3', 0.25, 4),
    ('660e8400-e29b-41d4-a716-446655440005', 'legendary', 'Legendary', '#FFD700', 0.10, 5);
