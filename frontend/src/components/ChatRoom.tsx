// frontend/src/components/ChatRoom.tsx
"use client";

import { useState } from "react";
import { useWebSocket } from "@/hooks/useWebSocket";
import { MessageList } from "./MessageList";
import { MessageInput } from "./MessageInput";

interface ChatRoomProps {
  initialRoomId: string;
  initialUsername: string;
}

export function ChatRoom({ initialRoomId, initialUsername }: ChatRoomProps) {
  const [username] = useState(initialUsername || "guest");
  const [roomId] = useState(initialRoomId || "general");

  const { messages, status, sendMessage } = useWebSocket({
    username,
    roomId,
  });

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
