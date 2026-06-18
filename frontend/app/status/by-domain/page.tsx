import { ServiceStatusView } from "@/components/public/service-status-view";

export const dynamic = "force-dynamic";

export default async function DomainStatusPage({
  searchParams,
}: {
  searchParams: Promise<{ domain?: string }>;
}) {
  const { domain } = await searchParams;
  const host = domain ?? "";
  return (
    <ServiceStatusView
      endpoint={`/api/v1/by-domain?domain=${encodeURIComponent(host)}`}
      queryKey={["dedicated-page", "domain", host]}
    />
  );
}
