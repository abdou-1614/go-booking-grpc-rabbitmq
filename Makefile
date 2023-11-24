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
	migrate -database postgresql://postgres:B4C*3AE3B5FdE5c55Edc*13eb56GFDc1@roundhouse.proxy.rlwy.net:57236/railway -path user/migrations up