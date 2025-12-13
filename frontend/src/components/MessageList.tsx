// frontend/src/components/MessageList.tsx
"use client";

import type { ChatMessage } from "@/types";

interface MessageListProps {
  messages: ChatMessage[];
}

export function MessageList({ messages }: MessageListProps) {
  return (
    <div
      style={{
        border: "1px solid #ccc",
        borderRadius: 4,
        padding: 8,
        height: 300,
        overflowY: "auto",
        marginBottom: 8,
      }}
    >
      {messages.length === 0 && (
        <div style={{ opacity: 0.6 }}>No messages yet. Say hi!</div>
      )}
      {messages.map((m, idx) => (
        <div key={idx} style={{ marginBottom: 4, fontSize: 14 }}>
          <strong>{m.username}</strong>{" "}
          <span style={{ opacity: 0.7, fontSize: 12 }}>
            [{m.roomId}] {new Date(m.timestamp).toLocaleTimeString()}
          </span>
          <div>{m.text}</div>
        </div>
      ))}
    </div>
  );
}