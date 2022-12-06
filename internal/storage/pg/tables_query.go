package pg

const queryCreateTypeOrderStatus = `
DO $$ BEGIN
	CREATE TYPE order_status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');
EXCEPTION
	WHEN duplicate_object
	THEN null;
END $$;
`

const queryCreateTableUsers = `
CREATE TABLE IF NOT EXISTS users
(
	id       bigint  PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	login    varchar UNIQUE NOT NULL,
	password varchar NOT NULL
);
`

const queryCreateTableRefreshSessions = `
CREATE TABLE IF NOT EXISTS refreshSessions
(
    id           bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id      bigint REFERENCES users(id) ON DELETE CASCADE,
    refreshToken uuid NOT NULL,
    expiresIn    timestamp NOT NULL
);
`

const queryCreateTableOrders = `
CREATE TABLE IF NOT EXISTS orders
(
	id            bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	user_id       bigint REFERENCES users(id) ON DELETE CASCADE,
	number        varchar NOT NULL UNIQUE,
	status        order_status NOT NULL,
	accrual       double precision NOT NULL,
	uploaded_at   timestamp NOT NULL
);
`

const queryCreateTableBalance = `
CREATE TABLE IF NOT EXISTS balance
(
	id             bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	user_id        bigint REFERENCES users(id) ON DELETE CASCADE,
	sum            double precision NOT NULL
);
`

const queryCreateTableWithdrawals = `
CREATE TABLE IF NOT EXISTS withdrawals
(
	id             bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	user_id        bigint REFERENCES users(id) ON DELETE CASCADE,
	order_number   varchar NOT NULL UNIQUE,
	sum            double precision NOT NULL,
	processed_at   timestamp NOT NULL
);
`
