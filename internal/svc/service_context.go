package svc

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/iluyuns/alpha-trade/internal/config"
	"github.com/iluyuns/alpha-trade/internal/gateway/binance"
	"github.com/iluyuns/alpha-trade/internal/domain/port"
	orderrepo "github.com/iluyuns/alpha-trade/internal/infra/order"
	riskrepo "github.com/iluyuns/alpha-trade/internal/infra/risk"
	"github.com/iluyuns/alpha-trade/internal/core/oms"
	risklogic "github.com/iluyuns/alpha-trade/internal/core/risk"
	"github.com/iluyuns/alpha-trade/internal/middleware"
	"github.com/iluyuns/alpha-trade/internal/pkg/email"
	"github.com/iluyuns/alpha-trade/internal/pkg/revocation"
	"github.com/iluyuns/alpha-trade/internal/query"
	"github.com/iluyuns/alpha-trade/internal/strategy"
	"github.com/iluyuns/alpha-trade/internal/domain/model"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config            config.Config
	Auth              rest.Middleware
	MFA               rest.Middleware
	MFAStepUp         rest.Middleware
	Email             email.EmailService
	DB                *sql.DB
	RevocationManager revocation.RevocationManager

	// Query 访问器
	Users               *query.UsersCustom
	WebauthnCredentials *query.WebauthnCredentialsCustom
	AuditLogs           *query.AuditLogsCustom
	UserAccessLogs      *query.UserAccessLogsCustom

	// 交易组件（可选，仅在启用交易时初始化）
	BinanceSpotClient *binance.SpotClient
	BinanceWSClient   *binance.WSClient
	OrderRepo         port.OrderRepo
	RiskRepo          port.RiskRepo
	RiskManager       *risklogic.Manager
	OMSManager        *oms.Manager
	StrategyEngine    *strategy.Engine
	TradingLoop       *TradingLoop
}

func (sc *ServiceContext) Close() error {
	var errs []error

	// 停止交易循环
	if sc.TradingLoop != nil {
		sc.TradingLoop.Stop()
	}

	// 停止 OMS 自动同步
	if sc.OMSManager != nil {
		sc.OMSManager.StopAutoSync()
	}

	// 关闭 WebSocket 连接
	if sc.BinanceWSClient != nil {
		if err := sc.BinanceWSClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close websocket client: %w", err))
		}
	}

	// 关闭数据库连接
	if sc.DB != nil {
		if err := sc.DB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close database: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}
	return nil
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	// 打开数据库连接
	db, err := sql.Open("postgres", c.Database.DataSource)
	if err != nil {
		return nil, err
	}

	// 创建 sqlx 连接
	sqlConn := sqlx.NewSqlConnFromDB(db)

	// 初始化 Query 访问器
	usersQuery := query.NewUsers(sqlConn)
	auditLogsQuery := query.NewAuditLogs(sqlConn)
	webauthnQuery := query.NewWebauthnCredentials(sqlConn)
	userAccessLogsQuery := query.NewUserAccessLogs(sqlConn)

	// 初始化撤销管理器
	revocationManager, err := revocation.NewCachedRevocationManager(usersQuery)
	if err != nil {
		return nil, err
	}

	ctx := &ServiceContext{
		Config:            c,
		Auth:              middleware.NewAuthMiddleware(c.Auth.AuthSecret, auditLogsQuery, revocationManager).Handle, // 基础/MFA 认证密钥
		MFA:               middleware.NewMFAMiddleware().Handle,                                                      // MFA 状态校验
		MFAStepUp:         middleware.NewMFAStepUpMiddleware(c.Auth.SudoSecret).Handle,                               // 提级认证校验
		Email:             email.NewAWSSES(&c.AWS),
		DB:                db,
		RevocationManager: revocationManager,

		// Query 访问器
		Users:               usersQuery,
		WebauthnCredentials: webauthnQuery,
		AuditLogs:           auditLogsQuery,
		UserAccessLogs:      userAccessLogsQuery,
	}

	// 如果启用交易，初始化交易组件
	if c.Trading.Enabled {
		if err := initTradingComponents(ctx, c); err != nil {
			return nil, fmt.Errorf("init trading components: %w", err)
		}
	}

	return ctx, nil
}

// initTradingComponents 初始化交易组件
func initTradingComponents(ctx *ServiceContext, c config.Config) error {
	// 1. 初始化 Binance Gateway
	if c.Binance.APIKey == "" || c.Binance.APISecret == "" {
		return fmt.Errorf("binance API key and secret are required when trading is enabled")
	}

	binanceCfg := binance.Config{
		APIKey:    c.Binance.APIKey,
		APISecret: c.Binance.APISecret,
		Testnet:   c.Binance.Testnet,
	}

	spotClient := binance.NewSpotClient(binanceCfg)
	wsClient := binance.NewWSClient(binanceCfg)

	ctx.BinanceSpotClient = spotClient
	ctx.BinanceWSClient = wsClient

	// 2. 初始化 OrderRepo
	ctx.OrderRepo = orderrepo.NewPostgresRepo(ctx.DB)

	// 3. 初始化 RiskRepo（根据配置选择 Redis 或 Postgres）
	var riskRepo port.RiskRepo
	if strings.ToLower(c.Risk.RepoType) == "redis" {
		// 解析 Redis URL
		opt, err := redis.ParseURL(c.Redis.URL)
		if err != nil {
			return fmt.Errorf("parse redis URL: %w", err)
		}
		redisClient := redis.NewClient(opt)
		riskRepo = riskrepo.NewRedisRepo(redisClient)
		logx.Infof("Using Redis for risk state storage: %s", c.Redis.URL)
	} else {
		riskRepo = riskrepo.NewPostgresRepo(ctx.DB)
		logx.Infof("Using PostgreSQL for risk state storage")
	}
	ctx.RiskRepo = riskRepo

	// 4. 初始化 RiskManager
	riskConfig := risklogic.RiskConfig{
		MaxConsecutiveLosses:       c.Risk.MaxConsecutiveLosses,
		MaxDailyDrawdown:           c.Risk.MaxDailyDrawdown,
		MaxTotalMDD:                c.Risk.MaxTotalMDD,
		MaxSinglePositionPercent:   c.Risk.MaxSinglePositionPercent,
		MaxTotalExposurePercent:    c.Risk.MaxTotalExposurePercent,
		MinCashReservePercent:      c.Risk.MinCashReservePercent,
		MaxLeverage:                c.Risk.MaxLeverage,
	}
	ctx.RiskManager = risklogic.NewManager(riskRepo, riskConfig)

	// 5. 初始化 OMS Manager
	omsConfig := oms.Config{
		SyncInterval: 5 * time.Second,
		AutoSync:     true,
	}
	ctx.OMSManager = oms.NewManager(spotClient, ctx.OrderRepo, ctx.RiskManager, omsConfig)

	// 6. 初始化 Strategy Engine
	// 默认使用 SimpleVolatility 策略
	var strategyInstance strategy.Strategy
	if c.Trading.StrategyType == "simple_volatility" || c.Trading.StrategyType == "" {
		// 从配置获取阈值，默认 2%
		threshold := model.MustMoney("0.02")
		if thresholdVal, ok := c.Trading.StrategyParams["threshold"].(string); ok {
			threshold = model.MustMoney(thresholdVal)
		}
		// 使用第一个交易对作为默认标的
		symbol := "BTCUSDT"
		if len(c.Trading.Symbols) > 0 {
			symbol = c.Trading.Symbols[0]
		}
		strategyInstance = strategy.NewSimpleVolatility(symbol, threshold)
	} else {
		return fmt.Errorf("unsupported strategy type: %s", c.Trading.StrategyType)
	}

	// 创建策略引擎（通过 OMS 下单，集成风控）
	// 使用适配器将 OMS Manager 适配到 Strategy Engine 接口
	omsAdapter := oms.NewStrategyOMSAdapter(ctx.OMSManager)
	accountID := "default-account" // 默认账户ID，后续可从配置读取
	ctx.StrategyEngine = strategy.NewEngineWithOMS(strategyInstance, omsAdapter, accountID)

	// 7. 初始化 TradingLoop
	ctx.TradingLoop = NewTradingLoop(wsClient, ctx.StrategyEngine, c.Trading.Symbols, c.Trading.KlineInterval)

	// 启动 OMS 自动同步
	ctx.OMSManager.StartAutoSync(context.Background())

	logx.Infof("Trading components initialized: symbols=%v, interval=%s, strategy=%s",
		c.Trading.Symbols, c.Trading.KlineInterval, strategyInstance.Name())

	return nil
}
