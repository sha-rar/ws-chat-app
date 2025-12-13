// frontend/src/hooks/useWebSocket.ts
"use client";

import { useEffect, useRef, useState } from "react";
import type { ChatMessage } from "@/types";

type ConnectionStatus = "connecting" | "open" | "closed" | "error";

interface UseWebSocketOptions {
  username: string;
  roomId: string;
}

export function useWebSocket({ username, roomId }: UseWebSocketOptions) {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [status, setStatus] = useState<ConnectionStatus>("connecting");
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    const wsUrl = `ws://localhost:8080/ws?roomId=${encodeURIComponent(
      roomId
    )}`;

    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    setStatus("connecting");

    ws.onopen = () => {
      setStatus("open");
    };

    ws.onclose = () => {
      setStatus("closed");
    };

    ws.onerror = () => {
      setStatus("error");
    };

    ws.onmessage = (event: MessageEvent) => {
      try {
        const msg = JSON.parse(event.data) as ChatMessage;
        setMessages((prev) => [...prev, msg]);
      } catch (err) {
        // ignore malformed data for now
        console.error("Invalid message", err);
      }
    };

    return () => {
      ws.close();
      wsRef.current = null;
    };
  }, [roomId]);

  const sendMessage = (text: string) => {
    if (!text.trim()) return;
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      console.warn("WebSocket not open");
      return;
    }

    const payload = {
      username,
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