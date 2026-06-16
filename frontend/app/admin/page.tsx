"use client";

import { useAllUptime } from "@/lib/hooks/use-uptime";
import { useServices } from "@/lib/hooks/use-services";
import { useIncidents } from "@/lib/hooks/use-incidents";
import { UptimeBar } from "@/components/uptime-bar";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { AlertTriangle, CheckCircle2, Activity } from "lucide-react";

export default function AdminOverviewPage() {
  const { data: services, isLoading: svcsLoading } = useServices();
  const { data: uptime, isLoading: uptimeLoading } = useAllUptime();
  const { data: incidents } = useIncidents();

  const openIncidents = incidents?.filter((i) => i.status === "open") ?? [];
  const activeServices = services?.filter((s) => !s.archived_at) ?? [];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Overview</h1>
        <p className="text-sm text-muted-foreground mt-0.5">Uptime across all services</p>
      </div>

      <div className="grid grid-cols-3 gap-4">
        <Card>
          <CardContent className="flex items-center gap-3 py-4">
            <Activity className="h-5 w-5 text-primary" />
            <div>
              <p className="text-2xl font-semibold">{activeServices.length}</p>
              <p className="text-xs text-muted-foreground">Services</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="flex items-center gap-3 py-4">
            <AlertTriangle className={`h-5 w-5 ${openIncidents.length > 0 ? "text-red-400" : "text-muted-foreground"}`} />
            <div>
              <p className="text-2xl font-semibold">{openIncidents.length}</p>
              <p className="text-xs text-muted-foreground">Open Incidents</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="flex items-center gap-3 py-4">
            <CheckCircle2 className="h-5 w-5 text-emerald-400" />
            <div>
              <p className="text-2xl font-semibold">{activeServices.filter((s) => s.status === "operational").length}</p>
              <p className="text-xs text-muted-foreground">Operational</p>
            </div>
          </CardContent>
        </Card>
      </div>

      <div className="space-y-4">
        <h2 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">Uptime History</h2>
        {uptimeLoading || svcsLoading ? (
          <div className="space-y-3">
            {[...Array(4)].map((_, i) => <Skeleton key={i} className="h-20 w-full" />)}
          </div>
        ) : !uptime || uptime.length === 0 ? (
          <Card>
            <CardContent className="py-8 text-center">
              <p className="text-sm text-muted-foreground">No uptime data yet.</p>
            </CardContent>
          </Card>
        ) : (
          <div className="space-y-3">
            {uptime.map((svcUptime) => {
              const svc = services?.find((s) => s.id === svcUptime.service_id);
              return (
                <Card key={svcUptime.service_id}>
                  <CardHeader className="pb-2 pt-4 px-4">
                    <CardTitle className="text-sm font-medium">{svc?.name ?? svcUptime.service_id}</CardTitle>
                  </CardHeader>
                  <CardContent className="px-4 pb-4 space-y-3">
                    {svcUptime.monitors.map((m) => (
                      <div key={m.monitor_id}>
                        <p className="text-xs text-muted-foreground mb-1">{m.monitor_name}</p>
                        <UptimeBar days={m.days} />
                      </div>
                    ))}
                  </CardContent>
                </Card>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
