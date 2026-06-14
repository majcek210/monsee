import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { webhooksApi, type CreateWebhookInput } from "@/lib/api/webhooks";
import { toast } from "sonner";

export const webhookKeys = {
  all: ["webhooks"] as const,
  logs: (id: string) => ["webhook-logs", id] as const,
};

export function useWebhooks() {
  return useQuery({ queryKey: webhookKeys.all, queryFn: webhooksApi.list });
}

export function useWebhookLogs(id: string) {
  return useQuery({
    queryKey: webhookKeys.logs(id),
    queryFn: () => webhooksApi.logs(id),
    enabled: !!id,
  });
}

export function useCreateWebhook() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateWebhookInput) => webhooksApi.create(data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: webhookKeys.all });
      toast.success("Webhook created");
    },
    onError: (e: Error) => toast.error(e.message),
  });
}

export function useUpdateWebhook() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      id,
      ...data
    }: {
      id: string;
      name?: string;
      url?: string;
      secret?: string;
      events?: string[];
      enabled?: boolean;
    }) => webhooksApi.update(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: webhookKeys.all });
      toast.success("Webhook updated");
    },
    onError: (e: Error) => toast.error(e.message),
  });
}

export function useArchiveWebhook() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: webhooksApi.archive,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: webhookKeys.all });
      toast.success("Webhook deleted");
    },
    onError: (e: Error) => toast.error(e.message),
  });
}
