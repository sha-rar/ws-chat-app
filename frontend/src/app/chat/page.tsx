// frontend/src/app/chat/page.tsx
"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useState, useSyncExternalStore } from "react";
import { getStoredAuth } from "@/lib/auth";
import { LogoutButton } from "@/components/LogoutButton";

function subscribeToStorage(onChange: () => void) {
  window.addEventListener("storage", onChange);
  return () => window.removeEventListener("storage", onChange);
}

function ChatLobby() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const forceGuest = searchParams.get("guest") === "1";

  const [roomId, setRoomId] = useState("general");
  // Read from localStorage on the client only; the server render has no user
  const username = useSyncExternalStore(
    subscribeToStorage,
    () => (forceGuest ? null : getStoredAuth()?.user.username ?? null),
    () => null
  );

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
      {username && (
        <p style={{ fontSize: 12, opacity: 0.7, marginTop: 4 }}>
          Logged in as <strong>{username}</strong> · <LogoutButton />
        </p>
      )}

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

// useSearchParams() needs a Suspense boundary so the page can be prerendered
export default function ChatLobbyPage() {
  return (
    <Suspense>
      <ChatLobby />
    </Suspense>
  );
}
