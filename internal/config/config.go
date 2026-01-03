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
}
