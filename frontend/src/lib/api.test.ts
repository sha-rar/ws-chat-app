import { describe, expect, it, vi } from "vitest";
import {
  fetchCurrentUser,
  fetchRoomMessages,
  loginUser,
  logoutUser,
  registerUser,
} from "./api";
import { API_BASE_URL } from "./config";

const message = {
  id: 1,
  username: "alice",
  roomId: "general",
  text: "hello",
  timestamp: "2026-11-02T12:00:00Z",
};

// Replaces fetch with a stub that answers every request with the given response
function respondWith(body: string | null, status = 200) {
  const fetchMock = vi.fn(async () => new Response(body, { status }));
  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
}
const respondWithJSON = (body: unknown, status = 200) =>
  respondWith(JSON.stringify(body), status);

function failWithNetworkError() {
  vi.stubGlobal(
    "fetch",
    vi.fn(async () => {
      throw new TypeError("Failed to fetch");
    })
  );
}

describe("fetchRoomMessages", () => {
  it("requests the room's history with the limit", async () => {
    const fetchMock = respondWithJSON([]);

    await fetchRoomMessages("my room/1", 25);

    expect(fetchMock).toHaveBeenCalledWith(
      `${API_BASE_URL}/api/rooms/my%20room%2F1/messages?limit=25`,
      expect.objectContaining({ method: "GET" })
    );
  });

  it("returns the messages", async () => {
    respondWithJSON([message, { ...message, id: 2 }]);

    expect(await fetchRoomMessages("general")).toEqual([
      message,
      { ...message, id: 2 },
    ]);
  });

  it("returns an empty list for an empty room", async () => {
    respondWithJSON([]);
    expect(await fetchRoomMessages("general")).toEqual([]);
  });

  it("treats a null body (from older servers) as an empty room", async () => {
    respondWithJSON(null);
    expect(await fetchRoomMessages("general")).toEqual([]);
  });

  it("drops malformed entries but keeps valid ones", async () => {
    respondWithJSON([message, { text: "no id or author" }, null, "hello"]);
    expect(await fetchRoomMessages("general")).toEqual([message]);
  });

  it.each([
    ["an object", { messages: [message] }],
    ["a string", "hello"],
    ["a number", 42],
  ])("rejects a response that is %s rather than a list", async (_label, body) => {
    respondWithJSON(body);
    await expect(fetchRoomMessages("general")).rejects.toThrow("unexpected response");
  });

  it("rejects a body that isn't JSON", async () => {
    respondWith("<html>Bad Gateway</html>");
    await expect(fetchRoomMessages("general")).rejects.toThrow();
  });

  it("rejects a server error with its status", async () => {
    respondWith("internal error", 500);
    await expect(fetchRoomMessages("general")).rejects.toThrow("500");
  });

  it("rejects when the server is unreachable", async () => {
    failWithNetworkError();
    await expect(fetchRoomMessages("general")).rejects.toThrow("Failed to fetch");
  });
});

describe("fetchCurrentUser", () => {
  it("sends the token and returns the user", async () => {
    const fetchMock = respondWithJSON({ id: 1, username: "alice" });

    expect(await fetchCurrentUser("tok")).toEqual({ id: 1, username: "alice" });
    expect(fetchMock).toHaveBeenCalledWith(
      `${API_BASE_URL}/api/auth/me`,
      expect.objectContaining({ headers: { Authorization: "Bearer tok" } })
    );
  });

  it("returns null when the session has ended", async () => {
    respondWith("invalid or expired token", 401);
    expect(await fetchCurrentUser("tok")).toBeNull();
  });

  // The reconnect logic relies on these throwing (keep retrying) rather than
  // returning null (log the user out)
  it("throws on a server error instead of reporting the session as ended", async () => {
    respondWith("internal error", 500);
    await expect(fetchCurrentUser("tok")).rejects.toThrow("500");
  });

  it("throws when the server is unreachable", async () => {
    failWithNetworkError();
    await expect(fetchCurrentUser("tok")).rejects.toThrow();
  });
});

describe("logoutUser", () => {
  it("posts with the token", async () => {
    const fetchMock = respondWith(null, 204);

    await logoutUser("tok");

    expect(fetchMock).toHaveBeenCalledWith(
      `${API_BASE_URL}/api/auth/logout`,
      expect.objectContaining({ method: "POST", headers: { Authorization: "Bearer tok" } })
    );
  });

  it("succeeds if the session was already gone", async () => {
    respondWith("invalid or expired token", 401);
    await expect(logoutUser("tok")).resolves.toBeUndefined();
  });

  it("throws on a server error", async () => {
    respondWith("failed to log out", 500);
    await expect(logoutUser("tok")).rejects.toThrow("500");
  });
});

describe("loginUser and registerUser", () => {
  it("returns the login response", async () => {
    const auth = { token: "tok", user: { id: 1, username: "alice" } };
    respondWithJSON(auth);
    expect(await loginUser("alice", "secret1")).toEqual(auth);
  });

  it("includes the server's reason when login fails", async () => {
    respondWith("invalid credentials\n", 401);
    await expect(loginUser("alice", "wrong")).rejects.toThrow(
      /401 invalid credentials/
    );
  });

  it("includes the server's reason when registration fails", async () => {
    respondWith("password too short (min 6)\n", 400);
    await expect(registerUser("alice", "123")).rejects.toThrow(
      /400 password too short/
    );
  });

  it("sends credentials as JSON", async () => {
    const fetchMock = respondWithJSON({ id: 1, username: "alice" }, 201);

    await registerUser("alice", "secret1");

    expect(fetchMock).toHaveBeenCalledWith(
      `${API_BASE_URL}/api/auth/register`,
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ username: "alice", password: "secret1" }),
      })
    );
  });
});
