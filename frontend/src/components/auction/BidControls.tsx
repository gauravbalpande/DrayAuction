"use client";

import { Gavel } from "lucide-react";
import { Button } from "@/components/ui/button";

interface BidControlsProps {
  disabled: boolean;
  onBid: (increment: number) => void;
  onPass: () => void;
}

export function BidControls({ disabled, onBid, onPass }: BidControlsProps) {
  return (
    <div className="grid grid-cols-4 gap-3">
      {[5, 10, 20].map((inc) => (
        <Button
          key={inc}
          size="lg"
          disabled={disabled}
          onClick={() => onBid(inc)}
          className="font-bold text-base"
        >
          <Gavel className="h-4 w-4 mr-1" />+{inc}M
        </Button>
      ))}
      <Button size="lg" variant="outline" disabled={disabled} onClick={onPass} className="font-bold">
        Pass
      </Button>
    </div>
  );
}
