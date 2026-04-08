"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { updateLocation } from "@/lib/actions/profile";
import { advanceStep } from "@/lib/actions/onboarding";
import type { User } from "@/lib/db/ensure-user";

export default function StepLocation({ initial }: { initial: User }) {
  const router = useRouter();
  const [lat, setLat] = useState(initial.latitude ?? 0);
  const [lng, setLng] = useState(initial.longitude ?? 0);
  const [status, setStatus] = useState<"idle" | "locating" | "done" | "error">(
    initial.latitude ? "done" : "idle",
  );
  const [saving, setSaving] = useState(false);

  function locate() {
    if (!navigator.geolocation) {
      setStatus("error");
      return;
    }
    setStatus("locating");
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        setLat(pos.coords.latitude);
        setLng(pos.coords.longitude);
        setStatus("done");
      },
      () => setStatus("error"),
      { enableHighAccuracy: false, timeout: 10000 },
    );
  }

  async function handleNext() {
    setSaving(true);
    if (lat && lng) await updateLocation({ latitude: lat, longitude: lng });
    await advanceStep(2);
    router.push("/onboarding/step/3");
  }

  return (
    <div className="step-form">
      <h1>Where are you based?</h1>
      <p className="muted">
        This helps us connect you with people nearby. Your exact location is
        never shared — only your general area.
      </p>

      <button
        type="button"
        onClick={locate}
        disabled={status === "locating"}
        className="primary"
      >
        {status === "locating"
          ? "Finding your location..."
          : status === "done"
            ? "Location set"
            : "Use my location"}
      </button>

      {status === "done" && (
        <p className="ok">
          Got it — approximate area: {lat.toFixed(2)}, {lng.toFixed(2)}
        </p>
      )}
      {status === "error" && (
        <p className="err">
          Could not get your location. You can skip this for now.
        </p>
      )}

      <div className="step-nav">
        <button onClick={() => router.push("/onboarding/step/1")}>Back</button>
        <button onClick={handleNext} disabled={saving} className="primary">
          {saving ? "Saving..." : "Next"}
        </button>
      </div>
    </div>
  );
}
