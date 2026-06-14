import { Badge } from "@/components/ui/badge";
import type { ServiceStatus, MonitorStatus, IncidentStatus, IncidentSeverity } from "@/types";

export function ServiceStatusBadge({ status }: { status: ServiceStatus }) {
  const map: Record<ServiceStatus, { label: string; variant: "success" | "warning" | "danger" | "info" }> = {
    operational: { label: "Operational", variant: "success" },
    degraded: { label: "Degraded", variant: "warning" },
    outage: { label: "Outage", variant: "danger" },
    maintenance: { label: "Maintenance", variant: "info" },
  };
  const { label, variant } = map[status] ?? { label: status, variant: "info" };
  return <Badge variant={variant}>{label}</Badge>;
}

export function MonitorStatusBadge({ status }: { status: MonitorStatus }) {
  const map: Record<MonitorStatus, { label: string; variant: "success" | "warning" | "danger" }> = {
    up: { label: "Up", variant: "success" },
    degraded: { label: "Degraded", variant: "warning" },
    down: { label: "Down", variant: "danger" },
  };
  const { label, variant } = map[status] ?? { label: status, variant: "warning" };
  return <Badge variant={variant}>{label}</Badge>;
}

export function IncidentStatusBadge({ status }: { status: IncidentStatus }) {
  return (
    <Badge variant={status === "resolved" ? "success" : "danger"}>
      {status === "resolved" ? "Resolved" : "Open"}
    </Badge>
  );
}

export function SeverityBadge({ severity }: { severity: IncidentSeverity }) {
  const map: Record<IncidentSeverity, "danger" | "warning" | "info"> = {
    high: "danger",
    medium: "warning",
    low: "info",
  };
  return <Badge variant={map[severity]}>{severity}</Badge>;
}
