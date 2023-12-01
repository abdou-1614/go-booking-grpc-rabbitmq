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
	migrate -database postgresql://postgres:E2e5E11dFDGGFGgCBDF5e4Dedbaf4da4@viaduct.proxy.rlwy.net:22852/railway -path user/migrations up