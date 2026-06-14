"use client";

import { useState } from "react";
import { Plus, CheckCircle } from "lucide-react";
import { useIncidents, useCreateIncident, useResolveIncident } from "@/lib/hooks/use-incidents";
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
import { IncidentStatusBadge, SeverityBadge } from "@/components/admin/status-badge";
import { formatDistanceToNow } from "date-fns";

export default function IncidentsPage() {
  const { data: incidents, isLoading } = useIncidents();
  const { data: services } = useServices();
  const createMutation = useCreateIncident();
  const resolveMutation = useResolveIncident();

  const [open, setOpen] = useState(false);
  const [title, setTitle] = useState("");
  const [serviceId, setServiceId] = useState("");
  const [severity, setSeverity] = useState<"low" | "medium" | "high">("high");

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    await createMutation.mutateAsync({ title, service_id: serviceId, severity });
    setOpen(false);
    setTitle("");
    setServiceId("");
    setSeverity("high");
  }

  const active = incidents?.filter((i) => !i.resolved_at) ?? [];
  const resolved = incidents?.filter((i) => i.resolved_at) ?? [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Incidents</h1>
          <p className="text-sm text-muted-foreground mt-0.5">
            {active.length} open · {resolved.length} resolved
          </p>
        </div>
        <Button onClick={() => setOpen(true)} size="sm">
          <Plus className="h-4 w-4" />
          New Incident
        </Button>
      </div>

      {isLoading ? (
        <div className="space-y-2">
          {[...Array(4)].map((_, i) => (
            <Skeleton key={i} className="h-16 w-full" />
          ))}
        </div>
      ) : incidents?.length === 0 ? (
        <Card>
          <CardContent className="flex items-center justify-center py-12">
            <p className="text-muted-foreground text-sm">No incidents. All systems are operational.</p>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-2">
          {[...active, ...resolved].map((inc) => (
            <Card key={inc.id}>
              <CardContent className="flex items-center gap-4 py-4">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 flex-wrap">
                    <span className="font-medium text-sm">{inc.title}</span>
                    <IncidentStatusBadge status={inc.status} />
                    <SeverityBadge severity={inc.severity} />
                  </div>
                  <p className="text-xs text-muted-foreground mt-0.5">
                    {formatDistanceToNow(new Date(inc.created_at), { addSuffix: true })}
                    {inc.resolved_at &&
                      ` · resolved ${formatDistanceToNow(new Date(inc.resolved_at), { addSuffix: true })}`}
                  </p>
                </div>
                {inc.status === "open" && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => resolveMutation.mutate(inc.id)}
                    disabled={resolveMutation.isPending}
                  >
                    <CheckCircle className="h-4 w-4" />
                    Resolve
                  </Button>
                )}
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create Incident</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleCreate} className="space-y-4">
            <div className="space-y-1.5">
              <Label>Title</Label>
              <Input
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="API latency spike"
                required
              />
            </div>
            <div className="space-y-1.5">
              <Label>Service</Label>
              <Select value={serviceId} onValueChange={(v) => v && setServiceId(v)}>
                <SelectTrigger>
                  <SelectValue placeholder="Select service" />
                </SelectTrigger>
                <SelectContent>
                  {services?.filter((s) => !s.archived_at).map((s) => (
                    <SelectItem key={s.id} value={s.id}>{s.name}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1.5">
              <Label>Severity</Label>
              <Select
                value={severity}
                onValueChange={(v) => v && setSeverity(v as "low" | "medium" | "high")}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="low">Low</SelectItem>
                  <SelectItem value="medium">Medium</SelectItem>
                  <SelectItem value="high">High</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <DialogFooter>
              <Button variant="outline" type="button" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button type="submit" disabled={!serviceId || createMutation.isPending}>
                Create
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
