import { api } from "./client";
import type { APIKey, CreatedAPIKey } from "@/types";

export const apiKeysApi = {
  list: () => api.get<APIKey[]>("/admin/api-keys"),
  create: (data: { name: string }) =>
    api.post<CreatedAPIKey>("/admin/api-keys", data),
  revoke: (id: string) => api.delete<void>(`/admin/api-keys/${id}`),
};
