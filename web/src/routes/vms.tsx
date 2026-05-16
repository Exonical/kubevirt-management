import { useParams } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { api, type VirtualMachineList } from "@/lib/api";

export function VirtualMachinesPage() {
  const { namespace } = useParams({ from: "/vms/$namespace" });
  const vms = useQuery({
    queryKey: ["vms", namespace],
    queryFn: () => api<VirtualMachineList>(`/api/namespaces/${namespace}/virtualmachines`),
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle>Virtual Machines &mdash; {namespace}</CardTitle>
      </CardHeader>
      <CardContent>
        {vms.isLoading && <p className="text-sm text-muted-foreground">Loading…</p>}
        {vms.error && (
          <p className="text-sm text-destructive">
            Could not list virtual machines. Your token may lack the required RBAC, or KubeVirt is not installed.
          </p>
        )}
        {vms.data && vms.data.items.length === 0 && (
          <p className="text-sm text-muted-foreground">No VirtualMachines in this namespace.</p>
        )}
        {vms.data && vms.data.items.length > 0 && (
          <table className="w-full text-sm">
            <thead className="text-left text-muted-foreground">
              <tr>
                <th className="py-2">Name</th>
                <th className="py-2">Status</th>
                <th className="py-2">Run strategy</th>
                <th className="py-2">Ready</th>
              </tr>
            </thead>
            <tbody>
              {vms.data.items.map((vm) => (
                <tr key={vm.uid} className="border-t">
                  <td className="py-2 font-medium">{vm.name}</td>
                  <td className="py-2">{vm.status || "—"}</td>
                  <td className="py-2">{vm.runStrategy || "—"}</td>
                  <td className="py-2">
                    <Badge variant={vm.ready ? "success" : "secondary"}>
                      {vm.ready ? "Ready" : "Not ready"}
                    </Badge>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </CardContent>
    </Card>
  );
}
