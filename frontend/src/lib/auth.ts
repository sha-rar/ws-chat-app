// frontend/src/lib/auth.ts
import type { AuthResponse } from "@/types";

const STORAGE_KEY = "auth";

export function getStoredAuth(): AuthResponse | null {
  if (typeof window === "undefined") return null;
  const stored = localStorage.getItem(STORAGE_KEY);
  if (!stored) return null;
  try {
    return JSON.parse(stored) as AuthResponse;
  } catch (err) {
    console.error("Invalid auth in localStorage", err);
    return null;
  }
}

export function setStoredAuth(auth: AuthResponse) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(auth));
}

export function clearStoredAuth() {
  localStorage.removeItem(STORAGE_KEY);
}
