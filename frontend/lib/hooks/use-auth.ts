import { useQuery } from "@tanstack/react-query";
import { authApi } from "@/lib/api/auth";

export function useCurrentUser() {
  return useQuery({
    queryKey: ["auth", "me"],
    queryFn: authApi.me,
    staleTime: 5 * 60_000,
    retry: false,
  });
}
