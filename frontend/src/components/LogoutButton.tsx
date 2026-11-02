// frontend/src/components/LogoutButton.tsx
"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { logoutUser } from "@/lib/api";
import { clearStoredAuth, getStoredAuth } from "@/lib/auth";

export function LogoutButton() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);

  const handleLogout = async () => {
    setLoading(true);
    const auth = getStoredAuth();
    if (auth) {
      try {
        await logoutUser(auth.token);
      } catch (err) {
        // Still log out locally; the token expires on its own eventually
        console.error("Logout request failed", err);
      }
    }
    clearStoredAuth();
    router.push("/login");
  };

  return (
    <button
      type="button"
      onClick={handleLogout}
      disabled={loading}
      style={{
        padding: 0,
        border: "none",
        background: "none",
        color: "inherit",
        textDecoration: "underline",
        cursor: "pointer",
        font: "inherit",
      }}
    >
      {loading ? "Logging out..." : "Log out"}
    </button>
  );
}
