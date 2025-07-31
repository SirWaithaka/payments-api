CREATE TABLE public."webhook_requests"
(
    "id"         UUID,
    "action"     text,
    "partner"    text,
    "payload"    JSONB,
    "created_at" timestamptz,
    PRIMARY KEY ("id")
);