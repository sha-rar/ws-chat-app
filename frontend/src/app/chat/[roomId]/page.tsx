// frontend/src/app/chat/[roomId]/page.tsx
"use client";

import { useParams, useSearchParams } from "next/navigation";
import { Suspense } from "react";
import { ChatRoom } from "@/components/ChatRoom";

function ChatRoomPage() {
  const params = useParams<{ roomId: string }>();
  const searchParams = useSearchParams();

  const roomIdRaw = params.roomId;
  const roomId =
    typeof roomIdRaw === "string"
      ? roomIdRaw
      : Array.isArray(roomIdRaw)
      ? roomIdRaw[0]
      : "general";

  const forceGuest = searchParams.get("guest") === "1";

  return <ChatRoom initialRoomId={roomId} forceGuest={forceGuest} />;
}

// useSearchParams() needs a Suspense boundary so the page can be prerendered
export default function ChatPage() {
  return (
    <Suspense>
      <ChatRoomPage />
    </Suspense>
  );
}
