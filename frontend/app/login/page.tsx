"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Image from "next/image";
import { authApi } from "@/lib/api/auth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { toast } from "sonner";

export default function LoginPage() {
  const router = useRouter();

  // Phase 1 — password auth
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  // Phase 2 — TOTP verify
  const [phase, setPhase] = useState<"password" | "totp">("password");
  const [pendingUserId, setPendingUserId] = useState("");
  const [totpCode, setTotpCode] = useState("");

  const [loading, setLoading] = useState(false);

  async function handlePasswordSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    try {
      const user = await authApi.login(email, password);
      if (user.totp_enabled) {
        setPendingUserId(user.id);
        setPhase("totp");
      } else {
        router.push("/admin");
      }
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : "Login failed");
    } finally {
      setLoading(false);
    }
  }

  async function handleTotpSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    try {
      await authApi.verify2FA(pendingUserId, totpCode);
      router.push("/admin");
    } catch (err: unknown) {
      toast.error(
        err instanceof Error ? err.message : "Invalid code, please try again"
      );
    } finally {
      setLoading(false);
    }
  }

  function handleBackToPassword() {
    setPhase("password");
    setTotpCode("");
    setPendingUserId("");
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-background p-4">
      <Card className="w-full max-w-sm">
        <CardHeader className="text-center">
          <div className="flex justify-center mb-3">
            <div className="flex items-center justify-center h-10 w-10 rounded-lg bg-primary/10">
              <Image src="/monsee.png" alt="monsee" width={24} height={24} className="rounded" />
            </div>
          </div>
          {phase === "password" ? (
            <>
              <CardTitle>Admin Login</CardTitle>
              <CardDescription>Sign in to manage your monitoring platform</CardDescription>
            </>
          ) : (
            <>
              <CardTitle>Two-Factor Authentication</CardTitle>
              <CardDescription>Enter the 6-digit code from your authenticator app</CardDescription>
            </>
          )}
        </CardHeader>
        <CardContent>
          {phase === "password" ? (
            <form onSubmit={handlePasswordSubmit} className="space-y-4">
              <div className="space-y-1.5">
                <Label htmlFor="email">Email</Label>
                <Input
                  id="email"
                  type="email"
                  placeholder="admin@example.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  autoComplete="email"
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="password">Password</Label>
                <Input
                  id="password"
                  type="password"
                  placeholder="••••••••"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  autoComplete="current-password"
                />
              </div>
              <Button type="submit" className="w-full" disabled={loading}>
                {loading ? "Signing in…" : "Sign in"}
              </Button>
            </form>
          ) : (
            <form onSubmit={handleTotpSubmit} className="space-y-4">
              <div className="space-y-1.5">
                <Label htmlFor="totp-code">Authentication Code</Label>
                <Input
                  id="totp-code"
                  type="text"
                  inputMode="numeric"
                  placeholder="000000"
                  value={totpCode}
                  onChange={(e) => setTotpCode(e.target.value)}
                  required
                  autoComplete="one-time-code"
                  maxLength={32}
                />
              </div>
              <Button type="submit" className="w-full" disabled={loading}>
                {loading ? "Verifying…" : "Verify"}
              </Button>
              <Button
                type="button"
                variant="ghost"
                className="w-full"
                onClick={handleBackToPassword}
                disabled={loading}
              >
                Back to sign in
              </Button>
            </form>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
