"use server";

import { db } from "@/lib/db";
import { interests } from "@/lib/db/schema";
import { ensureUser } from "@/lib/db/ensure-user";
import { currentClaims } from "@/lib/session";
import { eq, and } from "drizzle-orm";

export async function setInterests(labels: string[]) {
  const claims = await currentClaims();
  if (!claims) throw new Error("Not authenticated");
  const user = await ensureUser(claims.email);

  // Replace all: delete existing, insert new.
  db.delete(interests).where(eq(interests.userId, user.id)).run();
  for (const label of labels) {
    if (label.trim()) {
      db.insert(interests)
        .values({ userId: user.id, label: label.trim().toLowerCase() })
        .onConflictDoNothing()
        .run();
    }
  }
}

export async function getInterests(): Promise<string[]> {
  const claims = await currentClaims();
  if (!claims) throw new Error("Not authenticated");
  const user = await ensureUser(claims.email);

  return db
    .select({ label: interests.label })
    .from(interests)
    .where(eq(interests.userId, user.id))
    .all()
    .map((r) => r.label);
}
