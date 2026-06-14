import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { usersApi } from "@/lib/api/users";
import { toast } from "sonner";

const KEYS = { all: ["users"] as const };

export function useUsers() {
  return useQuery({ queryKey: KEYS.all, queryFn: usersApi.list });
}

export function useCreateUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: usersApi.create,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.all });
      toast.success("User created");
    },
    onError: (e: Error) => toast.error(e.message),
  });
}

export function useUpdateUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...data }: { id: string; role: "admin" | "viewer" }) =>
      usersApi.update(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.all });
      toast.success("User updated");
    },
    onError: (e: Error) => toast.error(e.message),
  });
}

export function useArchiveUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: usersApi.archive,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.all });
      toast.success("User archived");
    },
    onError: (e: Error) => toast.error(e.message),
  });
}
