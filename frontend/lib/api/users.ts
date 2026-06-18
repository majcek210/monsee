import { api } from "./client";
import type { User } from "@/types";

export const usersApi = {
  list: () => api.get<User[]>("/admin/users"),
  create: (data: { email: string; password: string; role: "admin" | "viewer" }) =>
    api.post<User>("/admin/users", data),
  update: (id: string, data: { role?: "admin" | "viewer"; email?: string; password?: string }) =>
    api.patch<User>(`/admin/users/${id}`, data),
  archive: (id: string) => api.delete<void>(`/admin/users/${id}`),
  disableTwoFactor: (id: string) => api.post<void>(`/admin/users/${id}/disable-2fa`, {}),
};
