package pg

const (
	queryAddUser = `INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id`

	queryGetUser = `SELECT id, login, password FROM users WHERE login = $1 AND password = $2`

	queryGetUserPassword = `SELECT password FROM users WHERE login = $1`
)
