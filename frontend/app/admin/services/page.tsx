"use client";

import { useState } from "react";
import Link from "next/link";
import { Plus, MoreHorizontal, Pencil, Archive, ChevronRight } from "lucide-react";
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
import type { Service } from "@/types";

type DialogMode = "create" | "edit" | null;

export default function ServicesPage() {
  const { data: services, isLoading } = useServices();
  const createMutation = useCreateService();
  const updateMutation = useUpdateService();
  const archiveMutation = useArchiveService();

  const [dialogMode, setDialogMode] = useState<DialogMode>(null);
  const [editing, setEditing] = useState<Service | null>(null);
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");

  function openCreate() {
    setName("");
    setDescription("");
    setEditing(null);
    setDialogMode("create");
  }

  function openEdit(svc: Service) {
    setName(svc.name);
    setDescription(svc.description ?? "");
    setEditing(svc);
    setDialogMode("edit");
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (dialogMode === "create") {
      await createMutation.mutateAsync({ name, description: description || undefined });
    } else if (editing) {
      await updateMutation.mutateAsync({ id: editing.id, name, description });
    }
    setDialogMode(null);
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
          {active.map((svc) => (
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
                </div>

                <Link
                  href={`/admin/services/${svc.id}`}
                  className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground transition-colors"
                >
                  Monitors <ChevronRight className="h-3 w-3" />
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
                      Edit
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
          ))}
        </div>
      )}

      <Dialog open={dialogMode !== null} onOpenChange={(o) => !o && setDialogMode(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{dialogMode === "create" ? "New Service" : "Edit Service"}</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-1.5">
              <Label htmlFor="svc-name">Name</Label>
              <Input
                id="svc-name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="API Gateway"
                required
              />
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
