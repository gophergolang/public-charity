import { db } from "./index";
import { users } from "./schema";
import { eq } from "drizzle-orm";

export type User = typeof users.$inferSelect;

/**
 * Ensures a user row exists for the given email (from JWT claims).
 * Returns the user record. Called at the start of every authenticated action.
 */
export async function ensureUser(email: string): Promise<User> {
  // Insert if not exists.
  db.insert(users)
    .values({ email })
    .onConflictDoNothing({ target: users.email })
    .run();

  const user = db.select().from(users).where(eq(users.email, email)).get();
  if (!user) throw new Error("Failed to create user");
  return user;
}
