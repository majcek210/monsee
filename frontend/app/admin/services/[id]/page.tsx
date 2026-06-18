"use client";

import { useEffect, useState, use } from "react";
import Link from "next/link";
import {
  ChevronLeft,
  Plus,
  MoreHorizontal,
  Pencil,
  Archive,
  Save,
  Copy,
  ExternalLink,
} from "lucide-react";
import { useService, useUpdateService } from "@/lib/hooks/use-services";
import {
  useMonitors,
  useCreateMonitor,
  useUpdateMonitor,
  useArchiveMonitor,
} from "@/lib/hooks/use-monitors";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from "@/components/ui/select";
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
} from "@/components/ui/dropdown-menu";
import { Switch } from "@/components/ui/switch";
import { MonitorStatusBadge, ServiceStatusBadge } from "@/components/admin/status-badge";
import { toast } from "sonner";
import type { Monitor, Service } from "@/types";

type DialogMode = "create" | "edit" | null;

const defaultForm = {
  name: "",
  type: "http" as import("@/types").MonitorType,
  url: "",
  host: "",
  port: "",
  interval_seconds: "60",
  timeout_ms: "5000",
  retry_count: "2",
  degraded_threshold_ms: "",
  http_method: "GET",
  http_expected_status: "200",
  enabled: true,
  ssl_expiry_threshold_days: undefined as number | undefined,
  keyword_match: "",
  keyword_should_exist: true,
  dns_record_type: "A",
  dns_expected_value: "",
};

export default function ServiceDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const { data: service, isLoading: serviceLoading } = useService(id);

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Link href="/admin/services">
          <Button variant="ghost" size="icon" className="h-8 w-8">
            <ChevronLeft className="h-4 w-4" />
          </Button>
        </Link>
        <div className="flex-1 flex items-center gap-2">
          <h1 className="text-2xl font-semibold">{service?.name ?? "Service"}</h1>
          {service && <ServiceStatusBadge status={service.status} />}
        </div>
      </div>

      <Tabs defaultValue="monitors">
        <TabsList>
          <TabsTrigger value="monitors">Monitors</TabsTrigger>
          <TabsTrigger value="settings">Settings</TabsTrigger>
          <TabsTrigger value="public">Public Page</TabsTrigger>
        </TabsList>

        <TabsContent value="monitors">
          <MonitorsTab serviceId={id} />
        </TabsContent>

        <TabsContent value="settings">
          {serviceLoading || !service ? (
            <Skeleton className="h-64 w-full" />
          ) : (
            <SettingsTab service={service} />
          )}
        </TabsContent>

        <TabsContent value="public">
          {serviceLoading || !service ? (
            <Skeleton className="h-64 w-full" />
          ) : (
            <PublicPageTab service={service} />
          )}
        </TabsContent>
      </Tabs>
    </div>
  );
}

/* ───────────────────────────── Settings tab ───────────────────────────── */

function SettingsTab({ service }: { service: Service }) {
  const updateMutation = useUpdateService();
  const [name, setName] = useState(service.name);
  const [description, setDescription] = useState(service.description ?? "");
  const [statusOverride, setStatusOverride] = useState(service.status_override ?? "");
  const [publicVisible, setPublicVisible] = useState(service.public_visible);
  const [showUptime, setShowUptime] = useState(service.show_uptime);
  const [uptimeRangeDays, setUptimeRangeDays] = useState(service.uptime_range_days);

  useEffect(() => {
    setName(service.name);
    setDescription(service.description ?? "");
    setStatusOverride(service.status_override ?? "");
    setPublicVisible(service.public_visible);
    setShowUptime(service.show_uptime);
    setUptimeRangeDays(service.uptime_range_days);
  }, [service]);

  async function handleSave(e: React.FormEvent) {
    e.preventDefault();
    await updateMutation.mutateAsync({
      id: service.id,
      name,
      description: description || undefined,
      status_override: statusOverride || null,
      public_visible: publicVisible,
      show_uptime: showUptime,
      uptime_range_days: uptimeRangeDays,
    });
  }

  return (
    <Card>
      <CardContent className="py-6">
        <form onSubmit={handleSave} className="space-y-6 max-w-xl">
          <div className="space-y-3">
            <div className="space-y-1.5">
              <Label htmlFor="svc-name">Name</Label>
              <Input id="svc-name" value={name} onChange={(e) => setName(e.target.value)} required />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="svc-desc">Description</Label>
              <Input
                id="svc-desc"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Optional description"
              />
            </div>
          </div>

          <div className="space-y-3">
            <p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
              Visibility
            </p>
            <div className="flex items-center justify-between">
              <Label htmlFor="svc-public-visible" className="cursor-pointer">
                Show on public status page
              </Label>
              <Switch id="svc-public-visible" checked={publicVisible} onCheckedChange={setPublicVisible} />
            </div>
            <div className="flex items-center justify-between">
              <Label htmlFor="svc-show-uptime" className="cursor-pointer">
                Show uptime bars
              </Label>
              <Switch id="svc-show-uptime" checked={showUptime} onCheckedChange={setShowUptime} />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="svc-status-override">Status Override</Label>
              <Select
                value={statusOverride || "none"}
                onValueChange={(v) => setStatusOverride(v === "none" ? "" : v)}
              >
                <SelectTrigger id="svc-status-override">
                  <SelectValue placeholder="None (Auto)" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">None (Auto)</SelectItem>
                  <SelectItem value="operational">Operational</SelectItem>
                  <SelectItem value="degraded">Degraded</SelectItem>
                  <SelectItem value="outage">Outage</SelectItem>
                  <SelectItem value="maintenance">Maintenance</SelectItem>
                </SelectContent>
              </Select>
              <p className="text-xs text-muted-foreground">
                Force a status regardless of monitor checks. Leave on Auto for normal behaviour.
              </p>
            </div>
          </div>

          <div className="space-y-3">
            <p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
              Uptime
            </p>
            <div className="space-y-1.5">
              <Label htmlFor="svc-uptime-days">Uptime History (days)</Label>
              <Input
                id="svc-uptime-days"
                type="number"
                min={1}
                max={365}
                value={uptimeRangeDays}
                onChange={(e) => setUptimeRangeDays(Number(e.target.value))}
              />
            </div>
          </div>

          <Button type="submit" disabled={updateMutation.isPending}>
            <Save className="h-4 w-4" />
            Save Settings
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}

/* ─────────────────────────── Public Page tab ──────────────────────────── */

function PublicPageTab({ service }: { service: Service }) {
  const updateMutation = useUpdateService();
  const [enabled, setEnabled] = useState(service.dedicated_page_enabled);
  const [slug, setSlug] = useState(service.slug ?? "");
  const [customDomain, setCustomDomain] = useState(service.custom_domain ?? "");

  useEffect(() => {
    setEnabled(service.dedicated_page_enabled);
    setSlug(service.slug ?? "");
    setCustomDomain(service.custom_domain ?? "");
  }, [service]);

  async function handleSave(e: React.FormEvent) {
    e.preventDefault();
    await updateMutation.mutateAsync({
      id: service.id,
      dedicated_page_enabled: enabled,
      slug: slug || undefined,
      custom_domain: customDomain || undefined,
    });
  }

  const slugUrl = service.slug ? `/status/${service.slug}` : null;
  const domainUrl = service.custom_domain ? `https://${service.custom_domain}/` : null;

  function copy(text: string) {
    navigator.clipboard.writeText(text);
    toast.success("Copied to clipboard");
  }

  return (
    <Card>
      <CardContent className="py-6">
        <form onSubmit={handleSave} className="space-y-6 max-w-xl">
          <div className="flex items-center justify-between">
            <div>
              <Label htmlFor="svc-dedicated-page" className="cursor-pointer">
                Enable dedicated page
              </Label>
              <p className="text-xs text-muted-foreground mt-0.5">
                Give this service its own public status page at a slug or custom domain.
              </p>
            </div>
            <Switch id="svc-dedicated-page" checked={enabled} onCheckedChange={setEnabled} />
          </div>

          {enabled && (
            <div className="space-y-4 pl-3 border-l border-border">
              <div className="space-y-1.5">
                <Label htmlFor="svc-slug">Slug</Label>
                <Input
                  id="svc-slug"
                  value={slug}
                  onChange={(e) => setSlug(e.target.value)}
                  placeholder="my-service"
                />
                <p className="text-xs text-muted-foreground">
                  Lowercase letters, numbers and hyphens. Served at <code>/status/&lt;slug&gt;</code>.
                </p>
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="svc-custom-domain">Custom Domain</Label>
                <Input
                  id="svc-custom-domain"
                  value={customDomain}
                  onChange={(e) => setCustomDomain(e.target.value)}
                  placeholder="status.example.com"
                />
                <p className="text-xs text-muted-foreground">
                  Point this domain at us with a CNAME. See the DNS guide under{" "}
                  <Link href="/admin/settings" className="text-primary underline">
                    Settings → Custom Domains
                  </Link>
                  . Custom domains must be enabled there first.
                </p>
              </div>
            </div>
          )}

          {/* Live links (reflect saved values, not the in-progress form) */}
          {service.dedicated_page_enabled && (slugUrl || domainUrl) && (
            <div className="space-y-2 rounded-md bg-muted/50 p-3">
              <p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                Public links
              </p>
              {slugUrl && (
                <LinkRow href={slugUrl} label={slugUrl} onCopy={() => copy(slugUrl)} external />
              )}
              {domainUrl && (
                <LinkRow href={domainUrl} label={domainUrl} onCopy={() => copy(domainUrl)} external />
              )}
            </div>
          )}

          <Button type="submit" disabled={updateMutation.isPending}>
            <Save className="h-4 w-4" />
            Save Public Page
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}

function LinkRow({
  href,
  label,
  onCopy,
  external,
}: {
  href: string;
  label: string;
  onCopy: () => void;
  external?: boolean;
}) {
  return (
    <div className="flex items-center gap-2">
      <a
        href={href}
        target={external ? "_blank" : undefined}
        rel={external ? "noopener noreferrer" : undefined}
        className="flex items-center gap-1.5 text-sm text-primary hover:underline truncate"
      >
        <ExternalLink className="h-3.5 w-3.5 shrink-0" />
        <span className="truncate">{label}</span>
      </a>
      <Button type="button" variant="ghost" size="icon" className="h-7 w-7" onClick={onCopy}>
        <Copy className="h-3.5 w-3.5" />
      </Button>
    </div>
  );
}

/* ───────────────────────────── Monitors tab ───────────────────────────── */

function MonitorsTab({ serviceId }: { serviceId: string }) {
  const { data: monitors, isLoading } = useMonitors(serviceId);
  const createMutation = useCreateMonitor(serviceId);
  const updateMutation = useUpdateMonitor(serviceId);
  const archiveMutation = useArchiveMonitor(serviceId);

  const [dialogMode, setDialogMode] = useState<DialogMode>(null);
  const [editing, setEditing] = useState<Monitor | null>(null);
  const [form, setForm] = useState(defaultForm);

  function set(k: keyof typeof defaultForm, v: string | boolean | number | undefined) {
    setForm((p) => ({ ...p, [k]: v }));
  }

  function openCreate() {
    setForm(defaultForm);
    setEditing(null);
    setDialogMode("create");
  }

  function openEdit(m: Monitor) {
    setForm({
      name: m.name,
      type: m.type,
      url: m.url ?? "",
      host: m.host ?? "",
      port: m.port?.toString() ?? "",
      interval_seconds: m.interval_seconds.toString(),
      timeout_ms: m.timeout_ms.toString(),
      retry_count: m.retry_count.toString(),
      degraded_threshold_ms: m.degraded_threshold_ms?.toString() ?? "",
      http_method: m.http_method ?? "GET",
      http_expected_status: m.http_expected_status?.toString() ?? "200",
      enabled: m.enabled,
      ssl_expiry_threshold_days: undefined,
      keyword_match: "",
      keyword_should_exist: true,
      dns_record_type: "A",
      dns_expected_value: "",
    });
    setEditing(m);
    setDialogMode("edit");
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const payload = {
      name: form.name,
      type: form.type,
      url: form.type === "http" || form.type === "keyword" ? form.url : undefined,
      host: form.type === "tcp" ? form.host : undefined,
      port: form.type === "tcp" && form.port ? parseInt(form.port) : undefined,
      interval_seconds: parseInt(form.interval_seconds),
      timeout_ms: parseInt(form.timeout_ms),
      retry_count: parseInt(form.retry_count),
      degraded_threshold_ms: form.degraded_threshold_ms
        ? parseInt(form.degraded_threshold_ms)
        : undefined,
      http_method: form.type === "http" ? form.http_method : undefined,
      http_expected_status: form.type === "http" ? parseInt(form.http_expected_status) : undefined,
      enabled: form.enabled,
      ssl_expiry_threshold_days:
        form.type === "ssl" && form.ssl_expiry_threshold_days !== undefined
          ? form.ssl_expiry_threshold_days
          : undefined,
      keyword_match: form.type === "keyword" ? form.keyword_match : undefined,
      keyword_should_exist: form.type === "keyword" ? form.keyword_should_exist : undefined,
      dns_record_type: form.type === "dns" ? form.dns_record_type : undefined,
      dns_expected_value: form.type === "dns" ? form.dns_expected_value : undefined,
    };
    if (dialogMode === "create") {
      await createMutation.mutateAsync(payload);
    } else if (editing) {
      await updateMutation.mutateAsync({ id: editing.id, ...payload });
    }
    setDialogMode(null);
  }

  const active = monitors?.filter((m) => !m.archived_at) ?? [];
  const lastStatus = (m: Monitor): "up" | "down" | "degraded" =>
    m.consecutive_failures > 0 ? "down" : "up";

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <p className="text-sm text-muted-foreground">Checks that determine this service&apos;s status</p>
        <Button onClick={openCreate} size="sm">
          <Plus className="h-4 w-4" />
          New Monitor
        </Button>
      </div>

      {isLoading ? (
        <div className="space-y-2">
          {[...Array(3)].map((_, i) => (
            <Skeleton key={i} className="h-16 w-full" />
          ))}
        </div>
      ) : active.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12 text-center">
            <p className="text-muted-foreground text-sm">No monitors yet.</p>
            <Button className="mt-4" size="sm" onClick={openCreate}>
              <Plus className="h-4 w-4" />
              Add monitor
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-2">
          {active.map((m) => (
            <Card key={m.id}>
              <CardContent className="flex items-center gap-4 py-4">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="font-medium text-sm">{m.name}</span>
                    <MonitorStatusBadge status={lastStatus(m)} />
                    <span className="text-xs text-muted-foreground uppercase">{m.type}</span>
                    {!m.enabled && <span className="text-xs text-muted-foreground">(disabled)</span>}
                  </div>
                  <p className="text-xs text-muted-foreground mt-0.5 truncate">
                    {m.url ?? `${m.host}:${m.port}`} · every {m.interval_seconds}s
                  </p>
                </div>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="icon" className="h-8 w-8">
                      <MoreHorizontal className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem onClick={() => openEdit(m)}>
                      <Pencil className="h-4 w-4" />
                      Edit
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem
                      className="text-destructive focus:text-destructive"
                      onClick={() => archiveMutation.mutate(m.id)}
                    >
                      <Archive className="h-4 w-4" />
                      Archive
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <Dialog open={dialogMode !== null} onOpenChange={(o) => !o && setDialogMode(null)}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>{dialogMode === "create" ? "New Monitor" : "Edit Monitor"}</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5 col-span-2">
                <Label>Name</Label>
                <Input
                  value={form.name}
                  onChange={(e) => set("name", e.target.value)}
                  placeholder="API Health"
                  required
                />
              </div>

              <div className="space-y-1.5">
                <Label>Type</Label>
                <Select value={form.type} onValueChange={(v) => v && set("type", v)}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="http">HTTP</SelectItem>
                    <SelectItem value="tcp">TCP</SelectItem>
                    <SelectItem value="ssl">SSL</SelectItem>
                    <SelectItem value="keyword">Keyword</SelectItem>
                    <SelectItem value="dns">DNS</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-1.5">
                <Label>Interval (seconds)</Label>
                <Input
                  type="number"
                  min={10}
                  value={form.interval_seconds}
                  onChange={(e) => set("interval_seconds", e.target.value)}
                />
              </div>

              {form.type === "http" && (
                <>
                  <div className="space-y-1.5 col-span-2">
                    <Label>URL</Label>
                    <Input
                      value={form.url}
                      onChange={(e) => set("url", e.target.value)}
                      placeholder="https://api.example.com/health"
                      required
                    />
                  </div>
                  <div className="space-y-1.5">
                    <Label>HTTP Method</Label>
                    <Select value={form.http_method} onValueChange={(v) => v && set("http_method", v)}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {["GET", "POST", "HEAD"].map((m) => (
                          <SelectItem key={m} value={m}>
                            {m}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-1.5">
                    <Label>Expected Status</Label>
                    <Input
                      type="number"
                      value={form.http_expected_status}
                      onChange={(e) => set("http_expected_status", e.target.value)}
                    />
                  </div>
                </>
              )}

              {form.type === "tcp" && (
                <>
                  <div className="space-y-1.5">
                    <Label>Host</Label>
                    <Input
                      value={form.host}
                      onChange={(e) => set("host", e.target.value)}
                      placeholder="db.example.com"
                      required
                    />
                  </div>
                  <div className="space-y-1.5">
                    <Label>Port</Label>
                    <Input
                      type="number"
                      value={form.port}
                      onChange={(e) => set("port", e.target.value)}
                      placeholder="5432"
                      required
                    />
                  </div>
                </>
              )}

              {form.type === "ssl" && (
                <div className="space-y-1.5 col-span-2">
                  <Label>Days Before Expiry Warning</Label>
                  <Input
                    type="number"
                    value={form.ssl_expiry_threshold_days ?? ""}
                    onChange={(e) =>
                      setForm((p) => ({
                        ...p,
                        ssl_expiry_threshold_days: e.target.value ? parseInt(e.target.value) : undefined,
                      }))
                    }
                    placeholder="14"
                  />
                </div>
              )}

              {form.type === "keyword" && (
                <>
                  <div className="space-y-1.5 col-span-2">
                    <Label>URL</Label>
                    <Input
                      value={form.url}
                      onChange={(e) => set("url", e.target.value)}
                      placeholder="https://api.example.com/health"
                      required
                    />
                  </div>
                  <div className="space-y-1.5 col-span-2">
                    <Label>Keyword to Match</Label>
                    <Input
                      value={form.keyword_match}
                      onChange={(e) => set("keyword_match", e.target.value)}
                      placeholder="OK"
                    />
                  </div>
                  <div className="col-span-2 flex items-center gap-2">
                    <Switch
                      id="keyword_should_exist"
                      checked={form.keyword_should_exist}
                      onCheckedChange={(v) => set("keyword_should_exist", v)}
                    />
                    <Label htmlFor="keyword_should_exist">Keyword Should Exist</Label>
                  </div>
                </>
              )}

              {form.type === "dns" && (
                <>
                  <div className="space-y-1.5">
                    <Label>Record Type</Label>
                    <Select value={form.dns_record_type} onValueChange={(v) => v && set("dns_record_type", v)}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {["A", "AAAA", "CNAME", "MX", "TXT", "NS"].map((r) => (
                          <SelectItem key={r} value={r}>
                            {r}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-1.5">
                    <Label>Expected Value</Label>
                    <Input
                      value={form.dns_expected_value}
                      onChange={(e) => set("dns_expected_value", e.target.value)}
                      placeholder="1.2.3.4"
                    />
                  </div>
                </>
              )}

              <div className="space-y-1.5">
                <Label>Timeout (ms)</Label>
                <Input
                  type="number"
                  value={form.timeout_ms}
                  onChange={(e) => set("timeout_ms", e.target.value)}
                />
              </div>
              <div className="space-y-1.5">
                <Label>Retry Count</Label>
                <Input
                  type="number"
                  min={0}
                  value={form.retry_count}
                  onChange={(e) => set("retry_count", e.target.value)}
                />
              </div>
              <div className="space-y-1.5 col-span-2">
                <Label>Degraded Threshold (ms, optional)</Label>
                <Input
                  type="number"
                  value={form.degraded_threshold_ms}
                  onChange={(e) => set("degraded_threshold_ms", e.target.value)}
                  placeholder="Leave empty to disable"
                />
              </div>

              <div className="col-span-2 flex items-center gap-2">
                <Switch id="enabled" checked={form.enabled} onCheckedChange={(v) => set("enabled", v)} />
                <Label htmlFor="enabled">Enabled</Label>
              </div>
            </div>

            <DialogFooter>
              <Button variant="outline" type="button" onClick={() => setDialogMode(null)}>
                Cancel
              </Button>
              <Button type="submit" disabled={createMutation.isPending || updateMutation.isPending}>
                {dialogMode === "create" ? "Create" : "Save"}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
