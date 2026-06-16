import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { settingsApi, publicSettingsApi } from "@/lib/api/settings";
import { toast } from "sonner";

export function useSettings() {
  return useQuery({
    queryKey: ["settings"],
    queryFn: settingsApi.get,
    staleTime: 30_000,
  });
}

export function useUpdateSettings() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: settingsApi.update,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["settings"] });
      toast.success("Settings saved");
    },
    onError: (e: Error) => toast.error(e.message),
  });
}

export function usePublicSettings() {
  return useQuery({
    queryKey: ["public-settings"],
    queryFn: publicSettingsApi.get,
    staleTime: 30_000,
    refetchInterval: 60_000,
  });
}
