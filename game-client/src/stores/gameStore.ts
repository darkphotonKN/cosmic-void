import { create } from "zustand";

interface GameState {
  sessionId: string | null;
  setSessionId: (id: string | null) => void;
}

export const useGameStore = create<GameState>()((set) => ({
  sessionId: null,
  setSessionId: (id) => set({ sessionId: id }),
}));
