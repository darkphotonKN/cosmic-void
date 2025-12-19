import { create } from "zustand";
import { persist } from "zustand/middleware";

interface MemberInfo {
  id: string;
  name: string;
  email: string;
  status: number;
  average_rating: number;
  created_at: string;
  updated_at: string;
}

interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  memberInfo: MemberInfo | null;
  isAuthenticated: boolean;
  setAuth: (data: {
    accessToken: string;
    refreshToken: string;
    memberInfo: MemberInfo;
  }) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      accessToken: null,
      refreshToken: null,
      memberInfo: null,
      isAuthenticated: false,
      setAuth: (data) =>
        set({
          accessToken: data.accessToken,
          refreshToken: data.refreshToken,
          memberInfo: data.memberInfo,
          isAuthenticated: true,
        }),
      logout: () =>
        set({
          accessToken: null,
          refreshToken: null,
          memberInfo: null,
          isAuthenticated: false,
        }),
    }),
    {
      name: "auth-storage",
    }
  )
);
