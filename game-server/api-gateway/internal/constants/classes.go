package constants

import (
	"github.com/darkphotonKN/cosmic-void-server/api-gateway/internal/class"
	"github.com/google/uuid"
)

/**
* Seed Constants for classes and ascendancies.
**/

// --- Classes ---

// -- base --

var DefaultClasses []class.CreateDefaultClass = []class.CreateDefaultClass{
	{
		ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Name:        "Warrior",
		Description: "Brutal monster wielding melee weapons.",
		ImageURL:    "Placeholder.",
	},
	{
		ID:          uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		Name:        "Sorceror",
		Description: "Master of elemental and arcane magic.",
		ImageURL:    "Placeholder.",
	},
	{
		ID:          uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		Name:        "Witch",
		Description: "Dark caster who deals in curses and chaos.",
		ImageURL:    "Placeholder.",
	},
	{
		ID:          uuid.MustParse("44444444-4444-4444-4444-444444444444"),
		Name:        "Monk",
		Description: "A disciplined fighter using martial arts.",
		ImageURL:    "Placeholder.",
	},
	{
		ID:          uuid.MustParse("55555555-5555-5555-5555-555555555555"),
		Name:        "Ranger",
		Description: "Skilled archer with a deep connection to nature.",
		ImageURL:    "Placeholder.",
	},
	{
		ID:          uuid.MustParse("66666666-6666-6666-6666-666666666666"),
		Name:        "Mercenary",
		Description: "Versatile fighter who masters various weapons.",
		ImageURL:    "Placeholder.",
	},
}

// -- ascendancy --
var DefaultAscendancies = []class.CreateDefaultAscendancy{

	// Warrior Ascendancies
	{ID: uuid.MustParse("11111111-1111-1111-1111-111111111112"), Name: "Titan", ImageURL: "Placeholder.", ClassID: uuid.MustParse("11111111-1111-1111-1111-111111111111")},
	{ID: uuid.MustParse("11111111-1111-1111-1111-111111111113"), Name: "Warbringer", ImageURL: "Placeholder.", ClassID: uuid.MustParse("11111111-1111-1111-1111-111111111111")},

	// Sorceror Ascendancies
	{ID: uuid.MustParse("22222222-2222-2222-2222-222222222223"), Name: "Stormweaver", ImageURL: "Placeholder.", ClassID: uuid.MustParse("22222222-2222-2222-2222-222222222222")},
	{ID: uuid.MustParse("22222222-2222-2222-2222-222222222224"), Name: "Chronomancer", ImageURL: "Placeholder.", ClassID: uuid.MustParse("22222222-2222-2222-2222-222222222222")},

	// Ranger Ascendancies
	{ID: uuid.MustParse("55555555-5555-5555-5555-555555555556"), Name: "Deadeye", ImageURL: "Placeholder.", ClassID: uuid.MustParse("55555555-5555-5555-5555-555555555555")},
	{ID: uuid.MustParse("55555555-5555-5555-5555-555555555557"), Name: "Pathfinder", ImageURL: "Placeholder.", ClassID: uuid.MustParse("55555555-5555-5555-5555-555555555555")},

	// Mercenary Ascendancies
	{ID: uuid.MustParse("66666666-6666-6666-6666-666666666667"), Name: "Gemling Legionnaire", ImageURL: "Placeholder.", ClassID: uuid.MustParse("66666666-6666-6666-6666-666666666666")},
	{ID: uuid.MustParse("66666666-6666-6666-6666-666666666668"), Name: "Witchhunter", ImageURL: "Placeholder.", ClassID: uuid.MustParse("66666666-6666-6666-6666-666666666666")},

	// Monk Ascendancies
	{ID: uuid.MustParse("44444444-4444-4444-4444-444444444445"), Name: "Invoker", ImageURL: "Placeholder.", ClassID: uuid.MustParse("44444444-4444-4444-4444-444444444444")},
	{ID: uuid.MustParse("44444444-4444-4444-4444-444444444446"), Name: "Acolyte of Chayula", ImageURL: "Placeholder.", ClassID: uuid.MustParse("44444444-4444-4444-4444-444444444444")},

	// Witch Ascendancies
	{ID: uuid.MustParse("33333333-3333-3333-3333-333333333334"), Name: "Blood Mage", ImageURL: "Placeholder.", ClassID: uuid.MustParse("33333333-3333-3333-3333-333333333333")},
	{ID: uuid.MustParse("33333333-3333-3333-3333-333333333335"), Name: "Infernalist", ImageURL: "Placeholder.", ClassID: uuid.MustParse("33333333-3333-3333-3333-333333333333")},
}
