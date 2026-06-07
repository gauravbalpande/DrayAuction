"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { motion } from "framer-motion";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuthStore } from "@/stores";

export default function LoginPage() {
  const router = useRouter();
  const login = useAuthStore((s) => s.login);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      await login(email, password);
      router.push("/dashboard");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center px-4">
      <motion.div initial={{ opacity: 0, scale: 0.95 }} animate={{ opacity: 1, scale: 1 }} className="w-full max-w-md">
        <Card>
          <CardHeader>
            <CardTitle className="text-center">Welcome Back</CardTitle>
            <p className="text-center text-muted-foreground text-sm">Login to AuctionXI</p>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-4">
              {error && <p className="text-destructive text-sm text-center">{error}</p>}
              <input
                type="email" placeholder="Email" value={email} onChange={(e) => setEmail(e.target.value)} required
                className="w-full h-10 px-3 rounded-md bg-secondary border border-border text-foreground"
              />
              <input
                type="password" placeholder="Password" value={password} onChange={(e) => setPassword(e.target.value)} required
                className="w-full h-10 px-3 rounded-md bg-secondary border border-border text-foreground"
              />
              <Button type="submit" className="w-full" disabled={loading}>
                {loading ? "Logging in..." : "Login"}
              </Button>
            </form>
            <p className="text-center text-sm text-muted-foreground mt-4">
              No account? <Link href="/register" className="text-primary hover:underline">Register</Link>
            </p>
          </CardContent>
        </Card>
      </motion.div>
    </div>
  );
}
