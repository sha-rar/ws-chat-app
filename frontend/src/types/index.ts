// frontend/src/types/index.ts
export interface ChatMessage {
  username: string;
  roomId: string;
  text: string;
  timestamp: string; // ISO string from backend
}