"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { updateBasics } from "@/lib/actions/profile";
import { advanceStep } from "@/lib/actions/onboarding";
import type { User } from "@/lib/db/ensure-user";

export default function StepBasics({ initial }: { initial: User }) {
  const router = useRouter();
  const [displayName, setDisplayName] = useState(initial.displayName ?? "");
  const [bio, setBio] = useState(initial.bio ?? "");
  const [accountType, setAccountType] = useState(initial.accountType);
  const [saving, setSaving] = useState(false);

  async function handleNext() {
    setSaving(true);
    await updateBasics({ displayName, bio, accountType });
    await advanceStep(1);
    router.push("/onboarding/step/2");
  }

  return (
    <div className="step-form">
      <h1>Let's get started</h1>
      <p className="muted">Tell us a bit about yourself. You can change this later.</p>

      <label htmlFor="name">Display name</label>
      <input
        id="name"
        type="text"
        placeholder="How should we refer to you?"
        value={displayName}
        onChange={(e) => setDisplayName(e.target.value)}
      />

      <label htmlFor="bio">Short bio</label>
      <textarea
        id="bio"
        placeholder="A sentence or two about yourself..."
        maxLength={280}
        rows={3}
        value={bio}
        onChange={(e) => setBio(e.target.value)}
      />
      <p className="muted">{280 - bio.length} characters left</p>

      <label>Account type</label>
      <div className="toggle-group">
        <button
          type="button"
          className={accountType === "individual" ? "active" : ""}
          onClick={() => setAccountType("individual")}
        >
          Individual
        </button>
        <button
          type="button"
          className={accountType === "organisation" ? "active" : ""}
          onClick={() => setAccountType("organisation")}
        >
          Organisation
        </button>
      </div>

      <button onClick={handleNext} disabled={saving} className="primary">
        {saving ? "Saving..." : "Next"}
      </button>
    </div>
  );
}
