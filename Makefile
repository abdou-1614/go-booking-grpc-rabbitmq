.PHONY:

# ==============================================================================
# Start local dev environment

develop:
	echo "Starting develop environment"
	docker-compose -f docker-compose.yml up --build

local:
	echo "Starting local environment"
	docker-compose -f docker-compose.local.yml up --build

# ==============================================================================


# Run Listner For All Services

listner:
	cd user && golangci-lint run


mg_user_db:
	migrate -database postgresql://postgres:wEkkDLzNNjGUuecxUqAHUizVVchCvBhX@monorail.proxy.rlwy.net:23962/railway -path user/migrations up