// Server-only env. Throw early if misconfigured.
export const AUTH_URL = process.env.AUTH_URL ?? "http://localhost:8080";
export const SESSION_COOKIE = process.env.SESSION_COOKIE ?? "pc_session";
