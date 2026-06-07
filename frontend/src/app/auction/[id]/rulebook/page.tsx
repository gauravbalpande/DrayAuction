"use client";

import { useParams, useRouter } from "next/navigation";
import { motion } from "framer-motion";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { api } from "@/lib/api";
import { useAuctionStore } from "@/stores";

export default function RulebookPage() {
  const params = useParams();
  const router = useRouter();
  const auctionId = params.id as string;
  const setState = useAuctionStore((s) => s.setState);

  const handleBegin = async () => {
    const state = await api.startAuction(auctionId);
    setState(state);
    router.push(`/auction/${auctionId}/live`);
  };

  return (
    <div className="min-h-screen container mx-auto px-4 py-10 max-w-3xl">
      <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }}>
        <h1 className="text-3xl font-bold mb-2">Auction Rulebook</h1>
        <p className="text-muted-foreground mb-8">Read the rules before you begin.</p>

        <div className="space-y-4 mb-10">
          <Card>
            <CardHeader><CardTitle className="text-lg">Auction Rules</CardTitle></CardHeader>
            <CardContent className="text-sm text-muted-foreground space-y-2">
              <p>• Soft timer: 15s to start, resets to 10s on every bid</p>
              <p>• Auction ends only when timer hits zero with no new bids</p>
              <p>• Minimum bid increment: 5M</p>
              <p>• First bid must be at least the player&apos;s market value</p>
              <p>• Highest bidder when timer expires wins the player</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader><CardTitle className="text-lg">Squad Requirements</CardTitle></CardHeader>
            <CardContent className="text-sm text-muted-foreground space-y-2">
              <p>• Minimum 11 players, maximum 15</p>
              <p>• Must include: GK, Defenders, Midfielders, Attackers</p>
              <p>• Formations: 4-3-3, 4-4-2, 3-5-2, 4-2-3-1</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader><CardTitle className="text-lg">Scoring System</CardTitle></CardHeader>
            <CardContent className="text-sm text-muted-foreground space-y-2">
              <p>• Attack, Midfield, Defense scores (not rating-only)</p>
              <p>• Chemistry from nation/club links</p>
              <p>• Formation fit, bench strength, squad depth</p>
              <p>• Highest total score wins</p>
            </CardContent>
          </Card>
        </div>

        <Button size="xl" className="w-full" onClick={handleBegin}>
          BEGIN AUCTION
        </Button>
      </motion.div>
    </div>
  );
}
