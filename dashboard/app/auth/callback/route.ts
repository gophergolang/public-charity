import { NextRequest, NextResponse } from "next/server";
import { AUTH_URL, SESSION_COOKIE } from "@/lib/env";

function baseUrl(req: NextRequest): { base: string; secure: boolean } {
  const proto = req.headers.get("x-forwarded-proto") ?? "https";
  const host = req.headers.get("host") ?? "localhost:3000";
  return { base: `${proto}://${host}`, secure: proto === "https" };
}

export async function GET(req: NextRequest) {
  const token = req.nextUrl.searchParams.get("token");
  const { base, secure } = baseUrl(req);

  if (!token) {
    return NextResponse.redirect(new URL("/login?e=missing_token", base));
  }

  const res = await fetch(`${AUTH_URL}/auth/verify`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ token }),
  });
  if (!res.ok) {
    return NextResponse.redirect(new URL("/login?e=invalid_token", base));
  }

  const data = (await res.json()) as { jwt: string; email: string };

  let maxAge = 60 * 60 * 24 * 30;
  try {
    const payload = JSON.parse(
      Buffer.from(data.jwt.split(".")[1].replace(/-/g, "+").replace(/_/g, "/"), "base64").toString("utf8"),
    ) as { exp: number };
    const secondsLeft = payload.exp - Math.floor(Date.now() / 1000);
    if (secondsLeft > 0) maxAge = secondsLeft;
  } catch {
    // keep fallback
  }

  const redirect = NextResponse.redirect(new URL("/", base));
  redirect.cookies.set(SESSION_COOKIE, data.jwt, {
    httpOnly: true,
    sameSite: "lax",
    secure,
    path: "/",
    maxAge,
  });
  return redirect;
}
