package pg

const (
	queryAddOrder        = `INSERT INTO orders (user_id, number, status, accrual, uploaded_at) VALUES ($1, $2, $3, $4, $5)`
	queryGetOrder        = `SELECT user_id, number, status, accrual, uploaded_at FROM orders WHERE number=$1`
	queryGetOrdersByUser = `SELECT user_id, number, status, accrual, uploaded_at FROM orders WHERE user_id=$1 ORDER BY uploaded_at`
)
