# 交易控制 API 文档

## 概述

交易控制 API 提供了通过 HTTP 接口控制交易循环的功能，支持启动、停止和查询交易状态。

**基础路径**: `/api/v1/trading`

**认证要求**: 所有接口都需要：
1. 登录认证（`Auth` middleware）
2. MFA 认证（`MFA` middleware）

---

## API 接口

### 1. 查询交易状态

**接口**: `GET /api/v1/trading/status`

**描述**: 查询当前交易系统的状态，包括是否启用、是否运行、配置信息等。

**请求示例**:
```bash
curl -X GET "http://localhost:8888/api/v1/trading/status" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**响应示例**:
```json
{
  "enabled": true,
  "started": true,
  "mode": "hybrid",
  "symbols": ["BTCUSDT", "ETHUSDT"],
  "interval": "1m",
  "strategy": "simple_volatility",
  "message": "Trading loop is running"
}
```

**响应字段说明**:
- `enabled`: 交易功能是否在配置中启用
- `started`: 交易循环是否正在运行
- `mode`: 交易模式（`auto`/`manual`/`hybrid`）
- `symbols`: 配置的交易对列表
- `interval`: K线周期（如 `1m`, `5m`, `1h`）
- `strategy`: 当前使用的策略类型
- `message`: 状态描述信息

---

### 2. 启动交易循环

**接口**: `POST /api/v1/trading/start`

**描述**: 手动启动交易循环。仅在 `manual` 或 `hybrid` 模式下可用。`auto` 模式下交易会自动启动，无需手动调用。

**请求示例**:
```bash
curl -X POST "http://localhost:8888/api/v1/trading/start" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json"
```

**响应示例（成功）**:
```json
{
  "success": true,
  "message": "Trading loop started successfully"
}
```

**响应示例（失败）**:
```json
{
  "success": false,
  "message": "Trading is not enabled. Please enable it in configuration first."
}
```

**可能的错误消息**:
- `"Trading is not enabled"`: 交易功能未在配置中启用
- `"Trading mode is 'auto'"`: 当前为自动模式，无需手动启动
- `"Trading components are not initialized"`: 交易组件未初始化，检查配置
- `"Trading loop is already running"`: 交易循环已在运行
- `"Failed to start trading loop: ..."`: 启动失败的具体原因

---

### 3. 停止交易循环

**接口**: `POST /api/v1/trading/stop`

**描述**: 停止正在运行的交易循环。停止后不会禁用交易功能，可以再次通过 `/start` 接口启动。

**请求示例**:
```bash
curl -X POST "http://localhost:8888/api/v1/trading/stop" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json"
```

**响应示例（成功）**:
```json
{
  "success": true,
  "message": "Trading loop stopped successfully"
}
```

**响应示例（已停止）**:
```json
{
  "success": true,
  "message": "Trading loop is already stopped"
}
```

---

## 使用场景

### 场景 1: 手动模式（Manual Mode）

配置 `TRADING_MODE=manual`，交易不会自动启动：

```bash
# 1. 查询状态（确认未启动）
GET /api/v1/trading/status
# 响应: {"started": false, ...}

# 2. 手动启动交易
POST /api/v1/trading/start
# 响应: {"success": true, "message": "Trading loop started successfully"}

# 3. 确认已启动
GET /api/v1/trading/status
# 响应: {"started": true, ...}

# 4. 需要时停止交易
POST /api/v1/trading/stop
```

### 场景 2: 混合模式（Hybrid Mode）

配置 `TRADING_MODE=hybrid`，交易会自动启动，但也可以通过 API 控制：

```bash
# 服务启动后，交易自动运行
# 可以通过 API 临时停止
POST /api/v1/trading/stop

# 需要时重新启动
POST /api/v1/trading/start
```

### 场景 3: 自动模式（Auto Mode）

配置 `TRADING_MODE=auto`，交易自动启动且无法通过 API 停止（需要修改配置重启服务）。

---

## 错误处理

所有接口在出错时会返回 HTTP 状态码和错误信息：

- `200 OK`: 请求成功
- `400 Bad Request`: 请求参数错误
- `401 Unauthorized`: 未认证或认证失败
- `403 Forbidden`: 未完成 MFA 认证
- `500 Internal Server Error`: 服务器内部错误

---

## 安全说明

1. **认证要求**: 所有交易控制接口都需要登录且完成 MFA 认证
2. **权限控制**: 建议在生产环境中添加额外的权限检查（如管理员权限）
3. **审计日志**: 所有交易控制操作都会记录在审计日志中

---

## 配置说明

交易控制 API 的行为受以下配置影响：

- `TRADING_ENABLED`: 是否启用交易功能
- `TRADING_MODE`: 交易模式（`auto`/`manual`/`hybrid`）
- `TRADING_SYMBOLS`: 交易对列表
- `TRADING_KLINE_INTERVAL`: K线周期
- `TRADING_STRATEGY_TYPE`: 策略类型

详细配置说明请参考 `env.example` 和 `etc/alpha_trade.yaml`。
