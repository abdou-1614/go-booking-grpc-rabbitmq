package repository

const (
	createUserQuery = `INSERT INTO users (first_name, last_name, email, password, role)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, first_name, last_name, email, role ,updated_at, created_at`

	getUserByIDQuery = `SELECT id, first_name, last_name, email, role, created_at, updated_at FROM users WHERE id = $1`
)
