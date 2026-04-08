import { cookies } from "next/headers";
import { createVerify } from "crypto";
import { SESSION_COOKIE, JWT_PUBLIC_KEY } from "./env";

export type Claims = {
  email: string;
  sub: string;
  iat: number;
  exp: number;
};

// Verifies an RS256 JWT using the public key.
// The dashboard cannot forge tokens — it only has the public key.
function verifyJwt(token: string): Claims | null {
  const parts = token.split(".");
  if (parts.length !== 3) return null;

  try {
    const header = JSON.parse(Buffer.from(parts[0], "base64url").toString());
    if (header.alg !== "RS256") return null;

    const verifier = createVerify("RSA-SHA256");
    verifier.update(parts[0] + "." + parts[1]);
    const signatureValid = verifier.verify(
      JWT_PUBLIC_KEY,
      parts[2],
      "base64url",
    );
    if (!signatureValid) return null;

    const claims = JSON.parse(
      Buffer.from(parts[1], "base64url").toString(),
    ) as Claims;
    if (Date.now() / 1000 > claims.exp) return null;

    return claims;
  } catch {
    return null;
  }
}

export async function currentClaims(): Promise<Claims | null> {
  if (!JWT_PUBLIC_KEY) return null;
  const store = await cookies();
  const c = store.get(SESSION_COOKIE);
  if (!c) return null;
  return verifyJwt(c.value);
}
