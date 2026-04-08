"use server";

import { db } from "@/lib/db";
import { needScores, CATEGORIES } from "@/lib/db/schema";
import { ensureUser } from "@/lib/db/ensure-user";
import { currentClaims } from "@/lib/session";
import { eq } from "drizzle-orm";

export async function updateNeedScores(scores: Record<string, number>) {
  const claims = await currentClaims();
  if (!claims) throw new Error("Not authenticated");
  const user = await ensureUser(claims.email);

  for (const cat of CATEGORIES) {
    const score = Math.max(0, Math.min(1, scores[cat] ?? 0));
    db.insert(needScores)
      .values({ userId: user.id, category: cat, score })
      .onConflictDoUpdate({
        target: [needScores.userId, needScores.category],
        set: { score },
      })
      .run();
  }
}

export async function getNeedScores(): Promise<Record<string, number>> {
  const claims = await currentClaims();
  if (!claims) throw new Error("Not authenticated");
  const user = await ensureUser(claims.email);

  const rows = db
    .select()
    .from(needScores)
    .where(eq(needScores.userId, user.id))
    .all();

  const result: Record<string, number> = {};
  for (const cat of CATEGORIES) result[cat] = 0;
  for (const row of rows) result[row.category] = row.score;
  return result;
}
