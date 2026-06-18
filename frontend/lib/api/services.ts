import { api } from "./client";
import type { Service } from "@/types";

interface ServiceVisibilityFields {
  public_visible?: boolean;
  show_uptime?: boolean;
  dedicated_page_enabled?: boolean;
  slug?: string;
  custom_domain?: string;
  uptime_range_days?: number;
  status_override?: string | null;
}

export const servicesApi = {
  list: () => api.get<Service[]>("/admin/services"),
  get: (id: string) => api.get<Service>(`/admin/services/${id}`),
  create: (data: { name: string; description?: string } & ServiceVisibilityFields) =>
    api.post<Service>("/admin/services", data),
  update: (id: string, data: { name?: string; description?: string; status?: string } & ServiceVisibilityFields) =>
    api.patch<Service>(`/admin/services/${id}`, data),
  archive: (id: string) => api.delete<void>(`/admin/services/${id}`),
};
