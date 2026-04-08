import { notFound } from "next/navigation";
import Link from "next/link";
import { getMessage, markRead } from "@/lib/actions/inbox";

export default async function MessagePage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const msg = await getMessage(id);
  if (!msg) notFound();

  if (!msg.read) {
    await markRead(id);
  }

  const sender =
    msg.senderType === "system" || msg.senderType === "ai_agent"
      ? "Public Charity"
      : "User";

  return (
    <main>
      <Link href="/dashboard/inbox" className="muted">&larr; Back to inbox</Link>
      <h1>{msg.subject}</h1>
      <p className="muted">
        From: {sender} &middot;{" "}
        {new Date(msg.createdAt!).toLocaleString()}
      </p>
      <div className="message-body">
        {msg.body.split("\n").map((line, i) => (
          <p key={i}>{line}</p>
        ))}
      </div>
    </main>
  );
}
