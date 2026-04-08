"use client";

import { useState, FormEvent } from "react";

export default function LoginPage() {
  const [email, setEmail] = useState("");
  const [status, setStatus] = useState<"idle" | "sending" | "sent" | "error">("idle");
  const [message, setMessage] = useState("");

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setStatus("sending");
    setMessage("");
    try {
      const res = await fetch("/api/auth/request", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email }),
      });
      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        setStatus("error");
        setMessage(body.error ?? `request failed (${res.status})`);
        return;
      }
      setStatus("sent");
    } catch (err) {
      setStatus("error");
      setMessage((err as Error).message);
    }
  }

  return (
    <main>
      <h1>Sign in</h1>
      {status === "sent" ? (
        <p className="ok">
          Check your email for a sign-in link. You can close this tab.
        </p>
      ) : (
        <form onSubmit={onSubmit}>
          <label htmlFor="email">Email address</label>
          <input
            id="email"
            type="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="you@example.com"
            disabled={status === "sending"}
          />
          <button type="submit" disabled={status === "sending" || !email}>
            {status === "sending" ? "Sending..." : "Send magic link"}
          </button>
          {status === "error" && <p className="err">{message}</p>}
        </form>
      )}
    </main>
  );
}
