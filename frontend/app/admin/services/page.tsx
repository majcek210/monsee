"use client";

import { useState } from "react";
import Link from "next/link";
import {
  Plus,
  MoreHorizontal,
  Pencil,
  Archive,
  ChevronRight,
  ExternalLink,
  Copy,
} from "lucide-react";
import {
  useServices,
  useCreateService,
  useUpdateService,
  useArchiveService,
} from "@/lib/hooks/use-services";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Card, CardContent } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
} from "@/components/ui/dropdown-menu";
import { ServiceStatusBadge } from "@/components/admin/status-badge";
import { toast } from "sonner";
import type { Service } from "@/types";

type DialogMode = "create" | "edit" | null;

interface FormState {
  name: string;
  description: string;
}

const defaultForm: FormState = { name: "", description: "" };

export default function ServicesPage() {
  const { data: services, isLoading } = useServices();
  const createMutation = useCreateService();
  const updateMutation = useUpdateService();
  const archiveMutation = useArchiveService();

  const [dialogMode, setDialogMode] = useState<DialogMode>(null);
  const [editing, setEditing] = useState<Service | null>(null);
  const [form, setForm] = useState<FormState>(defaultForm);

  function setField<K extends keyof FormState>(key: K, value: FormState[K]) {
    setForm((prev) => ({ ...prev, [key]: value }));
  }

  function openCreate() {
    setForm(defaultForm);
    setEditing(null);
    setDialogMode("create");
  }

  // Quick rename/description edit. Full settings live on the service detail page.
  function openEdit(svc: Service) {
    setForm({ name: svc.name, description: svc.description ?? "" });
    setEditing(svc);
    setDialogMode("edit");
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const payload = {
      name: form.name,
      description: form.description || undefined,
    };
    if (dialogMode === "create") {
      await createMutation.mutateAsync(payload);
    } else if (editing) {
      await updateMutation.mutateAsync({ id: editing.id, ...payload });
    }
    setDialogMode(null);
  }

  function publicHref(svc: Service): string | null {
    if (!svc.dedicated_page_enabled) return null;
    if (svc.custom_domain) return `https://${svc.custom_domain}/`;
    if (svc.slug) return `/status/${svc.slug}`;
    return null;
  }

  function copy(text: string) {
    navigator.clipboard.writeText(text);
    toast.success("Copied to clipboard");
  }

  const active = services?.filter((s) => !s.archived_at) ?? [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Services</h1>
          <p className="text-sm text-muted-foreground mt-0.5">
            Monitor groups for your infrastructure
          </p>
        </div>
        <Button onClick={openCreate} size="sm">
          <Plus className="h-4 w-4" />
          New Service
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
            <p className="text-muted-foreground text-sm">No services yet.</p>
            <Button className="mt-4" size="sm" onClick={openCreate}>
              <Plus className="h-4 w-4" />
              Create your first service
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-2">
          {active.map((svc) => {
            const href = publicHref(svc);
            return (
              <Card key={svc.id} className="hover:bg-accent/30 transition-colors">
                <CardContent className="flex items-center gap-4 py-4">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-sm">{svc.name}</span>
                      <ServiceStatusBadge status={svc.status} />
                    </div>
                    {svc.description && (
                      <p className="text-xs text-muted-foreground mt-0.5 truncate">
                        {svc.description}
                      </p>
                    )}
                    {href && (
                      <div className="flex items-center gap-1.5 mt-1">
                        <a
                          href={href}
                          target="_blank"
                          rel="noopener noreferrer"
                          onClick={(e) => e.stopPropagation()}
                          className="flex items-center gap-1 text-xs text-primary hover:underline truncate"
                        >
                          <ExternalLink className="h-3 w-3 shrink-0" />
                          <span className="truncate">{href}</span>
                        </a>
                        <button
                          type="button"
                          onClick={(e) => {
                            e.stopPropagation();
                            copy(href);
                          }}
                          className="text-muted-foreground hover:text-foreground"
                          aria-label="Copy public link"
                        >
                          <Copy className="h-3 w-3" />
                        </button>
                      </div>
                    )}
                  </div>

                  <Link
                    href={`/admin/services/${svc.id}`}
                    className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground transition-colors"
                  >
                    Manage <ChevronRight className="h-3 w-3" />
                  </Link>

                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="ghost" size="icon" className="h-8 w-8">
                        <MoreHorizontal className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem onClick={() => openEdit(svc)}>
                        <Pencil className="h-4 w-4" />
                        Rename
                      </DropdownMenuItem>
                      <DropdownMenuItem asChild>
                        <Link href={`/admin/services/${svc.id}`}>
                          <ChevronRight className="h-4 w-4" />
                          Manage settings
                        </Link>
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem
                        className="text-destructive focus:text-destructive"
                        onClick={() => archiveMutation.mutate(svc.id)}
                      >
                        <Archive className="h-4 w-4" />
                        Archive
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}

      <Dialog open={dialogMode !== null} onOpenChange={(o) => !o && setDialogMode(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{dialogMode === "create" ? "New Service" : "Rename Service"}</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-1.5">
              <Label htmlFor="svc-name">Name</Label>
              <Input
                id="svc-name"
                value={form.name}
                onChange={(e) => setField("name", e.target.value)}
                placeholder="API Gateway"
                required
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="svc-desc">Description</Label>
              <Input
                id="svc-desc"
                value={form.description}
                onChange={(e) => setField("description", e.target.value)}
                placeholder="Optional description"
              />
            </div>
            {dialogMode === "create" && (
              <p className="text-xs text-muted-foreground">
                Visibility, dedicated page and uptime settings are configured on the service page
                after it&apos;s created.
              </p>
            )}

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
