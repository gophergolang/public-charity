"use client";

import { useState, useRef, useCallback } from "react";
import { updateBasics, updateLocation } from "@/lib/actions/profile";
import type { User } from "@/lib/db/ensure-user";

export default function ProfileBasics({ initial }: { initial: User }) {
  const [name, setName] = useState(initial.displayName ?? "");
  const [bio, setBio] = useState(initial.bio ?? "");
  const [accountType, setAccountType] = useState(initial.accountType);
  const [locStatus, setLocStatus] = useState<"idle" | "locating" | "done">(
    initial.latitude ? "done" : "idle",
  );
  const [saving, setSaving] = useState(false);
  const timer = useRef<ReturnType<typeof setTimeout>>(undefined);

  const saveBasics = useCallback(
    (n: string, b: string, t: string) => {
      clearTimeout(timer.current);
      timer.current = setTimeout(async () => {
        setSaving(true);
        await updateBasics({ displayName: n, bio: b, accountType: t });
        setSaving(false);
      }, 800);
    },
    [],
  );

  function locate() {
    if (!navigator.geolocation) return;
    setLocStatus("locating");
    navigator.geolocation.getCurrentPosition(
      async (pos) => {
        await updateLocation({
          latitude: pos.coords.latitude,
          longitude: pos.coords.longitude,
        });
        setLocStatus("done");
      },
      () => setLocStatus("idle"),
      { enableHighAccuracy: false, timeout: 10000 },
    );
  }

  return (
    <div className="profile-basics">
      <div className="profile-field">
        <label htmlFor="name">Your name</label>
        <input
          id="name"
          type="text"
          placeholder="How should we refer to you?"
          value={name}
          onChange={(e) => {
            setName(e.target.value);
            saveBasics(e.target.value, bio, accountType);
          }}
        />
      </div>

      <div className="profile-field">
        <label htmlFor="bio">A bit about you</label>
        <textarea
          id="bio"
          rows={3}
          maxLength={280}
          placeholder="e.g. father of 2, 38 years old, software engineer. Tell us whatever feels right."
          value={bio}
          onChange={(e) => {
            setBio(e.target.value);
            saveBasics(name, e.target.value, accountType);
          }}
        />
      </div>

      <div className="profile-meta">
        <div className="segmented-control">
          <button
            type="button"
            className={accountType === "individual" ? "seg-active" : ""}
            onClick={() => { setAccountType("individual"); saveBasics(name, bio, "individual"); }}
          >Individual</button>
          <button
            type="button"
            className={accountType === "organisation" ? "seg-active" : ""}
            onClick={() => { setAccountType("organisation"); saveBasics(name, bio, "organisation"); }}
          >Organisation</button>
        </div>

        <button
          type="button"
          className="location-btn"
          onClick={locate}
          disabled={locStatus === "locating"}
        >
          {locStatus === "done" ? "📍 Location set" : locStatus === "locating" ? "📍 Finding..." : "📍 Set your location"}
        </button>

        {saving && <span className="auto-saved">Saved</span>}
      </div>
    </div>
  );
}
