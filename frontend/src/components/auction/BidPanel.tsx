"use client";

import { motion, AnimatePresence } from "framer-motion";
import { cn, formatMoney } from "@/lib/utils";
import type { BidRecord } from "@/lib/api";

interface BidPanelProps {
  currentBid: number;
  marketValue: number;
  highestBidderName?: string;
  isUserHighest: boolean;
  bidHistory: BidRecord[];
}

export function BidPanel({ currentBid, marketValue, highestBidderName, isUserHighest, bidHistory }: BidPanelProps) {
  const displayBid = currentBid || marketValue;

  return (
    <div className="grid md:grid-cols-3 gap-4">
      <motion.div
        key={displayBid}
        initial={{ scale: 1.1 }}
        animate={{ scale: 1 }}
        className="rounded-xl border border-border bg-card p-4 text-center"
      >
        <p className="text-xs text-muted-foreground uppercase tracking-wider mb-1">Current Bid</p>
        <p className="text-3xl font-black text-primary">{formatMoney(displayBid)}</p>
      </motion.div>

      <motion.div
        animate={{ scale: isUserHighest ? [1, 1.05, 1] : 1 }}
        transition={{ duration: 0.4 }}
        className={cn(
          "rounded-xl border p-4 text-center",
          isUserHighest ? "border-primary bg-primary/10" : "border-border bg-card"
        )}
      >
        <p className="text-xs text-muted-foreground uppercase tracking-wider mb-1">Highest Bidder</p>
        <p className="text-xl font-bold">
          {highestBidderName ? (
            <>{isUserHighest ? "🔥 You" : `⚡ ${highestBidderName}`}</>
          ) : (
            <span className="text-muted-foreground">—</span>
          )}
        </p>
      </motion.div>

      <div className="rounded-xl border border-border bg-card p-4">
        <p className="text-xs text-muted-foreground uppercase tracking-wider mb-2">Bid History</p>
        <div className="space-y-1 max-h-20 overflow-y-auto">
          <AnimatePresence mode="popLayout">
            {bidHistory.length === 0 ? (
              <p className="text-xs text-muted-foreground">No bids yet</p>
            ) : (
              [...bidHistory].reverse().slice(0, 5).map((b, i) => (
                <motion.p
                  key={`${b.timestamp}-${i}`}
                  initial={{ opacity: 0, x: -10 }}
                  animate={{ opacity: 1, x: 0 }}
                  className="text-xs text-muted-foreground"
                >
                  {b.participant_name} · {formatMoney(b.amount)}
                </motion.p>
              ))
            )}
          </AnimatePresence>
        </div>
      </div>
    </div>
  );
}
