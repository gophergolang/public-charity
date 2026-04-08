"use client";

import { useState, useRef, useCallback } from "react";
import { updateNeedScores } from "@/lib/actions/needs";
import { SECTIONS, type Category } from "@/lib/db/schema";

const LABELS: Record<string, { name: string; desc: string; emoji: string }> = {
  // Social
  companionship:      { name: "Companionship",      desc: "Someone to spend time with",                    emoji: "🤝" },
  community:          { name: "Community",           desc: "Feeling part of a group or neighbourhood",      emoji: "🏘️" },
  family:             { name: "Family",              desc: "Support with family or caring responsibilities", emoji: "👨‍👩‍👧" },
  language:           { name: "Language",             desc: "Help communicating or learning a language",     emoji: "💬" },
  // Health
  mental_health:      { name: "Mental Health",       desc: "Emotional or mental wellbeing support",          emoji: "🧠" },
  physical_health:    { name: "Physical Health",     desc: "Staying active or managing a condition",         emoji: "💪" },
  nutrition:          { name: "Nutrition",            desc: "Access to healthy food or meals",                emoji: "🥗" },
  substance_recovery: { name: "Recovery",            desc: "Support with recovery or substance issues",      emoji: "🌱" },
  // Daily Life
  housing:            { name: "Housing",             desc: "Help with housing or your living situation",     emoji: "🏠" },
  transport:          { name: "Transport",           desc: "Help getting around",                            emoji: "🚌" },
  errands:            { name: "Errands",             desc: "Help with day-to-day tasks",                     emoji: "📋" },
  personal_care:      { name: "Personal Care",       desc: "Support with personal care or daily routine",    emoji: "🪥" },
  // Economic
  employment:         { name: "Employment",          desc: "Finding or keeping work",                        emoji: "💼" },
  training:           { name: "Training",            desc: "Learning new skills for work or life",           emoji: "📚" },
  benefits:           { name: "Benefits",            desc: "Understanding or claiming what you're owed",     emoji: "📝" },
  budgeting:          { name: "Budgeting",           desc: "Help managing money or debt",                    emoji: "💰" },
  // Practical
  digital:            { name: "Digital",             desc: "Help with technology or the internet",           emoji: "💻" },
  admin:              { name: "Admin",               desc: "Paperwork, forms, or phone calls",               emoji: "📄" },
  legal:              { name: "Legal",               desc: "Support with legal issues or rights",            emoji: "⚖️" },
  childcare:          { name: "Childcare",           desc: "Help with childcare or parenting",               emoji: "👶" },
};

const DOTS = [0, 0.25, 0.5, 0.75, 1.0];

export default function NeedSliders({
  initial,
}: {
  initial: Record<string, number>;
}) {
  const [scores, setScores] = useState(initial);
  const [saving, setSaving] = useState(false);
  const timer = useRef<ReturnType<typeof setTimeout>>(undefined);

  const save = useCallback((next: Record<string, number>) => {
    clearTimeout(timer.current);
    timer.current = setTimeout(async () => {
      setSaving(true);
      await updateNeedScores(next);
      setSaving(false);
    }, 600);
  }, []);

  function set(cat: string, val: number) {
    const next = { ...scores, [cat]: val };
    setScores(next);
    save(next);
  }

  return (
    <div className="need-sections">
      {SECTIONS.map((section) => (
        <div
          key={section.id}
          className="need-section-card"
          style={{ "--section-color": section.color } as React.CSSProperties}
        >
          <div className="section-header">
            <h2>{section.label}</h2>
            <p>{section.description}</p>
          </div>
          <div className="section-sliders">
            {section.categories.map((cat: Category) => {
              const info = LABELS[cat];
              const val = scores[cat] ?? 0;
              return (
                <div key={cat} className="need-slider">
                  <div className="need-info">
                    <span className="need-emoji">{info?.emoji}</span>
                    <div>
                      <span className="need-name">{info?.name}</span>
                      <span className="need-desc">{info?.desc}</span>
                    </div>
                  </div>
                  <div className="need-dots">
                    {DOTS.map((d) => (
                      <button
                        key={d}
                        type="button"
                        className={`dot ${val >= d ? "active" : ""} ${val === d ? "selected" : ""}`}
                        onClick={() => set(cat, d)}
                        aria-label={`${d * 100}%`}
                      />
                    ))}
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      ))}
      {saving && <p className="saving-float">Saving your changes...</p>}
    </div>
  );
}
