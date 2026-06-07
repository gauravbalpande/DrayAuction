"use client";

import { useRouter } from "next/navigation";
import { motion } from "framer-motion";
import { Users, Brain, Trophy } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuctionStore, AUCTION_TIER_PREVIEW } from "@/stores";
import { api } from "@/lib/api";
import { cn } from "@/lib/utils";

const difficulties = [
  { id: "easy", label: "Easy", desc: "Casual managers — great for learning" },
  { id: "medium", label: "Medium", desc: "Position-aware AI opponents" },
  { id: "hard", label: "Hard", desc: "Budget-smart, squad-building AI" },
  { id: "legendary", label: "Legendary", desc: "Full optimization — elite challenge" },
];

export default function AuctionSetupPage() {
  const router = useRouter();
  const { setup, setSetup, setAuctionId } = useAuctionStore();
  const tierPreview = AUCTION_TIER_PREVIEW[setup.difficulty];

  const handleCreate = async () => {
    const res = await api.createAuction({
      ai_opponents: setup.aiOpponents,
      difficulty: setup.difficulty,
    });
    setAuctionId(res.id);
    router.push(`/auction/${res.id}/rulebook`);
  };

  return (
    <div className="min-h-screen container mx-auto px-4 py-10 max-w-2xl">
      <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }}>
        <h1 className="text-3xl font-bold mb-2">Enter the Auction Room</h1>
        <p className="text-muted-foreground mb-8">Choose your opponents. The game sets budget and player pool.</p>

        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <Users className="h-5 w-5 text-primary" /> AI Managers
              </CardTitle>
            </CardHeader>
            <CardContent className="flex gap-2 flex-wrap">
              {[1, 2, 3, 4, 5].map((n) => (
                <Button
                  key={n}
                  variant={setup.aiOpponents === n ? "default" : "outline"}
                  onClick={() => setSetup({ aiOpponents: n })}
                >
                  {n} Manager{n > 1 ? "s" : ""}
                </Button>
              ))}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <Brain className="h-5 w-5 text-primary" /> Difficulty
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              {difficulties.map((d) => (
                <button
                  key={d.id}
                  onClick={() => setSetup({ difficulty: d.id })}
                  className={cn(
                    "w-full text-left p-3 rounded-lg border transition-colors",
                    setup.difficulty === d.id
                      ? "border-primary bg-primary/10"
                      : "border-border hover:border-primary/40"
                  )}
                >
                  <p className="font-semibold">{d.label}</p>
                  <p className="text-xs text-muted-foreground">{d.desc}</p>
                </button>
              ))}
            </CardContent>
          </Card>

          {tierPreview && (
            <Card className="border-primary/30 bg-primary/5">
              <CardHeader>
                <CardTitle className="text-lg flex items-center gap-2">
                  <Trophy className="h-5 w-5 text-primary" /> Your Auction
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-3 gap-4 text-center">
                  <div>
                    <p className="text-xs text-muted-foreground">Type</p>
                    <p className="font-semibold text-sm">{tierPreview.type}</p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground">Budget</p>
                    <p className="font-semibold text-primary">{tierPreview.budget}</p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground">Player Pool</p>
                    <p className="font-semibold">{tierPreview.pool} players</p>
                  </div>
                </div>
                <p className="text-xs text-muted-foreground text-center mt-3">
                  Real footballers from Premier League, La Liga, Serie A, Bundesliga & Ligue 1
                </p>
              </CardContent>
            </Card>
          )}

          <Button size="xl" className="w-full" onClick={handleCreate}>
            Enter Auction Room
          </Button>
        </div>
      </motion.div>
    </div>
  );
}
