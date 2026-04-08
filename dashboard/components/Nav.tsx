import Link from "next/link";
import { currentClaims } from "@/lib/session";
import { unreadCount } from "@/lib/actions/inbox";

export default async function Nav() {
  const claims = await currentClaims();
  if (!claims) return null;

  const unread = await unreadCount();

  return (
    <nav className="top-nav">
      <Link href="/dashboard" className="nav-brand">Public Charity</Link>
      <div className="nav-links">
        <Link href="/dashboard">My State</Link>
        <Link href="/dashboard/inbox">
          Inbox{unread > 0 && <span className="badge">{unread}</span>}
        </Link>
        <Link href="/dashboard/offers">Offers</Link>
        <Link href="/dashboard/availability">Free times</Link>
      </div>
      <div className="nav-right">
        <span className="muted">{claims.email}</span>
        <form action="/api/auth/logout" method="post" style={{ display: "inline" }}>
          <button type="submit" className="link">Sign out</button>
        </form>
      </div>
    </nav>
  );
}
