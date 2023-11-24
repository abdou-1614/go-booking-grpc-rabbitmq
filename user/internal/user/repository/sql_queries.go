package repository

const (
	createUserQuery = `INSERT INTO users (first_name, last_name, email, password)
	VALUES ($1, $2, $3, $4)
	RETURNING id, first_name, last_name, email, updated_at, created_at`
)
