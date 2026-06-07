"use client";

import Link from "next/link";
import { motion } from "framer-motion";
import { Trophy, Users, Brain, Zap } from "lucide-react";
import { Button } from "@/components/ui/button";

const features = [
  { icon: Brain, title: "AI Managers", desc: "Compete against 4 difficulty tiers of algorithmic football managers." },
  { icon: Zap, title: "Live Auctions", desc: "15-second bidding windows with real-time activity feeds." },
  { icon: Trophy, title: "Deep Scoring", desc: "Win on strategy — chemistry, formation fit, and squad depth matter." },
  { icon: Users, title: "Unique Every Time", desc: "Procedurally generated player pools ensure no two auctions feel the same." },
];

export default function LandingPage() {
  return (
    <div className="min-h-screen">
      <nav className="border-b border-border/50 backdrop-blur-sm sticky top-0 z-50">
        <div className="container mx-auto px-4 h-16 flex items-center justify-between">
          <span className="text-xl font-bold text-primary">AuctionXI</span>
          <div className="flex gap-3">
            <Button variant="ghost" asChild><Link href="/login">Login</Link></Button>
            <Button asChild><Link href="/register">Register</Link></Button>
          </div>
        </div>
      </nav>

      <section className="container mx-auto px-4 py-24 text-center">
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.6 }}>
          <p className="text-primary font-medium mb-4 tracking-widest uppercase text-sm">Football Auction Strategy Game</p>
          <h1 className="text-5xl md:text-7xl font-bold mb-6 leading-tight">
            Build Your Dream Squad.<br />
            <span className="text-primary">Outbid the AI.</span>
          </h1>
          <p className="text-muted-foreground text-lg max-w-2xl mx-auto mb-10">
            Enter live football player auctions. Manage your budget, read the room,
            and assemble a squad that wins on strategy — not just star ratings.
          </p>
          <div className="flex gap-4 justify-center flex-wrap">
            <Button size="xl" asChild><Link href="/register">Start Playing</Link></Button>
            <Button size="xl" variant="outline" asChild><Link href="/login">Login</Link></Button>
          </div>
        </motion.div>
      </section>

      <section className="container mx-auto px-4 py-20">
        <h2 className="text-3xl font-bold text-center mb-12">Why AuctionXI?</h2>
        <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6">
          {features.map((f, i) => (
            <motion.div
              key={f.title}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              transition={{ delay: i * 0.1 }}
              viewport={{ once: true }}
              className="rounded-xl border border-border bg-card p-6 hover:border-primary/50 transition-colors"
            >
              <f.icon className="h-8 w-8 text-primary mb-4" />
              <h3 className="font-semibold text-lg mb-2">{f.title}</h3>
              <p className="text-muted-foreground text-sm">{f.desc}</p>
            </motion.div>
          ))}
        </div>
      </section>

      <section className="container mx-auto px-4 py-20">
        <div className="rounded-2xl border border-border bg-card p-8 md:p-12 text-center">
          <Trophy className="h-12 w-12 text-primary mx-auto mb-4" />
          <h2 className="text-3xl font-bold mb-4">Climb the Rankings</h2>
          <p className="text-muted-foreground max-w-xl mx-auto mb-6">
            Earn coins, XP, and rank points. Progress from Bronze to Legend.
          </p>
          <div className="flex justify-center gap-4 flex-wrap text-sm">
            {["Bronze", "Silver", "Gold", "Platinum", "Diamond", "Legend"].map((rank) => (
              <span key={rank} className="px-3 py-1 rounded-full bg-secondary text-secondary-foreground">{rank}</span>
            ))}
          </div>
        </div>
      </section>

      <section className="container mx-auto px-4 py-20 text-center">
        <h2 className="text-3xl font-bold mb-6">Ready to Manage?</h2>
        <Button size="xl" asChild><Link href="/register">Create Free Account</Link></Button>
      </section>
    </div>
  );
}
