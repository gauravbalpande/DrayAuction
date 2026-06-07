"use client";

import { motion } from "framer-motion";
import { cn, formatMoney } from "@/lib/utils";
import type { AuctionParticipant } from "@/lib/api";

interface ManagerCardProps {
  manager: AuctionParticipant;
  isActive: boolean;
  position: "top-left" | "top-right" | "bottom" | "center";
}

export function ManagerCard({ manager, isActive, position }: ManagerCardProps) {
  const isHuman = manager.type === "human";

  return (
    <motion.div
      layout
      animate={{
        scale: isActive ? 1.05 : 1,
        boxShadow: isActive
          ? "0 0 24px rgba(34, 197, 94, 0.5)"
          : "0 0 0px rgba(0,0,0,0)",
      }}
      transition={{ duration: 0.3 }}
      className={cn(
        "rounded-xl border p-4 min-w-[160px] max-w-[200px] transition-colors",
        isHuman ? "border-primary/60 bg-primary/5" : "border-border bg-card/80 backdrop-blur",
        isActive && "border-primary ring-2 ring-primary/40",
        position === "center" && "mx-auto"
      )}
    >
      <div className="flex items-center gap-2 mb-2">
        <div className={cn(
          "w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold",
          isHuman ? "bg-primary text-primary-foreground" : "bg-secondary"
        )}>
          {isHuman ? "YOU" : manager.name.charAt(0)}
        </div>
        <div className="min-w-0">
          <p className={cn("font-semibold text-sm truncate", isHuman && "text-primary")}>
            {manager.name}
          </p>
          {manager.style && (
            <p className="text-[10px] text-muted-foreground truncate">{manager.style}</p>
          )}
        </div>
      </div>

      {manager.targets && manager.targets.length > 0 && (
        <div className="mb-2">
          <p className="text-[10px] text-muted-foreground uppercase tracking-wide mb-1">Targets</p>
          <div className="flex flex-wrap gap-1">
            {manager.targets.map((t) => (
              <span key={t} className="text-[10px] px-1.5 py-0.5 rounded bg-secondary text-secondary-foreground">
                {t}
              </span>
            ))}
          </div>
        </div>
      )}

      <div className="flex justify-between text-xs mt-2 pt-2 border-t border-border/50">
        <span className="text-muted-foreground">Budget</span>
        <span className="font-semibold text-primary">{formatMoney(manager.remaining_budget)}</span>
      </div>
      <div className="flex justify-between text-xs">
        <span className="text-muted-foreground">Players</span>
        <span className="font-semibold">{manager.squad_size}</span>
      </div>
      {manager.has_passed && (
        <p className="text-[10px] text-destructive mt-1 text-center">Passed</p>
      )}
    </motion.div>
  );
}
