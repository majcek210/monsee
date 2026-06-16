"use client";

import type { DailyUptimeStatus } from "@/types";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { format } from "date-fns";

const statusColor: Record<string, string> = {
  up: "bg-emerald-400",
  degraded: "bg-amber-400",
  down: "bg-red-400",
  no_data: "bg-muted",
};

interface Props {
  days: DailyUptimeStatus[];
  className?: string;
}

export function UptimeBar({ days, className }: Props) {
  const displayed = days.slice(-90);

  return (
    <TooltipProvider delayDuration={150}>
      <div className={`flex gap-px ${className ?? ""}`}>
        {displayed.map((day) => (
          <Tooltip key={day.date}>
            <TooltipTrigger asChild>
              <div
                className={`flex-1 h-8 rounded-sm ${statusColor[day.status] ?? "bg-muted"}`}
                style={{ minWidth: 2 }}
              />
            </TooltipTrigger>
            <TooltipContent>
              <p className="text-xs font-medium">{format(new Date(day.date), "MMM d, yyyy")}</p>
              <p className="text-xs text-muted-foreground capitalize">{day.status === "no_data" ? "No data" : `${day.uptime_percent.toFixed(1)}% uptime`}</p>
            </TooltipContent>
          </Tooltip>
        ))}
      </div>
    </TooltipProvider>
  );
}
