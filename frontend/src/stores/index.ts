import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { User, AuctionState, ActivityEvent } from "@/lib/api";
import { api } from "@/lib/api";

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (username: string, email: string, password: string) => Promise<void>;
  logout: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      isAuthenticated: false,
      login: async (email, password) => {
        const res = await api.login(email, password);
        api.setToken(res.access_token);
        set({ user: res.user, isAuthenticated: true });
      },
      register: async (username, email, password) => {
        const res = await api.register(username, email, password);
        api.setToken(res.access_token);
        set({ user: res.user, isAuthenticated: true });
      },
      logout: () => {
        api.setToken(null);
        set({ user: null, isAuthenticated: false });
      },
    }),
    { name: "auctionxi-auth", partialize: (s) => ({ user: s.user, isAuthenticated: s.isAuthenticated }) }
  )
);

interface AuctionSetup {
  aiOpponents: number;
  difficulty: string;
}

interface AuctionStore {
  auctionId: string | null;
  setup: AuctionSetup;
  state: AuctionState | null;
  activeBidderId: string | null;
  setSetup: (setup: Partial<AuctionSetup>) => void;
  setAuctionId: (id: string) => void;
  setState: (state: AuctionState) => void;
  setActiveBidder: (id: string | null) => void;
  reset: () => void;
}

const defaultSetup: AuctionSetup = {
  aiOpponents: 3,
  difficulty: "medium",
};

export const useAuctionStore = create<AuctionStore>((set) => ({
  auctionId: null,
  setup: defaultSetup,
  state: null,
  activeBidderId: null,
  setSetup: (partial) => set((s) => ({ setup: { ...s.setup, ...partial } })),
  setAuctionId: (id) => set({ auctionId: id }),
  setState: (state) => {
    const prevFeedLen = useAuctionStore.getState().state?.activity_feed?.length ?? 0;
    const newFeedLen = state.activity_feed?.length ?? 0;
    let activeBidderId = useAuctionStore.getState().activeBidderId;
    if (newFeedLen > prevFeedLen && state.activity_feed?.[0]?.type === "bid_placed") {
      const latest = state.activity_feed[0];
      const match = state.participants.find((p) => p.name === latest.participant_name);
      if (match) activeBidderId = match.id;
    }
    set({ state, activeBidderId });
  },
  setActiveBidder: (id) => set({ activeBidderId: id }),
  reset: () => set({ auctionId: null, state: null, activeBidderId: null, setup: defaultSetup }),
}));

export const AUCTION_TIER_PREVIEW: Record<string, { type: string; budget: string; pool: number }> = {
  easy: { type: "Casual Auction", budget: "500M", pool: 40 },
  medium: { type: "Competitive Auction", budget: "800M", pool: 50 },
  hard: { type: "Competitive Auction", budget: "800M", pool: 50 },
  legendary: { type: "Elite Auction", budget: "1200M", pool: 60 },
};
