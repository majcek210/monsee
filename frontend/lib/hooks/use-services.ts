import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { servicesApi } from "@/lib/api/services";
import { toast } from "sonner";

export const serviceKeys = {
  all: ["services"] as const,
  detail: (id: string) => ["services", id] as const,
};

export function useServices() {
  return useQuery({
    queryKey: serviceKeys.all,
    queryFn: servicesApi.list,
    staleTime: 30_000,
    refetchInterval: 60_000,
  });
}

export function useService(id: string) {
  return useQuery({
    queryKey: serviceKeys.detail(id),
    queryFn: () => servicesApi.get(id),
    staleTime: 30_000,
  });
}

export function useCreateService() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: servicesApi.create,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: serviceKeys.all });
      toast.success("Service created");
    },
    onError: (e: Error) => toast.error(e.message),
  });
}

export function useUpdateService() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...data }: { id: string; name?: string; description?: string; status?: string }) =>
      servicesApi.update(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: serviceKeys.all });
      toast.success("Service updated");
    },
    onError: (e: Error) => toast.error(e.message),
  });
}

export function useArchiveService() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: servicesApi.archive,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: serviceKeys.all });
      toast.success("Service archived");
    },
    onError: (e: Error) => toast.error(e.message),
  });
}
