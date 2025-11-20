package member

import (
	"fmt"

	"github.com/darkphotonKN/cosmic-void-server/auth-service/internal/models"
	commonhelpers "github.com/darkphotonKN/cosmic-void-server/common/utils"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	DB *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		DB: db,
	}
}

func (r *Repository) Create(name, email, password string) (uuid.UUID, error) {
	memberId := uuid.New()
	query := `INSERT INTO members (id, name, email, password) VALUES ($1, $2, $3, $4)`

	_, err := r.DB.Exec(query, memberId, name, email, password)
	if err != nil {
		fmt.Println("Error when creating member:", err)
		return uuid.Nil, commonhelpers.AnalyzeDBErr(err)
	}

	return memberId, nil
}

func (r *Repository) UpdatePassword(params MemberUpdatePasswordParams) error {
	query := `UPDATE members SET password = :password WHERE id = :id`

	result, err := r.DB.NamedExec(query, params)
	if err != nil {
		return commonhelpers.AnalyzeDBErr(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no member found with id: %v", params.ID)
	}

	return nil
}

func (r *Repository) UpdateMemberInfo(id uuid.UUID, name, status string) error {
	params := MemberUpdateInfoParams{
		ID:     id,
		Name:   name,
		Status: status,
	}

	query := `UPDATE members SET name = :name, status = :status WHERE id = :id`

	result, err := r.DB.NamedExec(query, params)
	if err != nil {
		return commonhelpers.AnalyzeDBErr(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no member found with id: %v", params.ID)
	}

	return nil
}

func (r *Repository) GetByIdWithPassword(id uuid.UUID) (*models.Member, error) {
	query := `SELECT * FROM members WHERE members.id = $1`

	var member models.Member
	err := r.DB.Get(&member, query, id)
	if err != nil {
		return nil, err
	}

	return &member, nil
}

func (r *Repository) GetById(id uuid.UUID) (*models.Member, error) {
	query := `SELECT * FROM members WHERE members.id = $1`

	var member models.Member
	err := r.DB.Get(&member, query, id)
	if err != nil {
		return nil, err
	}

	// Remove password from the struct
	member.Password = ""

	return &member, nil
}

func (r *Repository) GetMemberByEmail(email string) (*models.Member, error) {
	var member models.Member
	query := `SELECT * FROM members WHERE members.email = $1`

	err := r.DB.Get(&member, query, email)
	if err != nil {
		return nil, err
	}

	return &member, nil
}

func (r *Repository) VerifyCredentials(email, password string) (*models.Member, error) {
	// First get the member by email
	member, err := r.GetMemberByEmail(email)
	if err != nil {
		return nil, err
	}

	// The password validation will be handled in the service layer
	return member, nil
}

func (r *Repository) CreateDefaultMembers(members []CreateDefaultMember) error {
	query := `
	INSERT INTO members(id, email, name, password, status)
	VALUES(:id, :email, :name, :password, :status)
	ON CONFLICT (id) DO NOTHING
	`
	_, err := r.DB.NamedExec(query, members)

	if err != nil {
		return commonhelpers.AnalyzeDBErr(err)
	}

	return nil
}
