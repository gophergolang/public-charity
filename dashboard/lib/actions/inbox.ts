"use server";

import { db } from "@/lib/db";
import { messages } from "@/lib/db/schema";
import { ensureUser } from "@/lib/db/ensure-user";
import { currentClaims } from "@/lib/session";
import { eq, and, desc } from "drizzle-orm";

export async function getMessages() {
  const claims = await currentClaims();
  if (!claims) throw new Error("Not authenticated");
  const user = await ensureUser(claims.email);

  return db
    .select()
    .from(messages)
    .where(
      and(eq(messages.recipientId, user.id), eq(messages.archived, 0)),
    )
    .orderBy(desc(messages.createdAt))
    .all();
}

export async function getMessage(id: string) {
  const claims = await currentClaims();
  if (!claims) throw new Error("Not authenticated");
  const user = await ensureUser(claims.email);

  return db
    .select()
    .from(messages)
    .where(and(eq(messages.id, id), eq(messages.recipientId, user.id)))
    .get();
}

export async function markRead(id: string) {
  const claims = await currentClaims();
  if (!claims) throw new Error("Not authenticated");
  const user = await ensureUser(claims.email);

  db.update(messages)
    .set({ read: 1 })
    .where(and(eq(messages.id, id), eq(messages.recipientId, user.id)))
    .run();
}

export async function archiveMessage(id: string) {
  const claims = await currentClaims();
  if (!claims) throw new Error("Not authenticated");
  const user = await ensureUser(claims.email);

  db.update(messages)
    .set({ archived: 1 })
    .where(and(eq(messages.id, id), eq(messages.recipientId, user.id)))
    .run();
}

export async function unreadCount(): Promise<number> {
  const claims = await currentClaims();
  if (!claims) return 0;
  const user = await ensureUser(claims.email);

  const rows = db
    .select()
    .from(messages)
    .where(
      and(
        eq(messages.recipientId, user.id),
        eq(messages.read, 0),
        eq(messages.archived, 0),
      ),
    )
    .all();
  return rows.length;
}
