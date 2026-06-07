"use client";

import { motion } from "framer-motion";
import { cn, formatMoney } from "@/lib/utils";
import type { AuctionPlayer } from "@/lib/api";

function StatBar({ label, value }: { label: string; value: number }) {
  return (
    <div>
      <div className="flex justify-between text-xs mb-1">
        <span className="text-muted-foreground">{label}</span>
        <span className="font-medium">{value}</span>
      </div>
      <div className="h-1.5 rounded-full bg-secondary overflow-hidden">
        <motion.div
          className="h-full bg-primary rounded-full"
          initial={{ width: 0 }}
          animate={{ width: `${value}%` }}
          transition={{ duration: 0.6, delay: 0.1 }}
        />
      </div>
    </div>
  );
}

interface PlayerSpotlightProps {
  player: AuctionPlayer;
}

export function PlayerSpotlight({ player }: PlayerSpotlightProps) {
  const formColor = {
    Excellent: "text-primary",
    Good: "text-green-400",
    Average: "text-yellow-400",
    Poor: "text-destructive",
  }[player.form_label] ?? "text-muted-foreground";

  return (
    <motion.div
      key={player.id}
      initial={{ opacity: 0, y: 20, scale: 0.95 }}
      animate={{ opacity: 1, y: 0, scale: 1 }}
      exit={{ opacity: 0, y: -20, scale: 0.95 }}
      className="rounded-2xl border-2 border-primary/30 bg-card/90 backdrop-blur p-6 shadow-lg shadow-primary/5"
    >
      <div className="text-center mb-4">
        <span className="inline-block px-3 py-1 rounded-full bg-primary/20 text-primary text-xs font-semibold mb-2">
          {player.position} · {player.league}
        </span>
        <h2 className="text-3xl md:text-4xl font-bold tracking-tight">{player.name}</h2>
        <p className="text-muted-foreground mt-1">{player.club} · {player.nation} · Age {player.age}</p>
      </div>

      <div className="flex justify-center gap-8 mb-6">
        <div className="text-center">
          <p className="text-5xl font-black text-primary">{player.rating}</p>
          <p className="text-xs text-muted-foreground uppercase tracking-wider">Rating</p>
        </div>
        <div className="text-center">
          <p className="text-2xl font-bold">{formatMoney(player.market_value)}</p>
          <p className="text-xs text-muted-foreground uppercase tracking-wider">Market Value</p>
        </div>
        <div className="text-center">
          <p className={cn("text-lg font-bold", formColor)}>{player.form_label}</p>
          <p className="text-xs text-muted-foreground uppercase tracking-wider">Form ({player.form})</p>
        </div>
      </div>

      <div className="border-t border-border/50 pt-4">
        <p className="text-xs text-muted-foreground uppercase tracking-wider mb-3 text-center">Scouting Report</p>
        <div className="grid grid-cols-2 gap-3">
          <StatBar label="Attack" value={player.attack} />
          <StatBar label="Passing" value={player.passing} />
          <StatBar label="Defending" value={player.defending} />
          <StatBar label="Physical" value={player.physical} />
        </div>
      </div>
    </motion.div>
  );
}
