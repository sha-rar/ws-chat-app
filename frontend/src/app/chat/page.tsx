// frontend/src/app/chat/page.tsx
"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { useState } from "react";

export default function ChatLobbyPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const forceGuest = searchParams.get("guest") === "1";

  const [roomId, setRoomId] = useState("general");

  const handleJoin = (e: React.FormEvent) => {
    e.preventDefault();
    const trimmedRoomId = roomId.trim() || "general";

    const guestSuffix = forceGuest ? "?guest=1" : "";

    router.push(
      `/chat/${encodeURIComponent(trimmedRoomId)}${guestSuffix}`
    );
  };

  return (
    <div
      style={{
        maxWidth: 400,
        margin: "60px auto",
        fontFamily: "system-ui, sans-serif",
      }}
    >
      <h1>Chat lobby</h1>
      <p style={{ opacity: 0.7, fontSize: 14, marginTop: 8 }}>
        {forceGuest
          ? "You are in guest mode. You must log in to send messages."
          : "Choose a room to join."}
      </p>

      <form onSubmit={handleJoin} style={{ marginTop: 16 }}>
        <div style={{ marginBottom: 16 }}>
          <label style={{ display: "block", marginBottom: 4 }}>
            Room name
          </label>
          <input
            type="text"
            value={roomId}
            onChange={(e) => setRoomId(e.target.value)}
            placeholder="e.g. general, memes, cs231"
            style={{
              width: "100%",
              padding: 8,
              borderRadius: 4,
              border: "1px solid #ccc",
            }}
          />
        </div>

        <button
          type="submit"
          style={{
            padding: "8px 16px",
            borderRadius: 4,
            border: "none",
            background: "#2563eb",
            color: "#fff",
            cursor: "pointer",
          }}
        >
          Join room
        </button>
      </form>
    </div>
  );
}
