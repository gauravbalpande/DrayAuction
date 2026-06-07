"use client";

import { useEffect } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { motion } from "framer-motion";
import { Coins, Star, Trophy, Swords } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuthStore } from "@/stores";

export default function DashboardPage() {
  const router = useRouter();
  const { user, isAuthenticated, logout } = useAuthStore();

  useEffect(() => {
    if (!isAuthenticated) router.push("/login");
  }, [isAuthenticated, router]);

  if (!user) return null;

  const stats = [
    { icon: Coins, label: "Coins", value: user.coins.toLocaleString() },
    { icon: Star, label: "XP", value: user.xp.toLocaleString() },
    { icon: Trophy, label: "Rank", value: user.rank.charAt(0).toUpperCase() + user.rank.slice(1) },
    { icon: Swords, label: "Record", value: `${user.wins}W / ${user.losses}L` },
  ];

  return (
    <div className="min-h-screen">
      <nav className="border-b border-border/50">
        <div className="container mx-auto px-4 h-16 flex items-center justify-between">
          <Link href="/dashboard" className="text-xl font-bold text-primary">AuctionXI</Link>
          <div className="flex items-center gap-4">
            <span className="text-sm text-muted-foreground">{user.username}</span>
            <Button variant="ghost" size="sm" onClick={() => { logout(); router.push("/"); }}>Logout</Button>
          </div>
        </div>
      </nav>

      <div className="container mx-auto px-4 py-10">
        <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }}>
          <h1 className="text-3xl font-bold mb-2">Welcome, {user.username}</h1>
          <p className="text-muted-foreground mb-8">Ready for your next auction?</p>
        </motion.div>

        <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-10">
          {stats.map((s, i) => (
            <motion.div key={s.label} initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: i * 0.05 }}>
              <Card>
                <CardContent className="pt-6 flex items-center gap-4">
                  <s.icon className="h-8 w-8 text-primary" />
                  <div>
                    <p className="text-sm text-muted-foreground">{s.label}</p>
                    <p className="text-xl font-bold">{s.value}</p>
                  </div>
                </CardContent>
              </Card>
            </motion.div>
          ))}
        </div>

        <Card className="mb-10">
          <CardHeader>
            <CardTitle>Recent Auctions</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-muted-foreground text-sm">No recent auctions yet. Start your first one!</p>
          </CardContent>
        </Card>

        <div className="text-center">
          <Button size="xl" asChild>
            <Link href="/auction/setup">START AUCTION</Link>
          </Button>
        </div>
      </div>
    </div>
  );
}
