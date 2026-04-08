// Server-only env.
export const AUTH_URL = process.env.AUTH_URL ?? "http://localhost:8080";
export const SESSION_COOKIE = process.env.SESSION_COOKIE ?? "pc_session";
export const JWT_PUBLIC_KEY = process.env.JWT_PUBLIC_KEY ?? "";
