# Alpha-Trade 系统安全协议 (System Security Protocol)

**版本**: v1.0
**状态**: 强制执行 (Enforced)
**定位**: 本文档定义了系统基础设施、网络、密钥及访问控制的最高安全标准。

> **警示**: 任何违反本协议的操作都可能导致资产归零。安全不是功能，是生存的前提。

---

## 1. 资产与密钥安全 (Asset & Key Security)

### 1.1 API Key 核心管理原则
*   **权限最小化 (Least Privilege)**:
    *   **交易 Key**: 仅开启 `Enable Spot & Margin Trading` 和 `Enable Futures`。**严禁**开启 `Enable Withdrawals` (提现)。
    *   **行情 Key**: 仅开启 `Enable Reading`，用于只读数据流。
*   **IP 白名单绑定 (IP Whitelisting)**:
    *   所有生产环境 API Key **必须** 绑定服务器出口静态 IP。
    *   开发环境测试 Key 必须绑定开发者本地 IP，严禁开放 `0.0.0.0/0`。
*   **非对称加密存储**:
    *   数据库中**严禁**明文存储 API Secret。
    *   采用 **AES-256-GCM** 算法加密存储 `api_secret`。
    *   解密密钥 (Master Key) 不得落地到代码仓库或配置文件，必须通过 **环境变量 (Environment Variables)** 或 **AWS KMS / HashiCorp Vault** 在运行时注入。

### 1.2 签名安全
*   **本地签名**: 所有交易请求的签名动作必须在受信任的后端服务器内存中完成。
*   **严禁前端暴露**: 任何情况下，API Key/Secret 不得出现在前端代码、浏览器 LocalStorage 或客户端日志中。

---

## 2. 网络与架构安全 (Network & Infrastructure)

### 2.1 隔离网络架构 (VPC)
系统必须部署在私有网络 (VPC) 中，遵循以下分层：

| 层级 | 访问策略 | 部署组件 |
| :--- | :--- | :--- |
| **Public Subnet** | 仅允许 80/443 入站，SSH 仅限 Bastion | Load Balancer, Bastion Host |
| **Private Subnet** | 禁止公网直接入站，仅允许 NAT 出站 | **Go Trade Core**, AI Agent, OMS |
| **Data Subnet** | 禁止任何公网出入 | PostgreSQL, Redis, NATS |

### 2.2 堡垒机 (Bastion Host)
*   所有 SSH 运维操作必须通过堡垒机跳转。
*   **禁止**直接 SSH 登录核心交易服务器。
*   堡垒机必须启用 **MFA (多因素认证)** 和 **审计日志**。

### 2.3 防火墙策略 (Security Groups)
*   **默认拒绝 (Deny All)**: 所有入站端口默认关闭。
*   **精确放行**: 仅放行特定源 IP 的特定端口 (如: 仅允许 Admin IP 访问管理后台端口)。

---

## 3. 访问控制与认证 (Authentication & Access Control)

### 3.1 管理后台安全
*   **首选认证 (Primary): Passkeys (WebAuthn)**
    *   **根信任设备 (Root Trust)**: 初始管理员账号及所有核心操作权限，**必须**绑定于 Apple 生态设备 (iPhone/Mac/iPad) 的 Secure Enclave 芯片中。
    *   **信任链管理**: 任何新增设备或 Passkey 的注册，必须经过现有根信任设备的授权验证。
    *   **非信任设备限制**: Windows/Android 等非 Apple 设备禁止注册为管理端 Passkey，仅允许作为临时受限会话使用。
    *   **抗钓鱼**: 依赖硬件级密钥及域名绑定，杜绝中间人攻击。
*   **备用认证 (Backup): TOTP**
    *   仅用于不支持 WebAuthn 的旧设备或作为应急恢复手段。
    *   登录必须经过 TOTP (Google Authenticator) 二次验证。
*   **会话管理**:
    *   Token 有效期不得超过 2 小时。
    *   检测到异地登录或 IP 变更，立即销毁会话。
*   **操作审计**: 任何修改风控参数、启停策略的操作，必须记录 `OperatorID`, `IP`, `Timestamp`, `Action`, `Diff`。

### 3.2 敏感操作双人复核 (Four-Eyes Principle)
*   对于核心配置变更 (如修改 API Key、调整最大资金限额)，建议实施双人复核机制（需两个不同 Admin 账号确认）。

---

## 4. 依赖与代码安全 (Supply Chain Security)

### 4.1 依赖扫描
*   **CI/CD 集成**: 每次构建必须运行 `govulncheck` (Go) 和 `safety check` (Python)。
*   **版本锁定**: `go.sum` 和 `requirements.txt` 必须提交并锁定版本，防止恶意依赖注入。

### 4.2 敏感信息扫描
*   启用 `git-secrets` 或类似工具作为 Pre-commit Hook，防止误将 Key/Secret 提交到 Git 仓库。

---

## 5. 应急响应 (Incident Response)

### 5.1 资产冻结 (Kill Switch)
当检测到未授权的异常交易或系统被入侵迹象时：
1.  **Level 1 (软停机)**: 管理后台点击 "Global Halt"，停止所有策略开仓，取消所有挂单。
2.  **Level 2 (硬隔离)**: 运维脚本切断服务器公网出口网络，阻止数据外泄，仅保留 SSH 管理通道。
3.  **Level 3 (资产冻结)**: 使用备用独立通道（手机 APP 或 备用脚本）立即**删除/重置交易所 API Key**。

### 5.3 Passkey 丢失恢复 (Break-Glass Procedure)
*   **前提条件**: 管理员丢失所有受信任设备，无法登录后台。
*   **执行环境**: 必须通过 SSH 登录生产服务器终端 (TTY)。
*   **工具**: 使用 `admin-cli` 命令行工具。
*   **流程**:
    1.  运行 `admin-cli reset-passkey --user=admin`。
    2.  **强制验证**: 输入该用户的**静态登录密码**进行身份复核。
    3.  系统清空该用户的 Passkey 列表。
    4.  终端输出一次性注册链接 (Magic Link, 有效期 5 分钟)。
    5.  使用新设备访问链接完成根密钥重置。

---

## 6. 日志脱敏 (Log Sanitization)

*   **原则**: 日志是调试工具，也是泄密源头。
*   **规则**:
    *   **禁止**打印 API Secret。
    *   **禁止**打印用户密码 Hash。
    *   **脱敏**打印 API Key (仅显示前 4 位，如 `vmPU***`)。
    *   **脱敏**打印手机号/邮箱。

---

## 7. 更新历史

*   **v1.0**: 2025-12-27 - 初始发布，确立核心安全基线。

