package pg

const (
	queryAddRefreshSession = `INSERT INTO refreshsessions (user_id, refreshToken, expiresIn) VALUES ($1, $2, $3)`

	queryDeleteRefreshSessions = `DELETE FROM refreshsessions WHERE user_id = $1`

	queryGetRefreshSessionByToken = `SELECT user_id, expiresIn FROM refreshsessions WHERE refreshToken = $1`
)
