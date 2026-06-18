"use client";

import { useEffect, useState } from "react";
import { Save } from "lucide-react";
import { useSettings, useUpdateSettings } from "@/lib/hooks/use-settings";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Switch } from "@/components/ui/switch";
import { Skeleton } from "@/components/ui/skeleton";
import { BrandLogo } from "@/components/ui/brand-logo";

export default function SettingsPage() {
  const { data: settings, isLoading } = useSettings();
  const updateMutation = useUpdateSettings();

  const [siteTitle, setSiteTitle] = useState("");
  const [logoUrl, setLogoUrl] = useState("");
  const [publicStatusEnabled, setPublicStatusEnabled] = useState(true);
  const [customDomainsEnabled, setCustomDomainsEnabled] = useState(false);

  useEffect(() => {
    if (settings) {
      setSiteTitle(settings.site_title);
      setLogoUrl(settings.logo_url);
      setPublicStatusEnabled(settings.public_status_enabled);
      setCustomDomainsEnabled(settings.custom_domains_enabled);
    }
  }, [settings]);

  async function handleSave(e: React.FormEvent) {
    e.preventDefault();
    await updateMutation.mutateAsync({
      site_title: siteTitle,
      logo_url: logoUrl,
      public_status_enabled: publicStatusEnabled,
      custom_domains_enabled: customDomainsEnabled,
    });
  }

  const proxyHost = process.env.NEXT_PUBLIC_PROXY_HOST || "proxy.example.com";

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-40" />
        <Skeleton className="h-48 w-full" />
      </div>
    );
  }

  return (
    <div className="space-y-6 max-w-2xl">
      <div>
        <h1 className="text-2xl font-semibold">Settings</h1>
        <p className="text-sm text-muted-foreground mt-0.5">Branding and visibility configuration</p>
      </div>

      <form onSubmit={handleSave} className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Branding</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-1.5">
              <Label>Site Title</Label>
              <Input
                value={siteTitle}
                onChange={(e) => setSiteTitle(e.target.value)}
                placeholder="monsee"
                required
              />
            </div>
            <div className="space-y-1.5">
              <Label>Logo URL</Label>
              <Input
                value={logoUrl}
                onChange={(e) => setLogoUrl(e.target.value)}
                placeholder="https://example.com/logo.png"
              />
              {logoUrl && (
                <div className="mt-2">
                  <BrandLogo
                    src={logoUrl}
                    alt="Logo preview"
                    size={40}
                    className="h-10 w-10 rounded object-contain bg-muted"
                  />
                </div>
              )}
              <p className="text-xs text-muted-foreground">
                Leave empty to use the default monsee logo. If this URL fails to load, the default
                logo is shown instead automatically.
              </p>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Public Status Page</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium">Enable public status page</p>
                <p className="text-xs text-muted-foreground mt-0.5">
                  Allow visitors to view the public status page at /
                </p>
              </div>
              <Switch
                checked={publicStatusEnabled}
                onCheckedChange={setPublicStatusEnabled}
              />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Custom Domains</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium">Enable custom domains</p>
                <p className="text-xs text-muted-foreground mt-0.5">
                  Let services be served on their own domain (e.g. status.acme.com). Must be on for
                  dedicated pages and TLS issuance to work.
                </p>
              </div>
              <Switch checked={customDomainsEnabled} onCheckedChange={setCustomDomainsEnabled} />
            </div>

            {customDomainsEnabled && (
              <div className="rounded-md border border-border bg-muted/40 p-4 space-y-3 text-sm">
                <p className="font-medium">Setup guide</p>
                <ol className="list-decimal list-inside space-y-2 text-muted-foreground">
                  <li>
                    Open the service, go to the <span className="text-foreground">Public Page</span>{" "}
                    tab, enable the dedicated page, set a slug and enter the custom domain.
                  </li>
                  <li>
                    At your domain&apos;s DNS provider, add a CNAME record:
                    <pre className="mt-1 rounded bg-background p-2 text-xs text-foreground overflow-x-auto">
{`status.theircompany.com  CNAME  ${proxyHost}`}
                    </pre>
                    <span className="text-xs">
                      Keep it <span className="text-foreground">DNS-only / grey-cloud</span> if your
                      proxy issues the certificate itself.
                    </span>
                  </li>
                  <li>
                    TLS is issued automatically on the first visit (the proxy checks{" "}
                    <code>/api/v1/tls-check</code> before issuing). Allow a minute, then open the
                    domain.
                  </li>
                </ol>
                <p className="text-xs text-muted-foreground">
                  See <code>docs/custom-domains-infra.md</code> for the proxy / Cloudflare Tunnel
                  setup and the Cloudflare-for-SaaS upgrade path.
                </p>
              </div>
            )}
          </CardContent>
        </Card>

        <Button type="submit" disabled={updateMutation.isPending}>
          <Save className="h-4 w-4" />
          Save Settings
        </Button>
      </form>
    </div>
  );
}
