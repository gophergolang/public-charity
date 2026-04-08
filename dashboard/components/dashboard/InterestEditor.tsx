"use client";

import { useState, KeyboardEvent, useRef } from "react";
import { setInterests } from "@/lib/actions/interests";

const SUGGESTIONS = [
  "gardening", "cooking", "walking", "reading", "music", "art",
  "sports", "volunteering", "tech", "animals", "crafts", "photography",
  "games", "film", "languages", "yoga", "cycling", "sewing",
];

export default function InterestEditor({ initial }: { initial: string[] }) {
  const [tags, setTags] = useState<string[]>(initial);
  const [input, setInput] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);

  async function sync(next: string[]) {
    setTags(next);
    await setInterests(next);
  }

  function add(tag: string) {
    const t = tag.trim().toLowerCase();
    if (t && !tags.includes(t)) sync([...tags, t]);
    setInput("");
  }

  function remove(tag: string) {
    sync(tags.filter((t) => t !== tag));
  }

  function onKeyDown(e: KeyboardEvent) {
    if (e.key === "Enter" || e.key === ",") {
      e.preventDefault();
      add(input);
    }
    if (e.key === "Backspace" && input === "" && tags.length > 0) {
      remove(tags[tags.length - 1]);
    }
  }

  return (
    <div className="interest-section">
      <label>Things you enjoy</label>
      <div
        className="token-input"
        onClick={() => inputRef.current?.focus()}
      >
        {tags.map((t) => (
          <span key={t} className="token">
            {t}
            <button type="button" onClick={() => remove(t)}>&times;</button>
          </span>
        ))}
        <input
          ref={inputRef}
          type="text"
          className="token-text"
          placeholder={tags.length === 0 ? "Type an interest and press Enter" : "Add more..."}
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={onKeyDown}
        />
      </div>
      <div className="interest-suggestions">
        {SUGGESTIONS.filter((s) => !tags.includes(s))
          .slice(0, 8)
          .map((s) => (
            <button key={s} type="button" className="suggestion-chip" onClick={() => add(s)}>
              {s}
            </button>
          ))}
      </div>
    </div>
  );
}
