import { NextRequest, NextResponse } from "next/server";
import { timingSafeEqual } from "crypto";
import { db } from "@/lib/db";
import { users, needScores, interests, offers, availability } from "@/lib/db/schema";

const API_KEY = process.env.DASHBOARD_API_KEY;

function checkAuth(req: NextRequest): boolean {
  if (!API_KEY) return false;
  const header = req.headers.get("authorization") ?? "";
  const expected = `Bearer ${API_KEY}`;
  if (header.length !== expected.length) return false;
  return timingSafeEqual(Buffer.from(header), Buffer.from(expected));
}

export async function GET(req: NextRequest) {
  if (!checkAuth(req)) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  // Batch-load all related data in 5 queries (not 4N+1).
  const allUsers = db.select().from(users).all();
  const allScores = db.select().from(needScores).all();
  const allInterests = db.select().from(interests).all();
  const allOffers = db.select().from(offers).all();
  const allAvailability = db.select().from(availability).all();

  // Index by userId.
  const scoresByUser = new Map<string, Record<string, number>>();
  for (const row of allScores) {
    if (!scoresByUser.has(row.userId)) scoresByUser.set(row.userId, {});
    scoresByUser.get(row.userId)![row.category] = row.score;
  }

  const interestsByUser = new Map<string, string[]>();
  for (const row of allInterests) {
    if (!interestsByUser.has(row.userId)) interestsByUser.set(row.userId, []);
    interestsByUser.get(row.userId)!.push(row.label);
  }

  const offersByUser = new Map<string, { category: string; description: string }[]>();
  for (const row of allOffers) {
    if (!offersByUser.has(row.userId)) offersByUser.set(row.userId, []);
    offersByUser.get(row.userId)!.push({ category: row.category, description: row.description });
  }

  const availByUser = new Map<string, string[]>();
  for (const row of allAvailability) {
    if (!availByUser.has(row.userId)) availByUser.set(row.userId, []);
    availByUser.get(row.userId)!.push(`${row.day}-${row.slot}`);
  }

  const result = allUsers.map((u) => ({
    id: u.id,
    email: u.email,
    display_name: u.displayName ?? "",
    cell_id: u.cellId ?? "",
    interests: interestsByUser.get(u.id) ?? [],
    need_scores: scoresByUser.get(u.id) ?? {},
    offers: offersByUser.get(u.id) ?? [],
    availability: availByUser.get(u.id) ?? [],
  }));

  return NextResponse.json(result);
}
