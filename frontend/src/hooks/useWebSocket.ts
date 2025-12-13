// frontend/src/hooks/useWebSocket.ts
"use client";

import { useEffect, useRef, useState } from "react";
import type { ChatMessage } from "@/types";
import { fetchRoomMessages } from "@/lib/api";

type ConnectionStatus = "connecting" | "open" | "closed" | "error";

interface UseWebSocketOptions {
  username: string | null;
  roomId: string;
  token: string | null;
}

export function useWebSocket({ username, roomId, token }: UseWebSocketOptions) {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [status, setStatus] = useState<ConnectionStatus>("connecting");
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    let isCancelled = false;

    // If not authenticated yet, don't open a WebSocket
    if (!username || !token) {
      setMessages([]);
      setStatus("closed");
      return;
    }

    setMessages([]);
    setStatus("connecting");

    // fetch history
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

    const wsUrl = `ws://localhost:8080/ws?roomId=${encodeURIComponent(
      roomId
    )}&token=${encodeURIComponent(token)}`;

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
        const msg = JSON.parse(event.data) as ChatMessage;
        if (!isCancelled) {
          setMessages((prev) => [...prev, msg]);
        }
      } catch (err) {
        console.error("Invalid message", err);
      }
    };

    return () => {
      isCancelled = true;
      ws.close();
      wsRef.current = null;
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
      username, // backend will override it anyway with auth username
      roomId,
      text,
    };

    wsRef.current.send(JSON.stringify(payload));
  };

  return {
    messages,
    status,
    sendMessage,
  };
}