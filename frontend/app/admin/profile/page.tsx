"use client";

import { useEffect, useState } from "react";
import { ShieldCheck, ShieldOff, Save } from "lucide-react";
import { useCurrentUser } from "@/lib/hooks/use-auth";
import { useUpdateUser } from "@/lib/hooks/use-users";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { api } from "@/lib/api/client";
import { toast } from "sonner";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";

export default function ProfilePage() {
  const { data: me, refetch } = useCurrentUser();
  const updateMutation = useUpdateUser();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  useEffect(() => {
    if (me) setEmail(me.email);
  }, [me]);

  const [setupOpen, setSetupOpen] = useState(false);
  const [disableOpen, setDisableOpen] = useState(false);

  const [setupStep, setSetupStep] = useState<"init" | "confirm">("init");
  const [secret, setSecret] = useState("");
  const [otpauthUri, setOtpauthUri] = useState("");
  const [confirmCode, setConfirmCode] = useState("");
  const [backupCodes, setBackupCodes] = useState<string[]>([]);
  const [disablePassword, setDisablePassword] = useState("");

  async function handleAccountSave(e: React.FormEvent) {
    e.preventDefault();
    if (!me) return;
    await updateMutation.mutateAsync({
      id: me.id,
      email: email !== me.email ? email : undefined,
      password: password || undefined,
    });
    setPassword("");
    refetch();
  }

  async function initiateSetup() {
    try {
      const res = await api.post<{ secret: string; otpauth_uri: string }>("/admin/2fa/setup", {});
      setSecret(res.secret);
      setOtpauthUri(res.otpauth_uri);
      setSetupStep("confirm");
    } catch (e) {
      toast.error((e as Error).message);
    }
  }

  async function confirmSetup(e: React.FormEvent) {
    e.preventDefault();
    try {
      const res = await api.post<{ backup_codes: string[] }>("/admin/2fa/confirm", { code: confirmCode });
      setBackupCodes(res.backup_codes);
      setSetupStep("init");
      setConfirmCode("");
      refetch();
      toast.success("2FA enabled");
    } catch (e) {
      toast.error((e as Error).message);
    }
  }

  async function disableTwoFactor(e: React.FormEvent) {
    e.preventDefault();
    try {
      await api.post("/admin/2fa/disable", { password: disablePassword });
      setDisableOpen(false);
      setDisablePassword("");
      refetch();
      toast.success("2FA disabled");
    } catch (e) {
      toast.error((e as Error).message);
    }
  }

  const totpEnabled = me?.totp_enabled ?? false;

  return (
    <div className="space-y-6 max-w-2xl">
      <div>
        <h1 className="text-2xl font-semibold">My Profile</h1>
        <p className="text-sm text-muted-foreground mt-0.5">
          Account details and two-factor authentication
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Account</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleAccountSave} className="space-y-4">
            <div className="space-y-1.5">
              <Label htmlFor="profile-email">Email</Label>
              <Input
                id="profile-email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="profile-password">New Password</Label>
              <Input
                id="profile-password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Leave blank to keep your current password"
              />
            </div>
            <Button type="submit" disabled={updateMutation.isPending}>
              <Save className="h-4 w-4" />
              Save
            </Button>
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Two-Factor Authentication</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-start gap-3">
            {totpEnabled ? (
              <ShieldCheck className="h-5 w-5 text-emerald-400 mt-0.5" />
            ) : (
              <ShieldOff className="h-5 w-5 text-muted-foreground mt-0.5" />
            )}
            <div className="flex-1">
              <p className="text-sm font-medium">{totpEnabled ? "2FA is enabled" : "2FA is not enabled"}</p>
              <p className="text-xs text-muted-foreground mt-0.5">
                {totpEnabled
                  ? "Your account is protected with a TOTP authenticator app."
                  : "Add an extra layer of security to your account with a TOTP app."}
              </p>
            </div>
            {totpEnabled ? (
              <Button variant="outline" size="sm" onClick={() => setDisableOpen(true)}>
                Disable
              </Button>
            ) : (
              <Button size="sm" onClick={() => { setSetupOpen(true); initiateSetup(); }}>
                Enable 2FA
              </Button>
            )}
          </div>

          {backupCodes.length > 0 && (
            <div className="rounded-lg bg-muted p-4 space-y-2">
              <p className="text-sm font-medium">Backup Codes — save these now</p>
              <p className="text-xs text-muted-foreground">Each code can only be used once.</p>
              <div className="grid grid-cols-2 gap-1 font-mono text-xs">
                {backupCodes.map((code) => (
                  <span key={code} className="bg-background rounded px-2 py-1">{code}</span>
                ))}
              </div>
              <Button variant="outline" size="sm" onClick={() => setBackupCodes([])}>
                I&apos;ve saved these
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Setup Dialog */}
      <Dialog open={setupOpen} onOpenChange={(o) => { if (!o) { setSetupOpen(false); setSetupStep("init"); setConfirmCode(""); } }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Enable Two-Factor Authentication</DialogTitle>
          </DialogHeader>
          {setupStep === "confirm" ? (
            <form onSubmit={confirmSetup} className="space-y-4">
              <p className="text-sm text-muted-foreground">
                Scan this secret with your authenticator app, then enter a code to confirm.
              </p>
              <div className="rounded bg-muted px-3 py-2 font-mono text-sm break-all">{secret}</div>
              <p className="text-xs text-muted-foreground">
                Or use this URI: <span className="break-all">{otpauthUri}</span>
              </p>
              <div className="space-y-1.5">
                <Label>Verification Code</Label>
                <Input
                  value={confirmCode}
                  onChange={(e) => setConfirmCode(e.target.value)}
                  placeholder="000000"
                  maxLength={6}
                  required
                />
              </div>
              <DialogFooter>
                <Button variant="outline" type="button" onClick={() => setSetupOpen(false)}>Cancel</Button>
                <Button type="submit" disabled={confirmCode.length < 6}>Verify & Enable</Button>
              </DialogFooter>
            </form>
          ) : (
            <div className="py-4 text-center text-muted-foreground text-sm">Generating secret...</div>
          )}
        </DialogContent>
      </Dialog>

      {/* Disable Dialog */}
      <Dialog open={disableOpen} onOpenChange={setDisableOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Disable Two-Factor Authentication</DialogTitle>
          </DialogHeader>
          <form onSubmit={disableTwoFactor} className="space-y-4">
            <p className="text-sm text-muted-foreground">Enter your password to confirm.</p>
            <div className="space-y-1.5">
              <Label>Password</Label>
              <Input
                type="password"
                value={disablePassword}
                onChange={(e) => setDisablePassword(e.target.value)}
                required
              />
            </div>
            <DialogFooter>
              <Button variant="outline" type="button" onClick={() => setDisableOpen(false)}>Cancel</Button>
              <Button variant="destructive" type="submit" disabled={!disablePassword}>Disable 2FA</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
