"use server";

import { db } from "@/lib/db";
import { users, messages } from "@/lib/db/schema";
import { ensureUser } from "@/lib/db/ensure-user";
import { currentClaims } from "@/lib/session";
import { eq } from "drizzle-orm";

export async function getOnboardingStep(): Promise<number> {
  const claims = await currentClaims();
  if (!claims) return 0;
  const user = await ensureUser(claims.email);
  return user.onboardingStep;
}

export async function advanceStep(step: number) {
  const claims = await currentClaims();
  if (!claims) throw new Error("Not authenticated");
  const user = await ensureUser(claims.email);

  // Only advance forward, never regress.
  if (step > user.onboardingStep) {
    db.update(users)
      .set({ onboardingStep: step, updatedAt: new Date().toISOString() })
      .where(eq(users.id, user.id))
      .run();
  }
}

export async function completeOnboarding() {
  const claims = await currentClaims();
  if (!claims) throw new Error("Not authenticated");
  const user = await ensureUser(claims.email);

  db.update(users)
    .set({ onboardingStep: 7, updatedAt: new Date().toISOString() })
    .where(eq(users.id, user.id))
    .run();

  // Insert welcome message.
  db.insert(messages)
    .values({
      recipientId: user.id,
      senderType: "system",
      subject: "Welcome to Public Charity",
      body: "Thanks for setting up your profile! We'll use what you've shared to connect you with people nearby who can help — and who you can help too.\n\nYou can update your profile anytime from the dashboard.",
    })
    .run();
}
