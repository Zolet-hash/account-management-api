CREATE TABLE users (
	id uuid PRIMARY KEY, 
	username text NOT NULL UNIQUE CHECK (length(trim(username)) > 0), 
	date_of_birth date NOT NULL, 
	document_number text NOT NULL UNIQUE CHECK (length(trim(document_number)) > 0), 
	created_at timestamptz NOT NULL DEFAULT now(), 
	updated_at timestamptz NOT NULL DEFAULT now()
);


CREATE TABLE exchanges (
	id uuid PRIMARY KEY,
	name text NOT NULL UNIQUE, 
	minimum_age integer NOT NULL CHECK (minimum_age > 0), 
	maximum_transfer_amount numeric(18,2) NOT NULL CHECK (maximum_transfer_amount > 0), 
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE accounts (
	id uuid PRIMARY KEY, 
	user_id uuid NOT NULL REFERENCES users(id), 
	exchange_id uuid NOT NULL REFERENCES exchanges(id),
	balance numeric(18,2) NOT NULL DEFAULT 0 CHECK (balance >= 0), 
	created_at timestamptz NOT NULL DEFAULT now(), 
	updated_at timestamptz NOT NULL DEFAULT now(), UNIQUE(user_id, exchange_id)
);


CREATE TABLE transactions (
	id uuid PRIMARY KEY, 
	account_id uuid NOT NULL REFERENCES accounts(id), 
	type text NOT NULL CHECK (type IN ('deposit','withdrawal')),
	amount numeric(18,2) NOT NULL CHECK (amount > 0), 
	idempotency_key text NOT NULL CHECK (length(idempotency_key) BETWEEN 1 AND 128), 
	created_at timestamptz NOT NULL DEFAULT now(), UNIQUE(account_id, idempotency_key)
);

CREATE INDEX transactions_created_at_idx ON transactions(created_at);
CREATE INDEX transactions_account_created_idx ON transactions(account_id, created_at);
CREATE INDEX accounts_user_idx ON accounts(user_id);

CREATE FUNCTION reject_transaction_mutation() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 RAISE EXCEPTION 'financial transactions are immutable' USING ERRCODE = '55000';
END;
$$;
CREATE TRIGGER transactions_immutable BEFORE UPDATE OR DELETE ON transactions FOR EACH ROW EXECUTE FUNCTION reject_transaction_mutation();
