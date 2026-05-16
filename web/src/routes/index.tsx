import { Link } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { api, type NamespaceList } from "@/lib/api";
import { useCurrentUser, loginUrl } from "@/lib/auth";

export function IndexPage() {
  const me = useCurrentUser();
  const namespaces = useQuery({
    queryKey: ["namespaces"],
    queryFn: () => api<NamespaceList>("/api/namespaces"),
    enabled: !!me.data,
  });

  if (me.isLoading) {
    return <p className="text-sm text-muted-foreground">Loading…</p>;
  }
  if (!me.data) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Sign in required</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <p className="text-sm text-muted-foreground">
            Authenticate with your cluster's identity provider to manage virtual machines.
          </p>
          <Button asChild>
            <a href={loginUrl(window.location.pathname)}>Sign in</a>
          </Button>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="grid gap-4">
      <Card>
        <CardHeader>
          <CardTitle>Welcome, {me.data.username}</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            All actions you take here are executed against your cluster as <code>{me.data.username}</code>
            {me.data.groups?.length ? ` (groups: ${me.data.groups.join(", ")})` : ""}. RBAC is enforced by
            kube-apiserver.
          </p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Namespaces</CardTitle>
        </CardHeader>
        <CardContent>
          {namespaces.isLoading && <p className="text-sm text-muted-foreground">Loading namespaces…</p>}
          {namespaces.error && (
            <p className="text-sm text-destructive">
              Could not load namespaces. Your token may lack the required RBAC.
            </p>
          )}
          {namespaces.data && (
            <ul className="grid grid-cols-2 md:grid-cols-3 gap-2">
              {namespaces.data.items.map((ns) => (
                <li key={ns}>
                  <Link
                    to="/vms/$namespace"
                    params={{ namespace: ns }}
                    className="block rounded-md border px-3 py-2 text-sm hover:bg-muted"
                  >
                    {ns}
                  </Link>
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
