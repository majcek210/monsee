import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { notificationsApi, type CreateNotificationInput } from "@/lib/api/notifications";
import { toast } from "sonner";

const KEYS = { all: ["notifications"] as const };

export function useNotifications() {
  return useQuery({ queryKey: KEYS.all, queryFn: notificationsApi.list });
}

export function useCreateNotification() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateNotificationInput) => notificationsApi.create(data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.all });
      toast.success("Channel created");
    },
    onError: (e: Error) => toast.error(e.message),
  });
}

export function useUpdateNotification() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      id,
      ...data
    }: {
      id: string;
      name?: string;
      enabled?: boolean;
      config?: Record<string, string>;
    }) => notificationsApi.update(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.all });
      toast.success("Channel updated");
    },
    onError: (e: Error) => toast.error(e.message),
  });
}

export function useArchiveNotification() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: notificationsApi.archive,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.all });
      toast.success("Channel deleted");
    },
    onError: (e: Error) => toast.error(e.message),
  });
}
