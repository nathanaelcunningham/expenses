-- name: CreateAccount :one
INSERT INTO accounts (account_id,name,simplefin_id)
VALUES (?,?,?)
RETURNING *;

-- name: GetAccounts :many
SELECT * FROM accounts;

-- name: DeleteAccount :exec
DELETE FROM accounts where id = ?;

-- name: CreateTransaction :one
INSERT INTO transactions (account_id,posted_date,description,payee)
VALUES (?,?,?,?)
RETURNING *;

-- name: GetTransactionsByAccount :many
SELECT * FROM transactions WHERE account_id = ?;
