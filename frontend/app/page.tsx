import { redirect } from "next/navigation";
import { PublicStatusPage } from "./status-page";

export const dynamic = "force-dynamic";

export default function HomePage() {
  if (process.env.NEXT_PUBLIC_STATUS_PAGE !== "true") {
    redirect("/admin/services");
  }

  return <PublicStatusPage />;
}
