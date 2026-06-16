"use client";

import { useState } from "react";
import { use } from "react";
import Link from "next/link";
import { ChevronLeft, Plus, MoreHorizontal, Pencil, Archive } from "lucide-react";
import { useService } from "@/lib/hooks/use-services";
import {
  useMonitors,
  useCreateMonitor,
  useUpdateMonitor,
  useArchiveMonitor,
} from "@/lib/hooks/use-monitors";
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
import { Switch } from "@/components/ui/switch";
import { MonitorStatusBadge } from "@/components/admin/status-badge";
import type { Monitor } from "@/types";

type DialogMode = "create" | "edit" | null;

const defaultForm = {
  name: "",
  type: "http" as import("@/types").MonitorType,
  url: "",
  host: "",
  port: "",
  interval_seconds: "60",
  timeout_ms: "5000",
  retry_count: "2",
  degraded_threshold_ms: "",
  http_method: "GET",
  http_expected_status: "200",
  enabled: true,
};

export default function ServiceMonitorsPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const { data: service } = useService(id);
  const { data: monitors, isLoading } = useMonitors(id);
  const createMutation = useCreateMonitor(id);
  const updateMutation = useUpdateMonitor(id);
  const archiveMutation = useArchiveMonitor(id);

  const [dialogMode, setDialogMode] = useState<DialogMode>(null);
  const [editing, setEditing] = useState<Monitor | null>(null);
  const [form, setForm] = useState(defaultForm);

  function set(k: keyof typeof defaultForm, v: string | boolean) {
    setForm((p) => ({ ...p, [k]: v }));
  }

  function openCreate() {
    setForm(defaultForm);
    setEditing(null);
    setDialogMode("create");
  }

  function openEdit(m: Monitor) {
    setForm({
      name: m.name,
      type: m.type,
      url: m.url ?? "",
      host: m.host ?? "",
      port: m.port?.toString() ?? "",
      interval_seconds: m.interval_seconds.toString(),
      timeout_ms: m.timeout_ms.toString(),
      retry_count: m.retry_count.toString(),
      degraded_threshold_ms: m.degraded_threshold_ms?.toString() ?? "",
      http_method: m.http_method ?? "GET",
      http_expected_status: m.http_expected_status?.toString() ?? "200",
      enabled: m.enabled,
    });
    setEditing(m);
    setDialogMode("edit");
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const payload = {
      name: form.name,
      type: form.type,
      url: form.type === "http" ? form.url : undefined,
      host: form.type === "tcp" ? form.host : undefined,
      port: form.type === "tcp" && form.port ? parseInt(form.port) : undefined,
      interval_seconds: parseInt(form.interval_seconds),
      timeout_ms: parseInt(form.timeout_ms),
      retry_count: parseInt(form.retry_count),
      degraded_threshold_ms: form.degraded_threshold_ms
        ? parseInt(form.degraded_threshold_ms)
        : undefined,
      http_method: form.type === "http" ? form.http_method : undefined,
      http_expected_status: form.type === "http" ? parseInt(form.http_expected_status) : undefined,
      enabled: form.enabled,
    };
    if (dialogMode === "create") {
      await createMutation.mutateAsync(payload);
    } else if (editing) {
      await updateMutation.mutateAsync({ id: editing.id, ...payload });
    }
    setDialogMode(null);
  }

  const active = monitors?.filter((m) => !m.archived_at) ?? [];
  const lastStatus = (m: Monitor): "up" | "down" | "degraded" =>
    m.consecutive_failures > 0 ? "down" : "up";

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Link href="/admin/services">
          <Button variant="ghost" size="icon" className="h-8 w-8">
            <ChevronLeft className="h-4 w-4" />
          </Button>
        </Link>
        <div className="flex-1">
          <h1 className="text-2xl font-semibold">{service?.name ?? "Service"}</h1>
          <p className="text-sm text-muted-foreground mt-0.5">Monitors</p>
        </div>
        <Button onClick={openCreate} size="sm">
          <Plus className="h-4 w-4" />
          New Monitor
        </Button>
      </div>

      {isLoading ? (
        <div className="space-y-2">
          {[...Array(3)].map((_, i) => (
            <Skeleton key={i} className="h-16 w-full" />
          ))}
        </div>
      ) : active.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12 text-center">
            <p className="text-muted-foreground text-sm">No monitors yet.</p>
            <Button className="mt-4" size="sm" onClick={openCreate}>
              <Plus className="h-4 w-4" />
              Add monitor
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-2">
          {active.map((m) => (
            <Card key={m.id}>
              <CardContent className="flex items-center gap-4 py-4">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="font-medium text-sm">{m.name}</span>
                    <MonitorStatusBadge status={lastStatus(m)} />
                    <span className="text-xs text-muted-foreground uppercase">{m.type}</span>
                    {!m.enabled && (
                      <span className="text-xs text-muted-foreground">(disabled)</span>
                    )}
                  </div>
                  <p className="text-xs text-muted-foreground mt-0.5 truncate">
                    {m.url ?? `${m.host}:${m.port}`} · every {m.interval_seconds}s
                  </p>
                </div>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="icon" className="h-8 w-8">
                      <MoreHorizontal className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem onClick={() => openEdit(m)}>
                      <Pencil className="h-4 w-4" />
                      Edit
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem
                      className="text-destructive focus:text-destructive"
                      onClick={() => archiveMutation.mutate(m.id)}
                    >
                      <Archive className="h-4 w-4" />
                      Archive
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <Dialog open={dialogMode !== null} onOpenChange={(o) => !o && setDialogMode(null)}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>{dialogMode === "create" ? "New Monitor" : "Edit Monitor"}</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5 col-span-2">
                <Label>Name</Label>
                <Input
                  value={form.name}
                  onChange={(e) => set("name", e.target.value)}
                  placeholder="API Health"
                  required
                />
              </div>

              <div className="space-y-1.5">
                <Label>Type</Label>
                <Select value={form.type} onValueChange={(v) => v && set("type", v)}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="http">HTTP</SelectItem>
                    <SelectItem value="tcp">TCP</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-1.5">
                <Label>Interval (seconds)</Label>
                <Input
                  type="number"
                  min={10}
                  value={form.interval_seconds}
                  onChange={(e) => set("interval_seconds", e.target.value)}
                />
              </div>

              {form.type === "http" ? (
                <>
                  <div className="space-y-1.5 col-span-2">
                    <Label>URL</Label>
                    <Input
                      value={form.url}
                      onChange={(e) => set("url", e.target.value)}
                      placeholder="https://api.example.com/health"
                      required
                    />
                  </div>
                  <div className="space-y-1.5">
                    <Label>HTTP Method</Label>
                    <Select value={form.http_method} onValueChange={(v) => v && set("http_method", v)}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {["GET", "POST", "HEAD"].map((m) => (
                          <SelectItem key={m} value={m}>{m}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-1.5">
                    <Label>Expected Status</Label>
                    <Input
                      type="number"
                      value={form.http_expected_status}
                      onChange={(e) => set("http_expected_status", e.target.value)}
                    />
                  </div>
                </>
              ) : (
                <>
                  <div className="space-y-1.5">
                    <Label>Host</Label>
                    <Input
                      value={form.host}
                      onChange={(e) => set("host", e.target.value)}
                      placeholder="db.example.com"
                      required
                    />
                  </div>
                  <div className="space-y-1.5">
                    <Label>Port</Label>
                    <Input
                      type="number"
                      value={form.port}
                      onChange={(e) => set("port", e.target.value)}
                      placeholder="5432"
                      required
                    />
                  </div>
                </>
              )}

              <div className="space-y-1.5">
                <Label>Timeout (ms)</Label>
                <Input
                  type="number"
                  value={form.timeout_ms}
                  onChange={(e) => set("timeout_ms", e.target.value)}
                />
              </div>
              <div className="space-y-1.5">
                <Label>Retry Count</Label>
                <Input
                  type="number"
                  min={0}
                  value={form.retry_count}
                  onChange={(e) => set("retry_count", e.target.value)}
                />
              </div>
              <div className="space-y-1.5 col-span-2">
                <Label>Degraded Threshold (ms, optional)</Label>
                <Input
                  type="number"
                  value={form.degraded_threshold_ms}
                  onChange={(e) => set("degraded_threshold_ms", e.target.value)}
                  placeholder="Leave empty to disable"
                />
              </div>

              <div className="col-span-2 flex items-center gap-2">
                <Switch
                  id="enabled"
                  checked={form.enabled}
                  onCheckedChange={(v) => set("enabled", v)}
                />
                <Label htmlFor="enabled">Enabled</Label>
              </div>
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
