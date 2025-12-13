// frontend/src/components/ChatRoom.tsx
"use client";

import { useEffect, useState } from "react";
import { useWebSocket } from "@/hooks/useWebSocket";
import { MessageList } from "./MessageList";
import { MessageInput } from "./MessageInput";
import type { AuthResponse } from "@/types";
import Link from "next/link";

interface ChatRoomProps {
  initialRoomId: string;
}

export function ChatRoom({ initialRoomId }: ChatRoomProps) {
  const [roomId] = useState(initialRoomId || "general");
  const [username, setUsername] = useState<string | null>(null);
  const [token, setToken] = useState<string | null>(null);

  useEffect(() => {
    if (typeof window === "undefined") return;
    const stored = localStorage.getItem("auth");
    if (!stored) return;
    try {
      const parsed = JSON.parse(stored) as AuthResponse;
      setUsername(parsed.user.username);
      setToken(parsed.token);
    } catch (err) {
      console.error("Invalid auth in localStorage", err);
    }
  }, []);

  const { messages, status, sendMessage } = useWebSocket({
    username,
    roomId,
    token,
  });

  if (!username || !token) {
    return (
      <div
        style={{
          maxWidth: 1800,
          margin: "40px auto",
          fontFamily: "system-ui, sans-serif",
        }}
      >
        <h1>Chat room: {roomId}</h1>
        <p>You are not logged in.</p>
        <p>
          <Link href="/login">Go to login</Link>
        </p>
      </div>
    );
  }

  return (
    <div
      style={{
        maxWidth: 600,
        margin: "40px auto",
        fontFamily: "system-ui, sans-serif",
      }}
    >
      <h1>Chat room: {roomId}</h1>
      <div style={{ marginBottom: 4, fontSize: 12, opacity: 0.7 }}>
        Logged in as: <strong>{username}</strong>
      </div>
      <div style={{ marginBottom: 4, fontSize: 12, opacity: 0.7 }}>
        Status: {status}
      </div>

      <MessageList messages={messages} />
      <MessageInput onSend={sendMessage} />
    </div>
  );
}
