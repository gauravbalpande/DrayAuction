const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

export interface User {
  id: string;
  username: string;
  email: string;
  coins: number;
  xp: number;
  rank: string;
  rank_points: number;
  wins: number;
  losses: number;
}

export interface AuthResponse {
  user: User;
  access_token: string;
  refresh_token: string;
}

export interface AuctionPlayer {
  id: string;
  name: string;
  position: string;
  club: string;
  nation: string;
  league: string;
  rating: number;
  market_value: number;
  form: number;
  form_label: string;
  age: number;
  attack: number;
  passing: number;
  defending: number;
  physical: number;
}

export interface AuctionParticipant {
  id: string;
  name: string;
  type: "human" | "ai";
  remaining_budget: number;
  squad_size: number;
  has_passed: boolean;
  style?: string;
  targets?: string[];
  personality?: string;
}

export interface BidRecord {
  participant_id: string;
  participant_name: string;
  amount: number;
  timestamp: string;
}

export interface ActivityEvent {
  type: string;
  message: string;
  icon?: string;
  participant_name?: string;
  amount?: number;
  timestamp: string;
}

export interface AuctionState {
  id: string;
  status: string;
  auction_type: string;
  difficulty: string;
  budget: number;
  player_pool_size: number;
  current_player_index: number;
  current_player?: AuctionPlayer;
  current_bid: number;
  highest_bidder?: string;
  highest_bidder_name?: string;
  timer_seconds: number;
  timer_max: number;
  participants: AuctionParticipant[];
  bid_history: BidRecord[];
  activity_feed: ActivityEvent[];
  results_ready: boolean;
}

export interface CreateAuctionResponse {
  id: string;
  auction_type: string;
  budget: number;
  player_pool_size: number;
  ai_managers: Array<{
    id: string;
    name: string;
    style: string;
    targets: string[];
  }>;
}

class ApiClient {
  private accessToken: string | null = null;

  setToken(token: string | null) {
    this.accessToken = token;
    if (typeof window !== "undefined") {
      if (token) localStorage.setItem("access_token", token);
      else localStorage.removeItem("access_token");
    }
  }

  getToken(): string | null {
    if (this.accessToken) return this.accessToken;
    if (typeof window !== "undefined") {
      return localStorage.getItem("access_token");
    }
    return null;
  }

  private async request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...(options.headers as Record<string, string>),
    };
    const token = this.getToken();
    if (token) headers["Authorization"] = `Bearer ${token}`;

    const res = await fetch(`${API_BASE}${path}`, { ...options, headers });
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: { message: res.statusText } }));
      throw new Error(err.error?.message || "Request failed");
    }
    if (res.status === 204) return {} as T;
    return res.json();
  }

  register(username: string, email: string, password: string) {
    return this.request<AuthResponse>("/auth/register", {
      method: "POST",
      body: JSON.stringify({ username, email, password }),
    });
  }

  login(email: string, password: string) {
    return this.request<AuthResponse>("/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    });
  }

  createAuction(data: { ai_opponents: number; difficulty: string }) {
    return this.request<CreateAuctionResponse>("/auctions", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  getAuction(id: string) {
    return this.request<AuctionState>(`/auctions/${id}`);
  }

  startAuction(id: string) {
    return this.request<AuctionState>(`/auctions/${id}/start`, { method: "POST" });
  }

  bid(id: string, increment: number) {
    return this.request<{ accepted: boolean; amount: number }>(
      `/auctions/${id}/bids?increment=${increment}`,
      { method: "POST", body: JSON.stringify({}) }
    );
  }

  pass(id: string) {
    return this.request<{ passed: boolean }>(`/auctions/${id}/pass`, {
      method: "POST",
      body: JSON.stringify({}),
    });
  }

  getResults(id: string) {
    return this.request<{ ready: boolean; winner_name: string; teams: unknown[] }>(`/auctions/${id}/results`);
  }
}

export const api = new ApiClient();
