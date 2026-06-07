"use client";

import { AnimatePresence } from "framer-motion";
import { ManagerCard } from "./ManagerCard";
import { PlayerSpotlight } from "./PlayerSpotlight";
import { BidPanel } from "./BidPanel";
import { SoftTimer } from "./SoftTimer";
import { BidControls } from "./BidControls";
import { LiveFeed } from "./LiveFeed";
import type { AuctionState } from "@/lib/api";
import { formatMoney } from "@/lib/utils";

interface AuctionRoomProps {
  state: AuctionState;
  activeBidderId: string | null;
  bidding: boolean;
  onBid: (increment: number) => void;
  onPass: () => void;
}

export function AuctionRoom({ state, activeBidderId, bidding, onBid, onPass }: AuctionRoomProps) {
  const player = state.current_player!;
  const human = state.participants.find((p) => p.type === "human")!;
  const aiManagers = state.participants.filter((p) => p.type === "ai");
  const isUserHighest = state.highest_bidder === human.id;

  const topLeft = aiManagers[0];
  const topRight = aiManagers[1];
  const bottom = aiManagers[2];
  const extraManagers = aiManagers.slice(3);

  return (
    <div className="min-h-screen">
      {/* Header */}
      <div className="border-b border-border/50 bg-card/30 backdrop-blur sticky top-0 z-40">
        <div className="container mx-auto px-4 py-3 flex items-center justify-between">
          <div>
            <p className="text-xs text-primary font-semibold uppercase tracking-widest">Auction Room</p>
            <p className="text-sm text-muted-foreground">
              {state.auction_type} · {formatMoney(state.budget)} Budget
            </p>
          </div>
          <p className="text-sm font-medium">
            Player {state.current_player_index + 1} / {state.player_pool_size}
          </p>
        </div>
      </div>

      <div className="container mx-auto px-4 py-6 max-w-5xl space-y-6">
        <SoftTimer seconds={state.timer_seconds} maxSeconds={state.timer_max} />

        {/* Virtual auction room layout */}
        <div className="relative py-4">
          {/* Top row: AI managers flanking player */}
          <div className="grid grid-cols-3 gap-4 items-start mb-4">
            <div className="flex justify-start">
              {topLeft && (
                <ManagerCard manager={topLeft} isActive={activeBidderId === topLeft.id} position="top-left" />
              )}
            </div>
            <div className="flex justify-center">
              <AnimatePresence mode="wait">
                <PlayerSpotlight player={player} />
              </AnimatePresence>
            </div>
            <div className="flex justify-end">
              {topRight && (
                <ManagerCard manager={topRight} isActive={activeBidderId === topRight.id} position="top-right" />
              )}
            </div>
          </div>

          {/* Center: Human manager */}
          <div className="flex justify-center mb-4">
            <ManagerCard manager={human} isActive={activeBidderId === human.id} position="center" />
          </div>

          {/* Bottom row: remaining AI managers */}
          {(bottom || extraManagers.length > 0) && (
            <div className="flex justify-center gap-4 flex-wrap">
              {bottom && (
                <ManagerCard manager={bottom} isActive={activeBidderId === bottom.id} position="bottom" />
              )}
              {extraManagers.map((m) => (
                <ManagerCard key={m.id} manager={m} isActive={activeBidderId === m.id} position="bottom" />
              ))}
            </div>
          )}
        </div>

        <BidPanel
          currentBid={state.current_bid}
          marketValue={player.market_value}
          highestBidderName={state.highest_bidder_name}
          isUserHighest={isUserHighest}
          bidHistory={state.bid_history ?? []}
        />

        <BidControls
          disabled={bidding || state.status !== "live"}
          onBid={onBid}
          onPass={onPass}
        />

        <LiveFeed events={state.activity_feed ?? []} />
      </div>
    </div>
  );
}
