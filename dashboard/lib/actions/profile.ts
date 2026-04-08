"use server";

import { db } from "@/lib/db";
import { users } from "@/lib/db/schema";
import { ensureUser } from "@/lib/db/ensure-user";
import { currentClaims } from "@/lib/session";
import { eq } from "drizzle-orm";

async function authedUser() {
  const claims = await currentClaims();
  if (!claims) throw new Error("Not authenticated");
  return ensureUser(claims.email);
}

export async function updateBasics(data: {
  displayName: string;
  bio: string;
  accountType: string;
}) {
  const user = await authedUser();
  db.update(users)
    .set({
      displayName: data.displayName || null,
      bio: data.bio || null,
      accountType: data.accountType || "individual",
      updatedAt: new Date().toISOString(),
    })
    .where(eq(users.id, user.id))
    .run();
}

export async function updateLocation(data: {
  latitude: number;
  longitude: number;
}) {
  const user = await authedUser();
  // Simple cell ID: floor(lat) _ floor(lng) — good enough for ~69-mile cells
  const cellId = `${Math.floor(data.latitude)}_${Math.floor(data.longitude)}`;
  db.update(users)
    .set({
      latitude: data.latitude,
      longitude: data.longitude,
      cellId,
      updatedAt: new Date().toISOString(),
    })
    .where(eq(users.id, user.id))
    .run();
}

export async function getProfile() {
  const user = await authedUser();
  return user;
}
