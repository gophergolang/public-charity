import Link from "next/link";
import { getMessages } from "@/lib/actions/inbox";

export default async function InboxPage() {
  const msgs = await getMessages();

  return (
    <main>
      <h1>Inbox</h1>
      {msgs.length === 0 ? (
        <p className="muted">No messages yet.</p>
      ) : (
        <div className="message-list">
          {msgs.map((m) => (
            <Link
              key={m.id}
              href={`/dashboard/inbox/${m.id}`}
              className={`message-row ${m.read ? "" : "unread"}`}
            >
              <span className="message-sender">
                {m.senderType === "system"
                  ? "Public Charity"
                  : m.senderType === "ai_agent"
                    ? "Public Charity"
                    : "User"}
              </span>
              <span className="message-subject">{m.subject}</span>
              <span className="message-date muted">
                {new Date(m.createdAt!).toLocaleDateString()}
              </span>
            </Link>
          ))}
        </div>
      )}
    </main>
  );
}
