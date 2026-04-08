"use server";

import { db } from "@/lib/db";
import { offers } from "@/lib/db/schema";
import { ensureUser } from "@/lib/db/ensure-user";
import { currentClaims } from "@/lib/session";
import { eq } from "drizzle-orm";

export async function setOffers(
  items: { category: string; description: string }[],
) {
  const claims = await currentClaims();
  if (!claims) throw new Error("Not authenticated");
  const user = await ensureUser(claims.email);

  db.delete(offers).where(eq(offers.userId, user.id)).run();
  for (const item of items) {
    if (item.description.trim()) {
      db.insert(offers)
        .values({
          userId: user.id,
          category: item.category,
          description: item.description.trim(),
        })
        .run();
    }
  }
}

export async function getOffers() {
  const claims = await currentClaims();
  if (!claims) throw new Error("Not authenticated");
  const user = await ensureUser(claims.email);

  return db.select().from(offers).where(eq(offers.userId, user.id)).all();
}
