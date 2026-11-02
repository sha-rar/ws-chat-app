// frontend/src/components/ChatRoom.tsx
"use client";

import { useEffect, useState } from "react";
import { useWebSocket } from "@/hooks/useWebSocket";
import { MessageList } from "./MessageList";
import { MessageInput } from "./MessageInput";
import { TypingIndicator } from "./TypingIndicator";
import Link from "next/link";
import { getStoredAuth } from "@/lib/auth";
import { LogoutButton } from "./LogoutButton";

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

    const auth = getStoredAuth();
    if (!auth) return;
    setUsername(auth.user.username);
    setToken(auth.token);
  }, [forceGuest]);

  const isAuthed = !forceGuest && !!username && !!token;
  const displayName = isAuthed ? username : "guest";

  const { messages, status, typingUsers, serverError, sendMessage, sendTyping } =
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
            Logged in as: <strong>{displayName}</strong> ·{" "}
            <Link href="/chat">Switch room</Link> · <LogoutButton />
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

      {serverError && (
        <div style={{ color: "#f87171", fontSize: 13, marginBottom: 4 }}>
          {serverError}
        </div>
      )}

      <MessageInput
        onSend={sendMessage}
        onTyping={sendTyping}
        disabled={!isAuthed}
        disabledPlaceholder="You must log in to send messages."
      />
    </div>
  );
}