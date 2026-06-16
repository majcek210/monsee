import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { maintenanceApi } from "@/lib/api/maintenance";
import { toast } from "sonner";

export function useMaintenanceWindows(serviceId?: string) {
  return useQuery({
    queryKey: ["maintenance", serviceId],
    queryFn: () => maintenanceApi.list(serviceId),
    staleTime: 30_000,
  });
}

export function useCreateMaintenanceWindow() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: maintenanceApi.create,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["maintenance"] });
      toast.success("Maintenance window created");
    },
    onError: (e: Error) => toast.error(e.message),
  });
}

export function useUpdateMaintenanceWindow() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...data }: { id: string; title?: string; description?: string; starts_at?: string; ends_at?: string }) =>
      maintenanceApi.update(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["maintenance"] });
      toast.success("Maintenance window updated");
    },
    onError: (e: Error) => toast.error(e.message),
  });
}

export function useArchiveMaintenanceWindow() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: maintenanceApi.archive,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["maintenance"] });
      toast.success("Maintenance window archived");
    },
    onError: (e: Error) => toast.error(e.message),
  });
}
