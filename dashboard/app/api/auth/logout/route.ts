import { NextRequest, NextResponse } from "next/server";
import { SESSION_COOKIE } from "@/lib/env";

export async function POST(req: NextRequest) {
  const proto = req.headers.get("x-forwarded-proto") ?? "https";
  const host = req.headers.get("host") ?? "localhost:3000";
  const base = `${proto}://${host}`;

  const res = NextResponse.redirect(new URL("/", base), 303);
  res.cookies.set(SESSION_COOKIE, "", {
    httpOnly: true,
    sameSite: "lax",
    path: "/",
    maxAge: 0,
  });
  return res;
}
