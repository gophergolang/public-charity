import { NextRequest, NextResponse } from "next/server";

const PROTECTED = ["/onboarding", "/dashboard"];
const SESSION_COOKIE = process.env.SESSION_COOKIE ?? "pc_session";

export function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl;
  if (PROTECTED.some((p) => pathname.startsWith(p))) {
    if (!req.cookies.get(SESSION_COOKIE)) {
      return NextResponse.redirect(new URL("/login", req.url));
    }
  }
  return NextResponse.next();
}

export const config = {
  matcher: ["/onboarding/:path*", "/dashboard/:path*"],
};
