// frontend/src/lib/messages.ts
import type { ChatMessage, WSEvent } from "@/types";

// Combines two message lists, dropping duplicates by ID and keeping them in time order
export function mergeMessages(
  a: ChatMessage[],
  b: ChatMessage[]
): ChatMessage[] {
  const seen = new Set<number>();
  const merged: ChatMessage[] = [];
  for (const m of [...a, ...b]) {
    // ID 0 means the server couldn't save it, so there's nothing to de-duplicate on
    if (m.id) {
      if (seen.has(m.id)) continue;
      seen.add(m.id);
    }
    merged.push(m);
  }
  return merged.sort(
    (x, y) =>
      new Date(x.timestamp).getTime() - new Date(y.timestamp).getTime() ||
      x.id - y.id
  );
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

export function isChatMessage(value: unknown): value is ChatMessage {
  return (
    isRecord(value) &&
    typeof value.id === "number" &&
    typeof value.username === "string" &&
    typeof value.roomId === "string" &&
    typeof value.text === "string" &&
    typeof value.timestamp === "string" &&
    !Number.isNaN(Date.parse(value.timestamp))
  );
}

function isWSEvent(value: unknown): value is WSEvent {
  if (!isRecord(value)) return false;
  switch (value.type) {
    case "chat":
      return isChatMessage(value.message);
    case "typing":
      return typeof value.username === "string";
    case "error":
      return typeof value.message === "string";
    default:
      return false;
  }
}

// Parses a WebSocket frame into events. The server may batch several events
// into one frame, one per line; lines that aren't valid events are skipped.
export function parseEvents(data: string): WSEvent[] {
  const events: WSEvent[] = [];
  for (const line of data.split("\n")) {
    if (!line.trim()) continue;
    try {
      const event: unknown = JSON.parse(line);
      if (isWSEvent(event)) {
        events.push(event);
      } else {
        console.warn("Ignoring unrecognized WS event", event);
      }
    } catch (err) {
      console.error("Invalid WS event", err);
    }
  }
  return events;
}
