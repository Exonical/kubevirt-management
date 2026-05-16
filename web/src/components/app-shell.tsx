import { Link, useRouterState } from "@tanstack/react-router";
import { LogOut, Server } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useCurrentUser, logout } from "@/lib/auth";

export function AppShell({ children }: { children: React.ReactNode }) {
  const me = useCurrentUser();
  const router = useRouterState();
  const onLogin = router.location.pathname === "/login";

  return (
    <div className="min-h-screen flex flex-col">
      <header className="border-b">
        <div className="container flex h-14 items-center gap-4">
          <Link to="/" className="flex items-center gap-2 font-semibold">
            <Server className="h-5 w-5" />
            <span>KubeVirt Management</span>
          </Link>
          <nav className="flex items-center gap-3 text-sm text-muted-foreground">
            <Link to="/" className="hover:text-foreground">Overview</Link>
            <Link to="/vms/$namespace" params={{ namespace: "default" }} className="hover:text-foreground">
              VMs
            </Link>
          </nav>
          <div className="ml-auto flex items-center gap-3">
            {me.data ? (
              <>
                <span className="text-sm text-muted-foreground">{me.data.username}</span>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={async () => {
                    await logout();
                    window.location.assign("/login");
                  }}
                >
                  <LogOut className="h-4 w-4" />
                  Sign out
                </Button>
              </>
            ) : !onLogin ? (
              <Button size="sm" asChild>
                <a href="/api/auth/login">Sign in</a>
              </Button>
            ) : null}
          </div>
        </div>
      </header>
      <main className="container flex-1 py-6">{children}</main>
      <footer className="border-t py-4 text-center text-xs text-muted-foreground">
        kubevirt-management &mdash; manage VMs via your cluster RBAC
      </footer>
    </div>
  );
}
