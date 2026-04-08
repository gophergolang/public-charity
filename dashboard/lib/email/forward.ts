const RESEND_ENDPOINT = "https://api.resend.com/emails";

export async function forwardToEmail(
  to: string,
  subject: string,
  body: string,
): Promise<boolean> {
  const apiKey = process.env.RESEND_API_KEY;
  const from = process.env.EMAIL_FROM;
  if (!apiKey || !from) return false;

  try {
    const res = await fetch(RESEND_ENDPOINT, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${apiKey}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        from,
        to: [to],
        subject,
        text: body,
        html: `<p>${body.replace(/\n/g, "<br>")}</p>`,
      }),
    });
    return res.ok;
  } catch {
    return false;
  }
}
