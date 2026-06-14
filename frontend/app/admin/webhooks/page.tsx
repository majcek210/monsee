"use client";

import { useState } from "react";
import { Plus, MoreHorizontal, Pencil, Archive, List } from "lucide-react";
import {
  useWebhooks,
  useCreateWebhook,
  useUpdateWebhook,
  useArchiveWebhook,
  useWebhookLogs,
} from "@/lib/hooks/use-webhooks";
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
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
} from "@/components/ui/dropdown-menu";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { formatDistanceToNow } from "date-fns";
import type { Webhook } from "@/types";

const ALL_EVENTS = [
  "monitor.down",
  "monitor.recovered",
  "incident.created",
  "incident.resolved",
];

type DialogMode = "create" | "edit" | "logs" | null;

export default function WebhooksPage() {
  const { data: webhooks, isLoading } = useWebhooks();
  const createMutation = useCreateWebhook();
  const updateMutation = useUpdateWebhook();
  const archiveMutation = useArchiveWebhook();

  const [dialogMode, setDialogMode] = useState<DialogMode>(null);
  const [editing, setEditing] = useState<Webhook | null>(null);
  const [name, setName] = useState("");
  const [url, setUrl] = useState("");
  const [secret, setSecret] = useState("");
  const [events, setEvents] = useState<string[]>(ALL_EVENTS);
  const [enabled, setEnabled] = useState(true);
  const [logsWebhookId, setLogsWebhookId] = useState<string>("");

  const { data: logs } = useWebhookLogs(logsWebhookId);

  function openCreate() {
    setName(""); setUrl(""); setSecret("");
    setEvents(ALL_EVENTS); setEnabled(true);
    setEditing(null); setDialogMode("create");
  }

  function openEdit(w: Webhook) {
    setName(w.name); setUrl(""); setSecret("");
    setEvents(w.events); setEnabled(w.enabled);
    setEditing(w); setDialogMode("edit");
  }

  function openLogs(w: Webhook) {
    setLogsWebhookId(w.id);
    setDialogMode("logs");
  }

  function toggleEvent(ev: string) {
    setEvents((prev) =>
      prev.includes(ev) ? prev.filter((e) => e !== ev) : [...prev, ev]
    );
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (dialogMode === "create") {
      await createMutation.mutateAsync({ name, url, secret: secret || undefined, events, enabled });
    } else if (editing) {
      await updateMutation.mutateAsync({
        id: editing.id,
        name,
        url: url || undefined,
        secret: secret || undefined,
        events,
        enabled,
      });
    }
    setDialogMode(null);
  }

  const active = webhooks?.filter((w) => !w.archived_at) ?? [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Webhooks</h1>
          <p className="text-sm text-muted-foreground mt-0.5">
            Outgoing HTTP callbacks for platform events
          </p>
        </div>
        <Button onClick={openCreate} size="sm">
          <Plus className="h-4 w-4" />
          New Webhook
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
            <p className="text-muted-foreground text-sm">No webhooks configured.</p>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-2">
          {active.map((w) => (
            <Card key={w.id}>
              <CardContent className="flex items-center gap-4 py-4">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 flex-wrap">
                    <span className="font-medium text-sm">{w.name}</span>
                    {!w.enabled && <Badge variant="secondary">Disabled</Badge>}
                  </div>
                  <div className="flex gap-1 mt-1 flex-wrap">
                    {w.events.map((ev) => (
                      <Badge key={ev} variant="outline" className="text-xs">{ev}</Badge>
                    ))}
                  </div>
                </div>

                <Switch
                  checked={w.enabled}
                  onCheckedChange={(v) =>
                    updateMutation.mutate({ id: w.id, enabled: v })
                  }
                />

                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="icon" className="h-8 w-8">
                      <MoreHorizontal className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem onClick={() => openLogs(w)}>
                      <List className="h-4 w-4" />
                      Delivery Logs
                    </DropdownMenuItem>
                    <DropdownMenuItem onClick={() => openEdit(w)}>
                      <Pencil className="h-4 w-4" />
                      Edit
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem
                      className="text-destructive focus:text-destructive"
                      onClick={() => archiveMutation.mutate(w.id)}
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

      {/* Create/Edit dialog */}
      <Dialog
        open={dialogMode === "create" || dialogMode === "edit"}
        onOpenChange={(o) => !o && setDialogMode(null)}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {dialogMode === "create" ? "New Webhook" : "Edit Webhook"}
            </DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-1.5">
              <Label>Name</Label>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="My Webhook"
                required
              />
            </div>
            <div className="space-y-1.5">
              <Label>
                URL {dialogMode === "edit" && <span className="text-muted-foreground">(leave blank to keep)</span>}
              </Label>
              <Input
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                placeholder="https://example.com/webhook"
                required={dialogMode === "create"}
              />
            </div>
            <div className="space-y-1.5">
              <Label>Secret (optional HMAC)</Label>
              <Input
                value={secret}
                onChange={(e) => setSecret(e.target.value)}
                placeholder="Leave blank to keep existing"
              />
            </div>
            <div className="space-y-1.5">
              <Label>Events</Label>
              <div className="flex gap-2 flex-wrap">
                {ALL_EVENTS.map((ev) => (
                  <button
                    key={ev}
                    type="button"
                    onClick={() => toggleEvent(ev)}
                    className={`text-xs px-2 py-1 rounded border transition-colors ${
                      events.includes(ev)
                        ? "border-primary bg-primary/10 text-primary"
                        : "border-border text-muted-foreground"
                    }`}
                  >
                    {ev}
                  </button>
                ))}
              </div>
            </div>
            <div className="flex items-center gap-2">
              <Switch id="wh-enabled" checked={enabled} onCheckedChange={setEnabled} />
              <Label htmlFor="wh-enabled">Enabled</Label>
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

      {/* Logs dialog */}
      <Dialog open={dialogMode === "logs"} onOpenChange={(o) => !o && setDialogMode(null)}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Delivery Logs</DialogTitle>
          </DialogHeader>
          {!logs || logs.length === 0 ? (
            <p className="text-sm text-muted-foreground text-center py-6">No delivery logs yet.</p>
          ) : (
            <div className="max-h-96 overflow-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Event</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Duration</TableHead>
                    <TableHead>Delivered</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {logs.map((log) => (
                    <TableRow key={log.id}>
                      <TableCell className="text-xs font-mono">{log.event}</TableCell>
                      <TableCell>
                        {log.status_code ? (
                          <Badge
                            variant={log.status_code < 300 ? "success" : "danger"}
                          >
                            {log.status_code}
                          </Badge>
                        ) : (
                          <Badge variant="danger">error</Badge>
                        )}
                      </TableCell>
                      <TableCell className="text-xs text-muted-foreground">
                        {log.duration_ms != null ? `${log.duration_ms}ms` : "—"}
                      </TableCell>
                      <TableCell className="text-xs text-muted-foreground">
                        {formatDistanceToNow(new Date(log.delivered_at), { addSuffix: true })}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
