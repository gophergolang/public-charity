import { NextRequest, NextResponse } from "next/server";
import { AUTH_URL } from "@/lib/env";

// Proxies the magic-link request to the auth service so that AUTH_URL
// stays server-side and we avoid browser CORS entirely.
export async function POST(req: NextRequest) {
  let body: unknown;
  try {
    body = await req.json();
  } catch {
    return NextResponse.json({ error: "invalid json" }, { status: 400 });
  }

  const res = await fetch(`${AUTH_URL}/auth/request`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });

  const text = await res.text();
  return new NextResponse(text, {
    status: res.status,
    headers: { "Content-Type": res.headers.get("Content-Type") ?? "application/json" },
  });
}
