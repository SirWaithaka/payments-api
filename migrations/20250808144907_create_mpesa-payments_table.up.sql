CREATE TABLE IF NOT EXISTS public."mpesa_payments"
(
    "id"                         uuid,
    "payment_id"                 text NOT NULL,
    "type"                       text NOT NULL,
    "status"                     text NOT NULL,
    "client_transaction_id"      text NOT NULL,
    "idempotency_id"             text NOT NULL,
    "payment_reference"          text,
    "amount"                     text NOT NULL,
    "source_account_number"      text NOT NULL,
    "destination_account_number" text NOT NULL,
    "beneficiary"                text,
    "description"                text,
    "shortcode_id"               text,
    "created_at"                 timestamp,
    "updated_at"                 timestamp,
    PRIMARY KEY ("id"),
    CONSTRAINT "chk_mpesa_payments_description" CHECK (description <> '')
)