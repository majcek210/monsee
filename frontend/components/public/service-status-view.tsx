"use client";

import { useQuery } from "@tanstack/react-query";
import { CheckCircle2, AlertTriangle, XCircle, Clock } from "lucide-react";
import { api } from "@/lib/api/client";
import { usePublicSettings } from "@/lib/hooks/use-settings";
import { Skeleton } from "@/components/ui/skeleton";
import { Separator } from "@/components/ui/separator";
import { BrandLogo } from "@/components/ui/brand-logo";
import type { Service, ServiceUptime, DailyUptimeStatus } from "@/types";

interface DedicatedPagePayload {
  service: Service;
  uptime: ServiceUptime;
}

function effectiveStatus(svc: Service): string {
  return svc.status_override || svc.status;
}

function OverallBanner({ status }: { status: string }) {
  if (status === "outage") {
    return (
      <Banner
        tone="red"
        icon={<XCircle className="h-6 w-6 text-red-400 shrink-0" />}
        title="Service Disruption"
        subtitle="This service is experiencing issues."
      />
    );
  }
  if (status === "degraded") {
    return (
      <Banner
        tone="amber"
        icon={<AlertTriangle className="h-6 w-6 text-amber-400 shrink-0" />}
        title="Partial Degradation"
        subtitle="This service is experiencing degraded performance."
      />
    );
  }
  if (status === "maintenance") {
    return (
      <Banner
        tone="blue"
        icon={<Clock className="h-6 w-6 text-blue-400 shrink-0" />}
        title="Under Maintenance"
        subtitle="Scheduled maintenance is in progress."
      />
    );
  }
  return (
    <Banner
      tone="emerald"
      icon={<CheckCircle2 className="h-6 w-6 text-emerald-400 shrink-0" />}
      title="Operational"
      subtitle="Everything is running smoothly."
    />
  );
}

function Banner({
  tone,
  icon,
  title,
  subtitle,
}: {
  tone: "red" | "amber" | "blue" | "emerald";
  icon: React.ReactNode;
  title: string;
  subtitle: string;
}) {
  const toneClass = {
    red: "bg-red-500/10 border-red-500/20 text-red-400",
    amber: "bg-amber-500/10 border-amber-500/20 text-amber-400",
    blue: "bg-blue-500/10 border-blue-500/20 text-blue-400",
    emerald: "bg-emerald-500/10 border-emerald-500/20 text-emerald-400",
  }[tone];
  return (
    <div className={`flex items-center gap-3 p-4 rounded-lg border ${toneClass}`}>
      {icon}
      <div>
        <p className="font-semibold">{title}</p>
        <p className="text-sm text-muted-foreground">{subtitle}</p>
      </div>
    </div>
  );
}

function dayColor(d: DailyUptimeStatus): string {
  switch (d.status) {
    case "up":
      return "bg-emerald-400/80";
    case "degraded":
      return "bg-amber-400/80";
    case "down":
      return "bg-red-400/80";
    default:
      return "bg-muted";
  }
}

function UptimeBars({ days }: { days: DailyUptimeStatus[] }) {
  if (!days || days.length === 0) {
    return <p className="text-xs text-muted-foreground">No uptime data yet.</p>;
  }
  const overall =
    days.reduce((acc, d) => acc + (d.status === "no_data" ? 0 : d.uptime_percent), 0) /
    Math.max(days.filter((d) => d.status !== "no_data").length, 1);
  return (
    <div className="space-y-1.5">
      <div className="flex items-end gap-[2px] h-8">
        {days.map((d) => (
          <div
            key={d.date}
            className={`flex-1 rounded-sm ${dayColor(d)}`}
            style={{ height: "100%" }}
            title={`${d.date}: ${d.status === "no_data" ? "no data" : `${d.uptime_percent.toFixed(2)}% uptime`}`}
          />
        ))}
      </div>
      <div className="flex items-center justify-between text-[11px] text-muted-foreground">
        <span>{days.length} days ago</span>
        <span>{overall.toFixed(2)}% uptime</span>
        <span>today</span>
      </div>
    </div>
  );
}

export function ServiceStatusView({
  endpoint,
  queryKey,
}: {
  endpoint: string;
  queryKey: string[];
}) {
  const { data: settings } = usePublicSettings();
  const { data, isLoading, isError } = useQuery({
    queryKey,
    queryFn: () => api.get<DedicatedPagePayload>(endpoint),
    staleTime: 30_000,
    refetchInterval: 60_000,
    retry: false,
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

      <div className="max-w-3xl mx-auto px-4 py-10 space-y-8">
        {isLoading ? (
          <>
            <Skeleton className="h-20 w-full" />
            <Skeleton className="h-40 w-full" />
          </>
        ) : isError || !data ? (
          <div className="text-center py-16">
            <p className="text-sm text-muted-foreground">
              This status page is not available.
            </p>
          </div>
        ) : (
          <>
            <div className="space-y-1">
              <h1 className="text-2xl font-semibold">{data.service.name}</h1>
              {data.service.description && (
                <p className="text-sm text-muted-foreground">{data.service.description}</p>
              )}
            </div>

            <OverallBanner status={effectiveStatus(data.service)} />

            <div className="space-y-3">
              <h2 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">
                Monitors
              </h2>
              {data.uptime?.monitors && data.uptime.monitors.length > 0 ? (
                <div className="space-y-2">
                  {data.uptime.monitors.map((m) => (
                    <div key={m.monitor_id} className="rounded-lg border border-border bg-card p-4">
                      <div className="flex items-center justify-between">
                        <span className="font-medium text-sm">{m.monitor_name}</span>
                      </div>
                      {data.service.show_uptime && (
                        <>
                          <Separator className="my-3" />
                          <UptimeBars days={m.days} />
                        </>
                      )}
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-sm text-muted-foreground">No monitors to display.</p>
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
