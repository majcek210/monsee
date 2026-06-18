"use client";

import { useState } from "react";
import { Plus, MoreHorizontal, Pencil, Archive, ShieldCheck, ShieldOff, Eye, ShieldX } from "lucide-react";
import {
  useUsers,
  useCreateUser,
  useUpdateUser,
  useArchiveUser,
  useDisableUserTwoFactor,
} from "@/lib/hooks/use-users";
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
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
} from "@/components/ui/dropdown-menu";
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from "@/components/ui/select";
import { formatDistanceToNow } from "date-fns";
import type { User } from "@/types";

export default function UsersPage() {
  const { data: users, isLoading } = useUsers();
  const createMutation = useCreateUser();
  const updateMutation = useUpdateUser();
  const archiveMutation = useArchiveUser();
  const disableTwoFactorMutation = useDisableUserTwoFactor();

  const [open, setOpen] = useState(false);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState<"admin" | "viewer">("viewer");

  const [editing, setEditing] = useState<User | null>(null);
  const [editEmail, setEditEmail] = useState("");
  const [editPassword, setEditPassword] = useState("");

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    await createMutation.mutateAsync({ email, password, role });
    setOpen(false);
    setEmail(""); setPassword(""); setRole("viewer");
  }

  function openEdit(u: User) {
    setEditing(u);
    setEditEmail(u.email);
    setEditPassword("");
  }

  async function handleEditSave(e: React.FormEvent) {
    e.preventDefault();
    if (!editing) return;
    await updateMutation.mutateAsync({
      id: editing.id,
      email: editEmail !== editing.email ? editEmail : undefined,
      password: editPassword || undefined,
    });
    setEditing(null);
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
                    {u.totp_enabled ? (
                      <Badge variant="outline" className="text-emerald-400 border-emerald-400/30">
                        <ShieldCheck className="h-3 w-3 mr-1" />2FA On
                      </Badge>
                    ) : (
                      <Badge variant="outline" className="text-muted-foreground">
                        <ShieldOff className="h-3 w-3 mr-1" />2FA Off
                      </Badge>
                    )}
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

                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="icon" className="h-8 w-8">
                      <MoreHorizontal className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem onClick={() => openEdit(u)}>
                      <Pencil className="h-4 w-4" />
                      Edit
                    </DropdownMenuItem>
                    {u.totp_enabled && (
                      <DropdownMenuItem
                        onClick={() => disableTwoFactorMutation.mutate(u.id)}
                        disabled={disableTwoFactorMutation.isPending}
                      >
                        <ShieldX className="h-4 w-4" />
                        Force-disable 2FA
                      </DropdownMenuItem>
                    )}
                    <DropdownMenuSeparator />
                    <DropdownMenuItem
                      className="text-destructive focus:text-destructive"
                      onClick={() => archiveMutation.mutate(u.id)}
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

      {/* Create dialog */}
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

      {/* Edit dialog */}
      <Dialog open={editing !== null} onOpenChange={(o) => !o && setEditing(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit User</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleEditSave} className="space-y-4">
            <div className="space-y-1.5">
              <Label>Email</Label>
              <Input
                type="email"
                value={editEmail}
                onChange={(e) => setEditEmail(e.target.value)}
                required
              />
            </div>
            <div className="space-y-1.5">
              <Label>Reset Password</Label>
              <Input
                type="password"
                value={editPassword}
                onChange={(e) => setEditPassword(e.target.value)}
                placeholder="Leave blank to keep their current password"
              />
            </div>
            <DialogFooter>
              <Button variant="outline" type="button" onClick={() => setEditing(null)}>
                Cancel
              </Button>
              <Button type="submit" disabled={updateMutation.isPending}>
                Save
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
