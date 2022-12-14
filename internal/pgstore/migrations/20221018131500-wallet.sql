-- +migrate Up
CREATE TABLE IF NOT EXISTS wallet
(
    id               bigserial PRIMARY KEY                  NOT NULL,
    owner_id         int UNIQUE                             NOT NULL,
    balance          numeric(11,2)                          NOT NULL,
    reserved_balance numeric(11,2)                          NOT NULL,
    created_at       timestamp with time zone DEFAULT NOW() NOT NULL,
    updated_at       timestamp with time zone DEFAULT NOW() NOT NULL
    );

CREATE TABLE IF NOT EXISTS services
(
    id    bigserial PRIMARY KEY NOT NULL,
    title text                  NOT NULL
);

CREATE TABLE transaction
(
    id               bigserial PRIMARY KEY                  NOT NULL,
    idempotence_key  int UNIQUE                             NOT NULL,
    wallet_id        bigint REFERENCES wallet (id)          NOT NULL,
    amount           numeric(11, 2)                         NOT NULL,
    target_wallet_id bigint REFERENCES wallet (id),
    service_id       int REFERENCES services (id),
    comment          text                                   NOT NULL,
    timestamp        timestamp with time zone DEFAULT NOW() NOT NULL
);

CREATE TABLE reserved_funds
(
    id         bigserial PRIMARY KEY                  NOT NULL,
    order_id   int UNIQUE                             NOT NULL,
    owner_id   int REFERENCES wallet (owner_id)       NOT NULL,
    service_id int REFERENCES services (id)           NOT NULL,
    amount     numeric(11,2)                          NOT NULL,
    status     text                                   NOT NULL,
    created_at timestamp with time zone DEFAULT NOW() NOT NULL,
    updated_at timestamp with time zone DEFAULT NOW() NOT NULL
);

INSERT INTO services (title)
VALUES ('Услуга');


-- +migrate Down
DROP TABLE reserved_funds CASCADE;
DROP TABLE transaction CASCADE;
DROP TABLE services CASCADE;
DROP TABLE wallet CASCADE;