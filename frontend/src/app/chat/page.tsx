// frontend/src/app/chat/page.tsx
"use client";

import { useParams } from "next/navigation";
import { ChatRoom } from "@/components/ChatRoom";

export default function ChatPage() {
  const params = useParams<{ roomId: string }>();

  const roomIdRaw = params.roomId;
  const roomId =
    typeof roomIdRaw === "string"
      ? roomIdRaw
      : Array.isArray(roomIdRaw)
      ? roomIdRaw[0]
      : "general";

  return <ChatRoom initialRoomId={roomId} />;
}