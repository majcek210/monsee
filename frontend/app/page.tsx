import { redirect } from "next/navigation";
import { PublicStatusPage } from "./status-page";

export const dynamic = "force-dynamic";

async function getPublicSettings() {
  try {
    const backendUrl = process.env.BACKEND_URL || "http://localhost:8080";
    const res = await fetch(`${backendUrl}/api/v1/settings`, { cache: "no-store" });
    if (!res.ok) return null;
    return res.json() as Promise<{ site_title: string; logo_url: string; public_status_enabled?: boolean }>;
  } catch {
    return null;
  }
}

export default async function HomePage() {
  const settings = await getPublicSettings();

  // Fall back to env var if settings fetch fails (e.g. before DB is up)
  const enabled =
    settings?.public_status_enabled ??
    (process.env.NEXT_PUBLIC_STATUS_PAGE === "true");

  if (!enabled) {
    redirect("/login");
  }

  return <PublicStatusPage />;
}
