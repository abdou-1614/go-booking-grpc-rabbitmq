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
	migrate -database postgresql://postgres:-5aaC14F51DbgCg*ad6e-dC*532*5c*d@monorail.proxy.rlwy.net:58785/railway -path user/migrations up