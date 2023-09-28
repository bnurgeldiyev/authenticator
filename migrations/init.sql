CREATE SCHEMA IF NOT EXISTS extensions;
CREATE EXTENSION IF NOT EXISTS unaccent SCHEMA extensions;

CREATE TYPE state_t AS ENUM ('enabled', 'disabled', 'deleted');

CREATE TABLE IF NOT EXISTS tbl_user
(
    id        UUID PRIMARY KEY                     DEFAULT gen_random_uuid(),
    username  VARCHAR(64)                 NOT NULL,
    password  VARCHAR(256)                NOT NULL,
    state     state_t                     NOT NULL,
    create_ts TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    update_ts TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    version   INT                         NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX uq_user_username ON tbl_user (username) WHERE
    state != 'deleted'::state_t;
