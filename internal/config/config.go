package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf

	// Auth 认证配置
	Auth struct {
		AuthSecret string `json:",optional,env=AUTH_SECRET"`
		SudoSecret string `json:",optional,env=AUTH_SUDO_SECRET"` // Sudo 专用密钥 (提级认证用)
		BaseExpire int64  `json:",optional,env=AUTH_BASE_EXPIRE,default=300"`
		MfaExpire  int64  `json:",optional,env=AUTH_MFA_EXPIRE,default=86400"`
		SudoExpire int64  `json:",optional,env=AUTH_SUDO_EXPIRE,default=600"`
	}

	// Database 数据库配置
	// 映射逻辑：
	// 1. 优先读取环境变量 DB_URL
	// 2. 若环境变量不存在，则读取 etc/xxx.yaml 中的 Database.DataSource
	Database struct {
		DataSource string `json:",optional,env=DB_URL"`
	}

	// AWS 基础服务配置 (通常由 etc/xxx.yaml 提供)
	AWS AWSConfig `json:"aws"`

	// OAuth2 配置
	OAuth struct {
		Google struct {
			ClientID     string `json:",optional,env=OAUTH_GOOGLE_CLIENT_ID"`
			ClientSecret string `json:",optional,env=OAUTH_GOOGLE_CLIENT_SECRET"`
			RedirectURL  string `json:",optional,env=OAUTH_GOOGLE_REDIRECT_URL"`
		} `json:"google"`
		Github struct {
			RedirectURL  string `json:",optional,env=OAUTH_GITHUB_REDIRECT_URL"`
			ClientSecret string `json:",optional,env=OAUTH_GITHUB_CLIENT_SECRET"`
			ClientID     string `json:",optional,env=OAUTH_GITHUB_CLIENT_ID"`
		} `json:"github"`
	}

	// Trading 交易配置
	Trading struct {
		Enabled      bool     `json:",optional,env=TRADING_ENABLED,default=false"`      // 是否启用交易
		Mode         string   `json:",optional,env=TRADING_MODE,default=manual"`          // auto/manual/hybrid
		Symbols      []string `json:",optional,env=TRADING_SYMBOLS"`                      // 交易对列表（逗号分隔）
		KlineInterval string  `json:",optional,env=TRADING_KLINE_INTERVAL,default=1m"`  // K线周期
		StrategyType string  `json:",optional,env=TRADING_STRATEGY_TYPE,default=simple_volatility"` // 策略类型
		StrategyParams map[string]interface{} `json:",optional"` // 策略参数
	}

	// Binance API 配置
	Binance struct {
		APIKey    string `json:",optional,env=BINANCE_API_KEY"`
		APISecret string `json:",optional,env=BINANCE_API_SECRET"`
		Testnet   bool   `json:",optional,env=BINANCE_TESTNET,default=true"` // 默认使用测试网
	}

	// Risk 风控配置
	Risk struct {
		RepoType string `json:",optional,env=RISK_REPO_TYPE,default=redis"` // "redis" 或 "postgres"
		// 风控规则配置
		MaxConsecutiveLosses int     `json:",optional,default=3"`
		MaxDailyDrawdown     float64 `json:",optional,default=0.05"`  // 5%
		MaxTotalMDD          float64 `json:",optional,default=0.15"`  // 15%
		MaxSinglePositionPercent float64 `json:",optional,default=0.3"` // 30%
		MaxTotalExposurePercent  float64 `json:",optional,default=0.7"` // 70%
		MinCashReservePercent    float64 `json:",optional,default=0.3"` // 30%
		MaxLeverage int `json:",optional,default=2"`
	}

	// Redis 配置（用于 RiskRepo）
	Redis struct {
		URL string `json:",optional,env=REDIS_URL,default=redis://localhost:6379/0"`
	}
}
