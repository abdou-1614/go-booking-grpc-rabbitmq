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
	migrate -database postgresql://postgres:5Bd4dedeBBggDD*21eBDEAfEBc-CcAeE@monorail.proxy.rlwy.net:42888/railway -path user/migrations up