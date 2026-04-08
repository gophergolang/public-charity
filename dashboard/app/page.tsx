import { redirect } from "next/navigation";
import { currentClaims } from "@/lib/session";
import Link from "next/link";

export default async function Home() {
  const claims = await currentClaims();

  if (!claims) {
    return (
      <main>
        <h1>Public Charity</h1>
        <p className="muted">Connecting people who need help with people who can help.</p>
        <p>
          <Link href="/login">Sign in to get started</Link>
        </p>
      </main>
    );
  }

  redirect("/dashboard");
}
