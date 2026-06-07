"use client";

import { motion } from "framer-motion";
import { cn } from "@/lib/utils";

interface SoftTimerProps {
  seconds: number;
  maxSeconds: number;
}

export function SoftTimer({ seconds, maxSeconds }: SoftTimerProps) {
  const pct = maxSeconds > 0 ? (seconds / maxSeconds) * 100 : 0;
  const urgent = seconds <= 5;

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <span className="text-xs text-muted-foreground uppercase tracking-wider">
          Soft Timer {urgent && "· Closing!"}
        </span>
        <motion.span
          key={seconds}
          initial={{ scale: 1.3, color: urgent ? "#ef4444" : undefined }}
          animate={{ scale: 1 }}
          className={cn("text-2xl font-black tabular-nums", urgent && "text-destructive")}
        >
          {seconds}s
        </motion.span>
      </div>
      <div className="relative h-3 rounded-full bg-secondary overflow-hidden">
        <motion.div
          className={cn("h-full rounded-full", urgent ? "bg-destructive" : "bg-primary")}
          animate={{ width: `${pct}%` }}
          transition={{ duration: 0.4, ease: "easeOut" }}
        />
        {seconds <= 10 && seconds > 0 && (
          <motion.div
            className="absolute inset-0 bg-primary/20"
            animate={{ opacity: [0.3, 0.6, 0.3] }}
            transition={{ repeat: Infinity, duration: 1 }}
          />
        )}
      </div>
      <p className="text-[10px] text-muted-foreground text-center">
        Resets to 10s on every bid
      </p>
    </div>
  );
}
