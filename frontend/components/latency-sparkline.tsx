"use client";

import type { ResponseTimePoint } from "@/types";

interface Props {
  points: ResponseTimePoint[];
  height?: number;
  className?: string;
}

export function LatencySparkline({ points, height = 40, className }: Props) {
  if (!points.length) return null;

  const values = points.map((p) => p.response_time_ms).filter((v) => v > 0);
  if (!values.length) return null;

  const max = Math.max(...values);
  const min = Math.min(...values);
  const range = max - min || 1;

  const w = 200;
  const h = height;
  const step = w / (values.length - 1 || 1);

  const pts = values
    .map((v, i) => {
      const x = i * step;
      const y = h - ((v - min) / range) * (h - 4) - 2;
      return `${x},${y}`;
    })
    .join(" ");

  return (
    <svg
      viewBox={`0 0 ${w} ${h}`}
      width={w}
      height={h}
      className={className}
      aria-hidden
    >
      <polyline
        points={pts}
        fill="none"
        stroke="oklch(0.72 0.17 265)"
        strokeWidth={1.5}
        strokeLinejoin="round"
        strokeLinecap="round"
      />
    </svg>
  );
}
