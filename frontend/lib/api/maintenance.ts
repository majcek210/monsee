import { api } from "./client";
import type { MaintenanceWindow } from "@/types";

export const maintenanceApi = {
  list: (serviceId?: string) =>
    api.get<MaintenanceWindow[]>(`/admin/maintenance-windows${serviceId ? `?service_id=${serviceId}` : ""}`),
  get: (id: string) => api.get<MaintenanceWindow>(`/admin/maintenance-windows/${id}`),
  create: (data: {
    service_id: string;
    title: string;
    description?: string;
    starts_at: string;
    ends_at: string;
  }) => api.post<MaintenanceWindow>("/admin/maintenance-windows", data),
  update: (
    id: string,
    data: { title?: string; description?: string; starts_at?: string; ends_at?: string }
  ) => api.patch<MaintenanceWindow>(`/admin/maintenance-windows/${id}`, data),
  archive: (id: string) => api.delete<void>(`/admin/maintenance-windows/${id}`),
};
