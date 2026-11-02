DROP TRIGGER IF EXISTS transactions_immutable ON transactions;
DROP FUNCTION IF EXISTS reject_transaction_mutation();
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS exchanges;
DROP TABLE IF EXISTS users;
