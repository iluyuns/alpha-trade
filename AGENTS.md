# AGENTS.md

## Cursor Cloud specific instructions

### Services overview

| Service | Port | Required | Start command |
|---------|------|----------|---------------|
| PostgreSQL 16 | 5432 | Yes | `sudo docker compose up -d postgres` |
| Redis | 6379 | Yes | `sudo docker compose up -d redis` |
| Go API Server | 8888 (API), 9091 (metrics) | Yes | See below |
| Vue Frontend | 3000 | Optional | `npm run dev` in `web/` |

A `.env` file at the repo root is required for `docker compose`. Minimum contents:
```
POSTGRES_USER=alpha
POSTGRES_PASSWORD=alpha_pwd
POSTGRES_DB=alpha_trade
AUTH_SECRET=dev-auth-secret-32chars-padding00
AUTH_SUDO_SECRET=dev-sudo-secret-32chars-padding0
```

### Starting Docker (nested container environment)

The Cloud VM runs inside a container. Docker requires manual daemon start:
```bash
sudo dockerd &>/tmp/dockerd.log &
sleep 3
```
Docker is pre-configured with `fuse-overlayfs` storage driver and `iptables-legacy`.

### Database migrations

After PostgreSQL is up, run migrations with `golang-migrate`:
```bash
~/go/bin/migrate -path ./migrations -database "postgres://alpha:alpha_pwd@localhost:5432/alpha_trade?sslmode=disable" up
```

### Running the Go API server

The server needs these env vars (trading can be disabled for dev):
```bash
DB_URL="postgres://alpha:alpha_pwd@localhost:5432/alpha_trade?sslmode=disable" \
TRADING_ENABLED=false TRADING_MODE=manual \
TRADING_KLINE_INTERVAL=1m TRADING_STRATEGY_TYPE=simple_volatility \
BINANCE_API_KEY=test BINANCE_API_SECRET=test BINANCE_TESTNET=true \
RISK_REPO_TYPE=redis REDIS_URL="redis://localhost:6379/0" \
AUTH_SECRET=dev-auth-secret-32chars-padding00 \
go run alpha_trade.go -f etc/alpha_trade.yaml
```

### Running tests and lint

- **Go tests**: `go test ./internal/... -v` (backtest/mock tests need no external services; Redis repo tests connect to localhost:6379)
- **Go lint**: `go vet ./...`
- **Backtest (standalone, no deps)**: `go run ./cmd/backtest -csv testdata/sample_btc.csv -symbol BTCUSDT -threshold 0.02 -capital 10000`
- **Frontend type check**: `npx vue-tsc -b --noEmit` in `web/` (has pre-existing TS errors from unimplemented components)
- **Frontend dev**: `npm run dev` in `web/`

### Gotchas

- The `vue-tsc` build check has ~70 pre-existing TypeScript errors due to unimplemented stores/components/views. The Vite dev server still works.
- `internal/infra/risk` Redis repo tests have pre-existing failures (serialization mismatch). These are not environment issues.
- Binance WebSocket integration tests are skipped unless `BINANCE_API_KEY` and `BINANCE_API_SECRET` are set to real keys.
- The `sql/pg/` directory referenced by docker-compose for auto-init does not exist; use migrations instead.
- All API endpoints except `/api/v1/auth/*` require JWT authentication.
