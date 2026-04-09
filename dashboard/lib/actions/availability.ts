"use server";

import { db } from "@/lib/db";
import { availability } from "@/lib/db/schema";
import { ensureUser } from "@/lib/db/ensure-user";
import { currentClaims } from "@/lib/session";
import { eq, and } from "drizzle-orm";

export async function getAvailabilityList(): Promise<string[]> {
  const claims = await currentClaims();
  if (!claims) throw new Error("Not authenticated");
  const user = await ensureUser(claims.email);

  return db
    .select({ day: availability.day, slot: availability.slot })
    .from(availability)
    .where(eq(availability.userId, user.id))
    .all()
    .map((r) => `${r.day}-${r.slot}`);
}

export async function toggleSlot(day: string, slot: string): Promise<boolean> {
  const claims = await currentClaims();
  if (!claims) throw new Error("Not authenticated");
  const user = await ensureUser(claims.email);

  const existing = db
    .select()
    .from(availability)
    .where(
      and(
        eq(availability.userId, user.id),
        eq(availability.day, day),
        eq(availability.slot, slot),
      ),
    )
    .get();

  if (existing) {
    db.delete(availability)
      .where(
        and(
          eq(availability.userId, user.id),
          eq(availability.day, day),
          eq(availability.slot, slot),
        ),
      )
      .run();
    return false; // now off
  } else {
    db.insert(availability)
      .values({ userId: user.id, day, slot })
      .run();
    return true; // now on
  }
}
