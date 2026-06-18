"use client";

import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { ArrowLeft, AlertTriangle } from "lucide-react";
import { api } from "@/lib/api/client";
import { usePublicSettings } from "@/lib/hooks/use-settings";
import { Skeleton } from "@/components/ui/skeleton";
import { BrandLogo } from "@/components/ui/brand-logo";
import type { Incident, IncidentUpdate } from "@/types";
import { format, formatDistanceToNow } from "date-fns";

interface IncidentWithUpdates {
  incident: Incident;
  updates: IncidentUpdate[];
}

const statusDotColors: Record<string, string> = {
  investigating: "bg-red-400",
  identified: "bg-amber-400",
  monitoring: "bg-blue-400",
  resolved: "bg-emerald-400",
};

const statusCardColors: Record<string, string> = {
  investigating: "border-red-500/20 bg-red-500/5",
  identified: "border-amber-500/20 bg-amber-500/5",
  monitoring: "border-blue-500/20 bg-blue-500/5",
  resolved: "border-emerald-500/20 bg-emerald-500/5",
};

export default function PublicIncidentPage() {
  const { id } = useParams<{ id: string }>();
  const { data: settings } = usePublicSettings();

  const { data, isLoading } = useQuery({
    queryKey: ["public-incident", id],
    queryFn: () => api.get<IncidentWithUpdates>(`/api/v1/incidents/${id}`),
    staleTime: 30_000,
    refetchInterval: 60_000,
  });

  const siteTitle = settings?.site_title || "monsee";

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b border-border">
        <div className="max-w-3xl mx-auto px-4 h-14 flex items-center gap-2">
          <BrandLogo src={settings?.logo_url} alt={siteTitle} className="rounded" />
          <span className="font-semibold">{siteTitle}</span>
        </div>
      </header>

      <div className="max-w-3xl mx-auto px-4 py-10 space-y-6">
        <Link href="/" className="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground transition-colors">
          <ArrowLeft className="h-4 w-4" />
          Back to status
        </Link>

        {isLoading ? (
          <div className="space-y-3">
            <Skeleton className="h-10 w-3/4" />
            <Skeleton className="h-32 w-full" />
          </div>
        ) : !data ? (
          <p className="text-muted-foreground">Incident not found.</p>
        ) : (
          <>
            <div>
              <div className="flex items-start gap-3">
                <AlertTriangle className="h-5 w-5 text-red-400 mt-0.5 shrink-0" />
                <div>
                  <h1 className="text-xl font-semibold">{data.incident.title}</h1>
                  <p className="text-sm text-muted-foreground mt-0.5">
                    {data.incident.status === "resolved" ? "Resolved" : "Open"} ·{" "}
                    {data.incident.severity} severity ·{" "}
                    Opened {formatDistanceToNow(new Date(data.incident.created_at), { addSuffix: true })}
                    {data.incident.resolved_at &&
                      ` · Resolved ${formatDistanceToNow(new Date(data.incident.resolved_at), { addSuffix: true })}`}
                  </p>
                </div>
              </div>
            </div>

            <div className="space-y-3">
              <h2 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">Timeline</h2>
              {data.updates.length === 0 ? (
                <p className="text-sm text-muted-foreground">No updates have been posted yet.</p>
              ) : (
                <div className="space-y-3">
                  {[...data.updates].reverse().map((u) => (
                    <div
                      key={u.id}
                      className={`rounded-lg border p-4 ${statusCardColors[u.status] ?? "border-border bg-card"}`}
                    >
                      <div className="flex items-center gap-2 mb-1.5">
                        <span className={`inline-block h-2 w-2 rounded-full ${statusDotColors[u.status] ?? "bg-muted-foreground"}`} />
                        <span className="text-xs font-semibold capitalize">{u.status}</span>
                        <span className="text-xs text-muted-foreground ml-auto">
                          {format(new Date(u.created_at), "MMM d, HH:mm")} UTC
                        </span>
                      </div>
                      <p className="text-sm">{u.message}</p>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </>
        )}
      </div>

      <footer className="border-t border-border mt-16">
        <div className="max-w-3xl mx-auto px-4 py-4 text-center text-xs text-muted-foreground">
          Powered by monsee
        </div>
      </footer>
    </div>
  );
}
