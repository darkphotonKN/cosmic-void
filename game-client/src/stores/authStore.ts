import { create } from "zustand";
import { persist } from "zustand/middleware";

interface MemberInfo {
  id: string;
  name: string;
  email: string;
  status: number;
  average_rating: number;
  avatar_url?: string;
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
  updateMemberInfo: (memberInfo: Partial<MemberInfo>) => void;
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
      updateMemberInfo: (updates) =>
        set((state) => ({
          memberInfo: state.memberInfo
            ? { ...state.memberInfo, ...updates }
            : null,
        })),
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
    },
  ),
);