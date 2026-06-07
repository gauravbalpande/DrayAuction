"use client";

import { motion, AnimatePresence } from "framer-motion";
import type { ActivityEvent } from "@/lib/api";

interface LiveFeedProps {
  events: ActivityEvent[];
}

export function LiveFeed({ events }: LiveFeedProps) {
  return (
    <div className="rounded-xl border border-border bg-card/80 backdrop-blur">
      <div className="px-4 py-3 border-b border-border/50">
        <h3 className="text-sm font-semibold uppercase tracking-wider">Live Feed</h3>
      </div>
      <div className="p-3 max-h-48 overflow-y-auto space-y-2">
        <AnimatePresence mode="popLayout">
          {events.length === 0 ? (
            <p className="text-sm text-muted-foreground text-center py-4">Waiting for the auction to begin...</p>
          ) : (
            events.map((event, i) => (
              <motion.div
                key={`${event.timestamp}-${i}`}
                initial={{ opacity: 0, x: 20, height: 0 }}
                animate={{ opacity: 1, x: 0, height: "auto" }}
                exit={{ opacity: 0, height: 0 }}
                transition={{ duration: 0.25 }}
                className="text-sm py-1.5 px-2 rounded-lg bg-secondary/50 border-l-2 border-primary/40"
              >
                {event.message}
              </motion.div>
            ))
          )}
        </AnimatePresence>
      </div>
    </div>
  );
}
