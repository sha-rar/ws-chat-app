// frontend/src/components/ChatRoom.tsx
"use client";

import { useEffect, useState } from "react";
import { useWebSocket } from "@/hooks/useWebSocket";
import { MessageList } from "./MessageList";
import { MessageInput } from "./MessageInput";
import { TypingIndicator } from "./TypingIndicator";
import type { AuthResponse } from "@/types";
import Link from "next/link";

interface ChatRoomProps {
  initialRoomId: string;
  forceGuest?: boolean;
}

export function ChatRoom({ initialRoomId, forceGuest = false }: ChatRoomProps) {
  const [roomId] = useState(initialRoomId || "general");
  const [username, setUsername] = useState<string | null>(null);
  const [token, setToken] = useState<string | null>(null);

  useEffect(() => {
    if (forceGuest) {
      setUsername(null);
      setToken(null);
      return;
    }

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
  }, [forceGuest]);

  const isAuthed = !forceGuest && !!username && !!token;
  const displayName = isAuthed ? username : "guest";

  const { messages, status, typingUsers, sendMessage, sendTyping } =
    useWebSocket({
      username: isAuthed ? username : null,
      roomId,
      token: isAuthed ? token : null,
    });

  return (
    <div
      style={{
        maxWidth: 800,
        margin: "40px auto",
        fontFamily: "system-ui, sans-serif",
      }}
    >
      <h1>Chat room: {roomId}</h1>

      <div style={{ marginBottom: 4, fontSize: 12, opacity: 0.7 }}>
        {isAuthed ? (
          <>
            Logged in as: <strong>{displayName}</strong>
          </>
        ) : (
          <>
            Viewing as <strong>guest</strong>.{" "}
            <Link href="/login">Login</Link> or{" "}
            <Link href="/register">register</Link> to send messages.
          </>
        )}
      </div>

      <div style={{ marginBottom: 4, fontSize: 12, opacity: 0.7 }}>
        Status: {status}
      </div>

      <MessageList messages={messages} />

      <TypingIndicator
        typingUsers={typingUsers}
        currentUsername={username}
      />

      <MessageInput
        onSend={sendMessage}
        onTyping={sendTyping}
        disabled={!isAuthed}
        disabledPlaceholder="You must log in to send messages."
      />
    </div>
  );
}