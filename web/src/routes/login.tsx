import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { loginUrl } from "@/lib/auth";

export function LoginPage() {
  return (
    <div className="max-w-md mx-auto">
      <Card>
        <CardHeader>
          <CardTitle>Sign in</CardTitle>
          <CardDescription>
            Use your organization's identity provider. Your cluster's RBAC controls what you can do.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Button asChild className="w-full">
            <a href={loginUrl("/")}>Continue with OIDC</a>
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
