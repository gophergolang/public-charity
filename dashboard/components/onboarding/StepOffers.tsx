"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { setOffers } from "@/lib/actions/offers";
import { advanceStep } from "@/lib/actions/onboarding";
import { CATEGORIES } from "@/lib/db/schema";

type OfferItem = { category: string; description: string };

export default function StepOffers({
  initial,
}: {
  initial: { category: string; description: string }[];
}) {
  const router = useRouter();
  const [items, setItems] = useState<OfferItem[]>(
    initial.length > 0 ? initial : [{ category: CATEGORIES[0], description: "" }],
  );
  const [saving, setSaving] = useState(false);

  function update(i: number, field: keyof OfferItem, value: string) {
    setItems((prev) => prev.map((it, idx) => (idx === i ? { ...it, [field]: value } : it)));
  }

  function addRow() {
    setItems([...items, { category: CATEGORIES[0], description: "" }]);
  }

  function removeRow(i: number) {
    setItems(items.filter((_, idx) => idx !== i));
  }

  async function handleNext() {
    setSaving(true);
    await setOffers(items.filter((it) => it.description.trim()));
    await advanceStep(5);
    router.push("/onboarding/step/6");
  }

  async function handleSkip() {
    setSaving(true);
    await advanceStep(5);
    router.push("/onboarding/step/6");
  }

  return (
    <div className="step-form">
      <h1>What can you offer?</h1>
      <p className="muted">
        It could be anything — a skill, some time, a spare tool. Every little
        bit helps.
      </p>

      {items.map((item, i) => (
        <div key={i} className="offer-row">
          <select value={item.category} onChange={(e) => update(i, "category", e.target.value)}>
            {CATEGORIES.map((c) => (
              <option key={c} value={c}>{c}</option>
            ))}
          </select>
          <input
            type="text"
            placeholder="What can you offer?"
            value={item.description}
            onChange={(e) => update(i, "description", e.target.value)}
          />
          {items.length > 1 && (
            <button type="button" onClick={() => removeRow(i)} className="remove">
              &times;
            </button>
          )}
        </div>
      ))}

      <button type="button" className="link" onClick={addRow}>
        + Add another
      </button>

      <div className="step-nav">
        <button onClick={() => router.push("/onboarding/step/4")}>Back</button>
        <button type="button" className="link" onClick={handleSkip}>
          Skip for now
        </button>
        <button onClick={handleNext} disabled={saving} className="primary">
          {saving ? "Saving..." : "Next"}
        </button>
      </div>
    </div>
  );
}
