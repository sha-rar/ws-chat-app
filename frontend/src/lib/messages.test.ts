import { describe, expect, it } from "vitest";
import type { ChatMessage } from "@/types";
import { mergeMessages } from "./messages";

const BASE_TIME = Date.parse("2026-11-02T12:00:00Z");

// msg(3) is message #3, sent 3 seconds after BASE_TIME
function msg(id: number, overrides: Partial<ChatMessage> = {}): ChatMessage {
  return {
    id,
    username: "alice",
    roomId: "general",
    text: `msg ${id}`,
    timestamp: new Date(BASE_TIME + id * 1000).toISOString(),
    ...overrides,
  };
}

const ids = (messages: ChatMessage[]) => messages.map((m) => m.id);

describe("mergeMessages", () => {
  it("returns an empty list when both inputs are empty", () => {
    expect(mergeMessages([], [])).toEqual([]);
  });

  it("appends a live message to the loaded history", () => {
    expect(ids(mergeMessages([msg(1), msg(2)], [msg(3)]))).toEqual([1, 2, 3]);
  });

  it("does not duplicate messages already shown when history is refetched after a reconnect", () => {
    const shownBeforeDisconnect = [msg(1), msg(2), msg(3)];
    const historyAfterReconnect = [msg(2), msg(3), msg(4), msg(5)];

    const merged = mergeMessages(historyAfterReconnect, shownBeforeDisconnect);

    expect(ids(merged)).toEqual([1, 2, 3, 4, 5]);
  });

  it("keeps older messages that have scrolled out of the refetched history window", () => {
    const shown = [msg(1), msg(2), msg(3)];
    const latestHistory = [msg(3), msg(4)];

    expect(ids(mergeMessages(latestHistory, shown))).toEqual([1, 2, 3, 4]);
  });

  it("keeps a live message that arrives before the history response", () => {
    // The socket opens, a new message arrives, then the history fetch resolves
    const liveFirst = [msg(6)];
    const history = [msg(4), msg(5)];

    expect(ids(mergeMessages(history, liveFirst))).toEqual([4, 5, 6]);
  });

  it("ignores a live message that is also in the history response", () => {
    const merged = mergeMessages([msg(1), msg(2)], [msg(2)]);

    expect(ids(merged)).toEqual([1, 2]);
  });

  it("sorts messages by timestamp regardless of input order", () => {
    expect(ids(mergeMessages([msg(3), msg(1)], [msg(4), msg(2)]))).toEqual([
      1, 2, 3, 4,
    ]);
  });

  it("orders messages with the same timestamp by ID", () => {
    const sameTime = new Date(BASE_TIME).toISOString();
    const merged = mergeMessages(
      [msg(9, { timestamp: sameTime }), msg(7, { timestamp: sameTime })],
      [msg(8, { timestamp: sameTime })]
    );

    expect(ids(merged)).toEqual([7, 8, 9]);
  });

  it("compares timestamps in different time zone formats correctly", () => {
    const earlier = msg(1, { timestamp: "2026-11-02T13:00:00-06:00" }); // 19:00:00Z
    const later = msg(2, { timestamp: "2026-11-02T19:00:01Z" });

    expect(ids(mergeMessages([later], [earlier]))).toEqual([1, 2]);
  });

  it("keeps every unsaved message (ID 0) since they can't be de-duplicated", () => {
    const unsaved = msg(0, { text: "not saved" });

    const merged = mergeMessages([unsaved], [{ ...unsaved }]);

    expect(merged).toHaveLength(2);
  });

  it("keeps the first copy when the same ID appears twice", () => {
    const merged = mergeMessages([msg(1, { text: "from history" })], [
      msg(1, { text: "from socket" }),
    ]);

    expect(merged).toEqual([msg(1, { text: "from history" })]);
  });

  it("does not modify its inputs", () => {
    const a = [msg(3), msg(1)];
    const b = [msg(2)];

    mergeMessages(a, b);

    expect(ids(a)).toEqual([3, 1]);
    expect(ids(b)).toEqual([2]);
  });
});
