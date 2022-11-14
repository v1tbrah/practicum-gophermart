package pg

const (
	queryAddWithdrawal  = `INSERT INTO withdrawals (user_id, order_number, sum, processed_at) VALUES ($1, $2, $3, $4)`
	queryGetWithdrawals = `SELECT order_number, sum, processed_at FROM withdrawals WHERE user_id=$1 ORDER BY processed_at`
)
