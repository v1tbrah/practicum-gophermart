package pg

const queryCreateStartingBalance = `INSERT INTO balance (user_id, sum) VALUES ($1, 0)`

const queryGetBalance = `
SELECT
	MAX(balance.sum) AS current,
	SUM(COALESCE(withdrawals.sum, 0)) AS withdrawn
FROM
	balance LEFT JOIN withdrawals ON
		balance.user_id = withdrawals.user_id
WHERE
	balance.user_id = $1
GROUP BY
	balance.user_id
`

const queryIncreaseBalance = `UPDATE balance SET sum = sum + $2 WHERE user_id=$1 RETURNING sum`

const queryReduceBalance = `UPDATE balance SET sum = sum - $2 WHERE user_id=$1 RETURNING sum`
