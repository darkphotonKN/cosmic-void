package constants

import (
	"github.com/darkphotonKN/cosmic-void-server/api-gateway/internal/member"
	"github.com/google/uuid"
)

/**
* Seed Constants for members.
**/

// --- Default Member ---

var DefaultMembers []member.CreateDefaultMember = []member.CreateDefaultMember{
	{
		ID:       uuid.MustParse("6f60f94a-6c90-45a1-96f6-32174cc0f908"),
		Email:    "communitybuildsmoderator@gmail.com",
		Name:     "Community Builds Moderator",
		Password: "QWE@asd123",
		Status:   1},
}
