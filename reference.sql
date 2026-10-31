-- users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- eg 23416
    username VARCHAR(100) NOT NULL UNIQUE, --eg sulwe
    date_of_birth DATE NOT NULL,
    document_number VARCHAR(100) NOT NULL UNIQUE, --eg DOC00875
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

--exchanges table (ie. binace, coinbase etc)
-- an exchange may have millions of users (binance may hold peter's acc, Mary's acc etc)
-- exchanges are created through seed script not API so no need for updated_at column
CREATE TABLE exchanges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), --eg 768001
    name VARCHAR(100) NOT NULL UNIQUE,  --eg coinbase
    minimum_age SMALLINT NOT NULL, -- eg 18
    maximum_transfer_amount NUMERIC(20, 2) NOT NULL, -- EG. 1000.00
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_minimum_age
        CHECK (minimum_age > 0),

    CONSTRAINT valid_maximum_transfer
        CHECK (maximum_transfer_amount > 0)
);

-- accounts table
-- a user may have multiple accounts (john owns binance account and coinbase account)
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id), -- eg john's id
    exchange_id UUID NOT NULL REFERENCES exchanges(id),
    balance NUMERIC(20,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_user_exchange
        UNIQUE (user_id, exchange_id), -- a user can't have two accounts in one exhange ( eg peter having 2 binance accounts)

    CONSTRAINT non_negative-balance
        CHECK (balance >= 0)
);