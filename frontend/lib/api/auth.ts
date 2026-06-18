import { api } from "./client";
import type { User } from "@/types";

export const authApi = {
  login: (email: string, password: string) =>
    api.post<User>("/auth/login", { email, password }),
  verify2FA: (user_id: string, code: string) =>
    api.post<User>("/auth/2fa/verify", { user_id, code }),
  logout: () => api.post<void>("/auth/logout"),
  me: () => api.get<User>("/auth/me"),
};
