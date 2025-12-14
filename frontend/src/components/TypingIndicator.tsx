"use client";

interface TypingIndicatorProps {
  typingUsers: string[];
  currentUsername: string | null;
}

export function TypingIndicator({
  typingUsers,
  currentUsername,
}: TypingIndicatorProps) {
  const others = typingUsers.filter((u) => u !== currentUsername);

  if (others.length === 0) return null;

  let text: string;
  if (others.length === 1) {
    text = `${others[0]} is typing...`;
  } else if (others.length === 2) {
    text = `${others[0]} and ${others[1]} are typing...`;
  } else {
    text = `${others[0]}, ${others[1]} and others are typing...`;
  }

  return (
    <div
      style={{
        fontSize: 12,
        opacity: 0.7,
        marginBottom: 4,
        minHeight: 16,
      }}
    >
      {text}
    </div>
  );
}
