import { ServiceStatusView } from "@/components/public/service-status-view";

export const dynamic = "force-dynamic";

export default async function DedicatedStatusPage({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;
  return (
    <ServiceStatusView
      endpoint={`/api/v1/pages/${encodeURIComponent(slug)}`}
      queryKey={["dedicated-page", "slug", slug]}
    />
  );
}
