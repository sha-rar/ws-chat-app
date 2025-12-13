// frontend/src/components/MessageList.tsx
"use client";

import { useEffect, useRef, useState } from "react";
import type { ChatMessage } from "@/types";

interface MessageListProps {
  messages: ChatMessage[] | null | undefined;
}

export function MessageList({ messages }: MessageListProps) {
  const safeMessages = Array.isArray(messages) ? messages : [];
  const containerRef = useRef<HTMLDivElement | null>(null);

  // Is the user currently at/near the bottom?
  const isAtBottomRef = useRef(true);
  // How many messages had been seen when user was last at bottom?
  const lastSeenCountRef = useRef(0);

  const [showJumpToLatest, setShowJumpToLatest] = useState(false);

  const BOTTOM_THRESHOLD = 40; // px

  const scrollToBottom = () => {
    const el = containerRef.current;
    if (!el) return;
    el.scrollTop = el.scrollHeight;
  };

  const handleScroll = () => {
    const el = containerRef.current;
    if (!el) return;

    const distanceFromBottom = el.scrollHeight - el.scrollTop - el.clientHeight;
    const atBottom = distanceFromBottom <= BOTTOM_THRESHOLD;

    isAtBottomRef.current = atBottom;

    if (atBottom) {
      // User manually scrolled to bottom; they've now "seen" all messages
      lastSeenCountRef.current = safeMessages.length;
      setShowJumpToLatest(false);
    }
    // If not at bottom, we do *not* show the button yet.
    // The effect below will decide that only when new messages arrive.
  };

  // On mount: start at bottom with whatever messages exist
  useEffect(() => {
    scrollToBottom();
    isAtBottomRef.current = true;
    lastSeenCountRef.current = safeMessages.length;
    setShowJumpToLatest(false);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // run once

  // When messages change:
  // - If at bottom: auto-scroll and mark all as seen
  // - If NOT at bottom: show button only if there are new messages beyond lastSeenCountRef
  useEffect(() => {
    if (isAtBottomRef.current) {
      scrollToBottom();
      lastSeenCountRef.current = safeMessages.length;
      setShowJumpToLatest(false);
    } else {
      if (safeMessages.length > lastSeenCountRef.current) {
        setShowJumpToLatest(true);
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [safeMessages.length]);

  return (
    <div style={{ position: "relative" }}>
      <div
        ref={containerRef}
        onScroll={handleScroll}
        style={{
          border: "1px solid #ccc",
          borderRadius: 4,
          padding: 8,
          height: "65vh",
          overflowY: "auto",
          marginBottom: 8,
        }}
      >
        {safeMessages.length === 0 && (
          <div style={{ opacity: 0.6 }}>No messages yet. Say hi!</div>
        )}
        {safeMessages.map((m, idx) => (
          <div key={idx} style={{ marginBottom: 4, fontSize: 14 }}>
            <strong>{m.username}</strong>{" "}
            <span style={{ opacity: 0.7, fontSize: 12 }}>
              [{m.roomId}] {new Date(m.timestamp).toLocaleTimeString()}
            </span>
            <div>{m.text}</div>
          </div>
        ))}
      </div>

      {showJumpToLatest && safeMessages.length > 0 && (
        <button
          type="button"
          onClick={() => {
            scrollToBottom();
            isAtBottomRef.current = true;
            lastSeenCountRef.current = safeMessages.length;
            setShowJumpToLatest(false);
          }}
          style={{
            position: "absolute",
            right: 12,
            bottom: 12 + 8,
            padding: "4px 8px",
            borderRadius: 999,
            border: "none",
            background: "#2563eb",
            color: "#fff",
            fontSize: 12,
            cursor: "pointer",
            boxShadow: "0 2px 4px rgba(0,0,0,0.2)",
          }}
        >
          Jump to latest
        </button>
      )}
    </div>
  );
}