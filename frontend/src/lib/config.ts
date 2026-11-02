// frontend/src/lib/config.ts

// NEXT_PUBLIC_* values are inlined at build time, so set them before `npm run build`
export const API_BASE_URL = (
  process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080"
).replace(/\/+$/, "");

// Defaults to the API host, switching http(s) to ws(s)
export const WS_BASE_URL = (
  process.env.NEXT_PUBLIC_WS_URL || API_BASE_URL.replace(/^http/, "ws")
).replace(/\/+$/, "");
