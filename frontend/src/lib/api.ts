// frontend/src/lib/api.ts
import type { ChatMessage, AuthResponse, User } from "@/types";
import { API_BASE_URL } from "@/lib/config";

export async function fetchRoomMessages(
  roomId: string,
  limit = 50
): Promise<ChatMessage[]> {
  const url = `${API_BASE_URL}/api/rooms/${encodeURIComponent(
    roomId
  )}/messages?limit=${limit}`;

  const res = await fetch(url, {
    method: "GET",
    cache: "no-store",
  });

  if (!res.ok) {
    throw new Error(`Failed to fetch messages: ${res.status}`);
  }

  return res.json();
}

export async function registerUser(username: string, password: string) {
  const res = await fetch(`${API_BASE_URL}/api/auth/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password }),
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(`Register failed: ${res.status} ${text}`);
  }

  return res.json(); // created user
}

export async function loginUser(
  username: string,
  password: string
): Promise<AuthResponse> {
  const res = await fetch(`${API_BASE_URL}/api/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password }),
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(`Login failed: ${res.status} ${text}`);
  }

  return res.json();
}

export async function logoutUser(token: string) {
  const res = await fetch(`${API_BASE_URL}/api/auth/logout`, {
    method: "POST",
    headers: { Authorization: `Bearer ${token}` },
  });

  // 401 means the session is already gone, which is what we wanted
  if (!res.ok && res.status !== 401) {
    throw new Error(`Logout failed: ${res.status}`);
  }
}

// Returns the logged-in user, or null if the token has expired or been revoked
export async function fetchCurrentUser(token: string): Promise<User | null> {
  const res = await fetch(`${API_BASE_URL}/api/auth/me`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });

  if (res.status === 401) return null;
  if (!res.ok) {
    throw new Error(`Failed to fetch current user: ${res.status}`);
  }

  return res.json();
}
