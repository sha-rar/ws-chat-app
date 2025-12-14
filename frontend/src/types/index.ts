// frontend/src/types/index.ts
export interface ChatMessage {
  username: string;
  roomId: string;
  text: string;
  timestamp: string;
}

export interface User {
  id: number;
  username: string;
  createdAt?: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

// WebSocket events from the backend
export type WSChatEvent = {
  type: "chat";
  message: ChatMessage;
};

export type WSTypingEvent = {
  type: "typing";
  username: string;
  roomId: string;
  isTyping: boolean;
};

export type WSEvent = WSChatEvent | WSTypingEvent;