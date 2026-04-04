const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:7001";

interface ApiConfig {
  method?: string;
  headers?: Record<string, string>;
  body?: any;
}

class ApiClient {
  private getAuthToken(): string | null {
    // Get token from localStorage (Zustand persist)
    const authStorage = localStorage.getItem("auth-storage");
    if (authStorage) {
      try {
        const data = JSON.parse(authStorage);
        return data.state?.accessToken || null;
      } catch {
        return null;
      }
    }
    return null;
  }

  private async request(endpoint: string, config: ApiConfig = {}) {
    const token = this.getAuthToken();

    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...config.headers,
    };

    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
    }

    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      method: config.method || "GET",
      headers,
      body: config.body ? JSON.stringify(config.body) : undefined,
    });

    const data = await response.json();

    if (response.status === 401) {
      // Just throw the error, don't automatically redirect
      throw new Error(data.message || "Unauthorized");
    }

    if (!response.ok) {
      throw new Error(data.message || `Request failed with status ${response.status}`);
    }

    return data;
  }

  // Get current member profile
  async getMemberProfile() {
    return this.request("/api/member");
  }

  // Request avatar upload presigned URL
  async requestAvatarUpload(filename: string) {
    return this.request("/api/member/avatar/upload-request", {
      method: "POST",
      body: { filename },
    });
  }

  // Confirm avatar upload
  async confirmAvatarUpload(uploadId: string) {
    return this.request("/api/member/avatar/confirm", {
      method: "POST",
      body: { upload_id: uploadId },
    });
  }

  // Subscribe to a product (backend auto-creates Stripe customer)
  async subscribe(productId: string, email: string) {
    return this.request("/api/payment/subscribe", {
      method: "POST",
      body: { product_id: productId, email },
    });
  }

  // Check subscription permission (polling endpoint)
  async checkSubscriptionPermission(): Promise<{ has_permission: boolean }> {
    return this.request("/api/payment/subscription/permission");
  }

  // Upload file directly to S3
  async uploadToS3(presignedUrl: string, file: File) {
    const response = await fetch(presignedUrl, {
      method: "PUT",
      body: file,
      headers: {
        "Content-Type": file.type,
      },
    });

    if (!response.ok) {
      throw new Error(`S3 upload failed with status ${response.status}`);
    }

    return response;
  }

  // Get leaderboard
  async getLeaderboard(limit: number = 50, offset: number = 0) {
    const params = new URLSearchParams({
      limit: limit.toString(),
      offset: offset.toString(),
    });
    return this.request(`/api/stats/leaderboard?${params.toString()}`);
  }

  // Get notifications for current user
  async getNotifications() {
    return this.request("/api/notification/");
  }

  // Mark notification as read
  async markNotificationAsRead(notificationId: string) {
    return this.request(`/api/notification/${notificationId}/read`, {
      method: "PATCH",
    });
  }

  // Mark all notifications as read
  async markAllNotificationsAsRead() {
    return this.request("/api/notification/read-all", {
      method: "PATCH",
    });
  }
}

export const apiClient = new ApiClient();