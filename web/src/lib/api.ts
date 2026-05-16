/**
 * Thin fetch wrapper for the backend API. Throws the raw Response on non-2xx
 * so React Query can branch on status codes (401 -> redirect to /login, etc).
 */
export async function api<T = unknown>(
  path: string,
  init: RequestInit = {},
): Promise<T> {
  const { headers: initHeaders, ...restInit } = init;
  const res = await fetch(path, {
    ...restInit,
    credentials: "include",
    headers: {
      Accept: "application/json",
      ...(init.body ? { "Content-Type": "application/json" } : {}),
      ...(initHeaders as Record<string, string>),
    },
  });
  if (!res.ok) {
    throw res;
  }
  if (res.status === 204) {
    return undefined as T;
  }
  const ct = res.headers.get("Content-Type") ?? "";
  if (ct.includes("application/json")) {
    return (await res.json()) as T;
  }
  return (await res.text()) as unknown as T;
}

export interface CurrentUser {
  subject: string;
  username: string;
  email?: string;
  groups?: string[];
  expires_at: string;
}

export interface VirtualMachineSummary {
  namespace: string;
  name: string;
  uid: string;
  status: string;
  ready: boolean;
  runStrategy?: string;
  labels?: Record<string, string>;
  nodeName?: string;
}

export interface NamespaceList {
  items: string[];
}

export interface VirtualMachineList {
  items: VirtualMachineSummary[];
}
