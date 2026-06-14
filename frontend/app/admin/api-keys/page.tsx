"use client";

import { useState } from "react";
import { Plus, Copy, Trash2 } from "lucide-react";
import { useAPIKeys, useCreateAPIKey, useRevokeAPIKey } from "@/lib/hooks/use-apikeys";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogDescription,
} from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import { toast } from "sonner";
import { formatDistanceToNow } from "date-fns";
import type { CreatedAPIKey } from "@/types";

export default function APIKeysPage() {
  const { data: keys, isLoading } = useAPIKeys();
  const createMutation = useCreateAPIKey();
  const revokeMutation = useRevokeAPIKey();

  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [createdKey, setCreatedKey] = useState<CreatedAPIKey | null>(null);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    const result = await createMutation.mutateAsync({ name });
    setCreatedKey(result);
    setOpen(false);
    setName("");
  }

  function copyKey() {
    if (!createdKey) return;
    navigator.clipboard.writeText(createdKey.key);
    toast.success("Copied to clipboard");
  }

  const active = keys?.filter((k) => !k.archived_at) ?? [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">API Keys</h1>
          <p className="text-sm text-muted-foreground mt-0.5">
            Bearer tokens for the public REST API
          </p>
        </div>
        <Button onClick={() => setOpen(true)} size="sm">
          <Plus className="h-4 w-4" />
          New Key
        </Button>
      </div>

      {isLoading ? (
        <div className="space-y-2">
          {[...Array(3)].map((_, i) => (
            <Skeleton key={i} className="h-14 w-full" />
          ))}
        </div>
      ) : active.length === 0 ? (
        <Card>
          <CardContent className="flex items-center justify-center py-12">
            <p className="text-muted-foreground text-sm">No API keys yet.</p>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-2">
          {active.map((k) => (
            <Card key={k.id}>
              <CardContent className="flex items-center gap-4 py-4">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="font-medium text-sm">{k.name}</span>
                    <Badge variant="outline" className="font-mono text-xs">
                      {k.prefix}••••••••
                    </Badge>
                  </div>
                  <p className="text-xs text-muted-foreground mt-0.5">
                    Created {formatDistanceToNow(new Date(k.created_at), { addSuffix: true })}
                    {k.last_used &&
                      ` · last used ${formatDistanceToNow(new Date(k.last_used), { addSuffix: true })}`}
                  </p>
                </div>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8 text-destructive hover:text-destructive"
                  onClick={() => revokeMutation.mutate(k.id)}
                  disabled={revokeMutation.isPending}
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Create dialog */}
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create API Key</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleCreate} className="space-y-4">
            <div className="space-y-1.5">
              <Label>Name</Label>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="CI integration"
                required
              />
            </div>
            <DialogFooter>
              <Button variant="outline" type="button" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button type="submit" disabled={createMutation.isPending}>
                Create
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Show key once dialog */}
      <Dialog open={!!createdKey} onOpenChange={(o) => !o && setCreatedKey(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>API Key Created</DialogTitle>
            <DialogDescription>
              Copy this key now. It will not be shown again.
            </DialogDescription>
          </DialogHeader>
          <div className="flex items-center gap-2 p-3 rounded-md bg-muted font-mono text-sm break-all">
            <span className="flex-1">{createdKey?.key}</span>
            <Button variant="ghost" size="icon" className="h-7 w-7 shrink-0" onClick={copyKey}>
              <Copy className="h-3.5 w-3.5" />
            </Button>
          </div>
          <DialogFooter>
            <Button onClick={() => setCreatedKey(null)}>Done</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
