package member

import (
	"github.com/google/uuid"
)

type MemberUpdatePasswordParams struct {
	ID       uuid.UUID `db:"id" json:"id"`
	Password string    `db:"password" json:"password"`
}

type MemberUpdateInfoParams struct {
	ID     uuid.UUID `db:"id" json:"id"`
	Name   string    `db:"name" json:"name"`
	Status string    `db:"status" json:"status"`
}

type CreateDefaultMember struct {
	ID       uuid.UUID `db:"id"`
	Email    string    `db:"email"`
	Name     string    `db:"name"`
	Password string    `db:"password"`
	Status   int       `db:"status"`
}
