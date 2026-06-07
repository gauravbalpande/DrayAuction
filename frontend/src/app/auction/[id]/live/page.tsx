"use client";

import { useEffect, useCallback, useState, useRef } from "react";
import { useParams, useRouter } from "next/navigation";
import { AuctionRoom } from "@/components/auction/AuctionRoom";
import { api } from "@/lib/api";
import { useAuctionStore } from "@/stores";

export default function LiveAuctionPage() {
  const params = useParams();
  const router = useRouter();
  const auctionId = params.id as string;
  const { state, setState } = useAuctionStore();
  const [bidding, setBidding] = useState(false);
  const activeBidderId = useAuctionStore((s) => s.activeBidderId);
  const navigatingRef = useRef(false);

  const pollState = useCallback(async () => {
    if (navigatingRef.current) return;
    try {
      const s = await api.getAuction(auctionId);
      setState(s);

      if (s.status === "completed") {
        navigatingRef.current = true;
        try {
          await api.getResults(auctionId);
          router.push(`/auction/${auctionId}/results`);
        } catch {
          navigatingRef.current = false;
        }
      }
    } catch {
      /* ignore transient errors */
    }
  }, [auctionId, setState, router]);

  useEffect(() => {
    pollState();
    const interval = setInterval(pollState, 1000);
    return () => clearInterval(interval);
  }, [pollState]);

  const handleBid = async (increment: number) => {
    setBidding(true);
    try {
      await api.bid(auctionId, increment);
      await pollState();
    } finally {
      setBidding(false);
    }
  };

  const handlePass = async () => {
    setBidding(true);
    try {
      await api.pass(auctionId);
      await pollState();
    } finally {
      setBidding(false);
    }
  };

  if (!state?.current_player) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center gap-4">
        <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin" />
        <p className="text-muted-foreground">Entering the auction room...</p>
      </div>
    );
  }

  return (
    <AuctionRoom
      state={state}
      activeBidderId={activeBidderId}
      bidding={bidding}
      onBid={handleBid}
      onPass={handlePass}
    />
  );
}
