// frontend/src/hooks/useWebSocket.ts
"use client";

import { useEffect, useRef, useState } from "react";
import type { ChatMessage, WSEvent } from "@/types";
import { fetchCurrentUser, fetchRoomMessages } from "@/lib/api";
import { WS_BASE_URL } from "@/lib/config";
import { mergeMessages, parseEvents } from "@/lib/messages";
import { createReconnectingSocket, type ConnectionStatus } from "@/lib/reconnect";

interface UseWebSocketOptions {
  username: string | null;
  roomId: string;
  token: string | null;
  // Called when the server no longer accepts our token (logged out elsewhere or expired)
  onAuthInvalid?: () => void;
}

export function useWebSocket({
  username,
  roomId,
  token,
  onAuthInvalid,
}: UseWebSocketOptions) {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [status, setStatus] = useState<ConnectionStatus>("connecting");
  const [typingUsers, setTypingUsers] = useState<string[]>([]);
  const [serverError, setServerError] = useState<string | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  const typingTimeoutsRef = useRef<Map<string, ReturnType<typeof setTimeout>>>(
    new Map()
  );

  // Kept in a ref so a new callback identity doesn't force a reconnect
  const onAuthInvalidRef = useRef(onAuthInvalid);
  useEffect(() => {
    onAuthInvalidRef.current = onAuthInvalid;
  }, [onAuthInvalid]);

  useEffect(() => {
    let isCancelled = false;
    const typingTimeouts = typingTimeoutsRef.current;

    setMessages([]);
    setStatus("connecting");
    setTypingUsers([]);
    setServerError(null);

    // Fetch history on every (re)connect to fill in anything missed while offline
    const loadHistory = async () => {
      try {
        const history = await fetchRoomMessages(roomId, 50);
        if (!isCancelled && Array.isArray(history)) {
          setMessages((prev) => mergeMessages(history, prev));
        }
      } catch (err) {
        console.error("Failed to fetch history", err);
      }
    };

    const handleEvent = (data: WSEvent) => {
      if (data.type === "chat") {
        const msg = data.message;
        setMessages((prev) => mergeMessages(prev, [msg]));
        // remove chatter from typing list
        setTypingUsers((prev) => prev.filter((u) => u !== msg.username));
      } else if (data.type === "typing") {
        const otherUser = data.username;
        // ignore our own typing events
        if (otherUser === username) return;

        // add / keep in typing list
        setTypingUsers((prev) =>
          prev.includes(otherUser) ? prev : [...prev, otherUser]
        );

        // reset timeout for this user
        const existing = typingTimeouts.get(otherUser);
        if (existing) {
          clearTimeout(existing);
        }
        const timeoutId = setTimeout(() => {
          setTypingUsers((prev) => prev.filter((u) => u !== otherUser));
          typingTimeouts.delete(otherUser);
        }, 3000);

        typingTimeouts.set(otherUser, timeoutId);
      } else if (data.type === "error") {
        setServerError(data.message);
      }
    };

    // without a token the server treats us as a read-only guest
    const tokenParam = token ? `&token=${encodeURIComponent(token)}` : "";
    const wsUrl = `${WS_BASE_URL}/ws?roomId=${encodeURIComponent(
      roomId
    )}${tokenParam}`;

    const connection = createReconnectingSocket({
      connect: ({ onOpen, onClose, onMessage }) => {
        const ws = new WebSocket(wsUrl);
        ws.onopen = onOpen;
        ws.onclose = onClose;
        ws.onmessage = (event: MessageEvent) => onMessage(event.data);
        wsRef.current = ws;
        return ws;
      },
      onOpen: loadHistory,
      onMessage: (data) => parseEvents(String(data)).forEach(handleEvent),
      onStatusChange: setStatus,
      // A closed socket may mean our session ended; don't retry with a dead token
      shouldReconnect: async () => {
        wsRef.current = null;
        setTypingUsers([]);
        if (!token) return true;

        // Throws if the server is unreachable, which keeps us retrying
        const user = await fetchCurrentUser(token);
        if (user || isCancelled) return true;
        onAuthInvalidRef.current?.();
        return false;
      },
    });

    // Retry right away when the browser comes back online
    window.addEventListener("online", connection.retryNow);

    return () => {
      isCancelled = true;
      window.removeEventListener("online", connection.retryNow);
      connection.stop();
      wsRef.current = null;

      // clear typing timeouts
      typingTimeouts.forEach((timeoutId) => clearTimeout(timeoutId));
      typingTimeouts.clear();
    };
  }, [roomId, username, token]);

  // Returns false if the message couldn't be sent, so the caller can keep the text
  const sendMessage = (text: string): boolean => {
    if (!text.trim()) return false;
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      console.warn("WebSocket not open");
      return false;
    }
    if (!username || !token) {
      console.warn("Not authenticated");
      return false;
    }

    const payload = {
      type: "chat",
      text,
    };

    setServerError(null);
    wsRef.current.send(JSON.stringify(payload));
    return true;
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
    serverError,
    sendMessage,
    sendTyping,
  };
}
