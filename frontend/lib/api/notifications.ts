import { api } from "./client";
import type { NotificationChannel } from "@/types";

export interface CreateNotificationInput {
  name: string;
  type: "discord" | "email";
  config: Record<string, string>;
  enabled?: boolean;
}

export const notificationsApi = {
  list: () => api.get<NotificationChannel[]>("/admin/notifications"),
  get: (id: string) => api.get<NotificationChannel>(`/admin/notifications/${id}`),
  create: (data: CreateNotificationInput) =>
    api.post<NotificationChannel>("/admin/notifications", data),
  update: (
    id: string,
    data: { name?: string; enabled?: boolean; config?: Record<string, string> }
  ) => api.patch<NotificationChannel>(`/admin/notifications/${id}`, data),
  archive: (id: string) => api.delete<void>(`/admin/notifications/${id}`),
};
