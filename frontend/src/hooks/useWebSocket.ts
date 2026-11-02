// frontend/src/hooks/useWebSocket.ts
"use client";

import { useEffect, useRef, useState } from "react";
import type { ChatMessage, WSEvent } from "@/types";
import { fetchRoomMessages } from "@/lib/api";
import { WS_BASE_URL } from "@/lib/config";

type ConnectionStatus = "connecting" | "open" | "closed" | "error";

interface UseWebSocketOptions {
  username: string | null;
  roomId: string;
  token: string | null;
}

export function useWebSocket({ username, roomId, token }: UseWebSocketOptions) {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [status, setStatus] = useState<ConnectionStatus>("connecting");
  const [typingUsers, setTypingUsers] = useState<string[]>([]);
  const wsRef = useRef<WebSocket | null>(null);

  const typingTimeoutsRef = useRef<Map<string, ReturnType<typeof setTimeout>>>(
    new Map()
  );

  useEffect(() => {
    let isCancelled = false;

    setMessages([]);
    setStatus("connecting");
    setTypingUsers([]);

    // 1) fetch history
    (async () => {
      try {
        const history = await fetchRoomMessages(roomId, 50);
        if (!isCancelled && Array.isArray(history)) {
          setMessages(history);
        }
      } catch (err) {
        console.error("Failed to fetch history", err);
      }
    })();

    // 2) open WebSocket; without a token the server treats us as a read-only guest
    const tokenParam = token ? `&token=${encodeURIComponent(token)}` : "";
    const wsUrl = `${WS_BASE_URL}/ws?roomId=${encodeURIComponent(
      roomId
    )}${tokenParam}`;
    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      if (!isCancelled) setStatus("open");
    };

    ws.onclose = () => {
      if (!isCancelled) setStatus("closed");
    };

    ws.onerror = () => {
      if (!isCancelled) setStatus("error");
    };

    ws.onmessage = (event: MessageEvent) => {
      try {
        const data = JSON.parse(event.data) as WSEvent;

        if (data.type === "chat") {
          const msg = data.message;
          if (!isCancelled) {
            setMessages((prev) => [...prev, msg]);
            // remove chatter from typing list
            setTypingUsers((prev) => prev.filter((u) => u !== msg.username));
          }
        } else if (data.type === "typing") {
          const otherUser = data.username;
          // ignore our own typing events
          if (otherUser === username) return;

          if (!isCancelled) {
            // add / keep in typing list
            setTypingUsers((prev) =>
              prev.includes(otherUser) ? prev : [...prev, otherUser]
            );

            // reset timeout for this user
            const timeouts = typingTimeoutsRef.current;
            const existing = timeouts.get(otherUser);
            if (existing) {
              clearTimeout(existing);
            }
            const timeoutId = setTimeout(() => {
              setTypingUsers((prev) =>
                prev.filter((u) => u !== otherUser)
              );
              timeouts.delete(otherUser);
            }, 3000);

            timeouts.set(otherUser, timeoutId);
          }
        }
      } catch (err) {
        console.error("Invalid WS event", err);
      }
    };

    return () => {
      isCancelled = true;
      ws.close();
      wsRef.current = null;

      // clear typing timeouts
      typingTimeoutsRef.current.forEach((timeoutId) =>
        clearTimeout(timeoutId)
      );
      typingTimeoutsRef.current.clear();
    };
  }, [roomId, username, token]);

  const sendMessage = (text: string) => {
    if (!text.trim()) return;
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      console.warn("WebSocket not open");
      return;
    }
    if (!username || !token) {
      console.warn("Not authenticated");
      return;
    }

    const payload = {
      type: "chat",
      text,
    };

    wsRef.current.send(JSON.stringify(payload));
  };

  const sendTyping = () => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) return;
    if (!username || !token) return;

    const payload = {
      type: "typing",
    };

    wsRef.current.send(JSON.stringify(payload));
  };

  return {
    messages,
    status,
    typingUsers,
    sendMessage,
    sendTyping,
  };
}