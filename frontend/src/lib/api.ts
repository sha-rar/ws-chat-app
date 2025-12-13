// frontend/src/lib/api.ts
import type { ChatMessage, AuthResponse } from "@/types";

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080";

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