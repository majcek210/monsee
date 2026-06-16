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

export default function SettingsPage() {
  const { data: settings, isLoading } = useSettings();
  const updateMutation = useUpdateSettings();

  const [siteTitle, setSiteTitle] = useState("");
  const [logoUrl, setLogoUrl] = useState("");
  const [publicStatusEnabled, setPublicStatusEnabled] = useState(true);

  useEffect(() => {
    if (settings) {
      setSiteTitle(settings.site_title);
      setLogoUrl(settings.logo_url);
      setPublicStatusEnabled(settings.public_status_enabled);
    }
  }, [settings]);

  async function handleSave(e: React.FormEvent) {
    e.preventDefault();
    await updateMutation.mutateAsync({
      site_title: siteTitle,
      logo_url: logoUrl,
      public_status_enabled: publicStatusEnabled,
    });
  }

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
                  {/* eslint-disable-next-line @next/next/no-img-element */}
                  <img
                    src={logoUrl}
                    alt="Logo preview"
                    className="h-10 w-10 rounded object-contain bg-muted"
                  />
                </div>
              )}
              <p className="text-xs text-muted-foreground">Leave empty to use the default monsee logo.</p>
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

        <Button type="submit" disabled={updateMutation.isPending}>
          <Save className="h-4 w-4" />
          Save Settings
        </Button>
      </form>
    </div>
  );
}
