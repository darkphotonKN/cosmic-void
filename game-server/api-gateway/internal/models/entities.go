package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

/**
* Types here are shared model entities that are imported by more than one package.
**/

/**
* Member
**/
type Member struct {
	BaseDBDateModel
	Email         string  `db:"email" json:"email"`
	Name          string  `db:"name" json:"name"`
	Password      string  `db:"password" json:"password,omitempty"`
	Status        string  `db:"status" json:"status"`
	AverageRating float64 `db:"average_rating"`
}

/**
* Class
**/
type Class struct {
	BaseDBDateModel
	Name        string `db:"name" json:"name"`
	Description string `db:"description" json:"description"`
	ImageURL    string `db:"image_url" json:"imageUrl"`
}

/**
* Ascendancy
**/
type Ascendancy struct {
	BaseDBDateModel
	ClassID     uuid.UUID `db:"class_id" json:"classId"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	ImageURL    string    `db:"image_url" json:"imageUrl"`
}

/**
* Items
**/
type Item struct {
	BaseDBDateModel
	MemberID    uuid.UUID `json:"memberId" db:"member_id"`
	BaseItemId  uuid.UUID `json:"baseItemId,omitempty" db:"base_item_id"`
	ImageUrl    string    `json:"imageUrl" db:"image_url"`
	Category    string    `json:"category" db:"category"`
	Class       string    `json:"class" db:"class"`
	Name        string    `json:"name" db:"name"`
	Type        string    `json:"type" db:"type"`
	Description string    `json:"description" db:"description"`
	UniqueItem  bool      `json:"uniqueItem" db:"unique_item"`
	Slot        string    `json:"slot" db:"slot"`
	// armor
	RequiredLevel        string `json:"requiredLevel,omitempty" db:"required_level"`
	RequiredStrength     string `json:"requiredStrength,omitempty" db:"required_strength"`
	RequiredDexterity    string `json:"requiredDexterity,omitempty" db:"required_dexterity"`
	RequiredIntelligence string `json:"requiredIntelligence,omitempty" db:"required_intelligence"`
	Armour               string `json:"armour,omitempty" db:"armour"`
	EnergyShield         string `json:"energyShield,omitempty" db:"energy_shield"`
	Evasion              string `json:"evasion,omitempty" db:"evasion"`
	Block                string `json:"block,omitempty" db:"block"`
	Ward                 string `json:"ward,omitempty" db:"ward"`
	// weapon
	Damage string `json:"damage,omitempty" db:"damage"`
	APS    string `json:"aps,omitempty" db:"aps"`
	Crit   string `json:"crit,omitempty" db:"crit"`
	PDPS   string `json:"pdps,omitempty" db:"pdps"`
	EDPS   string `json:"edps,omitempty" db:"edps"`
	DPS    string `json:"dps,omitempty" db:"dps"`
	// poison
	Life     string `json:"life,omitempty" db:"life"`
	Mana     string `json:"mana,omitempty" db:"mana"`
	Duration string `json:"duration,omitempty" db:"duration"`
	Usage    string `json:"usage,omitempty" db:"usage"`
	Capacity string `json:"capacity,omitempty" db:"capacity"`
	// common
	Additional string         `json:"additional,omitempty" db:"additional"`
	Stats      pq.StringArray `json:"stats" db:"stats"`
	Implicit   pq.StringArray `json:"implicit" db:"implicit"`
}

/**
* Base Items
**/
type BaseItem struct {
	BaseDBDateModel
	ImageUrl   string `json:"imageUrl" db:"image_url"`
	Category   string `json:"category" db:"category"`
	Class      string `json:"class" db:"class"`
	Name       string `json:"name" db:"name"`
	Type       string `json:"type" db:"type"`
	EquipType  string `json:"equipType" db:"equip_type"`
	IsTwoHands bool   `json:"isTwoHands" db:"is_two_hands"`
	Slot       string `json:"slot" db:"slot"`

	RequiredLevel        string `json:"requiredLevel,omitempty" db:"required_level"`
	RequiredStrength     string `json:"requiredStrength,omitempty" db:"required_strength"`
	RequiredDexterity    string `json:"requiredDexterity,omitempty" db:"required_dexterity"`
	RequiredIntelligence string `json:"requiredIntelligence,omitempty" db:"required_intelligence"`

	Damage string `json:"damage,omitempty" db:"damage"`
	APS    string `json:"aps,omitempty" db:"aps"`
	Crit   string `json:"crit,omitempty" db:"crit"`
	DPS    string `json:"dps,omitempty" db:"dps"`

	Armour       string `json:"armour,omitempty" db:"armour"`
	Evasion      string `json:"evasion,omitempty" db:"evasion"`
	EnergyShield string `json:"energyShield,omitempty" db:"energy_shield"`
	Ward         string `json:"ward,omitempty" db:"ward"`

	Implicit pq.StringArray `json:"implicit" db:"implicit"`
}

/**
* Item Mods
**/
type ItemMod struct {
	BaseDBDateModel
	Affix string `json:"affix" db:"affix"`
	Name  string `json:"name" db:"name"`
	Level string `json:"level" db:"level"`
	Stat  string `json:"stat" db:"stat"`
	Tags  string `json:"tags" db:"tags"`
}

/**
* Skills
**/
type Skill struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Type      string    `db:"type" json:"type"` // "active" or "support"
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}

/**
* Ratings
**/
type Rating struct {
	BaseIDModel
	BuildID  uuid.UUID `db:"build_id" json:"buildId"`
	MemberID uuid.UUID `db:"member_id" json:"memberId"`
	Category string    `db:"category" json:"category"`
	Value    int       `db:"value" json:"value"`
}

/**
* Build
**/
type Build struct {
	BaseDBDateModel
	MemberId           uuid.UUID `db:"member_id" json:"memberId"`
	MainSkillId        uuid.UUID `db:"main_skill_id" json:"mainSkill"`
	ClassId            uuid.UUID `db:"class_id" json:"classId"`
	AscendancyId       uuid.UUID `db:"ascendancy_id" json:"ascendancyId"`
	Title              string    `db:"title" json:"title"`
	Description        string    `db:"description" json:"description"`
	AvgEndGameRating   *float32  `db:"avg_end_game_rating" json:"avgEndGameRating,omitempty"`
	AvgFunRating       *float32  `db:"avg_fun_rating" json:"avgFunRating,omitempty"`
	AvgCreativeRating  *float32  `db:"avg_creative_rating" json:"avgCreativeRating,omitempty"`
	AvgSpeedFarmRating *float32  `db:"avg_speed_farm_rating" json:"avgSpeedFarmRating,omitempty"`
	AvgBossingRating   *float32  `db:"avg_bossing_rating" json:"avgBossingRating,omitempty"`
	Views              int       `db:"views" json:"views"`
	Status             int       `db:"status" json:"status"` // 0: Edit, 1: Published, 2: Archived
}

type BuildItem struct {
	ID      uuid.UUID `db:"id" json:"id"`
	BuildID uuid.UUID `db:"build_id" json:"buildId"`
	ItemID  uuid.UUID `db:"item_id" json:"itemId"`
	Slot    string    `db:"slot" json:"slot"`
}

type BuildSkill struct {
	ID      uuid.UUID `db:"id" json:"id"`
	BuildID uuid.UUID `db:"build_id" json:"buildId"`
	SkillID uuid.UUID `db:"skill_id" json:"skillId"`
}

/**
* Items
**/
type Tag struct {
	BaseDBDateModel
	Name string `db:"name" json:"name"`
}

/**
* Items
**/
type Article struct {
	BaseDBDateModel
	Content string `db:"content" json:"content"`
}

/**
* Base models for default table columns.
**/

type BaseIDModel struct {
	ID        uuid.UUID `db:"id" json:"id"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

type BaseDBMemberModel struct {
	Id            uuid.UUID `db:"id" json:"id"`
	UpdatedMember uuid.UUID `db:"updated_member" json:"updatedMember"`
	CreatedMember uuid.UUID `db:"created_member" json:"createdMember"`
}

type BaseDBDateModel struct {
	Id        uuid.UUID `db:"id" json:"id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type BaseDBMemberDateModel struct {
	Id            uuid.UUID `db:"id" json:"id"`
	UpdatedMember uuid.UUID `db:"updated_member" json:"updatedMember"`
	CreatedMember uuid.UUID `db:"created_member" json:"createdMember"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

/**
* Other helper models shared between two or more packages.
**/

// holds temporary data for joining skills skill links and builds tables.
type SkillRow struct {
	SkillLinkId     string    `db:"skill_link_id"`
	SkillLinkName   string    `db:"skill_link_name"`
	SkillLinkIsMain bool      `db:"skill_link_is_main"`
	SkillId         uuid.UUID `db:"skill_id"`
	SkillName       string    `db:"skill_name"`
	SkillType       string    `db:"skill_type"`
}

type BuildItemSetItem struct {
	ID             uuid.UUID `db:"id" json:"id"`
	BuildItemSetId uuid.UUID `db:"build_item_set_id" json:"buildItemSetId"`
	ItemId         uuid.UUID `db:"item_id" json:"itemId"`
	Slot           string    `db:"slot" json:"slot"`
}
