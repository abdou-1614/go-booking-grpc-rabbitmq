package repository

const (
	createUserQuery = `INSERT INTO users (first_name, last_name, email, password, role)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, first_name, last_name, email, role ,updated_at, created_at`

	getUserByEmail = `SELECT id, first_name, last_name, email, role, updated_at, created_at FROM users WHERE email = $1`

	getUserByIDQuery  = `SELECT id, first_name, last_name, email, role, created_at, updated_at FROM users WHERE id = $1`
	updateAvatarQuery = `UPDATE users SET avatar = $1 WHERE id = $2 `
)
