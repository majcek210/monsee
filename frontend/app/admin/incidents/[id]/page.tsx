"use client";

import { useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { ArrowLeft, Plus } from "lucide-react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { IncidentStatusBadge, SeverityBadge } from "@/components/admin/status-badge";
import type { Incident, IncidentUpdate } from "@/types";
import { formatDistanceToNow, format } from "date-fns";
import { toast } from "sonner";

function useIncidentDetail(id: string) {
  return useQuery({
    queryKey: ["incidents", id],
    queryFn: () => api.get<Incident>(`/admin/incidents/${id}`),
    staleTime: 30_000,
  });
}

function useIncidentUpdates(id: string) {
  return useQuery({
    queryKey: ["incident-updates", id],
    queryFn: () => api.get<IncidentUpdate[]>(`/admin/incidents/${id}/updates`),
    staleTime: 30_000,
    refetchInterval: 60_000,
  });
}

const statusColors: Record<string, string> = {
  investigating: "border-red-500/20 bg-red-500/5",
  identified: "border-amber-500/20 bg-amber-500/5",
  monitoring: "border-blue-500/20 bg-blue-500/5",
  resolved: "border-emerald-500/20 bg-emerald-500/5",
};

const statusDotColors: Record<string, string> = {
  investigating: "bg-red-400",
  identified: "bg-amber-400",
  monitoring: "bg-blue-400",
  resolved: "bg-emerald-400",
};

export default function IncidentDetailPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const qc = useQueryClient();

  const { data: incident, isLoading } = useIncidentDetail(id);
  const { data: updates } = useIncidentUpdates(id);

  const [open, setOpen] = useState(false);
  const [status, setStatus] = useState<string>("investigating");
  const [message, setMessage] = useState("");

  const postUpdate = useMutation({
    mutationFn: ({ s, m }: { s: string; m: string }) =>
      api.post<IncidentUpdate>(`/admin/incidents/${id}/updates`, { status: s, message: m }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["incident-updates", id] });
      qc.invalidateQueries({ queryKey: ["incidents", id] });
      qc.invalidateQueries({ queryKey: ["incidents"] });
      setOpen(false);
      setMessage("");
      toast.success("Update posted");
    },
    onError: (e: Error) => toast.error(e.message),
  });

  async function handlePost(e: React.FormEvent) {
    e.preventDefault();
    postUpdate.mutate({ s: status, m: message });
  }

  if (isLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-8 w-64" />
        <Skeleton className="h-32 w-full" />
      </div>
    );
  }

  if (!incident) return null;

  return (
    <div className="space-y-6 max-w-2xl">
      <div className="flex items-center gap-3">
        <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => router.back()}>
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div className="flex-1">
          <div className="flex items-center gap-2 flex-wrap">
            <h1 className="text-xl font-semibold">{incident.title}</h1>
            <IncidentStatusBadge status={incident.status} />
            <SeverityBadge severity={incident.severity} />
          </div>
          <p className="text-xs text-muted-foreground mt-0.5">
            Opened {formatDistanceToNow(new Date(incident.created_at), { addSuffix: true })}
            {incident.resolved_at && ` · Resolved ${formatDistanceToNow(new Date(incident.resolved_at), { addSuffix: true })}`}
          </p>
        </div>
        {incident.status === "open" && (
          <Button size="sm" onClick={() => setOpen(true)}>
            <Plus className="h-4 w-4" />
            Post Update
          </Button>
        )}
      </div>

      {/* Timeline */}
      <div className="space-y-3">
        <h2 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">Timeline</h2>
        {!updates || updates.length === 0 ? (
          <Card>
            <CardContent className="py-8 text-center">
              <p className="text-sm text-muted-foreground">No updates yet.</p>
            </CardContent>
          </Card>
        ) : (
          <div className="space-y-2">
            {[...updates].reverse().map((u) => (
              <div
                key={u.id}
                className={`rounded-lg border p-4 ${statusColors[u.status] ?? "border-border bg-card"}`}
              >
                <div className="flex items-center gap-2 mb-1.5">
                  <span className={`inline-block h-2 w-2 rounded-full ${statusDotColors[u.status] ?? "bg-muted-foreground"}`} />
                  <span className="text-xs font-semibold capitalize">{u.status}</span>
                  <span className="text-xs text-muted-foreground ml-auto">
                    {format(new Date(u.created_at), "MMM d, HH:mm")}
                  </span>
                </div>
                <p className="text-sm">{u.message}</p>
              </div>
            ))}
          </div>
        )}
      </div>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Post Incident Update</DialogTitle>
          </DialogHeader>
          <form onSubmit={handlePost} className="space-y-4">
            <div className="space-y-1.5">
              <Label>Status</Label>
              <Select value={status} onValueChange={(v) => v && setStatus(v)}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="investigating">Investigating</SelectItem>
                  <SelectItem value="identified">Identified</SelectItem>
                  <SelectItem value="monitoring">Monitoring</SelectItem>
                  <SelectItem value="resolved">Resolved</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1.5">
              <Label>Message</Label>
              <Textarea
                value={message}
                onChange={(e) => setMessage(e.target.value)}
                rows={4}
                placeholder="Describe the current situation..."
                required
              />
            </div>
            <DialogFooter>
              <Button variant="outline" type="button" onClick={() => setOpen(false)}>Cancel</Button>
              <Button type="submit" disabled={!message || postUpdate.isPending}>Post Update</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
