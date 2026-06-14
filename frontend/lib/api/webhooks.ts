import { api } from "./client";
import type { Webhook, WebhookLog } from "@/types";

export interface CreateWebhookInput {
  name: string;
  url: string;
  secret?: string;
  events: string[];
  enabled?: boolean;
}

export const webhooksApi = {
  list: () => api.get<Webhook[]>("/admin/webhooks"),
  get: (id: string) => api.get<Webhook>(`/admin/webhooks/${id}`),
  create: (data: CreateWebhookInput) =>
    api.post<Webhook>("/admin/webhooks", data),
  update: (
    id: string,
    data: { name?: string; url?: string; secret?: string; events?: string[]; enabled?: boolean }
  ) => api.patch<Webhook>(`/admin/webhooks/${id}`, data),
  archive: (id: string) => api.delete<void>(`/admin/webhooks/${id}`),
  logs: (id: string) => api.get<WebhookLog[]>(`/admin/webhooks/${id}/logs`),
};
