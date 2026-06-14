"use client";

import { useState } from "react";
import { Plus, Trash2, ShieldCheck, Eye } from "lucide-react";
import { useUsers, useCreateUser, useUpdateUser, useArchiveUser } from "@/lib/hooks/use-users";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
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
import { formatDistanceToNow } from "date-fns";

export default function UsersPage() {
  const { data: users, isLoading } = useUsers();
  const createMutation = useCreateUser();
  const updateMutation = useUpdateUser();
  const archiveMutation = useArchiveUser();

  const [open, setOpen] = useState(false);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState<"admin" | "viewer">("viewer");

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    await createMutation.mutateAsync({ email, password, role });
    setOpen(false);
    setEmail(""); setPassword(""); setRole("viewer");
  }

  const active = users?.filter((u) => !u.archived_at) ?? [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Users</h1>
          <p className="text-sm text-muted-foreground mt-0.5">
            Admin and viewer accounts
          </p>
        </div>
        <Button onClick={() => setOpen(true)} size="sm">
          <Plus className="h-4 w-4" />
          New User
        </Button>
      </div>

      {isLoading ? (
        <div className="space-y-2">
          {[...Array(3)].map((_, i) => (
            <Skeleton key={i} className="h-14 w-full" />
          ))}
        </div>
      ) : active.length === 0 ? (
        <Card>
          <CardContent className="flex items-center justify-center py-12">
            <p className="text-muted-foreground text-sm">No users found.</p>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-2">
          {active.map((u) => (
            <Card key={u.id}>
              <CardContent className="flex items-center gap-4 py-4">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="font-medium text-sm">{u.email}</span>
                    <Badge
                      variant={u.role === "admin" ? "default" : "secondary"}
                      className="capitalize"
                    >
                      {u.role === "admin" ? (
                        <><ShieldCheck className="h-3 w-3 mr-1" />Admin</>
                      ) : (
                        <><Eye className="h-3 w-3 mr-1" />Viewer</>
                      )}
                    </Badge>
                  </div>
                  <p className="text-xs text-muted-foreground mt-0.5">
                    Joined {formatDistanceToNow(new Date(u.created_at), { addSuffix: true })}
                  </p>
                </div>

                <Select
                  value={u.role}
                  onValueChange={(v) =>
                    v && updateMutation.mutate({ id: u.id, role: v as "admin" | "viewer" })
                  }
                >
                  <SelectTrigger className="w-28 h-8 text-xs">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="admin">Admin</SelectItem>
                    <SelectItem value="viewer">Viewer</SelectItem>
                  </SelectContent>
                </Select>

                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8 text-destructive hover:text-destructive"
                  onClick={() => archiveMutation.mutate(u.id)}
                  disabled={archiveMutation.isPending}
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>New User</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleCreate} className="space-y-4">
            <div className="space-y-1.5">
              <Label>Email</Label>
              <Input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="user@example.com"
                required
              />
            </div>
            <div className="space-y-1.5">
              <Label>Password</Label>
              <Input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Min 8 characters"
                required
              />
            </div>
            <div className="space-y-1.5">
              <Label>Role</Label>
              <Select value={role} onValueChange={(v) => v && setRole(v as "admin" | "viewer")}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="admin">Admin</SelectItem>
                  <SelectItem value="viewer">Viewer</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <DialogFooter>
              <Button variant="outline" type="button" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button type="submit" disabled={createMutation.isPending}>
                Create
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
