package game

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/models"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/websocket"
	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/game"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type service struct {
	Repo      *Repository
	publishCh *amqp.Channel
	wsHub     *websocket.Hub
}

func NewService(repo *Repository, ch *amqp.Channel, wsHub *websocket.Hub) *service {
	return &service{
		Repo:      repo,
		publishCh: ch,
		wsHub:     wsHub,
	}
}

func roomToProto(r *models.Room, players []*models.Player) *pb.Room {
	if r == nil {
		return nil
	}

	var protoPlayers []*pb.Player
	for _, player := range players {
		protoPlayers = append(protoPlayers, playerToProto(player))
	}

	return &pb.Room{
		Id:             r.ID.String(),
		Name:           r.Name,
		CreatorId:      r.CreatorID.String(),
		MaxPlayers:     int32(r.MaxPlayers),
		CurrentPlayers: int32(r.CurrentPlayers),
		GameMode:       r.GameMode,
		Status:         r.Status,
		CreatedAt:      timestamppb.New(r.CreatedAt),
		Players:        protoPlayers,
	}
}

func playerToProto(p *models.Player) *pb.Player {
	if p == nil {
		return nil
	}

	return &pb.Player{
		Id:        p.ID.String(),
		UserId:    p.UserID.String(),
		RoomId:    p.RoomID.String(),
		X:         p.X,
		Y:         p.Y,
		VelocityX: p.VelocityX,
		VelocityY: p.VelocityY,
		Health:    int32(p.Health),
		Score:     int32(p.Score),
		IsAlive:   p.IsAlive,
		JoinedAt:  timestamppb.New(p.JoinedAt),
	}
}

func (s *service) CreateRoom(ctx context.Context, req *pb.CreateRoomRequest) (*pb.Room, error) {
	creatorID, err := uuid.Parse(req.CreatorId)
	if err != nil {
		return nil, fmt.Errorf("invalid creator ID: %w", err)
	}

	params := CreateRoomParams{
		Name:       req.Name,
		CreatorID:  creatorID,
		MaxPlayers: int(req.MaxPlayers),
		GameMode:   req.GameMode,
	}

	// Create the room
	roomId, err := s.Repo.CreateRoom(params)
	if err != nil {
		return nil, err
	}

	// Publish to message broker
	payload := commonconstants.RoomCreatedEventPayload{
		RoomID:    roomId.String(),
		Name:      req.Name,
		CreatorID: req.CreatorId,
		GameMode:  req.GameMode,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	marshalledPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	err = s.publishCh.PublishWithContext(
		ctx,
		commonconstants.RoomCreatedEvent,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         marshalledPayload,
			DeliveryMode: amqp.Persistent,
		})

	if err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to publish room created event: %v\n", err)
	}

	// Get the created room
	room, err := s.Repo.GetRoomById(roomId)
	if err != nil {
		return nil, err
	}

	// Get players (should be empty for new room)
	players, _ := s.Repo.GetPlayersByRoomId(roomId)

	return roomToProto(room, players), nil
}

func (s *service) GetRoom(ctx context.Context, req *pb.GetRoomRequest) (*pb.Room, error) {
	roomID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid room ID: %w", err)
	}

	room, err := s.Repo.GetRoomById(roomID)
	if err != nil {
		return nil, err
	}

	players, err := s.Repo.GetPlayersByRoomId(roomID)
	if err != nil {
		return nil, err
	}

	return roomToProto(room, players), nil
}

func (s *service) ListRooms(ctx context.Context, req *pb.ListRoomsRequest) (*pb.ListRoomsResponse, error) {
	limit := int(req.Limit)
	offset := int(req.Offset)

	if limit <= 0 {
		limit = 10
	}

	var rooms []*models.Room
	var err error

	if req.GameMode != "" {
		rooms, err = s.Repo.GetRoomsByGameMode(req.GameMode, limit, offset)
	} else {
		rooms, err = s.Repo.GetAllRooms(limit, offset)
	}

	if err != nil {
		return nil, err
	}

	var protoRooms []*pb.Room
	for _, room := range rooms {
		players, _ := s.Repo.GetPlayersByRoomId(room.ID)
		protoRooms = append(protoRooms, roomToProto(room, players))
	}

	return &pb.ListRoomsResponse{
		Rooms: protoRooms,
		Total: int32(len(protoRooms)),
	}, nil
}

func (s *service) JoinRoom(ctx context.Context, req *pb.JoinRoomRequest) (*pb.JoinRoomResponse, error) {
	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, fmt.Errorf("invalid room ID: %w", err)
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Get room to check capacity
	room, err := s.Repo.GetRoomById(roomID)
	if err != nil {
		return nil, err
	}

	if room.CurrentPlayers >= room.MaxPlayers {
		return nil, fmt.Errorf("room is full")
	}

	// Create player
	params := JoinRoomParams{
		RoomID: roomID,
		UserID: userID,
		X:      0, // Default spawn position
		Y:      0,
	}

	playerID, err := s.Repo.CreatePlayer(params)
	if err != nil {
		return nil, err
	}

	// Update room current players count
	err = s.Repo.UpdateRoomCurrentPlayers(roomID, room.CurrentPlayers+1)
	if err != nil {
		return nil, err
	}

	// Get updated room and player
	room, err = s.Repo.GetRoomById(roomID)
	if err != nil {
		return nil, err
	}

	player, err := s.Repo.GetPlayerById(playerID)
	if err != nil {
		return nil, err
	}

	players, err := s.Repo.GetPlayersByRoomId(roomID)
	if err != nil {
		return nil, err
	}

	return &pb.JoinRoomResponse{
		Room:   roomToProto(room, players),
		Player: playerToProto(player),
	}, nil
}

func (s *service) LeaveRoom(ctx context.Context, req *pb.LeaveRoomRequest) (*pb.LeaveRoomResponse, error) {
	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, fmt.Errorf("invalid room ID: %w", err)
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Get room
	room, err := s.Repo.GetRoomById(roomID)
	if err != nil {
		return nil, err
	}

	// Get players in room to find the player to remove
	players, err := s.Repo.GetPlayersByRoomId(roomID)
	if err != nil {
		return nil, err
	}

	var playerToRemove *models.Player
	for _, player := range players {
		if player.UserID == userID {
			playerToRemove = player
			break
		}
	}

	if playerToRemove == nil {
		return &pb.LeaveRoomResponse{
			Success: false,
			Message: "Player not found in room",
		}, nil
	}

	// Remove player
	err = s.Repo.DeletePlayer(playerToRemove.ID)
	if err != nil {
		return nil, err
	}

	// Update room current players count
	newPlayerCount := room.CurrentPlayers - 1
	err = s.Repo.UpdateRoomCurrentPlayers(roomID, newPlayerCount)
	if err != nil {
		return nil, err
	}

	// If room is empty, optionally delete it or mark as inactive
	if newPlayerCount == 0 {
		err = s.Repo.UpdateRoomStatus(roomID, "inactive")
		if err != nil {
			return nil, err
		}
	}

	return &pb.LeaveRoomResponse{
		Success: true,
		Message: "Successfully left room",
	}, nil
}

func (s *service) StartGame(ctx context.Context, req *pb.StartGameRequest) (*pb.StartGameResponse, error) {
	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, fmt.Errorf("invalid room ID: %w", err)
	}

	// Update room status to "active"
	err = s.Repo.UpdateRoomStatus(roomID, "active")
	if err != nil {
		return nil, err
	}

	// Get room and players
	room, err := s.Repo.GetRoomById(roomID)
	if err != nil {
		return nil, err
	}

	players, err := s.Repo.GetPlayersByRoomId(roomID)
	if err != nil {
		return nil, err
	}

	// Publish game started event
	var playerIDs []string
	for _, player := range players {
		playerIDs = append(playerIDs, player.UserID.String())
	}

	payload := commonconstants.GameStartedEventPayload{
		RoomID:    roomID.String(),
		GameMode:  room.GameMode,
		PlayerIDs: playerIDs,
		StartedAt: time.Now().Format(time.RFC3339),
	}

	marshalledPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	err = s.publishCh.PublishWithContext(
		ctx,
		commonconstants.GameStartedEvent,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         marshalledPayload,
			DeliveryMode: amqp.Persistent,
		})

	if err != nil {
		fmt.Printf("Failed to publish game started event: %v\n", err)
	}

	return &pb.StartGameResponse{
		Room:      roomToProto(room, players),
		StartedAt: timestamppb.New(time.Now()),
	}, nil
}

func (s *service) EndGame(ctx context.Context, req *pb.EndGameRequest) (*pb.EndGameResponse, error) {
	roomID, err := uuid.Parse(req.RoomId)
	if err != nil {
		return nil, fmt.Errorf("invalid room ID: %w", err)
	}

	// Update room status to "finished"
	err = s.Repo.UpdateRoomStatus(roomID, "finished")
	if err != nil {
		return nil, err
	}

	// Get room
	room, err := s.Repo.GetRoomById(roomID)
	if err != nil {
		return nil, err
	}

	// Publish game ended event
	payload := commonconstants.GameEndedEventPayload{
		RoomID:   roomID.String(),
		GameMode: room.GameMode,
		WinnerID: req.WinnerId,
		EndedAt:  time.Now().Format(time.RFC3339),
	}

	marshalledPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	err = s.publishCh.PublishWithContext(
		ctx,
		commonconstants.GameEndedEvent,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         marshalledPayload,
			DeliveryMode: amqp.Persistent,
		})

	if err != nil {
		fmt.Printf("Failed to publish game ended event: %v\n", err)
	}

	players, _ := s.Repo.GetPlayersByRoomId(roomID)

	return &pb.EndGameResponse{
		Room:     roomToProto(room, players),
		EndedAt:  timestamppb.New(time.Now()),
		WinnerId: req.WinnerId,
	}, nil
}