import { sqliteTable, text, integer, real, uniqueIndex, index } from "drizzle-orm/sqlite-core";
import { sql } from "drizzle-orm";

const id = () =>
  text("id")
    .primaryKey()
    .$defaultFn(() => crypto.randomUUID().replace(/-/g, ""));

const timestamp = (name: string) =>
  text(name).default(sql`(datetime('now'))`);

// ── Users ──────────────────────────────────────────────────

export const users = sqliteTable("users", {
  id: id(),
  email: text("email").notNull().unique(),
  displayName: text("display_name"),
  bio: text("bio"),
  latitude: real("latitude"),
  longitude: real("longitude"),
  cellId: text("cell_id"),
  accountType: text("account_type").notNull().default("individual"),
  onboardingStep: integer("onboarding_step").notNull().default(0),
  contactPrefs: text("contact_prefs"), // JSON
  createdAt: timestamp("created_at").notNull(),
  updatedAt: timestamp("updated_at").notNull(),
});

// ── Need Scores ────────────────────────────────────────────

export const CATEGORIES = [
  // Social
  "companionship", "community", "family", "language",
  // Health & Wellbeing
  "mental_health", "physical_health", "nutrition", "substance_recovery",
  // Daily Life
  "housing", "transport", "errands", "personal_care",
  // Economic Opportunities
  "employment", "training", "benefits", "budgeting",
  // Practical
  "digital", "admin", "legal", "childcare",
] as const;

export type Section = {
  id: string;
  label: string;
  description: string;
  color: string;
  categories: readonly Category[];
};

export const SECTIONS: Section[] = [
  {
    id: "social",
    label: "Social & Connection",
    description: "Nobody should have to face things alone",
    color: "#f0e6ff",
    categories: ["companionship", "community", "family", "language"],
  },
  {
    id: "health",
    label: "Health & Wellbeing",
    description: "Your health matters — physical and mental",
    color: "#e6f7ed",
    categories: ["mental_health", "physical_health", "nutrition", "substance_recovery"],
  },
  {
    id: "daily",
    label: "Daily Life",
    description: "The everyday things that keep life running",
    color: "#fff4e6",
    categories: ["housing", "transport", "errands", "personal_care"],
  },
  {
    id: "economic",
    label: "Economic Opportunities",
    description: "Everyone deserves a fair chance",
    color: "#e6f0ff",
    categories: ["employment", "training", "benefits", "budgeting"],
  },
  {
    id: "practical",
    label: "Practical Support",
    description: "Sometimes you just need a hand with something specific",
    color: "#fce6e6",
    categories: ["digital", "admin", "legal", "childcare"],
  },
];

export type Category = (typeof CATEGORIES)[number];

export const needScores = sqliteTable(
  "need_scores",
  {
    userId: text("user_id")
      .notNull()
      .references(() => users.id, { onDelete: "cascade" }),
    category: text("category").notNull(),
    score: real("score").notNull().default(0),
  },
  (t) => [uniqueIndex("need_scores_pk").on(t.userId, t.category)],
);

// ── Interests ──────────────────────────────────────────────

export const interests = sqliteTable(
  "interests",
  {
    id: integer("id").primaryKey({ autoIncrement: true }),
    userId: text("user_id")
      .notNull()
      .references(() => users.id, { onDelete: "cascade" }),
    label: text("label").notNull(),
  },
  (t) => [uniqueIndex("interests_user_label").on(t.userId, t.label)],
);

// ── Offers ─────────────────────────────────────────────────

export const offers = sqliteTable(
  "offers",
  {
    id: integer("id").primaryKey({ autoIncrement: true }),
    userId: text("user_id")
      .notNull()
      .references(() => users.id, { onDelete: "cascade" }),
    category: text("category").notNull(),
    description: text("description").notNull(),
    available: integer("available").notNull().default(1),
    createdAt: timestamp("created_at").notNull(),
  },
  (t) => [index("idx_offers_user").on(t.userId)],
);

// ── Surplus ────────────────────────────────────────────────

export const surplus = sqliteTable(
  "surplus",
  {
    id: integer("id").primaryKey({ autoIncrement: true }),
    userId: text("user_id")
      .notNull()
      .references(() => users.id, { onDelete: "cascade" }),
    category: text("category").notNull(),
    description: text("description").notNull(),
    expiresAt: text("expires_at"),
    createdAt: timestamp("created_at").notNull(),
  },
  (t) => [index("idx_surplus_user").on(t.userId)],
);

// ── Availability ───────────────────────────────────────────
// A row exists = the user is free at that day+slot.
// No row = not available. Simple boolean presence model.

export const DAYS = ["mon", "tue", "wed", "thu", "fri", "sat", "sun"] as const;
export const SLOTS = ["morning", "afternoon", "evening"] as const;

export const availability = sqliteTable(
  "availability",
  {
    userId: text("user_id")
      .notNull()
      .references(() => users.id, { onDelete: "cascade" }),
    day: text("day").notNull(),
    slot: text("slot").notNull(),
    note: text("note"),
  },
  (t) => [uniqueIndex("availability_pk").on(t.userId, t.day, t.slot)],
);

// ── Messages (Inbox) ───────────────────────────────────────

export const messages = sqliteTable(
  "messages",
  {
    id: id(),
    recipientId: text("recipient_id")
      .notNull()
      .references(() => users.id, { onDelete: "cascade" }),
    senderId: text("sender_id").references(() => users.id, {
      onDelete: "set null",
    }),
    senderType: text("sender_type").notNull().default("user"), // user | system | ai_agent
    category: text("category"),
    subject: text("subject").notNull(),
    body: text("body").notNull(),
    ruleId: text("rule_id"),
    read: integer("read").notNull().default(0),
    archived: integer("archived").notNull().default(0),
    emailSent: integer("email_sent").notNull().default(0),
    createdAt: timestamp("created_at").notNull(),
  },
  (t) => [
    index("idx_messages_recipient").on(t.recipientId),
    index("idx_messages_unread").on(t.recipientId, t.read),
  ],
);
