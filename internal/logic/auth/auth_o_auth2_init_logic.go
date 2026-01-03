package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/iluyuns/alpha-trade/internal/svc"
	"github.com/iluyuns/alpha-trade/internal/types"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuthOAuth2InitLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuthOAuth2InitLogic(ctx context.Context, svcCtx *svc.ServiceContext) AuthOAuth2InitLogic {
	return AuthOAuth2InitLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AuthOAuth2InitLogic) AuthOAuth2Init(req *types.OAuth2Request) (resp *types.OAuth2InitResponse, err error) {
	// 从中间件注入的 Context 中获取 UID (用于绑定逻辑)
	uid, _ := l.ctx.Value("uid").(int64)

	var conf *oauth2.Config
	switch req.Provider {
	case "google":
		conf = &oauth2.Config{
			ClientID:     l.svcCtx.Config.OAuth.Google.ClientID,
			ClientSecret: l.svcCtx.Config.OAuth.Google.ClientSecret,
			RedirectURL:  l.svcCtx.Config.OAuth.Google.RedirectURL,
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
			Endpoint:     google.Endpoint,
		}
	case "github":
		conf = &oauth2.Config{
			ClientID:     l.svcCtx.Config.OAuth.Github.ClientID,
			ClientSecret: l.svcCtx.Config.OAuth.Github.ClientSecret,
			RedirectURL:  l.svcCtx.Config.OAuth.Github.RedirectURL,
			Scopes:       []string{"read:user", "user:email"},
			Endpoint:     github.Endpoint,
		}
	default:
		return nil, fmt.Errorf("unsupported provider: %s", req.Provider)
	}

	// 生成包含 Provider 和 UID 的加密 State，用于回调时的身份关联
	state := l.generateState(req.Provider, uid)
	url := conf.AuthCodeURL(state, oauth2.AccessTypeOffline)

	return &types.OAuth2InitResponse{
		RedirectURL: url,
	}, nil
}

func (l *AuthOAuth2InitLogic) generateState(provider string, uid int64) string {
	ts := time.Now().Unix()
	raw := fmt.Sprintf("%s:%d:%d", provider, uid, ts)

	// 使用 AuthSecret 进行简单的签名校验
	h := hmac.New(sha256.New, []byte(l.svcCtx.Config.Auth.AuthSecret))
	h.Write([]byte(raw))
	sig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", raw, sig)))
}
