import { db } from "@/lib/db";
import { surplus } from "@/lib/db/schema";
import { ensureUser } from "@/lib/db/ensure-user";
import { currentClaims } from "@/lib/session";
import { eq } from "drizzle-orm";

export default async function SurplusPage() {
  const claims = await currentClaims();
  if (!claims) return null;
  const user = await ensureUser(claims.email);

  const items = db
    .select()
    .from(surplus)
    .where(eq(surplus.userId, user.id))
    .all();

  return (
    <main>
      <h1>Surplus</h1>
      {items.length === 0 ? (
        <p className="muted">Nothing listed as surplus yet.</p>
      ) : (
        <div className="offer-list">
          {items.map((s) => (
            <div key={s.id} className="card">
              <span className="tag">{s.category}</span>
              <p>{s.description}</p>
              {s.expiresAt && (
                <p className="muted">Expires: {s.expiresAt}</p>
              )}
            </div>
          ))}
        </div>
      )}
    </main>
  );
}
