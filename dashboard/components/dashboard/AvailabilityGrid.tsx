"use client";

import { useState } from "react";
import { toggleSlot } from "@/lib/actions/availability";
import { DAYS, SLOTS } from "@/lib/db/schema";

const DAY_LABELS: Record<string, string> = {
  mon: "Mon", tue: "Tue", wed: "Wed", thu: "Thu",
  fri: "Fri", sat: "Sat", sun: "Sun",
};

const SLOT_LABELS: Record<string, string> = {
  morning: "Morning",
  afternoon: "Afternoon",
  evening: "Evening",
};

function key(day: string, slot: string) {
  return `${day}-${slot}`;
}

export default function AvailabilityGrid({
  initial,
}: {
  initial: string[];
}) {
  const [active, setActive] = useState<Set<string>>(new Set(initial));

  async function toggle(day: string, slot: string) {
    const k = key(day, slot);
    // Optimistic update.
    setActive((prev) => {
      const next = new Set(prev);
      if (next.has(k)) next.delete(k);
      else next.add(k);
      return next;
    });
    await toggleSlot(day, slot);
  }

  return (
    <div className="avail-grid">
      {/* Corner */}
      <div className="avail-corner" />
      {/* Day headers */}
      {DAYS.map((d) => (
        <div key={d} className="avail-day-header">{DAY_LABELS[d]}</div>
      ))}
      {/* Rows */}
      {SLOTS.map((slot) => (
        <>
          <div key={`label-${slot}`} className="avail-slot-label">
            {SLOT_LABELS[slot]}
          </div>
          {DAYS.map((day) => {
            const k = key(day, slot);
            const on = active.has(k);
            return (
              <button
                key={k}
                type="button"
                className={`avail-cell ${on ? "avail-on" : "avail-off"}`}
                onClick={() => toggle(day, slot)}
                aria-label={`${DAY_LABELS[day]} ${SLOT_LABELS[slot]}: ${on ? "available" : "not set"}`}
              >
                {on ? "✓" : ""}
              </button>
            );
          })}
        </>
      ))}
    </div>
  );
}
