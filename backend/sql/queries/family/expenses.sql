-- name: CreateExpense :one
INSERT INTO expenses (category_id, amount, name, day_of_month_due, is_autopay, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetExpenseByID :one
SELECT * FROM expenses WHERE id = ?;

-- name: UpdateExpense :one
UPDATE expenses 
SET category_id = ?, amount = ?, name = ?, day_of_month_due = ?, is_autopay = ?, updated_at = ?
WHERE id = ?
RETURNING *;

-- name: DeleteExpense :exec
DELETE FROM expenses WHERE id = ?;

-- name: ListExpenses :many
SELECT * FROM expenses 
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListExpensesByCategory :many
SELECT * FROM expenses 
WHERE category_id = ?
ORDER BY created_at DESC;

-- name: CountExpenses :one
SELECT COUNT(*) FROM expenses;

-- name: GetExpensesByDateRange :many
SELECT * FROM expenses
WHERE day_of_month_due BETWEEN ? AND ?
ORDER BY day_of_month_due ASC;
