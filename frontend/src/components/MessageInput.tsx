// frontend/src/components/MessageInput.tsx
"use client";

import { useRef, useState } from "react";
import { MAX_MESSAGE_LENGTH } from "@/lib/config";

interface MessageInputProps {
  // Return false to keep the text in the box (e.g. while disconnected)
  onSend: (text: string) => boolean | void;
  disabled?: boolean;
  disabledPlaceholder?: string;
  onTyping?: () => void;
}

export function MessageInput({
  onSend,
  disabled = false,
  disabledPlaceholder = "You must log in to send messages.",
  onTyping,
}: MessageInputProps) {
  const [value, setValue] = useState("");
  const lastTypingTimeRef = useRef<number>(0);

  const handleSend = () => {
    if (disabled) return;

    const text = value.trim();
    if (!text) return;
    if (onSend(text) === false) return;
    setValue("");
  };

  const handleKeyDown: React.KeyboardEventHandler<HTMLInputElement> = (e) => {
    if (disabled) {
      e.preventDefault();
      return;
    }
    if (e.key === "Enter") {
      e.preventDefault();
      handleSend();
    }
  };

  const handleChange: React.ChangeEventHandler<HTMLInputElement> = (e) => {
    const next = e.target.value;
    setValue(next);

    if (disabled || !onTyping) return;

    const now = Date.now();
    // send typing at most once per 800ms
    if (now - lastTypingTimeRef.current > 800) {
      lastTypingTimeRef.current = now;
      onTyping();
    }
  };

  const placeholder = disabled ? disabledPlaceholder : "Type a message";
  // Only show the counter once the user is getting close to the limit
  const remaining = MAX_MESSAGE_LENGTH - value.length;
  const showCounter = !disabled && remaining <= 100;

  return (
    <div>
      <div style={{ display: "flex", gap: 8 }}>
        <input
          type="text"
          value={value}
          maxLength={MAX_MESSAGE_LENGTH}
          onChange={handleChange}
          onKeyDown={handleKeyDown}
          placeholder={placeholder}
          disabled={disabled}
          style={{
            flex: 1,
            padding: 8,
            borderRadius: 4,
            border: "1px solid #444",
            backgroundColor: disabled ? "#111827" : "#020617",
            color: "#e5e7eb",
            opacity: disabled ? 0.6 : 1,
          }}
        />
        <button
          onClick={handleSend}
          disabled={disabled}
          style={{
            padding: "8px 16px",
            borderRadius: 4,
            border: "none",
            background: disabled ? "#4b5563" : "#2563eb",
            color: "#fff",
            cursor: disabled ? "not-allowed" : "pointer",
          }}
        >
          Send
        </button>
      </div>
      {showCounter && (
        <div
          style={{
            fontSize: 12,
            opacity: 0.7,
            marginTop: 4,
            textAlign: "right",
          }}
        >
          {remaining} characters left
        </div>
      )}
    </div>
  );
}