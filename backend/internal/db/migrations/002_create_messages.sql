CREATE TABLE IF NOT EXISTS messages (
    id        BIGSERIAL PRIMARY KEY,
    room_id   TEXT NOT NULL,
    username  TEXT NOT NULL,
    text      TEXT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_messages_room_timestamp
    ON messages (room_id, timestamp DESC, id DESC);
