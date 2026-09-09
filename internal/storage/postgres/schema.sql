CREATE TABLE IF NOT EXISTS links
(
    short_code   VARCHAR(10)
        CONSTRAINT links_short_code_pk PRIMARY KEY,
    original_url TEXT        NOT NULL
        CONSTRAINT links_original_url_uq UNIQUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);