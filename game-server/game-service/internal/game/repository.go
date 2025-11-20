package game

import (
	"fmt"

	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/models"
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

func (r *Repository) CreateRoom(params CreateRoomParams) (uuid.UUID, error) {
	roomId := uuid.New()
	query := `INSERT INTO rooms (id, name, creator_id, max_players, game_mode) VALUES ($1, $2, $3, $4, $5)`

	_, err := r.DB.Exec(query, roomId, params.Name, params.CreatorID, params.MaxPlayers, params.GameMode)
	if err != nil {
		fmt.Println("Error when creating room:", err)
		return uuid.Nil, commonhelpers.AnalyzeDBErr(err)
	}

	return roomId, nil
}

func (r *Repository) GetRoomById(id uuid.UUID) (*models.Room, error) {
	query := `SELECT * FROM rooms WHERE id = $1`

	var room models.Room
	err := r.DB.Get(&room, query, id)
	if err != nil {
		return nil, err
	}

	return &room, nil
}

func (r *Repository) GetRoomsByGameMode(gameMode string, limit, offset int) ([]*models.Room, error) {
	query := `SELECT * FROM rooms WHERE game_mode = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	var rooms []*models.Room
	err := r.DB.Select(&rooms, query, gameMode, limit, offset)
	if err != nil {
		return nil, err
	}

	return rooms, nil
}

func (r *Repository) GetAllRooms(limit, offset int) ([]*models.Room, error) {
	query := `SELECT * FROM rooms ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	var rooms []*models.Room
	err := r.DB.Select(&rooms, query, limit, offset)
	if err != nil {
		return nil, err
	}

	return rooms, nil
}

func (r *Repository) UpdateRoomStatus(roomId uuid.UUID, status string) error {
	query := `UPDATE rooms SET status = $1, updated_at = NOW() WHERE id = $2`

	result, err := r.DB.Exec(query, status, roomId)
	if err != nil {
		return commonhelpers.AnalyzeDBErr(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no room found with id: %v", roomId)
	}

	return nil
}

func (r *Repository) UpdateRoomCurrentPlayers(roomId uuid.UUID, currentPlayers int) error {
	query := `UPDATE rooms SET current_players = $1, updated_at = NOW() WHERE id = $2`

	result, err := r.DB.Exec(query, currentPlayers, roomId)
	if err != nil {
		return commonhelpers.AnalyzeDBErr(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no room found with id: %v", roomId)
	}

	return nil
}

func (r *Repository) CreatePlayer(params JoinRoomParams) (uuid.UUID, error) {
	playerId := uuid.New()
	query := `INSERT INTO players (id, user_id, room_id, x, y) VALUES ($1, $2, $3, $4, $5)`

	_, err := r.DB.Exec(query, playerId, params.UserID, params.RoomID, params.X, params.Y)
	if err != nil {
		fmt.Println("Error when creating player:", err)
		return uuid.Nil, commonhelpers.AnalyzeDBErr(err)
	}

	return playerId, nil
}

func (r *Repository) GetPlayerById(id uuid.UUID) (*models.Player, error) {
	query := `SELECT * FROM players WHERE id = $1`

	var player models.Player
	err := r.DB.Get(&player, query, id)
	if err != nil {
		return nil, err
	}

	return &player, nil
}

func (r *Repository) GetPlayersByRoomId(roomId uuid.UUID) ([]*models.Player, error) {
	query := `SELECT * FROM players WHERE room_id = $1 ORDER BY joined_at`

	var players []*models.Player
	err := r.DB.Select(&players, query, roomId)
	if err != nil {
		return nil, err
	}

	return players, nil
}

func (r *Repository) UpdatePlayerPosition(params UpdatePlayerPositionParams) error {
	query := `UPDATE players SET x = :x, y = :y, velocity_x = :velocity_x, velocity_y = :velocity_y WHERE id = :id`

	result, err := r.DB.NamedExec(query, params)
	if err != nil {
		return commonhelpers.AnalyzeDBErr(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no player found with id: %v", params.PlayerID)
	}

	return nil
}

func (r *Repository) UpdatePlayerHealth(params UpdatePlayerHealthParams) error {
	query := `UPDATE players SET health = :health WHERE id = :id`

	result, err := r.DB.NamedExec(query, params)
	if err != nil {
		return commonhelpers.AnalyzeDBErr(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no player found with id: %v", params.PlayerID)
	}

	return nil
}

func (r *Repository) UpdatePlayerScore(params UpdatePlayerScoreParams) error {
	query := `UPDATE players SET score = :score WHERE id = :id`

	result, err := r.DB.NamedExec(query, params)
	if err != nil {
		return commonhelpers.AnalyzeDBErr(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no player found with id: %v", params.PlayerID)
	}

	return nil
}

func (r *Repository) DeletePlayer(playerId uuid.UUID) error {
	query := `DELETE FROM players WHERE id = $1`

	result, err := r.DB.Exec(query, playerId)
	if err != nil {
		return commonhelpers.AnalyzeDBErr(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no player found with id: %v", playerId)
	}

	return nil
}

func (r *Repository) DeleteRoom(roomId uuid.UUID) error {
	query := `DELETE FROM rooms WHERE id = $1`

	result, err := r.DB.Exec(query, roomId)
	if err != nil {
		return commonhelpers.AnalyzeDBErr(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no room found with id: %v", roomId)
	}

	return nil
}