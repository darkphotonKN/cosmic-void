package constants

import (
	"github.com/darkphotonKN/cosmic-void-server/api-gateway/internal/models"
	"github.com/google/uuid"
)

var DefaultTags []models.Tag = []models.Tag{
	models.Tag{
		BaseDBDateModel: models.BaseDBDateModel{
			Id: uuid.MustParse("10000000-0000-0000-0000-000000000000"),
		},
		Name: "Leveling",
	},
	models.Tag{
		BaseDBDateModel: models.BaseDBDateModel{
			Id: uuid.MustParse("20000000-0000-0000-0000-000000000000"),
		},
		Name: "Bossing",
	},
	models.Tag{
		BaseDBDateModel: models.BaseDBDateModel{
			Id: uuid.MustParse("30000000-0000-0000-0000-000000000000"),
		},
		Name: "Endgame",
	},
	models.Tag{
		BaseDBDateModel: models.BaseDBDateModel{
			Id: uuid.MustParse("40000000-0000-0000-0000-000000000000"),
		},
		Name: "Speedfarm",
	},
}
