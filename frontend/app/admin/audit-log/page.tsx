"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { ChevronLeft, ChevronRight } from "lucide-react";
import type { AuditLogResponse } from "@/types";
import { formatDistanceToNow } from "date-fns";

const PAGE_SIZE = 50;

function useAuditLog(resource: string, offset: number) {
  const params = new URLSearchParams({ limit: String(PAGE_SIZE), offset: String(offset) });
  if (resource) params.set("resource", resource);

  return useQuery({
    queryKey: ["audit-log", resource, offset],
    queryFn: () => api.get<AuditLogResponse>(`/admin/audit-log?${params}`),
    staleTime: 30_000,
  });
}

const actionColors: Record<string, string> = {
  create: "text-emerald-400",
  update: "text-blue-400",
  archive: "text-amber-400",
  delete: "text-red-400",
};

export default function AuditLogPage() {
  const [resource, setResource] = useState("");
  const [resourceFilter, setResourceFilter] = useState("");
  const [offset, setOffset] = useState(0);

  const { data, isLoading } = useAuditLog(resourceFilter, offset);

  function applyFilter(e: React.FormEvent) {
    e.preventDefault();
    setResourceFilter(resource);
    setOffset(0);
  }

  const total = data?.total ?? 0;
  const page = Math.floor(offset / PAGE_SIZE) + 1;
  const pages = Math.ceil(total / PAGE_SIZE);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Audit Log</h1>
        <p className="text-sm text-muted-foreground mt-0.5">All admin write actions</p>
      </div>

      <form onSubmit={applyFilter} className="flex gap-2 max-w-sm">
        <Input
          value={resource}
          onChange={(e) => setResource(e.target.value)}
          placeholder="Filter by resource (e.g. monitor)"
        />
        <Button type="submit" variant="outline" size="sm">Filter</Button>
        {resourceFilter && (
          <Button type="button" variant="ghost" size="sm" onClick={() => { setResourceFilter(""); setResource(""); }}>
            Clear
          </Button>
        )}
      </form>

      {isLoading ? (
        <div className="space-y-2">
          {[...Array(8)].map((_, i) => <Skeleton key={i} className="h-14 w-full" />)}
        </div>
      ) : data?.entries.length === 0 ? (
        <Card>
          <CardContent className="flex items-center justify-center py-12">
            <p className="text-muted-foreground text-sm">No audit log entries found.</p>
          </CardContent>
        </Card>
      ) : (
        <>
          <div className="space-y-1.5">
            {data?.entries.map((entry) => (
              <Card key={entry.id}>
                <CardContent className="flex items-start gap-3 py-3">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 flex-wrap">
                      <span className={`text-xs font-mono font-semibold uppercase ${actionColors[entry.action] ?? "text-muted-foreground"}`}>
                        {entry.action}
                      </span>
                      <span className="text-sm font-medium">{entry.resource}</span>
                      {entry.resource_id && (
                        <span className="text-xs text-muted-foreground font-mono truncate max-w-[120px]">
                          {entry.resource_id}
                        </span>
                      )}
                    </div>
                    <div className="flex items-center gap-2 mt-0.5">
                      {entry.ip && <span className="text-xs text-muted-foreground">{entry.ip}</span>}
                      <span className="text-xs text-muted-foreground">
                        {formatDistanceToNow(new Date(entry.created_at), { addSuffix: true })}
                      </span>
                    </div>
                    {Array.isArray(entry.diff?.fields) && (
                      <p className="text-xs text-muted-foreground mt-0.5">
                        Fields: {(entry.diff!.fields as string[]).join(", ")}
                      </p>
                    )}
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>

          {pages > 1 && (
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">
                Page {page} of {pages} ({total} total)
              </span>
              <div className="flex gap-1">
                <Button
                  variant="outline"
                  size="icon"
                  className="h-8 w-8"
                  disabled={offset === 0}
                  onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
                >
                  <ChevronLeft className="h-4 w-4" />
                </Button>
                <Button
                  variant="outline"
                  size="icon"
                  className="h-8 w-8"
                  disabled={offset + PAGE_SIZE >= total}
                  onClick={() => setOffset(offset + PAGE_SIZE)}
                >
                  <ChevronRight className="h-4 w-4" />
                </Button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}
