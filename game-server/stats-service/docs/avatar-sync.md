# Avatar Sync Documentation

## Overview
This document describes how avatar URLs are synchronized from the auth-service to the stats-service for optimized leaderboard queries.

## Purpose
The `avatar_url` field is denormalized in the `player_ranking_stats` table to improve leaderboard query performance by avoiding joins with the members table in auth-service.

## Database Schema

### player_ranking_stats Table
The `avatar_url` column has been added to store member avatars alongside ranking data:

```sql
ALTER TABLE player_ranking_stats ADD COLUMN avatar_url TEXT;
```

This allows the leaderboard API to return complete player information in a single query:

```sql
SELECT member_id, username, wins, top_threes, rating, avatar_url
FROM player_ranking_stats
ORDER BY rating DESC
LIMIT 100;
```

## Sync Strategy

### When to Sync
Avatar URLs should be synchronized to the stats-service in the following scenarios:

1. **After Match Completion** - When updating player stats after a game ends
2. **Avatar Update Event** - When a member updates their avatar in auth-service
3. **Periodic Sync** - As a fallback, run a periodic sync job (e.g., daily)

### Sync Implementation (To Be Implemented)

The sync will be performed via gRPC calls from auth-service to stats-service:

1. **Auth-service publishes event** when avatar is updated
2. **Stats-service receives event** via message broker or direct gRPC
3. **Stats-service updates** the `avatar_url` in `player_ranking_stats`

Example gRPC method signature:
```protobuf
service StatsService {
  rpc UpdatePlayerAvatar(UpdatePlayerAvatarRequest) returns (UpdatePlayerAvatarResponse);
}

message UpdatePlayerAvatarRequest {
  string member_id = 1;
  string avatar_url = 2;
}
```

## Performance Benefits

### Without Denormalization
- Requires cross-service calls or joins
- Latency: ~50-100ms per leaderboard request
- Complex caching requirements

### With Denormalization
- Single table query
- Latency: ~5-10ms per leaderboard request
- Simple, efficient caching

## Data Consistency

### Eventual Consistency
The system accepts eventual consistency for avatar URLs:
- Avatar updates are non-critical for gameplay
- Temporary inconsistency (few seconds) is acceptable
- Failed syncs can be retried without impacting game functionality

### Handling Missing Avatars
When `avatar_url` is NULL:
- Frontend displays default avatar
- No cross-service calls for missing data
- Periodic sync job will eventually populate

## Migration Strategy

### Initial Population
After adding the column, populate existing records:

```sql
-- This will be done via gRPC batch update from auth-service
UPDATE player_ranking_stats prs
SET avatar_url = (
  SELECT avatar_url FROM auth_service.members m
  WHERE m.id = prs.member_id
);
```

### Rollback Plan
If issues arise, the column can be safely dropped:
```sql
ALTER TABLE player_ranking_stats DROP COLUMN avatar_url;
```
The application will continue functioning with default avatars.

## Monitoring

### Key Metrics to Track
1. **Sync Lag** - Time between avatar update and stats sync
2. **Sync Failures** - Number of failed sync attempts
3. **NULL Avatar Rate** - Percentage of rankings without avatars

### Alerts
Set up alerts for:
- Sync lag > 5 minutes
- Sync failure rate > 1%
- NULL avatar rate > 10%

## Future Enhancements

1. **Batch Sync API** - Update multiple avatars in single call
2. **Change Data Capture** - Use database triggers for real-time sync
3. **Avatar Caching** - CDN integration for avatar serving
4. **Fallback URLs** - Store multiple avatar sizes for responsive display