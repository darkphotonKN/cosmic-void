package validation

import (
	"github.com/darkphotonKN/cosmic-void-server/api-gateway/internal/gateway/item"
	"github.com/darkphotonKN/cosmic-void-server/api-gateway/internal/rating"
	"github.com/darkphotonKN/cosmic-void-server/api-gateway/internal/skill"
	"github.com/go-playground/validator/v10"
)

// RegisterValidators registers custom validators for categories, classes, and types
func RegisterValidators(v *validator.Validate) {
	// --- Item ---
	v.RegisterValidation("category", func(fl validator.FieldLevel) bool {
		// --- Item ---
		return item.IsValidCategory(fl.Field().String())
	})

	v.RegisterValidation("class", func(fl validator.FieldLevel) bool {
		return item.IsValidClass(fl.Field().String())
	})

	v.RegisterValidation("type", func(fl validator.FieldLevel) bool {
		return item.IsValidType(fl.Field().String())
	})

	// --- Skill ---
	v.RegisterValidation("skillType",
		func(fl validator.FieldLevel) bool {
			return skill.IsValidType(fl.Field().String())
		},
	)

	// --- Rating ---
	v.RegisterValidation("ratingCategory",
		func(fl validator.FieldLevel) bool {
			return rating.IsValidCategoryType(fl.Field().String())
		},
	)

}
