"use client";

import { useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/stores/authStore";
import { apiClient } from "@/utils/api";

interface UploadResponse {
  upload_id: string;
  presigned_url: string;
  s3_key: string;
  expires_at: string;
  max_file_size: number;
  allowed_content_types: string[];
}

export default function ProfilePage() {
  const router = useRouter();
  const { memberInfo, isAuthenticated, updateMemberInfo } = useAuthStore();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [isUploading, setIsUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  // Don't render if not authenticated (middleware handles redirect)
  if (!isAuthenticated || !memberInfo) {
    return (
      <div className="profile-loading">
        <div className="loading-spinner"></div>
      </div>
    );
  }

  const handleAvatarClick = () => {
    fileInputRef.current?.click();
  };

  const handleFileChange = async (
    event: React.ChangeEvent<HTMLInputElement>,
  ) => {
    const file = event.target.files?.[0];
    if (!file) return;

    setError("");
    setSuccess("");

    // Validate file type
    const allowedTypes = ["image/jpeg", "image/jpg", "image/png", "image/webp"];
    if (!allowedTypes.includes(file.type)) {
      setError("Please select a valid image file (JPEG, PNG, WebP)");
      return;
    }

    // Validate file size (5MB)
    const maxSize = 5 * 1024 * 1024;
    if (file.size > maxSize) {
      setError("File size must be less than 5MB");
      return;
    }

    setIsUploading(true);
    setUploadProgress(10);

    try {
      // Step 1: Request presigned URL
      const uploadRequestResponse = await apiClient.requestAvatarUpload(
        file.name,
      );
      const uploadData: UploadResponse = uploadRequestResponse.result;
      setUploadProgress(30);

      // Step 2: Upload to S3
      await apiClient.uploadToS3(uploadData.presigned_url, file);
      setUploadProgress(70);

      // Step 3: Confirm upload
      const confirmResponse = await apiClient.confirmAvatarUpload(
        uploadData.upload_id,
      );

      if (confirmResponse.success) {
        // Update local state with new avatar URL
        updateMemberInfo({ avatar_url: confirmResponse.avatar_url });
        setSuccess("Avatar updated successfully!");
        setUploadProgress(100);

        // Clear success message after 3 seconds
        setTimeout(() => setSuccess(""), 3000);
      } else {
        throw new Error(confirmResponse.message || "Failed to confirm upload");
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Upload failed");
    } finally {
      setIsUploading(false);
      setUploadProgress(0);
      // Reset file input
      if (fileInputRef.current) {
        fileInputRef.current.value = "";
      }
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("en-US", {
      year: "numeric",
      month: "long",
      day: "numeric",
    });
  };

  const getStatusLabel = (status: number) => {
    return status === 1 ? "Operator" : status === 2 ? "Commander" : "Unknown";
  };

  return (
    <main className="profile-container">
      {/* Background */}
      <div className="profile-bg" />

      {/* Header */}
      <div className="profile-header">
        <button
          onClick={() => router.push("/game")}
          className="profile-back-btn"
        >
          Back to Game
        </button>
        <h1 className="profile-title">OPERATOR PROFILE</h1>
      </div>

      {/* Profile Card */}
      <div className="profile-card">
        {/* Avatar Section */}
        <div className="profile-avatar-section">
          <div
            className={`profile-avatar ${isUploading ? "uploading" : ""}`}
            onClick={handleAvatarClick}
          >
            {memberInfo.avatar_url ? (
              <img
                src={memberInfo.avatar_url}
                alt="Profile Avatar"
                className="profile-avatar-img"
              />
            ) : (
              <div className="profile-avatar-placeholder">
                <svg
                  className="profile-avatar-icon"
                  viewBox="0 0 24 24"
                  fill="currentColor"
                >
                  <path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z" />
                </svg>
              </div>
            )}

            {/* Upload Overlay */}
            <div className="profile-avatar-overlay">
              {isUploading ? (
                <div className="profile-upload-progress">
                  <div className="profile-spinner" />
                  <span>{uploadProgress}%</span>
                </div>
              ) : (
                <div className="profile-avatar-text">
                  <svg
                    className="profile-camera-icon"
                    viewBox="0 0 24 24"
                    fill="currentColor"
                  >
                    <path d="M12 15.5A3.5 3.5 0 0 1 8.5 12A3.5 3.5 0 0 1 12 8.5a3.5 3.5 0 0 1 3.5 3.5a3.5 3.5 0 0 1-3.5 3.5m7.43-2.53c.04-.32.07-.64.07-.97c0-.33-.03-.66-.07-1l1.86-1.46c.25-.2.31-.57.13-.85l-1.86-3.22c-.12-.22-.39-.31-.61-.22l-2.19.87c-.26-.2-.55-.37-.85-.5l-.29-2.39C15.54 2.5 15.26 2.25 14.93 2.25h-3.86c-.33 0-.61.25-.66.58L10.12 5.23c-.3.13-.59.3-.85.5l-2.19-.87c-.22-.09-.49 0-.61.22L4.61 8.3c-.18.28-.12.65.13.85L6.6 10.61c-.04.34-.07.67-.07 1c0 .33.03.65.07.97L4.74 14.04c-.25.2-.31.57-.13.85l1.86 3.22c.12.22.39.31.61.22l2.19-.87c.26.2.55.37.85.5l.29 2.39c.05.33.33.58.66.58h3.86c.33 0 .61-.25.66-.58l.29-2.39c.3-.13.59-.3.85-.5l2.19.87c.22.09.49 0 .61-.22l1.86-3.22c.18-.28.12-.65-.13-.85L17.4 13.47z" />
                  </svg>
                  <span>Change Photo</span>
                </div>
              )}
            </div>
          </div>

          <input
            ref={fileInputRef}
            type="file"
            accept="image/jpeg,image/jpg,image/png,image/webp"
            onChange={handleFileChange}
            className="profile-file-input"
            disabled={isUploading}
          />
        </div>

        {/* User Info */}
        <div className="profile-info">
          <h2 className="profile-name">{memberInfo.name}</h2>
          <p className="profile-email">{memberInfo.email}</p>

          <div className="profile-details">
            <div className="profile-detail-item">
              <span className="profile-detail-label">Rank:</span>
              <span className="profile-detail-value">
                {getStatusLabel(memberInfo.status)}
              </span>
            </div>

            <div className="profile-detail-item">
              <span className="profile-detail-label">Rating:</span>
              <span className="profile-detail-value">
                {memberInfo.average_rating?.toFixed(1)}/5.0
              </span>
            </div>

            <div className="profile-detail-item">
              <span className="profile-detail-label">Enlisted:</span>
              <span className="profile-detail-value">
                {formatDate(memberInfo.created_at)}
              </span>
            </div>
          </div>
        </div>

        {/* Messages */}
        {error && (
          <div className="profile-message profile-error">
            <svg
              className="profile-message-icon"
              viewBox="0 0 24 24"
              fill="currentColor"
            >
              <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z" />
            </svg>
            {error}
          </div>
        )}

        {success && (
          <div className="profile-message profile-success">
            <svg
              className="profile-message-icon"
              viewBox="0 0 24 24"
              fill="currentColor"
            >
              <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z" />
            </svg>
            {success}
          </div>
        )}
      </div>
    </main>
  );
}
