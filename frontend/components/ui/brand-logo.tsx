"use client";

import { useEffect, useState } from "react";

interface BrandLogoProps {
  src?: string | null;
  alt: string;
  size?: number;
  className?: string;
}

// Falls back to the bundled monsee logo if no custom logo is configured, or
// if the configured URL fails to load (404, unreachable host, etc.).
export function BrandLogo({ src, alt, size = 24, className }: BrandLogoProps) {
  const [broken, setBroken] = useState(false);

  // Re-attempt the configured URL if it changes (e.g. settings load async
  // after the fallback was already shown for the initial undefined value).
  useEffect(() => setBroken(false), [src]);

  const resolved = !src || broken ? "/monsee.png" : src;

  return (
    // eslint-disable-next-line @next/next/no-img-element
    <img
      src={resolved}
      alt={alt}
      width={size}
      height={size}
      className={className}
      onError={() => setBroken(true)}
    />
  );
}
