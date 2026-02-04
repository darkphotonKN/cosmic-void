# Avatar Upload Flow

## Overview
This document describes the avatar upload flow using S3 presigned URLs for the auth-service.

## Architecture
The system uses a direct client-to-S3 upload pattern with presigned URLs to minimize server bandwidth and improve performance.

## Upload Flow

### 1. Request Presigned URL
**Endpoint**: `POST /members/{id}/avatar/upload-request`

**Request Body**:
```json
{
  "filename": "avatar.jpg"
}
```

**Response**:
```json
{
  "upload_id": "uuid",
  "presigned_url": "https://s3.amazonaws.com/...",
  "s3_key": "avatars/member_id/timestamp_random.jpg",
  "expires_at": "2024-01-01T00:05:00Z",
  "max_file_size": 5242880,
  "allowed_content_types": ["image/jpeg", "image/png", "image/webp"]
}
```

### 2. Client Uploads to S3
Client uses the presigned URL to upload directly to S3:
```javascript
const response = await fetch(presignedUrl, {
  method: 'PUT',
  body: file,
  headers: {
    'Content-Type': file.type
  }
});
```

### 3. Confirm Upload
**Endpoint**: `POST /members/{id}/avatar/confirm`

**Request Body**:
```json
{
  "upload_id": "uuid"
}
```

**Response**:
```json
{
  "success": true,
  "avatar_url": "https://cdn.example.com/avatars/member_id/timestamp_random.jpg"
}
```

## Environment Variables

Required environment variables for S3 configuration:

```bash
# AWS Configuration
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key

# S3 Bucket
S3_BUCKET_NAME=your-bucket-name

# Optional CDN URL (if using CloudFront or similar)
CDN_URL=https://cdn.example.com
```

## S3 Bucket Configuration

### CORS Configuration
The S3 bucket must have CORS configured to allow browser uploads:

```json
{
  "CORSRules": [
    {
      "AllowedOrigins": ["https://yourapp.com"],
      "AllowedMethods": ["PUT", "POST", "GET"],
      "AllowedHeaders": ["*"],
      "ExposeHeaders": ["ETag"],
      "MaxAgeSeconds": 3000
    }
  ]
}
```

### Bucket Policy
Ensure the IAM user has the following permissions:
- `s3:PutObject` - For generating upload presigned URLs
- `s3:GetObject` - For validating uploads
- `s3:HeadObject` - For checking object metadata

## Security Considerations

### File Size Limits
- Maximum file size: 5MB
- Enforced via presigned URL conditions and server-side validation

### Content Type Validation
- Allowed types: `image/jpeg`, `image/png`, `image/webp`
- Validated during presigned URL generation

### URL Expiration
- Presigned URLs expire after 5 minutes
- Prevents URL sharing and replay attacks

### Access Control
- Each member can only upload to their own avatar path
- S3 keys are namespaced by member ID: `avatars/{member_id}/...`

## Database Schema

### avatar_uploads Table
Tracks upload metadata and status:
- `id`: UUID primary key
- `member_id`: Foreign key to members table
- `s3_key`: S3 object key
- `upload_status`: pending, uploaded, synced, failed
- `file_size`: Size in bytes
- `content_type`: MIME type
- `presigned_url_expires_at`: Expiration timestamp
- `created_at`, `updated_at`: Timestamps

## Error Handling

Common error scenarios and responses:

### Invalid File Type
```json
{
  "error": "unsupported file type: .gif"
}
```

### Upload Not Found
```json
{
  "error": "upload not found"
}
```

### S3 Object Not Found
```json
{
  "error": "verifying S3 object: object not found"
}
```

### File Too Large
```json
{
  "error": "file size 6291456 exceeds maximum 5242880"
}
```

## Monitoring and Cleanup

### Failed Uploads
Periodically clean up uploads with status='failed' or expired presigned URLs:
```sql
DELETE FROM avatar_uploads
WHERE upload_status = 'failed'
   OR (upload_status = 'pending' AND presigned_url_expires_at < NOW() - INTERVAL '1 hour');
```

### Orphaned S3 Objects
Use S3 lifecycle policies to automatically delete orphaned objects older than 30 days.

## Future Enhancements

1. **Image Processing**: Automatic resizing and optimization using Lambda
2. **Multiple Sizes**: Generate thumbnail, medium, and large versions
3. **Backup**: Cross-region replication for disaster recovery
4. **Analytics**: Track upload metrics and success rates
5. **Rate Limiting**: Prevent abuse with per-user upload limits