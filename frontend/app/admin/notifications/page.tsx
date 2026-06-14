"use client";

import { useState } from "react";
import { Plus, MoreHorizontal, Pencil, Archive } from "lucide-react";
import {
  useNotifications,
  useCreateNotification,
  useUpdateNotification,
  useArchiveNotification,
} from "@/lib/hooks/use-notifications";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from "@/components/ui/select";
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
} from "@/components/ui/dropdown-menu";
import type { NotificationChannel, NotificationType } from "@/types";

type DialogMode = "create" | "edit" | null;

export default function NotificationsPage() {
  const { data: channels, isLoading } = useNotifications();
  const createMutation = useCreateNotification();
  const updateMutation = useUpdateNotification();
  const archiveMutation = useArchiveNotification();

  const [dialogMode, setDialogMode] = useState<DialogMode>(null);
  const [editing, setEditing] = useState<NotificationChannel | null>(null);
  const [name, setName] = useState("");
  const [type, setType] = useState<NotificationType>("discord");
  const [webhookUrl, setWebhookUrl] = useState(""); // discord
  const [smtpTo, setSmtpTo] = useState(""); // email
  const [enabled, setEnabled] = useState(true);

  function openCreate() {
    setName("");
    setType("discord");
    setWebhookUrl("");
    setSmtpTo("");
    setEnabled(true);
    setEditing(null);
    setDialogMode("create");
  }

  function openEdit(ch: NotificationChannel) {
    setName(ch.name);
    setType(ch.type);
    setWebhookUrl("");
    setSmtpTo("");
    setEnabled(ch.enabled);
    setEditing(ch);
    setDialogMode("edit");
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const config: Record<string, string> =
      type === "discord"
        ? { webhook_url: webhookUrl }
        : { to: smtpTo };

    if (dialogMode === "create") {
      await createMutation.mutateAsync({ name, type, config, enabled });
    } else if (editing) {
      await updateMutation.mutateAsync({
        id: editing.id,
        name,
        enabled,
        config: webhookUrl || smtpTo ? config : undefined,
      });
    }
    setDialogMode(null);
  }

  const active = channels?.filter((c) => !c.archived_at) ?? [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Notifications</h1>
          <p className="text-sm text-muted-foreground mt-0.5">
            Alert channels for monitor down/recovery events
          </p>
        </div>
        <Button onClick={openCreate} size="sm">
          <Plus className="h-4 w-4" />
          New Channel
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
            <p className="text-muted-foreground text-sm">No notification channels configured.</p>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-2">
          {active.map((ch) => (
            <Card key={ch.id}>
              <CardContent className="flex items-center gap-4 py-4">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="font-medium text-sm">{ch.name}</span>
                    <Badge variant="outline" className="capitalize">{ch.type}</Badge>
                    {!ch.enabled && (
                      <Badge variant="secondary">Disabled</Badge>
                    )}
                  </div>
                </div>

                <Switch
                  checked={ch.enabled}
                  onCheckedChange={(v) =>
                    updateMutation.mutate({ id: ch.id, enabled: v })
                  }
                />

                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="icon" className="h-8 w-8">
                      <MoreHorizontal className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem onClick={() => openEdit(ch)}>
                      <Pencil className="h-4 w-4" />
                      Edit
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem
                      className="text-destructive focus:text-destructive"
                      onClick={() => archiveMutation.mutate(ch.id)}
                    >
                      <Archive className="h-4 w-4" />
                      Delete
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <Dialog open={dialogMode !== null} onOpenChange={(o) => !o && setDialogMode(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {dialogMode === "create" ? "New Notification Channel" : "Edit Channel"}
            </DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-1.5">
              <Label>Name</Label>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Discord Alerts"
                required
              />
            </div>

            {dialogMode === "create" && (
              <div className="space-y-1.5">
                <Label>Type</Label>
                <Select value={type} onValueChange={(v) => v && setType(v as NotificationType)}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="discord">Discord</SelectItem>
                    <SelectItem value="email">Email</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            )}

            {(dialogMode === "create" ? type : editing?.type) === "discord" ? (
              <div className="space-y-1.5">
                <Label>Discord Webhook URL</Label>
                <Input
                  value={webhookUrl}
                  onChange={(e) => setWebhookUrl(e.target.value)}
                  placeholder="https://discord.com/api/webhooks/..."
                  required={dialogMode === "create"}
                />
              </div>
            ) : (
              <div className="space-y-1.5">
                <Label>Recipient Email</Label>
                <Input
                  type="email"
                  value={smtpTo}
                  onChange={(e) => setSmtpTo(e.target.value)}
                  placeholder="alerts@example.com"
                  required={dialogMode === "create"}
                />
              </div>
            )}

            <div className="flex items-center gap-2">
              <Switch id="notif-enabled" checked={enabled} onCheckedChange={setEnabled} />
              <Label htmlFor="notif-enabled">Enabled</Label>
            </div>

            <DialogFooter>
              <Button variant="outline" type="button" onClick={() => setDialogMode(null)}>
                Cancel
              </Button>
              <Button
                type="submit"
                disabled={createMutation.isPending || updateMutation.isPending}
              >
                {dialogMode === "create" ? "Create" : "Save"}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
