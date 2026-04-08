import { NextRequest, NextResponse } from "next/server";
import { db } from "@/lib/db";
import { users, messages } from "@/lib/db/schema";
import { eq } from "drizzle-orm";
import { forwardToEmail } from "@/lib/email/forward";

const API_KEY = process.env.DASHBOARD_API_KEY;

export async function POST(req: NextRequest) {
  const authHeader = req.headers.get("authorization");
  if (!API_KEY || authHeader !== `Bearer ${API_KEY}`) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  const body = (await req.json()) as {
    recipient_email: string;
    sender_type?: string;
    category?: string;
    subject: string;
    body: string;
    rule_id?: string;
  };

  if (!body.recipient_email || !body.subject || !body.body) {
    return NextResponse.json(
      { error: "recipient_email, subject, and body are required" },
      { status: 400 },
    );
  }

  const user = db
    .select()
    .from(users)
    .where(eq(users.email, body.recipient_email))
    .get();
  if (!user) {
    return NextResponse.json({ error: "user not found" }, { status: 404 });
  }

  const msg = db
    .insert(messages)
    .values({
      recipientId: user.id,
      senderType: body.sender_type ?? "ai_agent",
      category: body.category ?? null,
      subject: body.subject,
      body: body.body,
      ruleId: body.rule_id ?? null,
    })
    .returning()
    .get();

  // Forward to email.
  const sent = await forwardToEmail(user.email, body.subject, body.body);
  if (sent) {
    db.update(messages)
      .set({ emailSent: 1 })
      .where(eq(messages.id, msg.id))
      .run();
  }

  return NextResponse.json({ id: msg.id, email_sent: sent }, { status: 201 });
}
