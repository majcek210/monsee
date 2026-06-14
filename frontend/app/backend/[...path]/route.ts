import { NextRequest, NextResponse } from "next/server";

const BACKEND = process.env.BACKEND_URL ?? "http://localhost:8080";

async function handler(
  req: NextRequest,
  { params }: { params: Promise<{ path: string[] }> }
) {
  const { path } = await params;
  const target = `${BACKEND}/${path.join("/")}${req.nextUrl.search}`;

  const headers = new Headers();
  const ct = req.headers.get("content-type");
  if (ct) headers.set("content-type", ct);
  const auth = req.headers.get("authorization");
  if (auth) headers.set("authorization", auth);
  const cookie = req.headers.get("cookie");
  if (cookie) headers.set("cookie", cookie);

  const body =
    req.method !== "GET" && req.method !== "HEAD"
      ? await req.arrayBuffer()
      : undefined;

  let upstream: Response;
  try {
    upstream = await fetch(target, {
      method: req.method,
      headers,
      body: body as BodyInit | undefined,
    });
  } catch (err) {
    console.error("[proxy] backend unreachable:", target, err);
    return NextResponse.json({ error: "Backend unreachable" }, { status: 502 });
  }

  const res = new NextResponse(upstream.body, {
    status: upstream.status,
    statusText: upstream.statusText,
  });

  const skip = new Set(["content-encoding", "transfer-encoding", "connection"]);
  upstream.headers.forEach((value, key) => {
    if (skip.has(key.toLowerCase())) return;
    if (key.toLowerCase() === "set-cookie") {
      res.headers.append("set-cookie", value);
    } else {
      res.headers.set(key, value);
    }
  });

  return res;
}

export const GET = handler;
export const POST = handler;
export const PATCH = handler;
export const DELETE = handler;
export const PUT = handler;
