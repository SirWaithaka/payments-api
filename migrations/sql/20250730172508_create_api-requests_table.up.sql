CREATE TABLE public."api_requests"
(
    "id"          uuid,
    "request_id"  text,
    "external_id" text,
    "partner"     text,
    "status"      text,
    "latency_ms"  bigint,
    "response"    JSONB,
    "payment_id"  text,
    "created_at"  timestamptz,
    "updated_at"  timestamptz,
    PRIMARY KEY ("id"),
    CONSTRAINT "fk_api_requests_payment" FOREIGN KEY ("payment_id") REFERENCES "payment_requests" ("payment_id"),
    CONSTRAINT "uni_api_requests_request_id" UNIQUE ("request_id"),
    CONSTRAINT "chk_api_requests_request_id" CHECK (request_id <> ''),
    CONSTRAINT "chk_api_requests_partner" CHECK (partner <> ''),
    CONSTRAINT "chk_api_requests_external_id" CHECK (external_id <> ''),
    CONSTRAINT "chk_api_requests_status" CHECK (status <> '')
);