"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import {
  Activity,
  AlertTriangle,
  Bell,
  ClipboardList,
  Globe,
  Key,
  LayoutDashboard,
  LogOut,
  ScrollText,
  Settings,
  Shield,
  Users,
  Webhook,
  Wrench,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { authApi } from "@/lib/api/auth";
import { useCurrentUser } from "@/lib/hooks/use-auth";
import { toast } from "sonner";

const navItems = [
  { href: "/admin", label: "Overview", icon: LayoutDashboard, exact: true },
  { href: "/admin/services", label: "Services", icon: Globe },
  { href: "/admin/incidents", label: "Incidents", icon: AlertTriangle },
  { href: "/admin/maintenance", label: "Maintenance", icon: Wrench },
  { href: "/admin/api-keys", label: "API Keys", icon: Key },
  { href: "/admin/notifications", label: "Notifications", icon: Bell },
  { href: "/admin/webhooks", label: "Webhooks", icon: Webhook },
  { href: "/admin/users", label: "Users", icon: Users, adminOnly: true },
  { href: "/admin/security", label: "Security", icon: Shield },
  { href: "/admin/audit-log", label: "Audit Log", icon: ScrollText, adminOnly: true },
  { href: "/admin/settings", label: "Settings", icon: Settings, adminOnly: true },
];

export function Sidebar() {
  const pathname = usePathname();
  const router = useRouter();
  const { data: me } = useCurrentUser();

  async function handleLogout() {
    try {
      await authApi.logout();
      router.push("/login");
    } catch {
      toast.error("Logout failed");
    }
  }

  const items = navItems.filter((item) => !item.adminOnly || me?.role === "admin");

  return (
    <aside className="flex h-screen w-56 flex-col border-r border-border bg-card">
      <div className="flex h-14 items-center px-4 border-b border-border">
        <Activity className="h-5 w-5 text-primary mr-2" />
        <span className="font-semibold text-sm tracking-tight">monsee</span>
      </div>

      <nav className="flex-1 overflow-y-auto p-2 space-y-0.5">
        {items.map(({ href, label, icon: Icon, exact }) => {
          const active = exact ? pathname === href : (pathname === href || pathname.startsWith(href + "/"));
          return (
            <Link
              key={href}
              href={href}
              className={cn(
                "flex items-center gap-2.5 rounded-md px-3 py-2 text-sm transition-colors",
                active
                  ? "bg-primary/10 text-primary font-medium"
                  : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
              )}
            >
              <Icon className="h-4 w-4 shrink-0" />
              {label}
            </Link>
          );
        })}
      </nav>

      <div className="p-2 border-t border-border">
        <button
          onClick={handleLogout}
          className="flex w-full items-center gap-2.5 rounded-md px-3 py-2 text-sm text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors"
        >
          <LogOut className="h-4 w-4 shrink-0" />
          Logout
        </button>
      </div>
    </aside>
  );
}
