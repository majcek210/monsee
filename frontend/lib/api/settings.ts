import { api } from "./client";
import type { Settings, PublicSettings } from "@/types";

export const settingsApi = {
  get: () => api.get<Settings>("/admin/settings"),
  update: (data: { site_title?: string; logo_url?: string; public_status_enabled?: boolean }) =>
    api.patch<Settings>("/admin/settings", data),
};

export const publicSettingsApi = {
  get: () => api.get<PublicSettings>("/api/v1/settings"),
};
