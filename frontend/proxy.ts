import { NextRequest, NextResponse } from "next/server";

// Cheap presence check only — the backend independently validates the JWT
// signature/expiry/role on every /admin/* API call. This just keeps logged
// out users from landing on an empty admin shell.
export function proxy(req: NextRequest) {
  const session = req.cookies.get("session");
  if (!session) {
    const loginUrl = new URL("/login", req.url);
    return NextResponse.redirect(loginUrl);
  }
  return NextResponse.next();
}

export const config = {
  matcher: ["/admin/:path*"],
};
