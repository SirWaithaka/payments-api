CREATE TABLE IF NOT EXISTS public."mpesa_shortcodes"
(
    "id"                 text,
    "service"            text NOT NULL,
    "type"               text NOT NULL,
    "shortcode"          text NOT NULL,
    "initiator_name"     text,
    "initiator_password" text,
    "passphrase"         text,
    "key"                text NOT NULL,
    "secret"             text NOT NULL,
    "callback_url"       text NOT NULL,
    "created_at"         timestamp,
    "updated_at"         timestamp,
    PRIMARY KEY ("id"),
    CONSTRAINT "chk_mpesa_shortcodes_service" CHECK (service <> ''),
    CONSTRAINT "chk_mpesa_shortcodes_type" CHECK (type <> ''),
    CONSTRAINT "chk_mpesa_shortcodes_shortcode" CHECK (shortcode <> ''),
    CONSTRAINT "chk_mpesa_shortcodes_key" CHECK (key <> ''),
    CONSTRAINT "chk_mpesa_shortcodes_secret" CHECK (secret <> '')
);