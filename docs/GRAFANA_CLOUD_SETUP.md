# Grafana Cloud 配置指南

本文档说明如何在 Grafana Cloud 中配置 Alpha-Trade 监控 Dashboard。

---

## 1. 配置 Prometheus 数据源

### 1.1 在 Grafana Cloud 中添加数据源

1. 登录 Grafana Cloud: https://grafana.com/
2. 进入你的 Grafana 实例
3. 导航到 **Configuration** -> **Data Sources**
4. 点击 **Add data source**
5. 选择 **Prometheus**

### 1.2 配置 Prometheus 连接

**方式 A: 使用 Prometheus Remote Write（推荐）**

如果你的 Prometheus 实例支持 Remote Write：

1. **URL**: 使用 Grafana Cloud 提供的 Remote Write Endpoint
   ```
   https://prometheus-prod-XX.grafana.net/api/prom/push
   ```

2. **认证**: 使用 Grafana Cloud API Key
   - 在 Grafana Cloud 中生成 API Key
   - 配置为 Basic Auth 或 Bearer Token

**方式 B: 直接连接 Prometheus（本地/自托管）**

如果你的 Prometheus 运行在本地或私有网络：

1. **URL**: 你的 Prometheus 实例地址
   ```
   http://your-prometheus-host:9090
   ```

2. **访问方式**:
   - 如果 Prometheus 在公网：直接配置 URL
   - 如果 Prometheus 在内网：使用 Grafana Cloud Agent 或 VPN

### 1.3 配置 Grafana Cloud Agent（可选）

如果 Prometheus 在内网，可以使用 Grafana Cloud Agent 推送指标：

```yaml
# grafana-cloud-agent.yml
server:
  log_level: info

prometheus:
  configs:
    - name: alpha-trade
      remote_write:
        - url: https://prometheus-prod-XX.grafana.net/api/prom/push
          basic_auth:
            username: <your-instance-id>
            password: <your-api-key>
      scrape_configs:
        - job_name: 'alpha-trade-core'
          static_configs:
            - targets: ['localhost:9091']
```

---

## 2. 导入 Dashboard

### 2.1 导入 Dashboard JSON

1. 在 Grafana Cloud 中，导航到 **Dashboards** -> **Import**
2. 上传 `etc/grafana/dashboard.json` 文件
3. 选择刚才配置的 Prometheus 数据源
4. 点击 **Import**

### 2.2 Dashboard 面板说明

**核心指标面板**:
- **订单统计**: 下单/成交/拒绝速率
- **风控拦截率**: 订单被风控拦截的百分比
- **当前盈亏**: 总盈亏、当日盈亏、盈亏百分比
- **仓位敞口**: 总敞口和持仓数量

**趋势图**:
- **订单趋势**: 订单速率时间序列
- **风控检查统计**: 风控检查通过/拦截趋势
- **PnL 趋势**: 盈亏变化曲线
- **系统延迟分布**: 订单/风控/Gateway 延迟 P95

**告警指标**:
- **熔断器状态**: 熔断触发次数
- **连续亏损次数**: 当前连续亏损计数

---

## 3. 配置告警规则

### 3.1 在 Grafana Cloud 中创建告警

1. 导航到 **Alerting** -> **Alert rules**
2. 点击 **New alert rule**

### 3.2 关键告警规则示例

**告警 1: 熔断器触发**
```promql
sum(increase(alpha_trade_circuit_breaker_opened_total[5m])) > 0
```
- **条件**: 5 分钟内熔断器触发
- **严重性**: Critical
- **通知**: Telegram/Email

**告警 2: 连续亏损过多**
```promql
alpha_trade_consecutive_losses >= 3
```
- **条件**: 连续亏损 >= 3 次
- **严重性**: Warning
- **通知**: Telegram/Email

**告警 3: 风控拦截率过高**
```promql
sum(rate(alpha_trade_risk_checks_blocked_total[5m])) / sum(rate(alpha_trade_risk_checks_total[5m])) > 0.2
```
- **条件**: 拦截率 > 20%
- **严重性**: Warning
- **通知**: Telegram/Email

**告警 4: 系统延迟过高**
```promql
histogram_quantile(0.95, sum(rate(alpha_trade_order_latency_seconds_bucket[5m])) by (le)) > 1.0
```
- **条件**: 订单延迟 P95 > 1 秒
- **严重性**: Warning
- **通知**: Telegram/Email

**告警 5: 大额亏损**
```promql
alpha_trade_pnl_daily < -1000
```
- **条件**: 当日亏损 > $1000
- **严重性**: Critical
- **通知**: Telegram/Email

---

## 4. 配置通知渠道

### 4.1 Telegram Bot（推荐）

1. 创建 Telegram Bot（通过 @BotFather）
2. 获取 Bot Token
3. 在 Grafana Cloud 中配置 Notification Channel:
   - **Type**: Telegram
   - **Bot Token**: 你的 Bot Token
   - **Chat ID**: 你的 Telegram Chat ID

### 4.2 Email

1. 在 Grafana Cloud 中配置 SMTP
2. 创建 Email Notification Channel
3. 配置收件人列表

---

## 5. 验证配置

### 5.1 检查指标是否正常

在 Grafana Cloud Explore 中运行查询：

```promql
# 检查订单指标
alpha_trade_orders_total

# 检查风控指标
alpha_trade_risk_checks_total

# 检查 PnL 指标
alpha_trade_pnl_total
```

### 5.2 验证 Dashboard

1. 打开导入的 Dashboard
2. 检查所有面板是否显示数据
3. 验证时间范围和数据刷新

---

## 6. 最佳实践

1. **数据保留**: Grafana Cloud 免费版通常保留 14 天数据
2. **采样频率**: 建议 scrape_interval 设置为 15s
3. **告警频率**: 避免告警风暴，设置合理的告警间隔
4. **Dashboard 刷新**: 建议设置为 10s 或 30s

---

## 7. 故障排查

### 问题: Dashboard 显示 "No Data"

**可能原因**:
1. Prometheus 数据源未正确配置
2. 指标名称不匹配
3. 时间范围设置错误

**解决方案**:
1. 检查 Prometheus 数据源连接状态
2. 在 Explore 中验证指标是否存在
3. 检查时间范围设置

### 问题: 告警不触发

**可能原因**:
1. 告警规则配置错误
2. 通知渠道未配置
3. 告警阈值设置不合理

**解决方案**:
1. 检查告警规则语法
2. 验证通知渠道配置
3. 调整告警阈值

---

## 8. 参考资源

- [Grafana Cloud 文档](https://grafana.com/docs/grafana-cloud/)
- [Prometheus 查询语言](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [Grafana Dashboard JSON 格式](https://grafana.com/docs/grafana/latest/dashboards/json-dashboard/)
