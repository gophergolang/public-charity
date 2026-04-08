"use client";

import { useState, KeyboardEvent } from "react";
import { useRouter } from "next/navigation";
import { setInterests } from "@/lib/actions/interests";
import { advanceStep } from "@/lib/actions/onboarding";

const SUGGESTIONS = [
  "gardening", "cooking", "walking", "reading", "music", "art",
  "sports", "volunteering", "tech", "animals", "crafts", "photography",
  "games", "film", "languages", "yoga", "cycling", "sewing",
];

export default function StepInterests({ initial }: { initial: string[] }) {
  const router = useRouter();
  const [tags, setTags] = useState<string[]>(initial);
  const [input, setInput] = useState("");
  const [saving, setSaving] = useState(false);

  function addTag(tag: string) {
    const t = tag.trim().toLowerCase();
    if (t && !tags.includes(t)) setTags([...tags, t]);
    setInput("");
  }

  function removeTag(tag: string) {
    setTags(tags.filter((t) => t !== tag));
  }

  function onKeyDown(e: KeyboardEvent) {
    if (e.key === "Enter" || e.key === ",") {
      e.preventDefault();
      addTag(input);
    }
  }

  async function handleNext() {
    setSaving(true);
    await setInterests(tags);
    await advanceStep(4);
    router.push("/onboarding/step/5");
  }

  async function handleSkip() {
    setSaving(true);
    await advanceStep(4);
    router.push("/onboarding/step/5");
  }

  return (
    <div className="step-form">
      <h1>What are you interested in?</h1>
      <p className="muted">
        Add your interests so we can find people with things in common.
      </p>

      <div className="tag-input-wrap">
        <div className="tags">
          {tags.map((t) => (
            <span key={t} className="tag">
              {t}
              <button type="button" onClick={() => removeTag(t)}>
                &times;
              </button>
            </span>
          ))}
        </div>
        <input
          type="text"
          placeholder="Type an interest and press Enter"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={onKeyDown}
        />
      </div>

      <div className="suggestions">
        {SUGGESTIONS.filter((s) => !tags.includes(s)).map((s) => (
          <button key={s} type="button" className="chip" onClick={() => addTag(s)}>
            + {s}
          </button>
        ))}
      </div>

      <div className="step-nav">
        <button onClick={() => router.push("/onboarding/step/3")}>Back</button>
        <button type="button" className="link" onClick={handleSkip}>
          Skip for now
        </button>
        <button
          onClick={handleNext}
          disabled={saving || tags.length === 0}
          className="primary"
        >
          {saving ? "Saving..." : "Next"}
        </button>
      </div>
    </div>
  );
}
