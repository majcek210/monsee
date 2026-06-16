import { api } from "./client";
import type { Monitor, CheckResult } from "@/types";

export interface CreateMonitorInput {
  name: string;
  type: "http" | "tcp" | "ssl" | "keyword" | "dns";
  url?: string;
  host?: string;
  port?: number;
  interval_seconds?: number;
  timeout_ms?: number;
  retry_count?: number;
  degraded_threshold_ms?: number;
  http_method?: string;
  http_expected_status?: number;
  enabled?: boolean;
}

export const monitorsApi = {
  list: (serviceId: string) =>
    api.get<Monitor[]>(`/admin/monitors?service_id=${serviceId}`),
  get: (id: string) => api.get<Monitor>(`/admin/monitors/${id}`),
  create: (serviceId: string, data: CreateMonitorInput) =>
    api.post<Monitor>(`/admin/monitors`, { ...data, service_id: serviceId }),
  update: (id: string, data: Partial<CreateMonitorInput>) =>
    api.patch<Monitor>(`/admin/monitors/${id}`, data),
  archive: (id: string) => api.delete<void>(`/admin/monitors/${id}`),
  results: (id: string, limit = 90) =>
    api.get<CheckResult[]>(`/admin/monitors/${id}/results?limit=${limit}`),
};
