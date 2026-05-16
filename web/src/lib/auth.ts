import { useQuery } from "@tanstack/react-query";
import { api, type CurrentUser } from "./api";

export function useCurrentUser() {
  return useQuery({
    queryKey: ["auth", "me"],
    queryFn: () => api<CurrentUser>("/api/auth/me"),
    retry: false,
  });
}

export function loginUrl(redirectTo?: string) {
  const params = new URLSearchParams();
  if (redirectTo) params.set("redirect_to", redirectTo);
  const qs = params.toString();
  return `/api/auth/login${qs ? `?${qs}` : ""}`;
}

export async function logout() {
  await api("/api/auth/logout", { method: "POST" });
}
