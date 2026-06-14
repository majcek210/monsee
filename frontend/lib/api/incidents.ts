import { api } from "./client";
import type { Incident } from "@/types";

export const incidentsApi = {
  list: () => api.get<Incident[]>("/admin/incidents"),
  get: (id: string) => api.get<Incident>(`/admin/incidents/${id}`),
  create: (data: {
    service_id: string;
    monitor_id?: string;
    title: string;
    severity?: "low" | "medium" | "high";
  }) => api.post<Incident>("/admin/incidents", data),
  resolve: (id: string) =>
    api.post<Incident>(`/admin/incidents/${id}/resolve`, {}),
  update: (
    id: string,
    data: { title?: string; severity?: string; status?: string }
  ) => api.patch<Incident>(`/admin/incidents/${id}`, data),
};
