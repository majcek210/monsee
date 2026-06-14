import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { incidentsApi } from "@/lib/api/incidents";
import { toast } from "sonner";

export const incidentKeys = {
  all: ["incidents"] as const,
  detail: (id: string) => ["incidents", id] as const,
};

export function useIncidents() {
  return useQuery({
    queryKey: incidentKeys.all,
    queryFn: incidentsApi.list,
    staleTime: 30_000,
    refetchInterval: 30_000,
  });
}

export function useIncident(id: string) {
  return useQuery({
    queryKey: incidentKeys.detail(id),
    queryFn: () => incidentsApi.get(id),
    staleTime: 30_000,
  });
}

export function useCreateIncident() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: incidentsApi.create,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: incidentKeys.all });
      toast.success("Incident created");
    },
    onError: (e: Error) => toast.error(e.message),
  });
}

export function useResolveIncident() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: incidentsApi.resolve,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: incidentKeys.all });
      toast.success("Incident resolved");
    },
    onError: (e: Error) => toast.error(e.message),
  });
}
