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
  const [completionPhase, setCompletionPhase] = useState<
    "none" | "calculating" | "navigating" | "error"
  >("none");
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const pollAttemptsRef = useRef(0);
  const consecutiveErrorsRef = useRef(0);

  const navigateToResults = useCallback(async () => {
    if (navigatingRef.current) return;
    navigatingRef.current = true;
    setCompletionPhase("calculating");

    const maxAttempts = 30;
    pollAttemptsRef.current = 0;

    while (pollAttemptsRef.current < maxAttempts) {
      pollAttemptsRef.current++;
      try {
        const data = await api.getResults(auctionId);
        if (data.ready) {
          setCompletionPhase("navigating");
          router.push(`/auction/${auctionId}/results`);
          return;
        }
      } catch {
        // Results not ready yet, keep polling
      }
      await new Promise((r) => setTimeout(r, 1000));
    }
    setCompletionPhase("error");
    navigatingRef.current = false;
  }, [auctionId, router]);

  const pollState = useCallback(async () => {
    if (navigatingRef.current) return;
    try {
      const s = await api.getAuction(auctionId);
      setState(s);
      consecutiveErrorsRef.current = 0;
      setErrorMsg(null);

      if (
        s.status === "completed" ||
        s.status === "calculating_results" ||
        s.results_ready
      ) {
        navigateToResults();
      }
    } catch {
      consecutiveErrorsRef.current++;
      if (consecutiveErrorsRef.current >= 5) {
        const currentState = useAuctionStore.getState().state;
        const msg = currentState
          ? "Connection to the auction room was lost. The server may be restarting or experiencing issues."
          : "Unable to join the auction room. The auction may have expired or been deleted.";
        setErrorMsg(msg);
      }
    }
  }, [auctionId, setState, navigateToResults]);

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

  if (errorMsg) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center gap-4">
        <p className="text-destructive text-lg font-semibold text-center max-w-md px-4">
          {errorMsg}
        </p>
        <div className="flex gap-3">
          <button
            className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90"
            onClick={() => {
              setErrorMsg(null);
              consecutiveErrorsRef.current = 0;
              pollState();
            }}
          >
            Retry
          </button>
          <button
            className="px-4 py-2 border border-border rounded-md hover:bg-accent"
            onClick={() => router.push("/dashboard")}
          >
            Back to Dashboard
          </button>
        </div>
      </div>
    );
  }

  if (completionPhase === "error") {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center gap-4">
        <p className="text-destructive text-lg font-semibold">
          Results are taking longer than expected.
        </p>
        <p className="text-muted-foreground text-sm">
          The server may still be calculating results.
        </p>
        <div className="flex gap-3">
          <button
            className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90"
            onClick={() => {
              setCompletionPhase("none");
              navigatingRef.current = false;
              navigateToResults();
            }}
          >
            Retry
          </button>
          <button
            className="px-4 py-2 border border-border rounded-md hover:bg-accent"
            onClick={() => router.push("/dashboard")}
          >
            Back to Dashboard
          </button>
        </div>
      </div>
    );
  }

  if (completionPhase === "calculating" || completionPhase === "navigating") {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center gap-4">
        <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin" />
        <p className="text-muted-foreground">
          {completionPhase === "navigating"
            ? "Loading results..."
            : "Calculating final scores..."}
        </p>
        <p className="text-xs text-muted-foreground">
          Analyzing squad balance, chemistry & formation fit
        </p>
      </div>
    );
  }

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
