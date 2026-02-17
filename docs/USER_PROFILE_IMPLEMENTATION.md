# User Profile Implementation Guide

## Overview

This document outlines the implementation of user profiles for the Cosmic Void game, using a denormalized database design for optimal performance.

## Architecture Decision: Denormalized Design

### Why Denormalized?

- **Performance**: Single query fetches all user data (no JOINs required)
- **Caching**: Simpler cache invalidation strategy
- **Scalability**: Better horizontal scaling capabilities
- **Event-driven updates**: Game stats can be updated asynchronously via RabbitMQ events

## Backend Implementation (Auth Service)

### 1. Database Schema Migration

Create new migration file: `000004_add_user_profile_fields.up.sql`

```sql
-- Add profile fields to members table (denormalized approach)
ALTER TABLE members ADD COLUMN IF NOT EXISTS avatar_url TEXT;
ALTER TABLE members ADD COLUMN IF NOT EXISTS bio TEXT;
ALTER TABLE members ADD COLUMN IF NOT EXISTS display_name TEXT;
ALTER TABLE members ADD COLUMN IF NOT EXISTS total_games_played INTEGER DEFAULT 0;
ALTER TABLE members ADD COLUMN IF NOT EXISTS total_wins INTEGER DEFAULT 0;
ALTER TABLE members ADD COLUMN IF NOT EXISTS total_kills INTEGER DEFAULT 0;
ALTER TABLE members ADD COLUMN IF NOT EXISTS total_deaths INTEGER DEFAULT 0;
ALTER TABLE members ADD COLUMN IF NOT EXISTS win_rate DECIMAL(5,2) GENERATED ALWAYS AS (
    CASE
        WHEN total_games_played = 0 THEN 0
        ELSE (total_wins::DECIMAL / total_games_played * 100)
    END
) STORED;
ALTER TABLE members ADD COLUMN IF NOT EXISTS kd_ratio DECIMAL(5,2) GENERATED ALWAYS AS (
    CASE
        WHEN total_deaths = 0 THEN total_kills::DECIMAL
        ELSE (total_kills::DECIMAL / total_deaths)
    END
) STORED;
ALTER TABLE members ADD COLUMN IF NOT EXISTS favorite_weapon TEXT;
ALTER TABLE members ADD COLUMN IF NOT EXISTS play_time_minutes INTEGER DEFAULT 0;
ALTER TABLE members ADD COLUMN IF NOT EXISTS last_game_at TIMESTAMP;

-- Add indexes for performance
CREATE INDEX idx_members_total_wins ON members(total_wins DESC);
CREATE INDEX idx_members_win_rate ON members(win_rate DESC);
CREATE INDEX idx_members_kd_ratio ON members(kd_ratio DESC);
```

Down migration: `000004_add_user_profile_fields.down.sql`

```sql
ALTER TABLE members DROP COLUMN IF EXISTS avatar_url;
ALTER TABLE members DROP COLUMN IF EXISTS bio;
ALTER TABLE members DROP COLUMN IF EXISTS display_name;
ALTER TABLE members DROP COLUMN IF EXISTS total_games_played;
ALTER TABLE members DROP COLUMN IF EXISTS total_wins;
ALTER TABLE members DROP COLUMN IF EXISTS total_kills;
ALTER TABLE members DROP COLUMN IF EXISTS total_deaths;
ALTER TABLE members DROP COLUMN IF EXISTS win_rate;
ALTER TABLE members DROP COLUMN IF EXISTS kd_ratio;
ALTER TABLE members DROP COLUMN IF EXISTS favorite_weapon;
ALTER TABLE members DROP COLUMN IF EXISTS play_time_minutes;
ALTER TABLE members DROP COLUMN IF EXISTS last_game_at;

DROP INDEX IF EXISTS idx_members_total_wins;
DROP INDEX IF EXISTS idx_members_win_rate;
DROP INDEX IF EXISTS idx_members_kd_ratio;
```

### 2. Proto Definition Updates

Update `game-server/common/api/proto/auth/auth.proto`:

```protobuf
// Add to service definition
service AuthService {
  // ... existing methods ...

  // Get user profile
  rpc GetUserProfile(GetUserProfileRequest) returns (UserProfile) {}
  // Update user profile
  rpc UpdateUserProfile(UpdateUserProfileRequest) returns (UserProfile) {}
  // Get leaderboard
  rpc GetLeaderboard(GetLeaderboardRequest) returns (LeaderboardResponse) {}
}

// Add new message types
message GetUserProfileRequest {
  string member_id = 1;
}

message UserProfile {
  string id = 1;
  string email = 2;
  string name = 3;
  string display_name = 4;
  string avatar_url = 5;
  string bio = 6;
  int32 total_games_played = 7;
  int32 total_wins = 8;
  int32 total_kills = 9;
  int32 total_deaths = 10;
  float win_rate = 11;
  float kd_ratio = 12;
  string favorite_weapon = 13;
  int32 play_time_minutes = 14;
  google.protobuf.Timestamp last_game_at = 15;
  google.protobuf.Timestamp created_at = 16;
  google.protobuf.Timestamp updated_at = 17;
}

message UpdateUserProfileRequest {
  string member_id = 1;
  optional string display_name = 2;
  optional string avatar_url = 3;
  optional string bio = 4;
}

message GetLeaderboardRequest {
  enum LeaderboardType {
    WINS = 0;
    WIN_RATE = 1;
    KD_RATIO = 2;
    GAMES_PLAYED = 3;
  }
  LeaderboardType type = 1;
  int32 limit = 2; // default 10, max 100
}

message LeaderboardEntry {
  string member_id = 1;
  string display_name = 2;
  string avatar_url = 3;
  int32 rank = 4;
  float score = 5; // The relevant stat based on leaderboard type
}

message LeaderboardResponse {
  repeated LeaderboardEntry entries = 1;
  GetLeaderboardRequest.LeaderboardType type = 2;
}
```

### 3. Model Updates

Update `game-server/auth-service/internal/models/entities.go`:

```go
type Member struct {
    ID               uuid.UUID  `db:"id" json:"id"`
    Email            string     `db:"email" json:"email"`
    Name             string     `db:"name" json:"name"`
    Password         string     `db:"password" json:"password,omitempty"`
    Status           string     `db:"status" json:"status"`
    AverageRating    float64    `db:"average_rating"`

    // Profile fields
    AvatarURL        *string    `db:"avatar_url" json:"avatar_url"`
    Bio              *string    `db:"bio" json:"bio"`
    DisplayName      *string    `db:"display_name" json:"display_name"`
    TotalGamesPlayed int32      `db:"total_games_played" json:"total_games_played"`
    TotalWins        int32      `db:"total_wins" json:"total_wins"`
    TotalKills       int32      `db:"total_kills" json:"total_kills"`
    TotalDeaths      int32      `db:"total_deaths" json:"total_deaths"`
    WinRate          float32    `db:"win_rate" json:"win_rate"`
    KDRatio          float32    `db:"kd_ratio" json:"kd_ratio"`
    FavoriteWeapon   *string    `db:"favorite_weapon" json:"favorite_weapon"`
    PlayTimeMinutes  int32      `db:"play_time_minutes" json:"play_time_minutes"`
    LastGameAt       *time.Time `db:"last_game_at" json:"last_game_at"`

    CreatedAt        time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
}
```

### 4. Repository Implementation

Add to `game-server/auth-service/internal/member/repository.go`:

```go
func (r *Repository) GetUserProfile(ctx context.Context, memberID uuid.UUID) (*models.Member, error) {
    query := `
        SELECT
            id, email, name, status, average_rating,
            avatar_url, bio, display_name,
            total_games_played, total_wins, total_kills, total_deaths,
            win_rate, kd_ratio, favorite_weapon, play_time_minutes,
            last_game_at, created_at, updated_at
        FROM members
        WHERE id = $1
    `

    var member models.Member
    err := r.db.GetContext(ctx, &member, query, memberID)
    if err != nil {
        return nil, fmt.Errorf("get user profile: %w", err)
    }

    return &member, nil
}

func (r *Repository) UpdateUserProfile(ctx context.Context, memberID uuid.UUID, updates map[string]interface{}) error {
    // Build dynamic UPDATE query
    setClauses := []string{}
    values := []interface{}{}
    argCount := 1

    for field, value := range updates {
        setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, argCount))
        values = append(values, value)
        argCount++
    }

    if len(setClauses) == 0 {
        return nil // No updates
    }

    values = append(values, memberID)
    query := fmt.Sprintf(
        "UPDATE members SET %s, updated_at = CURRENT_TIMESTAMP WHERE id = $%d",
        strings.Join(setClauses, ", "),
        argCount,
    )

    _, err := r.db.ExecContext(ctx, query, values...)
    return err
}

func (r *Repository) UpdateGameStats(ctx context.Context, memberID uuid.UUID, stats *commontypes.PlayerMatchResults) error {
    query := `
        UPDATE members
        SET
            total_games_played = total_games_played + 1,
            total_wins = total_wins + $1,
            total_kills = total_kills + $2,
            total_deaths = total_deaths + $3,
            last_game_at = CURRENT_TIMESTAMP,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $4
    `

    winIncrement := 0
    if stats.Win {
        winIncrement = 1
    }

    _, err := r.db.ExecContext(ctx, query, winIncrement, stats.Kills, stats.Deaths, memberID)
    return err
}

func (r *Repository) GetLeaderboard(ctx context.Context, leaderboardType string, limit int) ([]*models.Member, error) {
    orderBy := "total_wins"
    switch leaderboardType {
    case "WIN_RATE":
        orderBy = "win_rate"
    case "KD_RATIO":
        orderBy = "kd_ratio"
    case "GAMES_PLAYED":
        orderBy = "total_games_played"
    }

    query := fmt.Sprintf(`
        SELECT
            id, display_name, avatar_url, %s as score
        FROM members
        WHERE total_games_played > 0
        ORDER BY %s DESC
        LIMIT $1
    `, orderBy, orderBy)

    var members []*models.Member
    err := r.db.SelectContext(ctx, &members, query, limit)
    return members, err
}
```

### 5. Service Implementation

Add to `game-server/auth-service/internal/member/service.go`:

```go
func (s *service) GetUserProfile(ctx context.Context, req *pb.GetUserProfileRequest) (*pb.UserProfile, error) {
    memberID, err := uuid.Parse(req.MemberId)
    if err != nil {
        return nil, fmt.Errorf("invalid member ID: %w", err)
    }

    // Try cache first
    cacheKey := fmt.Sprintf("profile:%s", memberID)
    cached, err := s.cache.Get(ctx, cacheKey)
    if err == nil && cached != "" {
        var profile pb.UserProfile
        if err := json.Unmarshal([]byte(cached), &profile); err == nil {
            return &profile, nil
        }
    }

    // Fetch from database
    member, err := s.Repo.GetUserProfile(ctx, memberID)
    if err != nil {
        return nil, err
    }

    profile := memberToProfileProto(member)

    // Cache for 5 minutes
    if profileJSON, err := json.Marshal(profile); err == nil {
        s.cache.Set(ctx, cacheKey, string(profileJSON), 5*time.Minute)
    }

    return profile, nil
}

func (s *service) UpdateUserProfile(ctx context.Context, req *pb.UpdateUserProfileRequest) (*pb.UserProfile, error) {
    memberID, err := uuid.Parse(req.MemberId)
    if err != nil {
        return nil, fmt.Errorf("invalid member ID: %w", err)
    }

    updates := make(map[string]interface{})

    if req.DisplayName != nil {
        // Validate display name
        if len(*req.DisplayName) < 3 || len(*req.DisplayName) > 20 {
            return nil, errors.New("display name must be between 3 and 20 characters")
        }
        updates["display_name"] = *req.DisplayName
    }

    if req.Bio != nil {
        if len(*req.Bio) > 500 {
            return nil, errors.New("bio must be 500 characters or less")
        }
        updates["bio"] = *req.Bio
    }

    if req.AvatarUrl != nil {
        // Could add URL validation here
        updates["avatar_url"] = *req.AvatarUrl
    }

    if err := s.Repo.UpdateUserProfile(ctx, memberID, updates); err != nil {
        return nil, err
    }

    // Invalidate cache
    cacheKey := fmt.Sprintf("profile:%s", memberID)
    s.cache.Delete(ctx, cacheKey)

    // Return updated profile
    return s.GetUserProfile(ctx, &pb.GetUserProfileRequest{MemberId: req.MemberId})
}

// Event handler for game match results
func (s *service) HandleMatchEndEvent(ctx context.Context, matchData *commontypes.MatchEndState) error {
    for _, result := range matchData.PlayerMatchResults {
        memberID, err := uuid.Parse(result.MemberID)
        if err != nil {
            slog.Error("Invalid member ID in match results", "error", err)
            continue
        }

        if err := s.Repo.UpdateGameStats(ctx, memberID, result); err != nil {
            slog.Error("Failed to update game stats", "error", err, "memberID", memberID)
        }

        // Invalidate cache
        cacheKey := fmt.Sprintf("profile:%s", memberID)
        s.cache.Delete(ctx, cacheKey)
    }

    return nil
}
```

### 6. RabbitMQ Consumer Setup

Add to auth service main.go or separate consumer:

```go
func setupMatchEndConsumer(ch *amqp.Channel, memberService *member.Service) {
    // Declare queue
    q, err := ch.QueueDeclare(
        "auth.match.ended", // Queue name
        true,               // Durable
        false,              // Auto-delete
        false,              // Exclusive
        false,              // No-wait
        nil,                // Arguments
    )
    if err != nil {
        log.Fatal("Failed to declare queue:", err)
    }

    // Bind to exchange
    err = ch.QueueBind(
        q.Name,
        "",                                    // Routing key
        commonconstants.GameMatchEndedEvent,  // Exchange
        false,
        nil,
    )
    if err != nil {
        log.Fatal("Failed to bind queue:", err)
    }

    // Consume messages
    msgs, err := ch.Consume(
        q.Name,
        "",    // Consumer
        false, // Auto-ack
        false, // Exclusive
        false, // No-local
        false, // No-wait
        nil,   // Args
    )
    if err != nil {
        log.Fatal("Failed to register consumer:", err)
    }

    go func() {
        for msg := range msgs {
            var matchData commontypes.MatchEndState
            if err := json.Unmarshal(msg.Body, &matchData); err != nil {
                slog.Error("Failed to unmarshal match data", "error", err)
                msg.Nack(false, false)
                continue
            }

            ctx := context.Background()
            if err := memberService.HandleMatchEndEvent(ctx, &matchData); err != nil {
                slog.Error("Failed to handle match end event", "error", err)
                msg.Nack(false, true) // Requeue
            } else {
                msg.Ack(false)
            }
        }
    }()
}
```

## Frontend Implementation

### 1. API Service

Create `game-client/src/services/profileService.ts`:

```typescript
import { API_BASE_URL } from "@/config/constants";

export interface UserProfile {
  id: string;
  email: string;
  name: string;
  displayName?: string;
  avatarUrl?: string;
  bio?: string;
  totalGamesPlayed: number;
  totalWins: number;
  totalKills: number;
  totalDeaths: number;
  winRate: number;
  kdRatio: number;
  favoriteWeapon?: string;
  playTimeMinutes: number;
  lastGameAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface UpdateProfileRequest {
  displayName?: string;
  avatarUrl?: string;
  bio?: string;
}

export interface LeaderboardEntry {
  memberId: string;
  displayName: string;
  avatarUrl?: string;
  rank: number;
  score: number;
}

export type LeaderboardType = "WINS" | "WIN_RATE" | "KD_RATIO" | "GAMES_PLAYED";

class ProfileService {
  private getAuthHeader(token: string) {
    return {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    };
  }

  async getUserProfile(memberId: string, token: string): Promise<UserProfile> {
    const response = await fetch(`${API_BASE_URL}/api/profile/${memberId}`, {
      headers: this.getAuthHeader(token),
    });

    if (!response.ok) {
      throw new Error("Failed to fetch profile");
    }

    return response.json();
  }

  async updateProfile(
    memberId: string,
    updates: UpdateProfileRequest,
    token: string,
  ): Promise<UserProfile> {
    const response = await fetch(`${API_BASE_URL}/api/profile/${memberId}`, {
      method: "PATCH",
      headers: this.getAuthHeader(token),
      body: JSON.stringify(updates),
    });

    if (!response.ok) {
      throw new Error("Failed to update profile");
    }

    return response.json();
  }

  async uploadAvatar(file: File, token: string): Promise<string> {
    const formData = new FormData();
    formData.append("avatar", file);

    const response = await fetch(`${API_BASE_URL}/api/profile/avatar`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
      },
      body: formData,
    });

    if (!response.ok) {
      throw new Error("Failed to upload avatar");
    }

    const data = await response.json();
    return data.avatarUrl;
  }

  async getLeaderboard(
    type: LeaderboardType,
    limit: number = 10,
    token: string,
  ): Promise<LeaderboardEntry[]> {
    const response = await fetch(
      `${API_BASE_URL}/api/leaderboard?type=${type}&limit=${limit}`,
      {
        headers: this.getAuthHeader(token),
      },
    );

    if (!response.ok) {
      throw new Error("Failed to fetch leaderboard");
    }

    return response.json();
  }
}

export const profileService = new ProfileService();
```

### 2. Profile Store

Create `game-client/src/stores/profileStore.ts`:

```typescript
import { create } from "zustand";
import {
  UserProfile,
  UpdateProfileRequest,
  LeaderboardEntry,
  LeaderboardType,
} from "@/services/profileService";
import { profileService } from "@/services/profileService";

interface ProfileState {
  currentProfile: UserProfile | null;
  isLoading: boolean;
  error: string | null;
  leaderboard: Record<LeaderboardType, LeaderboardEntry[]>;

  // Actions
  fetchProfile: (memberId: string, token: string) => Promise<void>;
  updateProfile: (
    memberId: string,
    updates: UpdateProfileRequest,
    token: string,
  ) => Promise<void>;
  uploadAvatar: (file: File, memberId: string, token: string) => Promise<void>;
  fetchLeaderboard: (type: LeaderboardType, token: string) => Promise<void>;
  clearProfile: () => void;
}

export const useProfileStore = create<ProfileState>((set, get) => ({
  currentProfile: null,
  isLoading: false,
  error: null,
  leaderboard: {
    WINS: [],
    WIN_RATE: [],
    KD_RATIO: [],
    GAMES_PLAYED: [],
  },

  fetchProfile: async (memberId: string, token: string) => {
    set({ isLoading: true, error: null });
    try {
      const profile = await profileService.getUserProfile(memberId, token);
      set({ currentProfile: profile, isLoading: false });
    } catch (error) {
      set({ error: error.message, isLoading: false });
    }
  },

  updateProfile: async (
    memberId: string,
    updates: UpdateProfileRequest,
    token: string,
  ) => {
    set({ isLoading: true, error: null });
    try {
      const updatedProfile = await profileService.updateProfile(
        memberId,
        updates,
        token,
      );
      set({ currentProfile: updatedProfile, isLoading: false });
    } catch (error) {
      set({ error: error.message, isLoading: false });
    }
  },

  uploadAvatar: async (file: File, memberId: string, token: string) => {
    set({ isLoading: true, error: null });
    try {
      const avatarUrl = await profileService.uploadAvatar(file, token);
      await get().updateProfile(memberId, { avatarUrl }, token);
    } catch (error) {
      set({ error: error.message, isLoading: false });
    }
  },

  fetchLeaderboard: async (type: LeaderboardType, token: string) => {
    try {
      const entries = await profileService.getLeaderboard(type, 10, token);
      set((state) => ({
        leaderboard: {
          ...state.leaderboard,
          [type]: entries,
        },
      }));
    } catch (error) {
      set({ error: error.message });
    }
  },

  clearProfile: () => {
    set({ currentProfile: null, error: null });
  },
}));
```

### 3. Profile Page Component

Create `game-client/src/app/profile/page.tsx`:

```tsx
"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { useAuthStore } from "@/stores/authStore";
import { useProfileStore } from "@/stores/profileStore";
import { ProfileCard } from "@/components/profile/ProfileCard";
import { StatsDisplay } from "@/components/profile/StatsDisplay";
import { EditProfileModal } from "@/components/profile/EditProfileModal";
import { Leaderboard } from "@/components/profile/Leaderboard";

export default function ProfilePage() {
  const { id } = useParams();
  const { accessToken, memberInfo } = useAuthStore();
  const { currentProfile, isLoading, fetchProfile } = useProfileStore();
  const [showEditModal, setShowEditModal] = useState(false);

  const profileId = id || memberInfo?.id;
  const isOwnProfile = profileId === memberInfo?.id;

  useEffect(() => {
    if (profileId && accessToken) {
      fetchProfile(profileId, accessToken);
    }
  }, [profileId, accessToken]);

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-screen">
        Loading...
      </div>
    );
  }

  if (!currentProfile) {
    return (
      <div className="flex justify-center items-center h-screen">
        Profile not found
      </div>
    );
  }

  return (
    <div className="container mx-auto p-4 max-w-6xl">
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Profile Card */}
        <div className="lg:col-span-1">
          <ProfileCard
            profile={currentProfile}
            isOwnProfile={isOwnProfile}
            onEditClick={() => setShowEditModal(true)}
          />
        </div>

        {/* Stats Display */}
        <div className="lg:col-span-2">
          <StatsDisplay profile={currentProfile} />
        </div>

        {/* Leaderboard */}
        <div className="lg:col-span-3">
          <Leaderboard />
        </div>
      </div>

      {/* Edit Modal */}
      {showEditModal && (
        <EditProfileModal
          profile={currentProfile}
          onClose={() => setShowEditModal(false)}
        />
      )}
    </div>
  );
}
```

### 4. Profile Components

Create `game-client/src/components/profile/ProfileCard.tsx`:

```tsx
import { UserProfile } from "@/services/profileService";
import Image from "next/image";

interface ProfileCardProps {
  profile: UserProfile;
  isOwnProfile: boolean;
  onEditClick: () => void;
}

export const ProfileCard: React.FC<ProfileCardProps> = ({
  profile,
  isOwnProfile,
  onEditClick,
}) => {
  const formatPlayTime = (minutes: number) => {
    const hours = Math.floor(minutes / 60);
    const mins = minutes % 60;
    return `${hours}h ${mins}m`;
  };

  return (
    <div className="bg-gray-800 rounded-lg p-6 text-white">
      {/* Avatar */}
      <div className="flex justify-center mb-4">
        <div className="relative w-32 h-32">
          {profile.avatarUrl ? (
            <Image
              src={profile.avatarUrl}
              alt={profile.displayName || profile.name}
              fill
              className="rounded-full object-cover"
            />
          ) : (
            <div className="w-32 h-32 bg-gray-600 rounded-full flex items-center justify-center text-4xl">
              {(profile.displayName || profile.name).charAt(0).toUpperCase()}
            </div>
          )}
        </div>
      </div>

      {/* Name and Bio */}
      <div className="text-center mb-4">
        <h2 className="text-2xl font-bold">
          {profile.displayName || profile.name}
        </h2>
        <p className="text-gray-400">@{profile.name}</p>
        {profile.bio && (
          <p className="mt-2 text-sm text-gray-300">{profile.bio}</p>
        )}
      </div>

      {/* Quick Stats */}
      <div className="space-y-2 text-sm">
        <div className="flex justify-between">
          <span className="text-gray-400">Play Time:</span>
          <span>{formatPlayTime(profile.playTimeMinutes)}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-gray-400">Member Since:</span>
          <span>{new Date(profile.createdAt).toLocaleDateString()}</span>
        </div>
        {profile.lastGameAt && (
          <div className="flex justify-between">
            <span className="text-gray-400">Last Game:</span>
            <span>{new Date(profile.lastGameAt).toLocaleDateString()}</span>
          </div>
        )}
      </div>

      {/* Edit Button */}
      {isOwnProfile && (
        <button
          onClick={onEditClick}
          className="mt-4 w-full bg-blue-600 hover:bg-blue-700 py-2 rounded-md transition"
        >
          Edit Profile
        </button>
      )}
    </div>
  );
};
```

## API Gateway Routes

Add to API Gateway to proxy gRPC calls:

```go
// In api-gateway routes configuration
router.GET("/api/profile/:id", authMiddleware, handlers.GetUserProfile)
router.PATCH("/api/profile/:id", authMiddleware, handlers.UpdateUserProfile)
router.POST("/api/profile/avatar", authMiddleware, handlers.UploadAvatar)
router.GET("/api/leaderboard", authMiddleware, handlers.GetLeaderboard)
```

## Testing Strategy

### Backend Tests

1. **Unit Tests**:

   - Repository methods with mock DB
   - Service methods with mock repository
   - Profile data validation

2. **Integration Tests**:

   - gRPC endpoints with test client
   - RabbitMQ event handling
   - Cache invalidation

3. **Load Tests**:
   - Profile fetching under load
   - Leaderboard queries performance
   - Cache hit ratio

### Frontend Tests

1. **Component Tests**:

   - Profile display with mock data
   - Edit form validation
   - Avatar upload flow

2. **E2E Tests**:
   - Full profile view/edit flow
   - Leaderboard navigation
   - Error handling

## Performance Considerations

### Caching Strategy

1. **Redis Cache**:

   - Cache profiles for 5 minutes
   - Cache leaderboards for 1 minute
   - Invalidate on updates

2. **Database Indexes**:

   - Index on game stats for leaderboards
   - Composite index for filtered queries

3. **Query Optimization**:
   - Use generated columns for calculated fields
   - Batch updates from game events

## Security Considerations

1. **Authorization**:

   - Users can only edit their own profile
   - Validate all input data
   - Rate limit profile updates

2. **Data Privacy**:

   - Optional fields (bio, avatar) are nullable
   - Email never exposed in public profiles
   - Configurable privacy settings (future)

3. **Avatar Upload**:
   - File size limits (max 5MB)
   - Image format validation
   - CDN/S3 storage (not local)

## Deployment Checklist

1. [ ] Run database migrations
2. [ ] Update proto and regenerate code
3. [ ] Deploy auth service with new handlers
4. [ ] Setup RabbitMQ consumer for game events
5. [ ] Update API gateway routes
6. [ ] Deploy frontend with profile pages
7. [ ] Configure Redis cache
8. [ ] Test profile CRUD operations
9. [ ] Verify game stats updates
10. [ ] Monitor performance metrics

## Future Enhancements

1. **Social Features**:

   - Friend system
   - Profile privacy settings
   - Activity feed

2. **Achievements**:

   - Achievement badges
   - Milestone tracking
   - Seasonal stats

3. **Advanced Stats**:

   - Per-weapon statistics
   - Heat maps
   - Match history

4. **Customization**:
   - Profile themes
   - Custom banners
   - Showcase items/achievements
