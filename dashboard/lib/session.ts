import { cookies } from "next/headers";
import { AUTH_URL, SESSION_COOKIE } from "./env";

export type Claims = {
  email: string;
  sub: string;
  iat: number;
  exp: number;
};

// Validates the JWT with the auth service. Returns null if invalid/expired.
export async function currentClaims(): Promise<Claims | null> {
  const store = await cookies();
  const c = store.get(SESSION_COOKIE);
  if (!c) return null;

  try {
    const res = await fetch(`${AUTH_URL}/auth/validate`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ jwt: c.value }),
    });
    if (!res.ok) return null;
    return (await res.json()) as Claims;
  } catch {
    return null;
  }
}
