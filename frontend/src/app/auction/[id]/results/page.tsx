"use client";

import { useEffect, useState, useRef } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { motion } from "framer-motion";
import { Trophy, ArrowLeft, Award, TrendingUp, Zap, Target } from "lucide-react";
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

interface AwardInfo {
  player_name?: string;
  participant_name: string;
  details: string;
}

interface Awards {
  best_signing: AwardInfo;
  biggest_overpay: AwardInfo;
  best_chemistry: AwardInfo;
  most_efficient: AwardInfo;
}

interface AuctionResults {
  auction_id: string;
  winner_id: string;
  winner_name: string;
  teams: TeamResult[];
  awards?: Awards;
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
          const data = await api.getResults(auctionId);
          if (!cancelled && data.ready) {
            setResults(data as unknown as AuctionResults);
            setPhase("ready");
            return;
          }
        } catch {
          // Also try fetching auction status to check if completed
          try {
            const auction = await api.getAuction(auctionId);
            if (
              auction.status === "completed" ||
              auction.results_ready
            ) {
              const data = await api.getResults(auctionId);
              if (!cancelled && data.ready) {
                setResults(data as unknown as AuctionResults);
                setPhase("ready");
                return;
              }
            }
          } catch {
            /* keep polling */
          }
        }
        await new Promise((r) => setTimeout(r, 1000));
      }
      if (!cancelled) setPhase("error");
    }

    poll();
    return () => {
      cancelled = true;
    };
  }, [auctionId]);

  if (phase === "waiting") {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center gap-4">
        <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin" />
        <p className="text-muted-foreground">Calculating final scores...</p>
        <p className="text-xs text-muted-foreground">
          Analyzing squad balance, chemistry & formation fit
        </p>
      </div>
    );
  }

  if (phase === "error" || !results) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center gap-4">
        <p className="text-muted-foreground">
          Results are taking longer than expected.
        </p>
        <Button
          onClick={() => {
            attemptsRef.current = 0;
            setPhase("waiting");
          }}
        >
          Retry
        </Button>
        <Button variant="outline" asChild>
          <Link href="/dashboard">Back to Dashboard</Link>
        </Button>
      </div>
    );
  }

  const humanTeam = results.teams.find((t) => t.type === "human");
  const isWin = results.winner_id === humanTeam?.participant_id;
  const sortedTeams = [...results.teams].sort(
    (a, b) => b.total_score - a.total_score
  );

  const scoreBarColor = (score: number) => {
    if (score >= 75) return "bg-emerald-500";
    if (score >= 50) return "bg-amber-500";
    return "bg-red-500";
  };

  return (
    <div className="min-h-screen container mx-auto px-4 py-10 max-w-4xl">
      <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }}>
        {/* Hero Banner */}
        <div className="text-center mb-10">
          <Trophy
            className={`h-16 w-16 mx-auto mb-4 ${
              isWin ? "text-primary" : "text-muted-foreground"
            }`}
          />
          <h1 className="text-4xl font-bold mb-2">
            {isWin ? "🏆 Victory!" : "Defeat"}
          </h1>
          <p className="text-muted-foreground">Winner: {results.winner_name}</p>
        </div>

        {/* Progression Rewards */}
        {results.rewards && (
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ delay: 0.3 }}
          >
            <Card className="mb-8 border-primary/30 bg-primary/5">
              <CardContent className="pt-6">
                <div className="flex items-center justify-center gap-8 flex-wrap">
                  <div className="text-center">
                    <p className="text-xs text-muted-foreground uppercase tracking-wide">
                      Coins
                    </p>
                    <p className="text-2xl font-bold text-primary">
                      +{results.rewards.coins}
                    </p>
                  </div>
                  <div className="text-center">
                    <p className="text-xs text-muted-foreground uppercase tracking-wide">
                      XP
                    </p>
                    <p className="text-2xl font-bold text-primary">
                      +{results.rewards.xp}
                    </p>
                  </div>
                  <div className="text-center">
                    <p className="text-xs text-muted-foreground uppercase tracking-wide">
                      Rank Points
                    </p>
                    <p className="text-2xl font-bold text-primary">
                      +{results.rewards.rank_points}
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>
          </motion.div>
        )}

        {/* Leaderboard */}
        <div className="space-y-4 mb-10">
          {sortedTeams.map((team, i) => (
            <motion.div
              key={team.participant_id}
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: 0.1 * i }}
            >
              <Card
                className={
                  team.participant_id === results.winner_id
                    ? "border-primary/50"
                    : ""
                }
              >
                <CardContent className="pt-6 flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <span className="text-2xl font-bold text-muted-foreground w-8">
                      #{i + 1}
                    </span>
                    <div>
                      <p className="font-semibold">
                        {team.name}{" "}
                        {team.type === "human" && (
                          <span className="text-primary">(You)</span>
                        )}
                      </p>
                      <p className="text-sm text-muted-foreground">
                        {team.formation} · {team.squad_size} players
                      </p>
                    </div>
                  </div>
                  <p className="text-2xl font-bold">
                    {team.total_score.toFixed(1)}
                  </p>
                </CardContent>
              </Card>
            </motion.div>
          ))}
        </div>

        {/* Awards */}
        {results.awards && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.5 }}
          >
            <Card className="mb-10">
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Award className="h-5 w-5 text-primary" /> Awards
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div className="flex items-start gap-3 p-3 rounded-lg bg-accent/30">
                    <Target className="h-5 w-5 text-emerald-500 mt-0.5 shrink-0" />
                    <div>
                      <p className="font-medium text-sm">Best Signing</p>
                      <p className="text-xs text-muted-foreground">
                        {results.awards.best_signing.details}
                      </p>
                      <p className="text-xs text-primary mt-1">
                        — {results.awards.best_signing.participant_name}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-start gap-3 p-3 rounded-lg bg-accent/30">
                    <TrendingUp className="h-5 w-5 text-red-500 mt-0.5 shrink-0" />
                    <div>
                      <p className="font-medium text-sm">Biggest Overpay</p>
                      <p className="text-xs text-muted-foreground">
                        {results.awards.biggest_overpay.details}
                      </p>
                      <p className="text-xs text-primary mt-1">
                        — {results.awards.biggest_overpay.participant_name}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-start gap-3 p-3 rounded-lg bg-accent/30">
                    <Zap className="h-5 w-5 text-amber-500 mt-0.5 shrink-0" />
                    <div>
                      <p className="font-medium text-sm">Best Chemistry</p>
                      <p className="text-xs text-muted-foreground">
                        {results.awards.best_chemistry.details}
                      </p>
                      <p className="text-xs text-primary mt-1">
                        — {results.awards.best_chemistry.participant_name}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-start gap-3 p-3 rounded-lg bg-accent/30">
                    <Award className="h-5 w-5 text-blue-500 mt-0.5 shrink-0" />
                    <div>
                      <p className="font-medium text-sm">Most Efficient</p>
                      <p className="text-xs text-muted-foreground">
                        {results.awards.most_efficient.details}
                      </p>
                      <p className="text-xs text-primary mt-1">
                        — {results.awards.most_efficient.participant_name}
                      </p>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </motion.div>
        )}

        {/* Strengths & Weaknesses */}
        {humanTeam && (
          <div className="grid md:grid-cols-2 gap-6 mb-10">
            <Card>
              <CardHeader>
                <CardTitle className="text-lg text-primary">
                  Strengths
                </CardTitle>
              </CardHeader>
              <CardContent>
                <ul className="space-y-1">
                  {humanTeam.strengths.map((s) => (
                    <li key={s} className="text-sm text-muted-foreground">
                      ✓ {s}
                    </li>
                  ))}
                </ul>
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <CardTitle className="text-lg text-destructive">
                  Weaknesses
                </CardTitle>
              </CardHeader>
              <CardContent>
                <ul className="space-y-1">
                  {humanTeam.weaknesses.map((w) => (
                    <li key={w} className="text-sm text-muted-foreground">
                      ✗ {w}
                    </li>
                  ))}
                </ul>
              </CardContent>
            </Card>
          </div>
        )}

        {/* Score Breakdown with bars */}
        {humanTeam && (
          <Card className="mb-10">
            <CardHeader>
              <CardTitle>Score Breakdown</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                {Object.entries(humanTeam.breakdown)
                  .filter(([k]) => k !== "total_score")
                  .map(([key, value]) => (
                    <div key={key}>
                      <div className="flex justify-between text-sm mb-1">
                        <span className="text-muted-foreground capitalize">
                          {key.replace(/_/g, " ")}
                        </span>
                        <span className="font-semibold">
                          {typeof value === "number" ? value.toFixed(1) : value}
                        </span>
                      </div>
                      <div className="h-2 bg-accent rounded-full overflow-hidden">
                        <div
                          className={`h-full rounded-full transition-all duration-700 ${scoreBarColor(
                            typeof value === "number" ? value : 0
                          )}`}
                          style={{
                            width: `${Math.min(
                              typeof value === "number" ? value : 0,
                              100
                            )}%`,
                          }}
                        />
                      </div>
                    </div>
                  ))}
              </div>
            </CardContent>
          </Card>
        )}

        <div className="flex gap-4 justify-center">
          <Button size="lg" asChild>
            <Link href="/auction/setup">Play Again</Link>
          </Button>
          <Button size="lg" variant="outline" asChild>
            <Link href="/dashboard">
              <ArrowLeft className="h-4 w-4 mr-2" />
              Dashboard
            </Link>
          </Button>
        </div>
      </motion.div>
    </div>
  );
}
