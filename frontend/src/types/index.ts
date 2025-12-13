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