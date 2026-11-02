import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  createReconnectingSocket,
  getReconnectDelay,
  RECONNECT_BASE_DELAY_MS,
  RECONNECT_MAX_DELAY_MS,
  type ConnectionStatus,
  type ReconnectingSocketOptions,
  type SocketHandlers,
} from "./reconnect";

// A stand-in for a WebSocket that the test opens, closes and messages by hand
class FakeSocket {
  closed = false;
  constructor(private handlers: SocketHandlers) {}
  open() {
    this.handlers.onOpen();
  }
  receive(data: unknown) {
    this.handlers.onMessage(data);
  }
  drop() {
    this.handlers.onClose();
  }
  close = vi.fn(() => {
    this.closed = true;
  });
}

function setup(options: Partial<ReconnectingSocketOptions> = {}) {
  const sockets: FakeSocket[] = [];
  const statuses: ConnectionStatus[] = [];
  const onOpen = vi.fn();
  const onMessage = vi.fn();

  const connection = createReconnectingSocket({
    connect: (handlers) => {
      const socket = new FakeSocket(handlers);
      sockets.push(socket);
      return socket;
    },
    onOpen,
    onMessage,
    onStatusChange: (s) => statuses.push(s),
    // Jitter always picks the full delay, so timings are predictable
    random: () => 1,
    ...options,
  });

  return {
    connection,
    sockets,
    statuses,
    onOpen,
    onMessage,
    latest: () => sockets[sockets.length - 1],
    status: () => statuses[statuses.length - 1],
  };
}

describe("getReconnectDelay", () => {
  it("doubles the delay with each attempt", () => {
    const full = () => 1;
    expect([0, 1, 2, 3].map((n) => getReconnectDelay(n, full))).toEqual([
      1000, 2000, 4000, 8000,
    ]);
  });

  it("caps the delay at the maximum", () => {
    expect(getReconnectDelay(4, () => 1)).toBe(RECONNECT_MAX_DELAY_MS);
    expect(getReconnectDelay(20, () => 1)).toBe(RECONNECT_MAX_DELAY_MS);
  });

  it("applies jitter between half and the full delay", () => {
    expect(getReconnectDelay(0, () => 0)).toBe(RECONNECT_BASE_DELAY_MS / 2);
    expect(getReconnectDelay(0, () => 0.5)).toBe(RECONNECT_BASE_DELAY_MS * 0.75);
    expect(getReconnectDelay(10, () => 0)).toBe(RECONNECT_MAX_DELAY_MS / 2);
  });

  it("uses Math.random by default", () => {
    vi.spyOn(Math, "random").mockReturnValue(0);
    expect(getReconnectDelay(1)).toBe(1000);
  });
});

describe("createReconnectingSocket", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it("connects immediately and reports when the socket opens", () => {
    const { sockets, latest, statuses, onOpen } = setup();

    expect(sockets).toHaveLength(1);
    expect(statuses).toEqual(["connecting"]);

    latest().open();

    expect(statuses).toEqual(["connecting", "open"]);
    expect(onOpen).toHaveBeenCalledTimes(1);
  });

  it("forwards messages from the open socket", () => {
    const { latest, onMessage } = setup();
    latest().open();

    latest().receive("hello");

    expect(onMessage).toHaveBeenCalledWith("hello");
  });

  it("retries after a dropped connection, waiting the backoff delay", async () => {
    const { sockets, latest, status } = setup();
    latest().open();

    latest().drop();
    await vi.advanceTimersByTimeAsync(0);
    expect(status()).toBe("reconnecting");

    await vi.advanceTimersByTimeAsync(RECONNECT_BASE_DELAY_MS - 1);
    expect(sockets).toHaveLength(1);

    await vi.advanceTimersByTimeAsync(1);
    expect(sockets).toHaveLength(2);
  });

  it("backs off exponentially while the server stays down", async () => {
    const { sockets, latest } = setup();
    const waits: number[] = [];

    // Each new socket fails before it ever opens
    for (let attempt = 0; attempt < 6; attempt++) {
      const before = sockets.length;
      latest().drop();
      let waited = 0;
      while (sockets.length === before) {
        await vi.advanceTimersByTimeAsync(100);
        waited += 100;
      }
      waits.push(waited);
    }

    expect(waits).toEqual([1000, 2000, 4000, 8000, 15000, 15000]);
  });

  it("resets the backoff after a successful reconnection", async () => {
    const { sockets, latest, status, onOpen } = setup();
    latest().open();

    // Fail twice so the next delay would be 4s
    latest().drop();
    await vi.advanceTimersByTimeAsync(1000);
    latest().drop();
    await vi.advanceTimersByTimeAsync(2000);
    expect(sockets).toHaveLength(3);

    latest().open();
    expect(status()).toBe("open");
    expect(onOpen).toHaveBeenCalledTimes(2);

    // Back to the base delay
    latest().drop();
    await vi.advanceTimersByTimeAsync(1000);
    expect(sockets).toHaveLength(4);
  });

  it("ignores events from a socket that has been replaced", async () => {
    const { latest, onMessage, onOpen, statuses } = setup();
    const first = latest();
    first.open();
    first.drop();
    await vi.advanceTimersByTimeAsync(1000);

    const statusCount = statuses.length;
    first.receive("stale");
    first.open();
    first.drop();
    await vi.advanceTimersByTimeAsync(0);

    expect(onMessage).not.toHaveBeenCalled();
    expect(onOpen).toHaveBeenCalledTimes(1);
    expect(statuses).toHaveLength(statusCount);
  });

  it("stops retrying when shouldReconnect returns false", async () => {
    const shouldReconnect = vi.fn().mockResolvedValue(false);
    const { sockets, latest, status } = setup({ shouldReconnect });
    latest().open();

    latest().drop();
    await vi.advanceTimersByTimeAsync(60_000);

    expect(shouldReconnect).toHaveBeenCalledTimes(1);
    expect(status()).toBe("closed");
    expect(sockets).toHaveLength(1);
  });

  it("keeps retrying when shouldReconnect throws", async () => {
    const shouldReconnect = vi.fn().mockRejectedValue(new Error("offline"));
    const { sockets, latest, status } = setup({ shouldReconnect });

    latest().drop();
    await vi.advanceTimersByTimeAsync(1000);

    expect(status()).toBe("reconnecting");
    expect(sockets).toHaveLength(2);
  });

  // retryNow is what the hook calls when the browser comes back online
  it("retries immediately with retryNow, skipping the wait", async () => {
    const { connection, sockets, latest } = setup();
    // Build up a long backoff first
    latest().drop();
    await vi.advanceTimersByTimeAsync(1000);
    latest().drop();
    await vi.advanceTimersByTimeAsync(0);
    expect(sockets).toHaveLength(2);

    connection.retryNow();
    expect(sockets).toHaveLength(3);

    // The pending timer was cancelled, and the backoff starts over
    latest().drop();
    await vi.advanceTimersByTimeAsync(1000);
    expect(sockets).toHaveLength(4);
  });

  it("does nothing on retryNow while connected", () => {
    const { connection, sockets, latest } = setup();
    latest().open();

    connection.retryNow();

    expect(sockets).toHaveLength(1);
  });

  describe("stop", () => {
    it("closes the current socket", () => {
      const { connection, latest } = setup();
      latest().open();

      connection.stop();

      expect(latest().close).toHaveBeenCalledTimes(1);
    });

    it("cancels a pending retry", async () => {
      const { connection, sockets, latest } = setup();
      latest().drop();
      await vi.advanceTimersByTimeAsync(0);

      connection.stop();
      await vi.advanceTimersByTimeAsync(60_000);

      expect(sockets).toHaveLength(1);
    });

    it("ignores a close that is still being checked by shouldReconnect", async () => {
      let resolveCheck: (retry: boolean) => void = () => {};
      const shouldReconnect = vi.fn(
        () => new Promise<boolean>((resolve) => (resolveCheck = resolve))
      );
      const { connection, sockets, latest, statuses } = setup({ shouldReconnect });
      latest().drop();

      connection.stop();
      const statusCount = statuses.length;
      resolveCheck(true);
      await vi.advanceTimersByTimeAsync(60_000);

      expect(sockets).toHaveLength(1);
      expect(statuses).toHaveLength(statusCount);
    });

    it("ignores events after stopping", () => {
      const { connection, latest, onMessage, onOpen } = setup();
      const socket = latest();

      connection.stop();
      socket.open();
      socket.receive("late");

      expect(onOpen).not.toHaveBeenCalled();
      expect(onMessage).not.toHaveBeenCalled();
    });

    it("makes retryNow a no-op", async () => {
      const { connection, sockets, latest } = setup();
      latest().drop();
      await vi.advanceTimersByTimeAsync(0);

      connection.stop();
      connection.retryNow();

      expect(sockets).toHaveLength(1);
    });
  });
});
