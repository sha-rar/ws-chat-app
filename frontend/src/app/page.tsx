// frontend/src/app/page.tsx
import Link from "next/link";

export default function HomePage() {
  return (
    <main
      style={{
        minHeight: "100vh",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        fontFamily: "system-ui, sans-serif",
        background: "#f3f4f6",
      }}
    >
      <div
        style={{
          padding: "32px 24px",
          borderRadius: 12,
          background: "#ffffff",
          boxShadow: "0 10px 30px rgba(0,0,0,0.08)",
          maxWidth: 420,
          width: "100%",
          textAlign: "center",
        }}
      >
        <h1 style={{ fontSize: 28, marginBottom: 8 }}>Real-time Chat</h1>
        <p style={{ fontSize: 14, opacity: 0.7, marginBottom: 24 }}>
          Welcome! Choose an option to get started.
        </p>

        <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
          <Link href="/login">
            <button
              type="button"
              style={{
                width: "100%",
                padding: "10px 16px",
                borderRadius: 8,
                border: "none",
                background: "#2563eb",
                color: "#fff",
                fontSize: 14,
                fontWeight: 500,
                cursor: "pointer",
              }}
            >
              Returning user? Login
            </button>
          </Link>

          <Link href="/register">
            <button
              type="button"
              style={{
                width: "100%",
                padding: "10px 16px",
                borderRadius: 8,
                border: "1px solid #d1d5db",
                background: "#ffffff",
                color: "#111827",
                fontSize: 14,
                fontWeight: 500,
                cursor: "pointer",
              }}
            >
              New user? Register
            </button>
          </Link>

          <Link href="/chat">
            <button
              type="button"
              style={{
                width: "100%",
                padding: "8px 16px",
                borderRadius: 8,
                border: "none",
                background: "transparent",
                color: "#4b5563",
                fontSize: 13,
                cursor: "pointer",
                textDecoration: "underline",
                marginTop: 4,
              }}
            >
              Or go straight to chat lobby
            </button>
          </Link>
        </div>
      </div>
    </main>
  );
}