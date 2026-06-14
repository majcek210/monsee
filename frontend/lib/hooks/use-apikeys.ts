import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiKeysApi } from "@/lib/api/apikeys";
import { toast } from "sonner";

const KEYS = { all: ["api-keys"] as const };

export function useAPIKeys() {
  return useQuery({ queryKey: KEYS.all, queryFn: apiKeysApi.list });
}

export function useCreateAPIKey() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: apiKeysApi.create,
    onSuccess: () => qc.invalidateQueries({ queryKey: KEYS.all }),
    onError: (e: Error) => toast.error(e.message),
  });
}

export function useRevokeAPIKey() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: apiKeysApi.revoke,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.all });
      toast.success("API key revoked");
    },
    onError: (e: Error) => toast.error(e.message),
  });
}
