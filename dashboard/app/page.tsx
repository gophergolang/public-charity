import Link from "next/link";
import { currentClaims } from "@/lib/session";

export default async function Home() {
  const claims = await currentClaims();

  if (!claims) {
    return (
      <main>
        <h1>Dashboard</h1>
        <p className="muted">You are not signed in.</p>
        <p>
          <Link href="/login">Sign in</Link>
        </p>
      </main>
    );
  }

  return (
    <main>
      <h1>Dashboard</h1>
      <p>
        Signed in as <strong>{claims.email}</strong>.
      </p>
      <p className="muted">
        Session expires {new Date(claims.exp * 1000).toLocaleString()}.
      </p>
      <form action="/api/auth/logout" method="post">
        <button type="submit">Sign out</button>
      </form>
    </main>
  );
}
