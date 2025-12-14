// frontend/src/app/chat/[roomId]/page.tsx
"use client";

import { useParams, useSearchParams } from "next/navigation";
import { ChatRoom } from "@/components/ChatRoom";

export default function ChatPage() {
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
