"use client";

import { useState } from "react";
import { Plus, Trash2, Wrench } from "lucide-react";
import {
  useMaintenanceWindows,
  useCreateMaintenanceWindow,
  useArchiveMaintenanceWindow,
} from "@/lib/hooks/use-maintenance";
import { useServices } from "@/lib/hooks/use-services";
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
import { Textarea } from "@/components/ui/textarea";
import { formatDistanceToNow, format } from "date-fns";

function isActive(mw: { starts_at: string; ends_at: string }) {
  const now = Date.now();
  return new Date(mw.starts_at).getTime() <= now && new Date(mw.ends_at).getTime() >= now;
}

export default function MaintenancePage() {
  const { data: windows, isLoading } = useMaintenanceWindows();
  const { data: services } = useServices();
  const createMutation = useCreateMaintenanceWindow();
  const archiveMutation = useArchiveMaintenanceWindow();

  const [open, setOpen] = useState(false);
  const [serviceId, setServiceId] = useState("");
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [startsAt, setStartsAt] = useState("");
  const [endsAt, setEndsAt] = useState("");

  function toISO(local: string) {
    return new Date(local).toISOString();
  }

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    await createMutation.mutateAsync({
      service_id: serviceId,
      title,
      description: description || undefined,
      starts_at: toISO(startsAt),
      ends_at: toISO(endsAt),
    });
    setOpen(false);
    setTitle("");
    setDescription("");
    setStartsAt("");
    setEndsAt("");
    setServiceId("");
  }

  const active = windows?.filter((w) => !w.archived_at && isActive(w)) ?? [];
  const upcoming = windows?.filter((w) => !w.archived_at && new Date(w.starts_at) > new Date()) ?? [];
  const past = windows?.filter((w) => !w.archived_at && new Date(w.ends_at) < new Date() && !isActive(w)) ?? [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Maintenance Windows</h1>
          <p className="text-sm text-muted-foreground mt-0.5">
            {active.length} active · {upcoming.length} upcoming
          </p>
        </div>
        <Button onClick={() => setOpen(true)} size="sm">
          <Plus className="h-4 w-4" />
          New Window
        </Button>
      </div>

      {isLoading ? (
        <div className="space-y-2">
          {[...Array(3)].map((_, i) => <Skeleton key={i} className="h-16 w-full" />)}
        </div>
      ) : windows?.filter((w) => !w.archived_at).length === 0 ? (
        <Card>
          <CardContent className="flex items-center justify-center py-12">
            <p className="text-muted-foreground text-sm">No maintenance windows scheduled.</p>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-4">
          {active.length > 0 && (
            <section>
              <h2 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-2">Active</h2>
              <div className="space-y-2">
                {active.map((mw) => <MaintenanceCard key={mw.id} mw={mw} services={services} onArchive={archiveMutation.mutate} />)}
              </div>
            </section>
          )}
          {upcoming.length > 0 && (
            <section>
              <h2 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-2">Upcoming</h2>
              <div className="space-y-2">
                {upcoming.map((mw) => <MaintenanceCard key={mw.id} mw={mw} services={services} onArchive={archiveMutation.mutate} />)}
              </div>
            </section>
          )}
          {past.length > 0 && (
            <section>
              <h2 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-2">Past</h2>
              <div className="space-y-2">
                {past.map((mw) => <MaintenanceCard key={mw.id} mw={mw} services={services} onArchive={archiveMutation.mutate} />)}
              </div>
            </section>
          )}
        </div>
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Schedule Maintenance Window</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleCreate} className="space-y-4">
            <div className="space-y-1.5">
              <Label>Service</Label>
              <Select value={serviceId} onValueChange={(v) => v && setServiceId(v)}>
                <SelectTrigger><SelectValue placeholder="Select service" /></SelectTrigger>
                <SelectContent>
                  {services?.filter((s) => !s.archived_at).map((s) => (
                    <SelectItem key={s.id} value={s.id}>{s.name}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1.5">
              <Label>Title</Label>
              <Input value={title} onChange={(e) => setTitle(e.target.value)} placeholder="Database maintenance" required />
            </div>
            <div className="space-y-1.5">
              <Label>Description (optional)</Label>
              <Textarea value={description} onChange={(e) => setDescription(e.target.value)} rows={2} />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label>Starts At</Label>
                <Input type="datetime-local" value={startsAt} onChange={(e) => setStartsAt(e.target.value)} required />
              </div>
              <div className="space-y-1.5">
                <Label>Ends At</Label>
                <Input type="datetime-local" value={endsAt} onChange={(e) => setEndsAt(e.target.value)} required />
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" type="button" onClick={() => setOpen(false)}>Cancel</Button>
              <Button type="submit" disabled={!serviceId || createMutation.isPending}>Schedule</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}

function MaintenanceCard({
  mw,
  services,
  onArchive,
}: {
  mw: { id: string; service_id: string; title: string; description?: string | null; starts_at: string; ends_at: string };
  services?: { id: string; name: string }[];
  onArchive: (id: string) => void;
}) {
  const svcName = services?.find((s) => s.id === mw.service_id)?.name ?? mw.service_id;
  const active = isActive(mw);

  return (
    <Card>
      <CardContent className="flex items-start gap-3 py-4">
        <Wrench className={`h-4 w-4 mt-0.5 shrink-0 ${active ? "text-blue-400" : "text-muted-foreground"}`} />
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-medium text-sm">{mw.title}</span>
            {active && <span className="text-xs text-blue-400 font-medium">Active</span>}
          </div>
          <p className="text-xs text-muted-foreground">
            {svcName} · {format(new Date(mw.starts_at), "MMM d, HH:mm")} – {format(new Date(mw.ends_at), "MMM d, HH:mm")}
          </p>
          {mw.description && <p className="text-xs text-muted-foreground mt-0.5">{mw.description}</p>}
        </div>
        <Button
          variant="ghost"
          size="icon"
          className="h-8 w-8 text-muted-foreground hover:text-destructive"
          onClick={() => onArchive(mw.id)}
        >
          <Trash2 className="h-3.5 w-3.5" />
        </Button>
      </CardContent>
    </Card>
  );
}

// Suppress unused import warning
const _f = formatDistanceToNow;
void _f;
