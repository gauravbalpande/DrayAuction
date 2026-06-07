"use client";

import { useEffect, useState, useRef } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { motion } from "framer-motion";
import { Trophy, ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { api } from "@/lib/api";

interface TeamResult {
  participant_id: string;
  name: string;
  type: string;
  formation: string;
  total_score: number;
  breakdown: Record<string, number>;
  strengths: string[];
  weaknesses: string[];
  squad_size: number;
}

interface AuctionResults {
  auction_id: string;
  winner_id: string;
  winner_name: string;
  teams: TeamResult[];
  ready: boolean;
  rewards?: { coins: number; xp: number; rank_points: number };
}

export default function ResultsPage() {
  const params = useParams();
  const router = useRouter();
  const auctionId = params.id as string;
  const [results, setResults] = useState<AuctionResults | null>(null);
  const [phase, setPhase] = useState<"waiting" | "ready" | "error">("waiting");
  const attemptsRef = useRef(0);

  useEffect(() => {
    let cancelled = false;
    const maxAttempts = 30;

    async function poll() {
      while (!cancelled && attemptsRef.current < maxAttempts) {
        attemptsRef.current++;
        try {
          const auction = await api.getAuction(auctionId);
          if (auction.status === "completed") {
            const data = await api.getResults(auctionId);
            if (!cancelled) {
              setResults(data as unknown as AuctionResults);
              setPhase("ready");
            }
            return;
          }
        } catch {
          try {
            const data = await api.getResults(auctionId);
            if (!cancelled) {
              setResults(data as unknown as AuctionResults);
              setPhase("ready");
            }
            return;
          } catch {
            /* keep polling */
          }
        }
        await new Promise((r) => setTimeout(r, 1000));
      }
      if (!cancelled) setPhase("error");
    }

    poll();
    return () => { cancelled = true; };
  }, [auctionId]);

  if (phase === "waiting") {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center gap-4">
        <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin" />
        <p className="text-muted-foreground">Calculating final scores...</p>
        <p className="text-xs text-muted-foreground">Analyzing squad balance, chemistry & formation fit</p>
      </div>
    );
  }

  if (phase === "error" || !results) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center gap-4">
        <p className="text-muted-foreground">Results are taking longer than expected.</p>
        <Button onClick={() => router.refresh()}>Retry</Button>
        <Button variant="outline" asChild><Link href="/dashboard">Back to Dashboard</Link></Button>
      </div>
    );
  }

  const humanTeam = results.teams.find((t) => t.type === "human");
  const isWin = results.winner_id === humanTeam?.participant_id;

  return (
    <div className="min-h-screen container mx-auto px-4 py-10 max-w-4xl">
      <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }}>
        <div className="text-center mb-10">
          <Trophy className={`h-16 w-16 mx-auto mb-4 ${isWin ? "text-primary" : "text-muted-foreground"}`} />
          <h1 className="text-4xl font-bold mb-2">{isWin ? "Victory!" : "Defeat"}</h1>
          <p className="text-muted-foreground">Winner: {results.winner_name}</p>
        </div>

        <div className="space-y-4 mb-10">
          {results.teams
            .sort((a, b) => b.total_score - a.total_score)
            .map((team, i) => (
              <Card key={team.participant_id} className={i === 0 ? "border-primary/50" : ""}>
                <CardContent className="pt-6 flex items-center justify-between">
                  <div>
                    <p className="font-semibold">{team.name} {team.type === "human" && "(You)"}</p>
                    <p className="text-sm text-muted-foreground">{team.formation} · {team.squad_size} players</p>
                  </div>
                  <p className="text-2xl font-bold">{team.total_score.toFixed(1)}</p>
                </CardContent>
              </Card>
            ))}
        </div>

        {humanTeam && (
          <div className="grid md:grid-cols-2 gap-6 mb-10">
            <Card>
              <CardHeader><CardTitle className="text-lg text-primary">Strengths</CardTitle></CardHeader>
              <CardContent>
                <ul className="space-y-1">
                  {humanTeam.strengths.map((s) => (
                    <li key={s} className="text-sm text-muted-foreground">✓ {s}</li>
                  ))}
                </ul>
              </CardContent>
            </Card>
            <Card>
              <CardHeader><CardTitle className="text-lg text-destructive">Weaknesses</CardTitle></CardHeader>
              <CardContent>
                <ul className="space-y-1">
                  {humanTeam.weaknesses.map((w) => (
                    <li key={w} className="text-sm text-muted-foreground">✗ {w}</li>
                  ))}
                </ul>
              </CardContent>
            </Card>
          </div>
        )}

        {humanTeam && (
          <Card className="mb-10">
            <CardHeader><CardTitle>Score Breakdown</CardTitle></CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
                {Object.entries(humanTeam.breakdown)
                  .filter(([k]) => k !== "total_score")
                  .map(([key, value]) => (
                    <div key={key}>
                      <p className="text-xs text-muted-foreground capitalize">{key.replace(/_/g, " ")}</p>
                      <p className="text-lg font-semibold">{typeof value === "number" ? value.toFixed(1) : value}</p>
                    </div>
                  ))}
              </div>
            </CardContent>
          </Card>
        )}

        <div className="flex gap-4 justify-center">
          <Button size="lg" asChild><Link href="/auction/setup">Play Again</Link></Button>
          <Button size="lg" variant="outline" asChild>
            <Link href="/dashboard"><ArrowLeft className="h-4 w-4 mr-2" />Dashboard</Link>
          </Button>
        </div>
      </motion.div>
    </div>
  );
}
