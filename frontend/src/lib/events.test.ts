import { beforeEach, describe, expect, it, vi } from "vitest";
import { isChatMessage, parseEvents } from "./messages";

const message = {
  id: 1,
  username: "alice",
  roomId: "general",
  text: "hello",
  timestamp: "2026-11-02T12:00:00Z",
};

const chat = { type: "chat", message };
const typing = { type: "typing", username: "bob", roomId: "general", isTyping: true };
const error = { type: "error", code: "message_too_long", message: "Too long" };

const frame = (...events: unknown[]) =>
  events.map((e) => (typeof e === "string" ? e : JSON.stringify(e))).join("\n");

describe("parseEvents", () => {
  beforeEach(() => {
    // Invalid input is logged; keep test output clean
    vi.spyOn(console, "error").mockImplementation(() => {});
    vi.spyOn(console, "warn").mockImplementation(() => {});
  });

  it("parses each supported event type", () => {
    expect(parseEvents(frame(chat))).toEqual([chat]);
    expect(parseEvents(frame(typing))).toEqual([typing]);
    expect(parseEvents(frame(error))).toEqual([error]);
  });

  it("splits a batched frame into separate events, in order", () => {
    expect(parseEvents(frame(typing, chat, error))).toEqual([typing, chat, error]);
  });

  it("ignores blank lines and trailing newlines", () => {
    expect(parseEvents(`\n${frame(chat)}\n\n`)).toEqual([chat]);
    expect(parseEvents("")).toEqual([]);
  });

  it("skips lines that aren't JSON but keeps the rest of the frame", () => {
    expect(parseEvents(frame(chat, "{not json", typing))).toEqual([chat, typing]);
    expect(console.error).toHaveBeenCalled();
  });

  it.each([
    ["an unknown type", { type: "presence", username: "bob" }],
    ["a missing type", { username: "bob" }],
    ["a chat event without a message", { type: "chat" }],
    ["a chat event with a malformed message", { type: "chat", message: { text: "hi" } }],
    ["a typing event without a username", { type: "typing" }],
    ["an error event without a message", { type: "error", code: "x" }],
    ["a JSON primitive", 42],
    ["JSON null", null],
  ])("skips %s", (_label, event) => {
    expect(parseEvents(frame(event, chat))).toEqual([chat]);
    expect(console.warn).toHaveBeenCalled();
  });
});

describe("isChatMessage", () => {
  it("accepts a complete message", () => {
    expect(isChatMessage(message)).toBe(true);
  });

  it("accepts an unsaved message with ID 0", () => {
    expect(isChatMessage({ ...message, id: 0 })).toBe(true);
  });

  it.each(["id", "username", "roomId", "text", "timestamp"])(
    "rejects a message missing %s",
    (field) => {
      const incomplete: Record<string, unknown> = { ...message };
      delete incomplete[field];
      expect(isChatMessage(incomplete)).toBe(false);
    }
  );

  it("rejects fields of the wrong type", () => {
    expect(isChatMessage({ ...message, id: "1" })).toBe(false);
    expect(isChatMessage({ ...message, text: null })).toBe(false);
  });

  it("rejects an unparseable timestamp", () => {
    expect(isChatMessage({ ...message, timestamp: "yesterday" })).toBe(false);
  });

  it("rejects non-objects", () => {
    expect(isChatMessage(null)).toBe(false);
    expect(isChatMessage("hello")).toBe(false);
    expect(isChatMessage([message])).toBe(false);
  });
});
