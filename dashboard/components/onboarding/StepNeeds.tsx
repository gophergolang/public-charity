"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { updateNeedScores } from "@/lib/actions/needs";
import { advanceStep } from "@/lib/actions/onboarding";
import { CATEGORIES } from "@/lib/db/schema";

const LABELS: Record<string, string> = {
  company: "Someone to spend time with",
  tools: "Access to tools or equipment",
  transport: "Help getting around",
  grocery: "Help with food shopping",
  errands: "Help with day-to-day tasks",
  admin: "Help with paperwork, forms, or phone calls",
  skills: "Learning something new",
  wellness: "Support for physical or mental health",
  digital: "Help with technology",
  community: "Feeling part of something",
};

const STEPS = [0, 0.25, 0.5, 0.75, 1.0];
const STEP_LABELS = ["I'm fine", "", "Sometimes", "", "I really need this"];

export default function StepNeeds({
  initial,
}: {
  initial: Record<string, number>;
}) {
  const router = useRouter();
  const [scores, setScores] = useState(initial);
  const [saving, setSaving] = useState(false);

  function setScore(cat: string, val: number) {
    setScores((s) => ({ ...s, [cat]: val }));
  }

  async function handleNext() {
    setSaving(true);
    await updateNeedScores(scores);
    await advanceStep(3);
    router.push("/onboarding/step/4");
  }

  return (
    <div className="step-form">
      <h1>What do you need?</h1>
      <p className="muted">
        Be honest — this helps us find the right matches for you.
      </p>

      <div className="needs-list">
        {CATEGORIES.map((cat) => (
          <div key={cat} className="need-item">
            <label>{LABELS[cat]}</label>
            <input
              type="range"
              min={0}
              max={1}
              step={0.25}
              value={scores[cat] ?? 0}
              onChange={(e) => setScore(cat, parseFloat(e.target.value))}
            />
            <div className="range-labels">
              {STEP_LABELS.map((l, i) => (
                <span key={i} className={scores[cat] === STEPS[i] ? "active" : ""}>
                  {l}
                </span>
              ))}
            </div>
          </div>
        ))}
      </div>

      <div className="step-nav">
        <button onClick={() => router.push("/onboarding/step/2")}>Back</button>
        <button onClick={handleNext} disabled={saving} className="primary">
          {saving ? "Saving..." : "Next"}
        </button>
      </div>
    </div>
  );
}
