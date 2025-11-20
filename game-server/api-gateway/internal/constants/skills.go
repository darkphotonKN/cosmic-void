package constants

import (
	"github.com/darkphotonKN/cosmic-void-server/api-gateway/internal/skill"
	"github.com/google/uuid"
)

/**
* Seed Constants for skills.
**/

// --- Skills ---

// -- Active Skills --
var ActiveSkills []skill.SeedSkill = []skill.SeedSkill{
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), Name: "Absolution", Type: "active"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000002"), Name: "Ancestral Protector", Type: "active"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000003"), Name: "Ancestral Warchief", Type: "active"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000004"), Name: "Animate Guardian", Type: "active"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000005"), Name: "Arc", Type: "active"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000006"), Name: "Arctic Armour", Type: "active"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000007"), Name: "Ball Lightning", Type: "active"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000008"), Name: "Barrage", Type: "active"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000009"), Name: "Bear Trap", Type: "active"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-00000000000A"), Name: "Blade Flurry", Type: "active"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-00000000000B"), Name: "Blade Vortex", Type: "active"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-00000000000C"), Name: "Bladefall", Type: "active"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-00000000000D"), Name: "Bladestorm", Type: "active"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-00000000000E"), Name: "Blast Rain", Type: "active"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-00000000000F"), Name: "Blazing Salvo", Type: "active"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000010"), Name: "Blink Arrow", Type: "active"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000011"), Name: "Blood Rage", Type: "active"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000012"), Name: "Burning Arrow", Type: "active"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000013"), Name: "Caustic Arrow", Type: "active"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000014"), Name: "Charged Dash", Type: "active"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000015"), Name: "Cobra Lash", Type: "active"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000016"), Name: "Cold Snap", Type: "active"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000017"), Name: "Cyclone", Type: "active"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000018"), Name: "Detonate Dead", Type: "active"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000019"), Name: "Double Strike", Type: "active"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-00000000001A"), Name: "Dual Strike", Type: "active"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-00000000001B"), Name: "Earthquake", Type: "active"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-00000000001C"), Name: "Elemental Hit", Type: "active"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-00000000001D"), Name: "Ethereal Knives", Type: "active"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-00000000001E"), Name: "Leap Slam", Type: "active"},
}

// -- Support Skills --
var SupportSkills []skill.SeedSkill = []skill.SeedSkill{
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000101"), Name: "Added Chaos Damage", Type: "support"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000102"), Name: "Added Cold Damage", Type: "support"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000103"), Name: "Added Fire Damage", Type: "support"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000104"), Name: "Added Lightning Damage", Type: "support"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000105"), Name: "Ancestral Call", Type: "support"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000106"), Name: "Arcane Surge", Type: "support"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000107"), Name: "Ballista Totem", Type: "support"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000108"), Name: "Barrage", Type: "support"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000109"), Name: "Blind", Type: "support"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-00000000010A"), Name: "Brutality", Type: "support"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-00000000010B"), Name: "Burning Damage", Type: "support"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-00000000010C"), Name: "Cast on Critical Strike", Type: "support"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-00000000010D"), Name: "Cast on Death", Type: "support"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-00000000010E"), Name: "Cast on Melee Kill", Type: "support"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-00000000010F"), Name: "Chain", Type: "support"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000110"), Name: "Chance to Flee", Type: "support"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000111"), Name: "Chance to Poison", Type: "support"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000112"), Name: "Chaos Damage", Type: "support"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000113"), Name: "Close Combat", Type: "support"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000114"), Name: "Cold to Fire", Type: "support"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000115"), Name: "Concentrated Effect", Type: "support"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000116"), Name: "Multistrike", Type: "support"},
	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000117"), Name: "Faster Attacks", Type: "support"},

	{ID: uuid.MustParse("00000000-0000-0000-0000-000000000118"), Name: "Sunder", Type: "active"},
}
