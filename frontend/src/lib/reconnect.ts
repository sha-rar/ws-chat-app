// frontend/src/lib/reconnect.ts

export type ConnectionStatus = "connecting" | "open" | "reconnecting" | "closed";

export const RECONNECT_BASE_DELAY_MS = 1000;
export const RECONNECT_MAX_DELAY_MS = 15000;

// Exponential backoff with jitter so clients don't all retry at once.
// attempt 0 waits 0.5-1s, attempt 1 waits 1-2s, and so on up to 7.5-15s.
export function getReconnectDelay(
  attempt: number,
  random: () => number = Math.random
): number {
  const ceiling = Math.min(
    RECONNECT_MAX_DELAY_MS,
    RECONNECT_BASE_DELAY_MS * 2 ** attempt
  );
  return ceiling * (0.5 + random() / 2);
}

export interface SocketHandlers {
  onOpen: () => void;
  // onerror is always followed by onclose, so only close needs handling
  onClose: () => void;
  onMessage: (data: unknown) => void;
}

export interface ReconnectingSocketOptions {
  // Opens a socket and wires its events to the given handlers
  connect: (handlers: SocketHandlers) => { close: () => void };
  onOpen?: () => void;
  onMessage?: (data: unknown) => void;
  onStatusChange?: (status: ConnectionStatus) => void;
  // Called after the socket closes unexpectedly; resolve false to give up.
  // If it throws (e.g. the server is unreachable), we keep retrying.
  shouldReconnect?: () => boolean | Promise<boolean>;
  random?: () => number;
}

export interface ReconnectingSocket {
  // Closes the socket and cancels any pending retry
  stop: () => void;
  // Skips the remaining wait if a retry is pending (e.g. when back online)
  retryNow: () => void;
}

export function createReconnectingSocket({
  connect,
  onOpen,
  onMessage,
  onStatusChange,
  shouldReconnect,
  random = Math.random,
}: ReconnectingSocketOptions): ReconnectingSocket {
  let stopped = false;
  let attempts = 0;
  let timer: ReturnType<typeof setTimeout> | null = null;
  let socket: { close: () => void } | null = null;
  // Incremented per socket so events from a replaced socket are ignored
  let generation = 0;

  const scheduleRetry = () => {
    if (stopped || timer) return;
    const delay = getReconnectDelay(attempts, random);
    attempts++;
    onStatusChange?.("reconnecting");
    timer = setTimeout(() => {
      timer = null;
      open();
    }, delay);
  };

  const handleClose = async () => {
    let retry = true;
    if (shouldReconnect) {
      try {
        retry = await shouldReconnect();
      } catch {
        retry = true;
      }
    }
    if (stopped) return;
    if (!retry) {
      onStatusChange?.("closed");
      return;
    }
    scheduleRetry();
  };

  const open = () => {
    if (stopped) return;
    const current = ++generation;
    const isCurrent = () => !stopped && current === generation;

    socket = connect({
      onOpen: () => {
        if (!isCurrent()) return;
        attempts = 0;
        onStatusChange?.("open");
        onOpen?.();
      },
      onClose: () => {
        if (!isCurrent()) return;
        socket = null;
        handleClose();
      },
      onMessage: (data) => {
        if (isCurrent()) onMessage?.(data);
      },
    });
  };

  onStatusChange?.("connecting");
  open();

  return {
    stop: () => {
      stopped = true;
      if (timer) clearTimeout(timer);
      timer = null;
      socket?.close();
      socket = null;
    },
    retryNow: () => {
      if (stopped || !timer) return;
      clearTimeout(timer);
      timer = null;
      attempts = 0;
      open();
    },
  };
}
