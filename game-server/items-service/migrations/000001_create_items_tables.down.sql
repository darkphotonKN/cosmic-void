DROP TRIGGER IF EXISTS item_templates_updated_at ON item_templates;
DROP TRIGGER IF EXISTS consumables_updated_at ON consumables;
DROP TRIGGER IF EXISTS armors_updated_at ON armors;
DROP TRIGGER IF EXISTS weapons_updated_at ON weapons;
DROP TRIGGER IF EXISTS item_rarities_updated_at ON item_rarities;
DROP TRIGGER IF EXISTS item_types_updated_at ON item_types;

DROP FUNCTION IF EXISTS update_updated_at_column();

-- item_templates 索引
DROP INDEX IF EXISTS idx_item_templates_itemable;
DROP INDEX IF EXISTS idx_item_templates_code;
DROP INDEX IF EXISTS idx_item_templates_rarity;
DROP INDEX IF EXISTS idx_item_templates_type;

-- consumables 索引
DROP INDEX IF EXISTS idx_consumables_rarity;
DROP INDEX IF EXISTS idx_consumables_type;

-- armors 索引
DROP INDEX IF EXISTS idx_armors_rarity;
DROP INDEX IF EXISTS idx_armors_type;

-- weapons 索引
DROP INDEX IF EXISTS idx_weapons_rarity;
DROP INDEX IF EXISTS idx_weapons_type;

-- item_rarities 索引
DROP INDEX IF EXISTS idx_item_rarities_sort;

-- item_types 索引
DROP INDEX IF EXISTS idx_item_types_sort;
DROP INDEX IF EXISTS idx_item_types_active;

-- 先刪除有外鍵依賴的表
DROP TABLE IF EXISTS item_templates;
DROP TABLE IF EXISTS consumables;
DROP TABLE IF EXISTS armors;
DROP TABLE IF EXISTS weapons;

-- 最後刪除被依賴的表
DROP TABLE IF EXISTS item_rarities;
DROP TABLE IF EXISTS item_types;