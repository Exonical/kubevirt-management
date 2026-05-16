import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  Outlet,
  RouterProvider,
  createRootRoute,
  createRoute,
  createRouter,
} from "@tanstack/react-router";
import { AppShell } from "@/components/app-shell";
import { IndexPage } from "@/routes/index";
import { LoginPage } from "@/routes/login";
import { VirtualMachinesPage } from "@/routes/vms";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: (failureCount, error) => {
        // Don't retry auth failures.
        if (error instanceof Response && (error.status === 401 || error.status === 403)) {
          return false;
        }
        return failureCount < 2;
      },
      staleTime: 5_000,
    },
  },
});

const rootRoute = createRootRoute({
  component: () => (
    <AppShell>
      <Outlet />
    </AppShell>
  ),
});

const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/",
  component: IndexPage,
});

const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/login",
  component: LoginPage,
});

const vmsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/vms/$namespace",
  component: VirtualMachinesPage,
});

const routeTree = rootRoute.addChildren([indexRoute, loginRoute, vmsRoute]);

const router = createRouter({ routeTree });

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  );
}
