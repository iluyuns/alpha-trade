SERVICE_NAME = alpha-trade

TEMPLATE_LINK = https://github.com/iluyuns/go-zero-template
DB_URL = postgres://alpha:alpha_pwd@localhost:5432/alpha_trade?sslmode=disable
comma := ,
empty :=
space := $(empty) $(empty)
TABLES = users,webauthn_credentials,exchange_accounts,audit_logs,orders,executions,risk_records,asset_snapshots,strategy_configs,settlements,user_access_logs

API_FILE = api/_.api
API_DIR = ./

# api 生成
.PHONY: api
api:
	@echo ">>> generator api file..."
	goctl api go -api $(API_FILE) -dir $(API_DIR) --style go_zero --remote $(TEMPLATE_LINK)
	@echo ">>> generator api file success"

# rpc 生成
.PHONY: rpc
rpc:
	@echo ">>> generator rpc file..."
	goctl rpc protoc pb/*.proto --go_out=./pb --go-grpc_out=./pb --zrpc_out=. --style go_zero --remote $(TEMPLATE_LINK)
	@echo ">>> generator rpc file success"

# 生成 migrate 文件
.PHONY: migrate
migrate:
	@echo ">>> generator migrate file..."
	@read -p "Enter migration description (e.g., init_schema): " DESC; \
	migrate create -ext sql -dir ./migrations -seq $${DESC}
	@echo ">>> generator migrate file success"

# 生成数据库实体和部分方法
.PHONY: model
model:
	@echo ">>> generate all models using gpmg..."
	../gpmg/gpmg --url $(DB_URL) --schema public --table $(TABLES) --dir ./internal/query --package query
	@echo ">>> generate model file success"