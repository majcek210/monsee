import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { monitorsApi, type CreateMonitorInput } from "@/lib/api/monitors";
import { toast } from "sonner";

export const monitorKeys = {
  all: (serviceId: string) => ["monitors", serviceId] as const,
  detail: (id: string) => ["monitor", id] as const,
  results: (id: string) => ["monitor-results", id] as const,
};

export function useMonitors(serviceId: string) {
  return useQuery({
    queryKey: monitorKeys.all(serviceId),
    queryFn: () => monitorsApi.list(serviceId),
    staleTime: 30_000,
    refetchInterval: 60_000,
  });
}

export function useMonitorResults(id: string) {
  return useQuery({
    queryKey: monitorKeys.results(id),
    queryFn: () => monitorsApi.results(id),
    staleTime: 60_000,
    refetchInterval: 60_000,
  });
}

export function useCreateMonitor(serviceId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateMonitorInput) =>
      monitorsApi.create(serviceId, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: monitorKeys.all(serviceId) });
      toast.success("Monitor created");
    },
    onError: (e: Error) => toast.error(e.message),
  });
}

export function useUpdateMonitor(serviceId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...data }: { id: string } & Partial<CreateMonitorInput>) =>
      monitorsApi.update(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: monitorKeys.all(serviceId) });
      toast.success("Monitor updated");
    },
    onError: (e: Error) => toast.error(e.message),
  });
}

export function useArchiveMonitor(serviceId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: monitorsApi.archive,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: monitorKeys.all(serviceId) });
      toast.success("Monitor archived");
    },
    onError: (e: Error) => toast.error(e.message),
  });
}
