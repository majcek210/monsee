import { NextRequest, NextResponse } from "next/server";

// The host this app is primarily served on (e.g. "status.cryvex.xyz").
// Any other Host hitting "/" is treated as a customer custom domain and
// rewritten to the dedicated by-domain status page. Configure via env.
const PRIMARY_HOST = (process.env.APP_PRIMARY_HOST || process.env.NEXT_PUBLIC_APP_HOST || "")
  .toLowerCase()
  .trim();

function hostname(req: NextRequest): string {
  const host = req.headers.get("host") || "";
  return host.split(":")[0].toLowerCase();
}

function isLocalHost(h: string): boolean {
  return h === "localhost" || h === "127.0.0.1" || h === "0.0.0.0";
}

export function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl;

  // Admin gate: cheap presence check only — the backend independently validates
  // the JWT on every /admin/* API call. This just keeps logged-out users from
  // landing on an empty admin shell.
  if (pathname.startsWith("/admin")) {
    const session = req.cookies.get("session");
    if (!session) {
      return NextResponse.redirect(new URL("/login", req.url));
    }
    return NextResponse.next();
  }

  // Custom-domain routing: a non-primary Host hitting the root renders that
  // domain's dedicated service page. The page itself resolves/validates the
  // domain against the backend, so middleware stays cheap (no fetch here).
  if (pathname === "/" && PRIMARY_HOST) {
    const host = hostname(req);
    if (host && host !== PRIMARY_HOST && !isLocalHost(host)) {
      const url = req.nextUrl.clone();
      url.pathname = "/status/by-domain";
      url.searchParams.set("domain", host);
      return NextResponse.rewrite(url);
    }
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/", "/admin/:path*"],
};
