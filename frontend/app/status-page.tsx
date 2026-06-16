"use client";

import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { CheckCircle2, AlertTriangle, XCircle, Clock } from "lucide-react";
import { api } from "@/lib/api/client";
import { usePublicSettings } from "@/lib/hooks/use-settings";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import type { PublicService, Incident } from "@/types";
import { formatDistanceToNow } from "date-fns";

function usePublicStatus() {
  return useQuery({
    queryKey: ["public-status"],
    queryFn: () => api.get<PublicService[]>("/api/v1/status"),
    staleTime: 30_000,
    refetchInterval: 60_000,
  });
}

function usePublicIncidents() {
  return useQuery({
    queryKey: ["public-incidents"],
    queryFn: () => api.get<Incident[]>("/api/v1/incidents?status=open"),
    staleTime: 30_000,
    refetchInterval: 60_000,
  });
}

function OverallStatus({ services }: { services: PublicService[] }) {
  const hasOutage = services.some((s) => s.status === "outage");
  const hasDegraded = services.some((s) => s.status === "degraded");

  if (hasOutage) {
    return (
      <div className="flex items-center gap-3 p-4 rounded-lg bg-red-500/10 border border-red-500/20">
        <XCircle className="h-6 w-6 text-red-400 shrink-0" />
        <div>
          <p className="font-semibold text-red-400">Service Disruption</p>
          <p className="text-sm text-muted-foreground">Some systems are experiencing issues.</p>
        </div>
      </div>
    );
  }

  if (hasDegraded) {
    return (
      <div className="flex items-center gap-3 p-4 rounded-lg bg-amber-500/10 border border-amber-500/20">
        <AlertTriangle className="h-6 w-6 text-amber-400 shrink-0" />
        <div>
          <p className="font-semibold text-amber-400">Partial Degradation</p>
          <p className="text-sm text-muted-foreground">Some systems are experiencing degraded performance.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex items-center gap-3 p-4 rounded-lg bg-emerald-500/10 border border-emerald-500/20">
      <CheckCircle2 className="h-6 w-6 text-emerald-400 shrink-0" />
      <div>
        <p className="font-semibold text-emerald-400">All Systems Operational</p>
        <p className="text-sm text-muted-foreground">Everything is running smoothly.</p>
      </div>
    </div>
  );
}

function MonitorDot({ status }: { status: string }) {
  const color =
    status === "up"
      ? "bg-emerald-400"
      : status === "degraded"
      ? "bg-amber-400"
      : "bg-red-400";
  return <span className={`inline-block h-2 w-2 rounded-full ${color}`} />;
}

function ServiceStatusIcon({ status }: { status: string }) {
  if (status === "outage") return <XCircle className="h-4 w-4 text-red-400" />;
  if (status === "degraded") return <AlertTriangle className="h-4 w-4 text-amber-400" />;
  if (status === "maintenance") return <Clock className="h-4 w-4 text-blue-400" />;
  return <CheckCircle2 className="h-4 w-4 text-emerald-400" />;
}

export function PublicStatusPage() {
  const { data: services, isLoading } = usePublicStatus();
  const { data: incidents } = usePublicIncidents();
  const { data: settings } = usePublicSettings();

  const openIncidents = incidents?.filter((i) => i.status === "open") ?? [];
  const logoUrl = settings?.logo_url || "/monsee.png";
  const siteTitle = settings?.site_title || "monsee";

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b border-border">
        <div className="max-w-3xl mx-auto px-4 h-14 flex items-center gap-2">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img src={logoUrl} alt={siteTitle} width={24} height={24} className="rounded" />
          <span className="font-semibold">{siteTitle}</span>
        </div>
      </header>

      <div className="max-w-3xl mx-auto px-4 py-10 space-y-8">
        {/* Overall status */}
        {isLoading ? (
          <Skeleton className="h-20 w-full" />
        ) : services ? (
          <OverallStatus services={services} />
        ) : null}

        {/* Active incidents */}
        {openIncidents.length > 0 && (
          <div className="space-y-3">
            <h2 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">
              Active Incidents
            </h2>
            {openIncidents.map((inc) => (
              <Link
                key={inc.id}
                href={`/incidents/${inc.id}`}
                className="block p-4 rounded-lg border border-red-500/20 bg-red-500/5 space-y-1 hover:bg-red-500/10 transition-colors"
              >
                <div className="flex items-center gap-2">
                  <AlertTriangle className="h-4 w-4 text-red-400 shrink-0" />
                  <span className="font-medium text-sm">{inc.title}</span>
                  <Badge variant="danger" className="ml-auto">{inc.severity}</Badge>
                </div>
                <p className="text-xs text-muted-foreground pl-6">
                  Opened {formatDistanceToNow(new Date(inc.created_at), { addSuffix: true })}
                </p>
              </Link>
            ))}
          </div>
        )}

        {/* Services */}
        <div className="space-y-3">
          <h2 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">
            Services
          </h2>
          {isLoading ? (
            <div className="space-y-2">
              {[...Array(4)].map((_, i) => (
                <Skeleton key={i} className="h-16 w-full" />
              ))}
            </div>
          ) : !services || services.length === 0 ? (
            <p className="text-sm text-muted-foreground">No services to display.</p>
          ) : (
            <div className="space-y-2">
              {services.map((svc) => (
                <div key={svc.id} className="rounded-lg border border-border bg-card p-4">
                  <div className="flex items-center gap-2">
                    <ServiceStatusIcon status={svc.status} />
                    <span className="font-medium text-sm flex-1">{svc.name}</span>
                    <span className="text-xs text-muted-foreground capitalize">{svc.status}</span>
                  </div>

                  {svc.monitors && svc.monitors.length > 0 && (
                    <>
                      <Separator className="my-3" />
                      <div className="space-y-1.5">
                        {svc.monitors.map((m) => (
                          <div key={m.id} className="flex items-center gap-2 text-xs">
                            <MonitorDot status={m.status} />
                            <span className="text-muted-foreground flex-1">{m.name}</span>
                            {m.response_time_ms != null && (
                              <span className="text-muted-foreground">{m.response_time_ms}ms</span>
                            )}
                          </div>
                        ))}
                      </div>
                    </>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      <footer className="border-t border-border mt-16">
        <div className="max-w-3xl mx-auto px-4 py-4 text-center text-xs text-muted-foreground">
          Powered by monsee
        </div>
      </footer>
    </div>
  );
}
