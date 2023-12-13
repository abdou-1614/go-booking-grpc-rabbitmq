package repository

const (
	createSession = `
	INSERT INTO sessions (
	  id,
	  email,
	  refresh_token,
	  user_agent,
	  client_ip,
	  is_blocked,
	  expires_at
	) VALUES (
	  $1, $2, $3, $4, $5, $6, $7
	) RETURNING id, email, refresh_token, user_agent, client_ip, is_blocked, expires_at, created_at
	`
)
