import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import type { ServiceUptime, ResponseTimePoint } from "@/types";

export function useServiceUptime(serviceId: string) {
  return useQuery({
    queryKey: ["uptime", serviceId],
    queryFn: () => api.get<ServiceUptime>(`/admin/services/${serviceId}/uptime`),
    staleTime: 60_000,
    refetchInterval: 120_000,
  });
}

export function useAllUptime() {
  return useQuery({
    queryKey: ["uptime"],
    queryFn: () => api.get<ServiceUptime[]>("/admin/uptime"),
    staleTime: 60_000,
    refetchInterval: 120_000,
  });
}

export function useMonitorLatency(monitorId: string) {
  return useQuery({
    queryKey: ["latency", monitorId],
    queryFn: () => api.get<ResponseTimePoint[]>(`/admin/monitors/${monitorId}/latency`),
    staleTime: 60_000,
    refetchInterval: 60_000,
  });
}
