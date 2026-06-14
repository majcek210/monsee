import { api } from "./client";
import type { Service } from "@/types";

export const servicesApi = {
  list: () => api.get<Service[]>("/admin/services"),
  get: (id: string) => api.get<Service>(`/admin/services/${id}`),
  create: (data: { name: string; description?: string }) =>
    api.post<Service>("/admin/services", data),
  update: (id: string, data: { name?: string; description?: string; status?: string }) =>
    api.patch<Service>(`/admin/services/${id}`, data),
  archive: (id: string) => api.delete<void>(`/admin/services/${id}`),
};
